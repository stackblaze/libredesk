package main

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	amodels "github.com/abhinavxd/libredesk/internal/auth/models"
	authzModels "github.com/abhinavxd/libredesk/internal/authz/models"
	cmodels "github.com/abhinavxd/libredesk/internal/conversation/models"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/abhinavxd/libredesk/internal/mcp"
	umodels "github.com/abhinavxd/libredesk/internal/user/models"
	"github.com/zerodha/fastglue"
)

func handleMCPOptions(r *fastglue.Request) error {
	setMCPHeaders(r)
	r.RequestCtx.SetStatusCode(http.StatusNoContent)
	return nil
}

func mcpAuth(handler fastglue.FastRequestHandler) fastglue.FastRequestHandler {
	return func(r *fastglue.Request) error {
		setMCPHeaders(r)
		app := r.Context.(*App)
		key, secret, ok := parseMCPAuth(r)
		if !ok {
			r.RequestCtx.Response.Header.Set("WWW-Authenticate", `Basic realm="libredesk-mcp"`)
			return r.SendErrorEnvelope(http.StatusUnauthorized, "API key required", nil, envelope.UnauthorizedError)
		}
		user, err := app.user.ValidateAPIKey(key, secret)
		if err != nil {
			return r.SendErrorEnvelope(http.StatusUnauthorized, "invalid API key", nil, envelope.UnauthorizedError)
		}
		r.RequestCtx.SetUserValue("user", amodels.User{
			ID:        user.ID,
			Email:     user.Email.String,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})
		r.RequestCtx.SetUserValue("auth_method", authMethodAPIKey)
		return handler(r)
	}
}

func parseMCPAuth(r *fastglue.Request) (string, string, bool) {
	key, secret, err := r.ParseAuthHeader(fastglue.AuthBasic | fastglue.AuthToken)
	if err == nil && len(key) > 0 && len(secret) > 0 {
		return string(key), string(secret), true
	}
	raw := strings.TrimSpace(string(r.RequestCtx.Request.Header.Peek("Authorization")))
	raw = strings.TrimPrefix(raw, "Bearer ")
	raw = strings.TrimPrefix(raw, "bearer ")
	if i := strings.Index(raw, ":"); i > 0 {
		return raw[:i], raw[i+1:], true
	}
	return "", "", false
}

func handleMCP(r *fastglue.Request) error {
	app := r.Context.(*App)
	if string(r.RequestCtx.Method()) == http.MethodGet || string(r.RequestCtx.Method()) == http.MethodDelete {
		return writeMCPJSON(r, map[string]any{
			"name":    "libredesk",
			"version": buildString,
		})
	}
	auser := r.RequestCtx.UserValue("user").(amodels.User)
	user, err := app.user.GetAgentCachedOrLoad(auser.ID)
	if err != nil {
		return r.SendErrorEnvelope(http.StatusUnauthorized, "invalid agent", nil, envelope.UnauthorizedError)
	}
	srv := mcp.Server{
		Name:    "libredesk",
		Version: buildString,
		Tools:   mcpTools(app, user),
	}
	resp := srv.Handle(r.RequestCtx.PostBody())
	if resp == nil {
		r.RequestCtx.SetStatusCode(http.StatusAccepted)
		return nil
	}
	return writeMCPJSON(r, resp)
}

func setMCPHeaders(r *fastglue.Request) {
	h := &r.RequestCtx.Response.Header
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Mcp-Session-Id, MCP-Protocol-Version")
	h.Set("Access-Control-Expose-Headers", "Mcp-Session-Id, MCP-Protocol-Version")
	h.Set("MCP-Protocol-Version", mcp.ProtocolVersion)
}

func writeMCPJSON(r *fastglue.Request, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return r.SendErrorEnvelope(http.StatusInternalServerError, "encode error", nil, envelope.GeneralError)
	}
	r.RequestCtx.SetContentType("application/json")
	r.RequestCtx.SetStatusCode(http.StatusOK)
	r.RequestCtx.SetBody(b)
	return nil
}

