package samlsso

import (
	"context"
	"encoding/xml"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/crewjam/saml"
	"github.com/zerodha/logf"
)

type testSessions struct{ email string }

func (s testSessions) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	return &saml.Session{ID: "s1", NameID: s.email, NameIDFormat: string(saml.EmailAddressNameIDFormat), UserEmail: s.email}
}

type testSPs struct{ sp *saml.ServiceProvider }

func (p testSPs) GetServiceProvider(r *http.Request, id string) (*saml.EntityDescriptor, error) {
	return p.sp.Metadata(), nil
}

// newTestIDP returns a crewjam IdP whose metadata XML can be fed to Manager.
func newTestIDP(t *testing.T, email string) (*saml.IdentityProvider, string) {
	t.Helper()
	certPEM, keyPEM, err := GenerateKeyPair("test-idp")
	if err != nil {
		t.Fatal(err)
	}
	key, cert, err := ParseKeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	meta, _ := url.Parse("https://idp.example.com/metadata")
	sso, _ := url.Parse("https://idp.example.com/sso")
	idp := &saml.IdentityProvider{
		Key:             key,
		Signer:          key,
		Certificate:     cert,
		MetadataURL:     *meta,
		SSOURL:          *sso,
		SessionProvider: testSessions{email: email},
	}
	b, err := xml.Marshal(idp.Metadata())
	if err != nil {
		t.Fatal(err)
	}
	return idp, string(b)
}

func newTestManager(t *testing.T, idpXML string, idpInitiated bool) *Manager {
	t.Helper()
	certPEM, keyPEM, err := GenerateKeyPair("test-sp")
	if err != nil {
		t.Fatal(err)
	}
	lo := logf.New(logf.Opts{})
	m := New(http.DefaultClient, &lo)
	if err := m.Load(Config{
		Enabled:           true,
		Name:              "Test",
		IDPMetadataXML:    idpXML,
		AllowIDPInitiated: idpInitiated,
		SPCertificate:     certPEM,
		SPPrivateKey:      keyPEM,
	}, "https://desk.example.com"); err != nil {
		t.Fatal(err)
	}
	return m
}

var samlResponseRe = regexp.MustCompile(`name="SAMLResponse" value="([^"]+)"`)

// runIDP sends the AuthnRequest in loginURL to the IdP and returns the base64 SAMLResponse it posts back.
func runIDP(t *testing.T, idp *saml.IdentityProvider, m *Manager, loginURL string) string {
	t.Helper()
	idp.ServiceProviderProvider = testSPs{sp: m.sp}
	w := httptest.NewRecorder()
	idp.ServeSSO(w, httptest.NewRequest(http.MethodGet, loginURL, nil))
	match := samlResponseRe.FindStringSubmatch(w.Body.String())
	if match == nil {
		t.Fatalf("idp did not return a SAMLResponse: %d %s", w.Code, w.Body.String())
	}
	return html.UnescapeString(match[1])
}

func TestSPInitiatedLogin(t *testing.T) {
	idp, idpXML := newTestIDP(t, "Agent@Example.com")
	m := newTestManager(t, idpXML, false)

	loginURL, reqID, err := m.LoginURL(context.Background(), "relay123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(loginURL, "https://idp.example.com/sso?SAMLRequest=") || !strings.Contains(loginURL, "RelayState=relay123") {
		t.Fatalf("unexpected login url %s", loginURL)
	}

	resp := runIDP(t, idp, m, loginURL)
	email, err := m.ParseResponse(context.Background(), resp, []string{reqID})
	if err != nil {
		t.Fatal(err)
	}
	if email != "agent@example.com" {
		t.Fatalf("email = %q", email)
	}

	// The same response must not validate against a different request ID.
	if _, err := m.ParseResponse(context.Background(), resp, []string{"id-other"}); err == nil {
		t.Fatal("response accepted for an unknown request id")
	}
}

func TestIDPInitiatedNeedsOptIn(t *testing.T) {
	idp, idpXML := newTestIDP(t, "agent@example.com")
	m := newTestManager(t, idpXML, false)
	loginURL, _, err := m.LoginURL(context.Background(), "r")
	if err != nil {
		t.Fatal(err)
	}
	resp := runIDP(t, idp, m, loginURL)
	if _, err := m.ParseResponse(context.Background(), resp, nil); err == nil {
		t.Fatal("unsolicited response accepted while IdP-initiated login is off")
	}
}

func TestResponseFromOtherIDPRejected(t *testing.T) {
	_, trustedXML := newTestIDP(t, "agent@example.com")
	rogue, _ := newTestIDP(t, "agent@example.com")
	m := newTestManager(t, trustedXML, false)
	loginURL, reqID, err := m.LoginURL(context.Background(), "r")
	if err != nil {
		t.Fatal(err)
	}
	resp := runIDP(t, rogue, m, loginURL)
	if _, err := m.ParseResponse(context.Background(), resp, []string{reqID}); err == nil {
		t.Fatal("response signed by an untrusted key accepted")
	}
}

func TestDisabledManager(t *testing.T) {
	lo := logf.New(logf.Opts{})
	m := New(http.DefaultClient, &lo)
	if err := m.Load(Config{}, "https://desk.example.com"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.LoginURL(context.Background(), "r"); err != ErrNotConfigured {
		t.Fatalf("err = %v", err)
	}
}

func TestEmailFromAssertion(t *testing.T) {
	a := &saml.Assertion{
		Subject: &saml.Subject{NameID: &saml.NameID{Value: "opaque-id"}},
		AttributeStatements: []saml.AttributeStatement{{Attributes: []saml.Attribute{
			{Name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress", Values: []saml.AttributeValue{{Value: " Bob@Example.com "}}},
			{Name: "workEmail", Values: []saml.AttributeValue{{Value: "work@example.com"}}},
		}}},
	}
	if got := EmailFromAssertion(a, ""); got != "bob@example.com" {
		t.Fatalf("default attrs: %q", got)
	}
	if got := EmailFromAssertion(a, "workEmail"); got != "work@example.com" {
		t.Fatalf("custom attr: %q", got)
	}
	a.AttributeStatements = nil
	if got := EmailFromAssertion(a, ""); got != "" {
		t.Fatalf("opaque nameid should not be an email: %q", got)
	}
}
