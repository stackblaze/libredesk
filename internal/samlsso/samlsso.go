// Package samlsso implements SAML 2.0 single sign-on for agents, with Libredesk acting as the service provider.
package samlsso

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/zerodha/logf"
)

// metadataRefreshInterval is how long a fetched IdP metadata document is trusted before it is fetched again.
const metadataRefreshInterval = 12 * time.Hour

// ErrNotConfigured is returned when SAML is disabled or has no usable IdP metadata.
var ErrNotConfigured = errors.New("saml sso is not configured")

// Config is the SAML configuration stored in the settings table under the `saml.` prefix.
type Config struct {
	Enabled           bool   `json:"saml.enabled"`
	Name              string `json:"saml.name"`
	IDPMetadataURL    string `json:"saml.idp_metadata_url"`
	IDPMetadataXML    string `json:"saml.idp_metadata_xml"`
	AllowIDPInitiated bool   `json:"saml.allow_idp_initiated"`
	EmailAttribute    string `json:"saml.email_attribute"`
	SPCertificate     string `json:"saml.sp_certificate"`
	SPPrivateKey      string `json:"saml.sp_private_key"`
}

// Manager holds the service provider built from the current configuration.
type Manager struct {
	mu         sync.RWMutex
	cfg        Config
	rootURL    string
	sp         *saml.ServiceProvider
	fetchedAt  time.Time
	httpClient *http.Client
	lo         *logf.Logger
}

// New returns an empty manager. Call Load to apply a configuration.
func New(httpClient *http.Client, lo *logf.Logger) *Manager {
	return &Manager{httpClient: httpClient, lo: lo}
}

// ACSURL returns the assertion consumer service URL for a root URL.
func ACSURL(rootURL string) string {
	return strings.TrimRight(rootURL, "/") + "/api/v1/saml/acs"
}

// MetadataURL returns the SP metadata URL, which is also used as the SP entity ID.
func MetadataURL(rootURL string) string {
	return strings.TrimRight(rootURL, "/") + "/api/v1/saml/metadata"
}

// Load applies a configuration. IdP metadata is fetched now when a URL is set.
func (m *Manager) Load(cfg Config, rootURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = cfg
	m.rootURL = rootURL
	m.sp = nil
	m.fetchedAt = time.Time{}
	if !cfg.Enabled {
		return nil
	}
	return m.buildLocked(context.Background())
}

// Config returns the loaded configuration.
func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// Enabled reports whether SAML sign-in should be offered.
func (m *Manager) Enabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg.Enabled
}

func (m *Manager) buildLocked(ctx context.Context) error {
	cfg := m.cfg
	if cfg.SPCertificate == "" || cfg.SPPrivateKey == "" {
		return errors.New("saml service provider key pair is missing")
	}
	key, cert, err := ParseKeyPair(cfg.SPCertificate, cfg.SPPrivateKey)
	if err != nil {
		return err
	}
	idp, err := m.idpMetadata(ctx, cfg)
	if err != nil {
		return err
	}
	acs, err := url.Parse(ACSURL(m.rootURL))
	if err != nil {
		return err
	}
	meta, err := url.Parse(MetadataURL(m.rootURL))
	if err != nil {
		return err
	}
	m.sp = &saml.ServiceProvider{
		EntityID:          meta.String(),
		Key:               key,
		Certificate:       cert,
		HTTPClient:        m.httpClient,
		MetadataURL:       *meta,
		AcsURL:            *acs,
		IDPMetadata:       idp,
		AllowIDPInitiated: cfg.AllowIDPInitiated,
		AuthnNameIDFormat: saml.EmailAddressNameIDFormat,
	}
	m.fetchedAt = time.Now()
	return nil
}

func (m *Manager) idpMetadata(ctx context.Context, cfg Config) (*saml.EntityDescriptor, error) {
	if u := strings.TrimSpace(cfg.IDPMetadataURL); u != "" {
		parsed, err := url.Parse(u)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return nil, fmt.Errorf("invalid idp metadata url")
		}
		fctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		desc, err := samlsp.FetchMetadata(fctx, m.httpClient, *parsed)
		if err == nil {
			return desc, nil
		}
		if strings.TrimSpace(cfg.IDPMetadataXML) == "" {
			return nil, fmt.Errorf("fetching idp metadata: %w", err)
		}
		m.lo.Warn("error fetching saml idp metadata, using stored xml", "error", err)
	}
	if strings.TrimSpace(cfg.IDPMetadataXML) == "" {
		return nil, errors.New("idp metadata url or xml is required")
	}
	desc, err := samlsp.ParseMetadata([]byte(cfg.IDPMetadataXML))
	if err != nil {
		return nil, fmt.Errorf("parsing idp metadata: %w", err)
	}
	return desc, nil
}

