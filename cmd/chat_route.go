package main

import "strings"

const (
	livechatIntentSales   = "sales"
	livechatIntentSupport = "support"
)

// normalizeLivechatIntent maps a widget starter (or the first message) to sales or
// support. The black "Start conversation" button and sales starters land on Sales;
// "Get help with my account" is Support.
func normalizeLivechatIntent(raw, message string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case livechatIntentSales:
		return livechatIntentSales
	case livechatIntentSupport:
		return livechatIntentSupport
	}
	msg := strings.ToLower(strings.TrimSpace(message))
	if strings.Contains(msg, "help with my account") {
		return livechatIntentSupport
	}
	return livechatIntentSales
}

func livechatRouteSubject(intent string) string {
	if intent == livechatIntentSupport {
		return "Account support"
	}
	return "Sales"
}

func livechatTeamName(intent string) string {
	if intent == livechatIntentSupport {
		return "Support"
	}
	return "Sales"
}

func (app *App) livechatTeamID(intent string) int {
	want := strings.ToLower(livechatTeamName(intent))
	teams, err := app.team.GetAll()
	if err != nil {
		app.lo.Error("error listing teams for livechat routing", "error", err)
		return 0
	}
	for _, t := range teams {
		if strings.ToLower(strings.TrimSpace(t.Name)) == want {
			return t.ID
		}
	}
	app.lo.Warn("livechat routing team not found", "intent", intent, "team", livechatTeamName(intent))
	return 0
}