func mcpTools(app *App, user umodels.User) []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "list_conversations",
			Description: "List tickets. Use list=assigned (default), unassigned, mentioned, or all.",
			InputSchema: objectSchema(map[string]any{
				"list":      prop("string", "assigned | unassigned | mentioned | all | team_unassigned"),
				"team_id":   prop("integer", "Required when list is team_unassigned"),
				"page":      prop("integer", "Page number, default 1"),
				"page_size": prop("integer", "Page size, default 20"),
			}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListConversations(app, user, args)
			},
		},
		{
			Name:        "get_conversation",
			Description: "Get one ticket by UUID.",
			InputSchema: objectSchema(map[string]any{
				"uuid": prop("string", "Conversation UUID"),
			}, []string{"uuid"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpGetConversation(app, user, args)
			},
		},
		{
			Name:        "list_messages",
			Description: "List messages in a ticket.",
			InputSchema: objectSchema(map[string]any{
				"uuid":      prop("string", "Conversation UUID"),
				"page":      prop("integer", "Page number, default 1"),
				"page_size": prop("integer", "Page size, default 30"),
			}, []string{"uuid"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpListMessages(app, user, args)
			},
		},
		{
			Name:        "search_conversations",
			Description: "Search tickets by subject or reference. Query must be at least 3 characters.",
			InputSchema: objectSchema(map[string]any{
				"query": prop("string", "Search query"),
			}, []string{"query"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpSearchConversations(app, user, mcp.StrArg(args, "query"))
			},
		},
		{
			Name:        "search_contacts",
			Description: "Search contacts by name or email. Query must be at least 3 characters.",
			InputSchema: objectSchema(map[string]any{
				"query": prop("string", "Search query"),
			}, []string{"query"}),
			Handler: func(args map[string]any) (any, error) {
				if !hasPerm(user, authzModels.PermContactsRead) && !hasPerm(user, authzModels.PermContactsReadAll) {
					return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
				}
				q := strings.TrimSpace(mcp.StrArg(args, "query"))
				if len(q) < 3 {
					return nil, envelope.NewError(envelope.InputError, "query must be at least 3 characters", nil)
				}
				return app.search.Contacts(q)
			},
		},
		{
			Name:        "send_reply",
			Description: "Send a public reply on a ticket.",
			InputSchema: objectSchema(map[string]any{
				"uuid":    prop("string", "Conversation UUID"),
				"message": prop("string", "Reply body (plain text or HTML)"),
			}, []string{"uuid", "message"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpSendMessage(app, user, args, false)
			},
		},
		{
			Name:        "send_note",
			Description: "Add a private agent note on a ticket.",
			InputSchema: objectSchema(map[string]any{
				"uuid":    prop("string", "Conversation UUID"),
				"message": prop("string", "Note body (plain text or HTML)"),
			}, []string{"uuid", "message"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpSendMessage(app, user, args, true)
			},
		},
		{
			Name:        "update_status",
			Description: "Set a ticket status (for example Open, Replied, Resolved, Closed, Snoozed).",
			InputSchema: objectSchema(map[string]any{
				"uuid":          prop("string", "Conversation UUID"),
				"status":        prop("string", "Status name"),
				"snoozed_until": prop("string", "Go duration such as 2h; required when status is Snoozed"),
			}, []string{"uuid", "status"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpUpdateStatus(app, user, args)
			},
		},
		{
			Name:        "get_me",
			Description: "Return the authenticated agent (id, name, email, teams, permissions).",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpGetMe(user), nil
			},
		},
		{
			Name:        "search_messages",
			Description: "Search message bodies. Query must be at least 3 characters.",
			InputSchema: objectSchema(map[string]any{
				"query": prop("string", "Search query"),
			}, []string{"query"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpSearchMessages(app, user, mcp.StrArg(args, "query"))
			},
		},
		{
			Name:        "update_priority",
			Description: "Set a ticket priority by name (use list_priorities).",
			InputSchema: objectSchema(map[string]any{
				"uuid":     prop("string", "Conversation UUID"),
				"priority": prop("string", "Priority name"),
			}, []string{"uuid", "priority"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpUpdatePriority(app, user, args)
			},
		},
		{
			Name:        "assign_conversation",
			Description: "Assign a ticket to an agent and/or team. Set assign_to_me true to take it yourself.",
			InputSchema: objectSchema(map[string]any{
				"uuid":         prop("string", "Conversation UUID"),
				"user_id":      prop("integer", "Agent id"),
				"team_id":      prop("integer", "Team id"),
				"assign_to_me": prop("boolean", "Assign to the authenticated agent"),
			}, []string{"uuid"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpAssign(app, user, args)
			},
		},
		{
			Name:        "unassign_conversation",
			Description: "Remove the assigned agent, team, or both. assignee is user, team, or both (default both).",
			InputSchema: objectSchema(map[string]any{
				"uuid":     prop("string", "Conversation UUID"),
				"assignee": prop("string", "user | team | both"),
			}, []string{"uuid"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpUnassign(app, user, args)
			},
		},
		{
			Name:        "update_tags",
			Description: "Change ticket tags. action is set (replace), add, or remove. Default set.",
			InputSchema: objectSchema(map[string]any{
				"uuid":   prop("string", "Conversation UUID"),
				"tags":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tag names"},
				"action": prop("string", "set | add | remove"),
			}, []string{"uuid", "tags"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpUpdateTags(app, user, args)
			},
		},
		{
			Name:        "create_conversation",
			Description: "Open a new email conversation. initiator is agent (default) or contact.",
			InputSchema: objectSchema(map[string]any{
				"inbox_id":      prop("integer", "Enabled email inbox id"),
				"contact_email": prop("string", "Contact email"),
				"first_name":    prop("string", "Contact first name"),
				"last_name":     prop("string", "Contact last name"),
				"subject":       prop("string", "Subject"),
				"content":       prop("string", "Initial message"),
				"initiator":     prop("string", "agent | contact"),
				"agent_id":      prop("integer", "Optional assigned agent"),
				"team_id":       prop("integer", "Optional assigned team"),
			}, []string{"inbox_id", "contact_email", "first_name", "content"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpCreateConversation(app, user, args)
			},
		},
		{
			Name:        "apply_macro",
			Description: "Apply a saved macro's actions. If it has message_content, send it as a reply unless private is true.",
			InputSchema: objectSchema(map[string]any{
				"uuid":     prop("string", "Conversation UUID"),
				"macro_id": prop("integer", "Macro id from list_macros"),
				"private":  prop("boolean", "Send macro message as a private note"),
			}, []string{"uuid", "macro_id"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpApplyMacro(app, user, args)
			},
		},
		{
			Name:        "get_contact",
			Description: "Get a contact by numeric id or email.",
			InputSchema: objectSchema(map[string]any{
				"id":    prop("integer", "Contact id"),
				"email": prop("string", "Contact email"),
			}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpGetContact(app, user, args)
			},
		},
		{
			Name:        "update_contact",
			Description: "Update a contact. Omitted fields are left unchanged.",
			InputSchema: objectSchema(map[string]any{
				"id":         prop("integer", "Contact id"),
				"first_name": prop("string", "First name"),
				"last_name":  prop("string", "Last name"),
				"email":      prop("string", "Email"),
				"phone":      prop("string", "Phone number"),
				"phone_code": prop("string", "Phone country code"),
				"country":    prop("string", "Country"),
			}, []string{"id"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpUpdateContact(app, user, args)
			},
		},
		{
			Name:        "get_contact_notes",
			Description: "List internal notes on a contact.",
			InputSchema: objectSchema(map[string]any{
				"id": prop("integer", "Contact id"),
			}, []string{"id"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpGetContactNotes(app, user, args)
			},
		},
		{
			Name:        "add_contact_note",
			Description: "Add an internal note to a contact profile.",
			InputSchema: objectSchema(map[string]any{
				"id":   prop("integer", "Contact id"),
				"note": prop("string", "Note text"),
			}, []string{"id", "note"}),
			Handler: func(args map[string]any) (any, error) {
				return mcpAddContactNote(app, user, args)
			},
		},
		{
			Name:        "list_statuses",
			Description: "List conversation statuses.",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListStatuses(app)
			},
		},
		{
			Name:        "list_priorities",
			Description: "List conversation priorities.",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListPriorities(app)
			},
		},
		{
			Name:        "list_tags",
			Description: "List tags that can be applied to conversations.",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListTags(app)
			},
		},
		{
			Name:        "list_teams",
			Description: "List teams for assignment.",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListTeams(app)
			},
		},
		{
			Name:        "list_agents",
			Description: "List agents for assignment.",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListAgents(app)
			},
		},
		{
			Name:        "list_inboxes",
			Description: "List inboxes (id, name, channel, enabled). Secrets are never returned.",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListInboxes(app)
			},
		},
		{
			Name:        "list_macros",
			Description: "List saved macros (id, name, visibility, actions, message_content).",
			InputSchema: objectSchema(map[string]any{}, nil),
			Handler: func(args map[string]any) (any, error) {
				return mcpListMacros(app)
			},
		},
	}
}

func mcpListConversations(app *App, user umodels.User, args map[string]any) (any, error) {
	page := mcp.IntArg(args, "page", 1)
	pageSize := mcp.IntArg(args, "page_size", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	list := strings.ToLower(strings.TrimSpace(mcp.StrArg(args, "list")))
	if list == "" {
		list = "assigned"
	}
	var (
		items []cmodels.ConversationListItem
		err   error
	)
	switch list {
	case "assigned":
		if !hasPerm(user, authzModels.PermConversationsReadAssigned) && !hasPerm(user, authzModels.PermConversationsRead) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		items, err = app.conversation.GetAssignedConversationsList(user.ID, user.ID, "", "", "", page, pageSize)
	case "unassigned":
		if !hasPerm(user, authzModels.PermConversationsReadUnassigned) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		items, err = app.conversation.GetUnassignedConversationsList(user.ID, "", "", "", page, pageSize)
	case "mentioned":
		if !hasPerm(user, authzModels.PermConversationsRead) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		items, err = app.conversation.GetMentionedConversationsList(user.ID, "", "", "", page, pageSize)
	case "all":
		if !hasPerm(user, authzModels.PermConversationsReadAll) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		items, err = app.conversation.GetAllConversationsList(user.ID, "", "", "", page, pageSize)
	case "team_unassigned":
		if !hasPerm(user, authzModels.PermConversationsReadTeamInbox) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		teamID := mcp.IntArg(args, "team_id", 0)
		if teamID < 1 {
			return nil, envelope.NewError(envelope.InputError, "team_id is required when list is team_unassigned", nil)
		}
		ok, teamErr := app.team.UserBelongsToTeam(teamID, user.ID)
		if teamErr != nil {
			return nil, teamErr
		}
		if !ok {
			return nil, envelope.NewError(envelope.PermissionError, "not a member of that team", nil)
		}
		items, err = app.conversation.GetTeamUnassignedConversationsList(user.ID, teamID, "", "", "", page, pageSize)
	default:
		return nil, envelope.NewError(envelope.InputError, "list must be assigned, unassigned, mentioned, all, or team_unassigned", nil)
	}
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(items))
	total := 0
	for _, item := range items {
		total = item.Total
		out = append(out, compactListItem(item))
	}
	return map[string]any{"page": page, "page_size": pageSize, "total": total, "conversations": out}, nil
}

func mcpListMessages(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermMessagesRead) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	if _, err := enforceConversationAccess(app, uuid, user); err != nil {
		return nil, err
	}
	page := mcp.IntArg(args, "page", 1)
	pageSize := mcp.IntArg(args, "page_size", 30)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 30
	}
	messages, pageSize, err := app.conversation.GetConversationMessages(uuid, page, pageSize, nil, nil)
	if err != nil {
		return nil, err
	}
	total := 0
	out := make([]map[string]any, 0, len(messages))
	for _, m := range messages {
		total = m.Total
		text := m.TextContent
		if text == "" {
			text = m.Content
		}
		out = append(out, map[string]any{
			"uuid":        m.UUID,
			"created_at":  m.CreatedAt,
			"type":        m.Type,
			"private":     m.Private,
			"sender_type": m.SenderType,
			"author":      m.Author.FullName(),
			"text":        text,
		})
	}
	return map[string]any{"page": page, "page_size": pageSize, "total": total, "messages": out}, nil
}

func mcpSearchConversations(app *App, user umodels.User, query string) (any, error) {
	if !hasPerm(user, authzModels.PermConversationsRead) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	q := strings.TrimSpace(query)
	if len(q) < 3 {
		return nil, envelope.NewError(envelope.InputError, "query must be at least 3 characters", nil)
	}
	results, err := app.search.Conversations(q)
	if err != nil {
		return nil, err
	}
	uuids := make([]string, len(results))
	for i, c := range results {
		uuids[i] = c.UUID
	}
	allowed, err := app.conversation.FilterAuthorizedListUUIDs(user.ID, uuids)
	if err != nil {
		return nil, err
	}
	set := uuidSet(allowed)
	out := make([]any, 0, len(allowed))
	for _, c := range results {
		if _, ok := set[c.UUID]; ok {
			out = append(out, c)
		}
	}
	return out, nil
}

func mcpSendMessage(app *App, user umodels.User, args map[string]any, private bool) (any, error) {
	perm := authzModels.PermMessagesWrite
	if private {
		perm = authzModels.PermMessagesWritePrivate
	}
	if !hasPerm(user, perm) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	body := strings.TrimSpace(mcp.StrArg(args, "message"))
	if body == "" {
		return nil, envelope.NewError(envelope.InputError, "message is required", nil)
	}
	conv, err := enforceConversationAccess(app, uuid, user)
	if err != nil {
		return nil, err
	}
	inbox, err := app.inbox.GetDBRecord(conv.InboxID)
	if err != nil {
		return nil, err
	}
	if !inbox.Enabled {
		return nil, envelope.NewError(envelope.InputError, "inbox is disabled", nil)
	}
	htmlBody := mcpHTML(body)
	if private {
		msg, err := app.conversation.SendPrivateNote(nil, user.ID, uuid, htmlBody, nil)
		if err != nil {
			return nil, err
		}
		return map[string]any{"ok": true, "uuid": msg.UUID, "private": true}, nil
	}
	msg, err := app.conversation.QueueReply(nil, conv.InboxID, user.ID, conv.ContactID, uuid, htmlBody, nil, nil, nil, map[string]any{})
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "uuid": msg.UUID, "private": false}, nil
}

func compactListItem(item cmodels.ConversationListItem) map[string]any {
	return map[string]any{
		"uuid":             item.UUID,
		"reference_number": item.ReferenceNumber,
		"subject":          item.Subject.String,
		"status":           item.Status.String,
		"priority":         item.Priority.String,
		"inbox":            item.InboxName,
		"contact":          strings.TrimSpace(item.Contact.FirstName + " " + item.Contact.LastName),
		"contact_email":    item.Contact.Email.String,
		"assigned_user_id": item.AssignedUserID,
		"assigned_team_id": item.AssignedTeamID,
		"last_message":     item.LastMessage.String,
		"updated_at":       item.UpdatedAt,
	}
}

func compactConversation(conv cmodels.Conversation) map[string]any {
	return map[string]any{
		"uuid":             conv.UUID,
		"reference_number": conv.ReferenceNumber,
		"subject":          conv.Subject.String,
		"status":           conv.Status.String,
		"priority":         conv.Priority.String,
		"inbox":            conv.InboxName,
		"channel":          conv.InboxChannel,
		"contact_id":       conv.ContactID,
		"contact":          conv.Contact.FullName(),
		"contact_email":    conv.Contact.Email.String,
		"assigned_user_id": conv.AssignedUserID,
		"assigned_team_id": conv.AssignedTeamID,
		"tags":             conv.Tags,
		"last_message":     conv.LastMessage.String,
		"updated_at":       conv.UpdatedAt,
	}
}

func mcpHTML(s string) string {
	if strings.Contains(s, "<") && strings.Contains(s, ">") {
		return s
	}
	return "<p>" + html.EscapeString(s) + "</p>"
}

func hasPerm(user umodels.User, perm string) bool {
	for _, p := range user.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{"type": "object", "properties": properties}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func prop(typ, description string) map[string]any {
	return map[string]any{"type": typ, "description": description}
}
