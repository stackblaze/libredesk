package main

import (
	"testing"

	authzModels "github.com/abhinavxd/libredesk/internal/authz/models"
	umodels "github.com/abhinavxd/libredesk/internal/user/models"
	"github.com/lib/pq"
	"github.com/volatiletech/null/v9"
)

func TestMCPStaffToolsRegistered(t *testing.T) {
	tools := mcpTools(&App{}, umodels.User{})
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{
		"list_conversations", "get_conversation", "list_messages", "search_conversations",
		"search_contacts", "send_reply", "send_note", "update_status",
		"get_me", "search_messages", "update_priority", "assign_conversation",
		"unassign_conversation", "update_tags", "create_conversation", "apply_macro",
		"get_contact", "update_contact", "get_contact_notes", "add_contact_note",
		"list_statuses", "list_priorities", "list_tags", "list_teams",
		"list_agents", "list_inboxes", "list_macros",
	} {
		if !names[want] {
			t.Fatalf("missing MCP tool %s", want)
		}
	}
}

func TestMCPGetMe(t *testing.T) {
	user := umodels.User{
		ID:          3,
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       null.StringFrom("ada@example.com"),
		Roles:       pq.StringArray{"Agent"},
		Permissions: pq.StringArray{authzModels.PermMessagesWrite},
	}
	got := mcpGetMe(user)
	if got["id"] != 3 || got["first_name"] != "Ada" {
		t.Fatalf("%#v", got)
	}
}

func TestHasPerm(t *testing.T) {
	user := umodels.User{Permissions: pq.StringArray{authzModels.PermMessagesWrite}}
	if !hasPerm(user, authzModels.PermMessagesWrite) {
		t.Fatal("expected perm")
	}
	if hasPerm(user, authzModels.PermConversationsReadAll) {
		t.Fatal("unexpected perm")
	}
}