// serviceProvider returns the current SP, refreshing URL-based IdP metadata when it is stale.
func (m *Manager) serviceProvider(ctx context.Context) (*saml.ServiceProvider, error) {
	m.mu.RLock()
	sp, enabled, stale := m.sp, m.cfg.Enabled, m.cfg.IDPMetadataURL != "" && time.Since(m.fetchedAt) > metadataRefreshInterval
	m.mu.RUnlock()
	if !enabled {
		return nil, ErrNotConfigured
	}
	if sp != nil && !stale {
		return sp, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.buildLocked(ctx); err != nil {
		if m.sp != nil {
			m.lo.Error("error refreshing saml idp metadata, keeping previous copy", "error", err)
			m.fetchedAt = time.Now()
			return m.sp, nil
		}
		return nil, err
	}
	return m.sp, nil
}

// LoginURL builds the IdP redirect URL. It returns the AuthnRequest ID, which must be checked against the response.
func (m *Manager) LoginURL(ctx context.Context, relayState string) (string, string, error) {
	sp, err := m.serviceProvider(ctx)
	if err != nil {
		return "", "", err
	}
	req, err := sp.MakeAuthenticationRequest(sp.GetSSOBindingLocation(saml.HTTPRedirectBinding), saml.HTTPRedirectBinding, saml.HTTPPostBinding)
	if err != nil {
		return "", "", err
	}
	u, err := req.Redirect(url.QueryEscape(relayState), sp)
	if err != nil {
		return "", "", err
	}
	return u.String(), req.ID, nil
}

// ParseResponse validates a base64 SAMLResponse posted to the ACS and returns the asserted email.
// possibleRequestIDs holds the IDs of AuthnRequests this browser started; it is empty for IdP-initiated logins.
func (m *Manager) ParseResponse(ctx context.Context, samlResponse string, possibleRequestIDs []string) (string, error) {
	sp, err := m.serviceProvider(ctx)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return "", fmt.Errorf("decoding saml response: %w", err)
	}
	assertion, err := sp.ParseXMLResponse(raw, possibleRequestIDs, sp.AcsURL)
	if err != nil {
		var ire *saml.InvalidResponseError
		if errors.As(err, &ire) {
			return "", fmt.Errorf("invalid saml response: %w", ire.PrivateErr)
		}
		return "", err
	}
	email := EmailFromAssertion(assertion, m.Config().EmailAttribute)
	if email == "" {
		return "", errors.New("saml assertion has no email")
	}
	return email, nil
}

// Metadata returns the SP metadata document.
func (m *Manager) Metadata(ctx context.Context) ([]byte, error) {
	sp, err := m.serviceProvider(ctx)
	if err != nil {
		return nil, err
	}
	return xml.MarshalIndent(sp.Metadata(), "", "  ")
}

// emailAttributeNames are the attribute names IdPs commonly use for the user's email.
var emailAttributeNames = []string{
	"email",
	"mail",
	"emailaddress",
	"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
	"urn:oid:0.9.2342.19200300.100.1.3",
	"user.email",
}

// EmailFromAssertion reads the email from a configured attribute, common email attributes, or an email-shaped NameID.
func EmailFromAssertion(a *saml.Assertion, attribute string) string {
	names := emailAttributeNames
	if attribute = strings.TrimSpace(attribute); attribute != "" {
		names = []string{attribute}
	}
	for _, stmt := range a.AttributeStatements {
		for _, attr := range stmt.Attributes {
			for _, name := range names {
				if !strings.EqualFold(attr.Name, name) && !strings.EqualFold(attr.FriendlyName, name) {
					continue
				}
				for _, v := range attr.Values {
					if e := normalizeEmail(v.Value); e != "" {
						return e
					}
				}
			}
		}
	}
	if a.Subject != nil && a.Subject.NameID != nil {
		return normalizeEmail(a.Subject.NameID.Value)
	}
	return ""
}

func normalizeEmail(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if !strings.Contains(v, "@") {
		return ""
	}
	return v
}

// GenerateKeyPair creates a self-signed RSA certificate for signing AuthnRequests and decrypting assertions.
func GenerateKeyPair(commonName string) (certPEM, keyPEM string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	if err != nil {
		return "", "", err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	return certPEM, keyPEM, nil
}

// ParseKeyPair parses PEM encoded SP certificate and RSA private key.
func ParseKeyPair(certPEM, keyPEM string) (crypto.Signer, *x509.Certificate, error) {
	cb, _ := pem.Decode([]byte(certPEM))
	if cb == nil {
		return nil, nil, errors.New("invalid sp certificate pem")
	}
	cert, err := x509.ParseCertificate(cb.Bytes)
	if err != nil {
		return nil, nil, err
	}
	kb, _ := pem.Decode([]byte(keyPEM))
	if kb == nil {
		return nil, nil, errors.New("invalid sp private key pem")
	}
	if k, err := x509.ParsePKCS1PrivateKey(kb.Bytes); err == nil {
		return k, cert, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(kb.Bytes)
	if err != nil {
		return nil, nil, err
	}
	signer, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("sp private key must be rsa")
	}
	return signer, cert, nil
}

// HTTPClient returns the client used to fetch IdP metadata.
func (m *Manager) HTTPClient() *http.Client {
	return m.httpClient
}
