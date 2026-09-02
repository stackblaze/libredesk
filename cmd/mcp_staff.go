package main

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	authzModels "github.com/abhinavxd/libredesk/internal/authz/models"
	autoModels "github.com/abhinavxd/libredesk/internal/automation/models"
	cmodels "github.com/abhinavxd/libredesk/internal/conversation/models"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/abhinavxd/libredesk/internal/mcp"
	"github.com/abhinavxd/libredesk/internal/stringutil"
	umodels "github.com/abhinavxd/libredesk/internal/user/models"
	wmodels "github.com/abhinavxd/libredesk/internal/webhook/models"
	"github.com/volatiletech/null/v9"
)

func mcpGetMe(user umodels.User) map[string]any {
	teams := make([]map[string]any, 0, len(user.Teams))
	for _, t := range user.Teams {
		teams = append(teams, map[string]any{"id": t.ID, "name": t.Name})
	}
	return map[string]any{
		"id":           user.ID,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"email":        user.Email.String,
		"availability": user.AvailabilityStatus,
		"roles":        user.Roles,
		"permissions":  user.Permissions,
		"teams":        teams,
	}
}

func mcpGetConversation(app *App, user umodels.User, args map[string]any) (any, error) {
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	conv, err := enforceConversationAccess(app, uuid, user)
	if err != nil {
		return nil, err
	}
	out := compactConversation(*conv)
	if prev, err := app.conversation.GetContactPreviousConversations(conv.ContactID, 10); err == nil {
		items := make([]map[string]any, 0, len(prev))
		for _, p := range prev {
			if p.UUID == conv.UUID {
				continue
			}
			items = append(items, map[string]any{
				"uuid":            p.UUID,
				"subject":         p.Subject,
				"last_message":    p.LastMessage.String,
				"last_message_at": p.LastMessageAt,
			})
		}
		out["previous"] = items
	}
	return out, nil
}

func mcpUpdateStatus(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermConversationsUpdateStatus) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	status := strings.TrimSpace(mcp.StrArg(args, "status"))
	snooze := strings.TrimSpace(mcp.StrArg(args, "snoozed_until"))
	if status == "" {
		return nil, envelope.NewError(envelope.InputError, "status is required", nil)
	}
	if status == cmodels.StatusSnoozed {
		if snooze == "" {
			return nil, envelope.NewError(envelope.InputError, "snoozed_until is required when status is Snoozed", nil)
		}
		if _, err := time.ParseDuration(snooze); err != nil {
			return nil, envelope.NewError(envelope.InputError, "snoozed_until must be a duration such as 2h or 24h", nil)
		}
	}
	conv, err := enforceConversationAccess(app, uuid, user)
	if err != nil {
		return nil, err
	}
	if err := app.conversation.UpdateConversationStatus(uuid, 0, status, snooze, user); err != nil {
		return nil, err
	}
	if status == cmodels.StatusResolved {
		if inbox, err := app.inbox.GetDBRecord(conv.InboxID); err == nil && inbox.CSATEnabled {
			_ = app.conversation.SendCSATReply(user.ID, *conv)
		}
	}
	return map[string]any{"ok": true, "uuid": uuid, "status": status}, nil
}

func mcpSearchMessages(app *App, user umodels.User, query string) (any, error) {
	if !hasPerm(user, authzModels.PermMessagesRead) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	q := strings.TrimSpace(query)
	if len(q) < 3 {
		return nil, envelope.NewError(envelope.InputError, "query must be at least 3 characters", nil)
	}
	results, err := app.search.Messages(q)
	if err != nil {
		return nil, err
	}
	uuids := make([]string, len(results))
	for i, m := range results {
		uuids[i] = m.ConversationUUID
	}
	allowed, err := app.conversation.FilterAuthorizedListUUIDs(user.ID, uuids)
	if err != nil {
		return nil, err
	}
	set := uuidSet(allowed)
	out := make([]any, 0, len(allowed))
	for _, m := range results {
		if _, ok := set[m.ConversationUUID]; ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func mcpUpdatePriority(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermConversationsUpdatePriority) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	priority := strings.TrimSpace(mcp.StrArg(args, "priority"))
	if priority == "" {
		return nil, envelope.NewError(envelope.InputError, "priority is required", nil)
	}
	if _, err := enforceConversationAccess(app, uuid, user); err != nil {
		return nil, err
	}
	if err := app.conversation.UpdateConversationPriority(uuid, 0, priority, user); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "uuid": uuid, "priority": priority}, nil
}

func mcpAssign(app *App, user umodels.User, args map[string]any) (any, error) {
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	conv, err := enforceConversationAccess(app, uuid, user)
	if err != nil {
		return nil, err
	}
	userID := mcp.IntArg(args, "user_id", 0)
	teamID := mcp.IntArg(args, "team_id", 0)
	if mcp.BoolArg(args, "assign_to_me") {
		userID = user.ID
	}
	if userID == 0 && teamID == 0 {
		return nil, envelope.NewError(envelope.InputError, "provide user_id, team_id, or assign_to_me", nil)
	}
	if teamID > 0 {
		if !hasPerm(user, authzModels.PermConversationsUpdateTeamAssignee) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		if _, err := app.team.Get(teamID); err != nil {
			return nil, err
		}
		if conv.AssignedTeamID.Int != teamID {
			if err := app.conversation.UpdateConversationTeamAssignee(uuid, teamID, user); err != nil {
				return nil, err
			}
		}
	}
	if userID > 0 {
		if !hasPerm(user, authzModels.PermConversationsUpdateUserAssignee) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		if conv.AssignedUserID.Int != userID {
			if err := app.conversation.UpdateConversationUserAssignee(uuid, userID, user); err != nil {
				return nil, err
			}
		}
	}
	return map[string]any{"ok": true, "uuid": uuid, "user_id": userID, "team_id": teamID}, nil
}

func mcpUnassign(app *App, user umodels.User, args map[string]any) (any, error) {
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	if _, err := enforceConversationAccess(app, uuid, user); err != nil {
		return nil, err
	}
	who := strings.ToLower(strings.TrimSpace(mcp.StrArg(args, "assignee")))
	if who == "" {
		who = "both"
	}
	switch who {
	case "user":
		if !hasPerm(user, authzModels.PermConversationsUpdateUserAssignee) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		if err := app.conversation.RemoveConversationAssignee(uuid, "user", user); err != nil {
			return nil, err
		}
	case "team":
		if !hasPerm(user, authzModels.PermConversationsUpdateTeamAssignee) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		if err := app.conversation.RemoveConversationAssignee(uuid, "team", user); err != nil {
			return nil, err
		}
	case "both":
		if !hasPerm(user, authzModels.PermConversationsUpdateUserAssignee) || !hasPerm(user, authzModels.PermConversationsUpdateTeamAssignee) {
			return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
		}
		if err := app.conversation.RemoveConversationAssignee(uuid, "user", user); err != nil {
			return nil, err
		}
		if err := app.conversation.RemoveConversationAssignee(uuid, "team", user); err != nil {
			return nil, err
		}
	default:
		return nil, envelope.NewError(envelope.InputError, "assignee must be user, team, or both", nil)
	}
	return map[string]any{"ok": true, "uuid": uuid, "unassigned": who}, nil
}

func mcpUpdateTags(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermConversationsUpdateTags) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	if _, err := enforceConversationAccess(app, uuid, user); err != nil {
		return nil, err
	}
	action := strings.TrimSpace(mcp.StrArg(args, "action"))
	switch action {
	case autoModels.ActionAddTags, autoModels.ActionRemoveTags, autoModels.ActionSetTags:
	case "":
		action = autoModels.ActionSetTags
	default:
		return nil, envelope.NewError(envelope.InputError, "action must be set, add, or remove", nil)
	}
	tags := mcp.StringSliceArg(args, "tags")
	if len(tags) > 0 {
		if err := app.tag.EnsureExist(tags); err != nil {
			return nil, err
		}
	}
	if err := app.conversation.SetConversationTags(uuid, action, tags, user); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "uuid": uuid, "action": action, "tags": tags}, nil
}

func mcpCreateConversation(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermConversationWrite) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	inboxID := mcp.IntArg(args, "inbox_id", 0)
	email := strings.TrimSpace(mcp.StrArg(args, "contact_email"))
	firstName := strings.TrimSpace(mcp.StrArg(args, "first_name"))
	lastName := strings.TrimSpace(mcp.StrArg(args, "last_name"))
	subject := strings.TrimSpace(mcp.StrArg(args, "subject"))
	content := strings.TrimSpace(mcp.StrArg(args, "content"))
	initiator := strings.TrimSpace(mcp.StrArg(args, "initiator"))
	if initiator == "" {
		initiator = umodels.UserTypeAgent
	}
	if inboxID <= 0 || email == "" || firstName == "" || content == "" {
		return nil, envelope.NewError(envelope.InputError, "inbox_id, contact_email, first_name, and content are required", nil)
	}
	if !stringutil.ValidEmail(email) {
		return nil, envelope.NewError(envelope.InputError, "invalid contact_email", nil)
	}
	if initiator != umodels.UserTypeAgent && initiator != umodels.UserTypeContact {
		return nil, envelope.NewError(envelope.InputError, "initiator must be agent or contact", nil)
	}
	inbox, err := app.inbox.GetDBRecord(inboxID)
	if err != nil {
		return nil, err
	}
	if !inbox.Enabled {
		return nil, envelope.NewError(envelope.InputError, "inbox is disabled", nil)
	}
	if inbox.Channel != "email" {
		return nil, envelope.NewError(envelope.InputError, "create_conversation only supports email inboxes", nil)
	}

	contact := umodels.User{
		Email:            null.StringFrom(email),
		FirstName:        firstName,
		LastName:         lastName,
		CustomAttributes: json.RawMessage(`{}`),
	}
	if err := app.user.ResolveContact(&contact, umodels.ContactReuse); err != nil {
		return nil, err
	}
	_, conversationUUID, err := app.conversation.CreateConversation(
		contact.ID, inboxID, "", time.Now(), subject, true, nil, nil, 0, 0,
	)
	if err != nil {
		return nil, err
	}
	if teamID := mcp.IntArg(args, "team_id", 0); teamID > 0 {
		_ = app.conversation.UpdateConversationTeamAssignee(conversationUUID, teamID, user)
	}
	if agentID := mcp.IntArg(args, "agent_id", 0); agentID > 0 {
		_ = app.conversation.UpdateConversationUserAssignee(conversationUUID, agentID, user)
	}

	htmlBody := mcpHTML(content)
	switch initiator {
	case umodels.UserTypeAgent:
		if _, err := app.conversation.QueueReply(nil, inboxID, user.ID, contact.ID, conversationUUID, htmlBody, []string{email}, nil, nil, map[string]any{}); err != nil {
			_ = app.conversation.DeleteConversation(conversationUUID)
			return nil, err
		}
		if c, err := app.conversation.GetConversation(0, conversationUUID, ""); err == nil {
			app.webhook.TriggerEvent(wmodels.EventConversationCreated, c)
		}
	case umodels.UserTypeContact:
		if _, err := app.conversation.CreateContactMessage(nil, contact.ID, conversationUUID, htmlBody, cmodels.ContentTypeHTML, true, ""); err != nil {
			_ = app.conversation.DeleteConversation(conversationUUID)
			return nil, err
		}
	}

	created, err := app.conversation.GetConversation(0, conversationUUID, "")
	if err != nil {
		return map[string]any{"ok": true, "uuid": conversationUUID}, nil
	}
	return compactConversation(created), nil
}

func mcpApplyMacro(app *App, user umodels.User, args map[string]any) (any, error) {
	uuid := strings.TrimSpace(mcp.StrArg(args, "uuid"))
	macroID := mcp.IntArg(args, "macro_id", 0)
	if macroID < 1 {
		return nil, envelope.NewError(envelope.InputError, "macro_id is required", nil)
	}
	conv, err := enforceConversationAccess(app, uuid, user)
	if err != nil {
		return nil, err
	}
	macro, err := app.macro.Get(macroID)
	if err != nil {
		return nil, err
	}
	var actions []autoModels.RuleAction
	if err := json.Unmarshal(macro.Actions, &actions); err != nil {
		return nil, envelope.NewError(envelope.InputError, "invalid macro actions", nil)
	}

	applied := 0
	skipped := make([]string, 0)
	for _, act := range actions {
		if !isMacroActionAllowed(act.Type) {
			skipped = append(skipped, act.Type+" (not allowed)")
			continue
		}
		required, ok := autoModels.ActionPermissions[act.Type]
		if !ok || !slices.Contains(user.Permissions, required) {
			skipped = append(skipped, act.Type+" (permission denied)")
			continue
		}
		if err := app.conversation.ApplyAction(act, *conv, user); err == nil {
			applied++
		} else {
			skipped = append(skipped, act.Type+" (failed)")
		}
	}

	messageSent := false
	if strings.TrimSpace(macro.MessageContent) != "" {
		perm := authzModels.PermMessagesWrite
		if mcp.BoolArg(args, "private") {
			perm = authzModels.PermMessagesWritePrivate
		}
		if hasPerm(user, perm) {
			body := mcpHTML(macro.MessageContent)
			if mcp.BoolArg(args, "private") {
				if _, err := app.conversation.SendPrivateNote(nil, user.ID, uuid, body, nil); err == nil {
					messageSent = true
				}
			} else if inbox, err := app.inbox.GetDBRecord(conv.InboxID); err == nil && inbox.Enabled {
				if _, err := app.conversation.QueueReply(nil, conv.InboxID, user.ID, conv.ContactID, uuid, body, nil, nil, nil, map[string]any{}); err == nil {
					messageSent = true
				}
			}
		}
	}
	_ = app.macro.IncrementUsageCount(macro.ID)
	if applied == 0 && !messageSent {
		return nil, envelope.NewError(envelope.GeneralError, "could not apply macro", nil)
	}
	return map[string]any{
		"ok":              true,
		"uuid":            uuid,
		"macro_id":        macro.ID,
		"macro_name":      macro.Name,
		"actions_applied": applied,
		"skipped":         skipped,
		"message_sent":    messageSent,
	}, nil
}

func mcpGetContact(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermContactsRead) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	id := mcp.IntArg(args, "id", 0)
	email := strings.TrimSpace(mcp.StrArg(args, "email"))
	if id <= 0 && email == "" {
		return nil, envelope.NewError(envelope.InputError, "id or email is required", nil)
	}
	contact, err := app.user.GetContactOrVisitor(id, email)
	if err != nil {
		return nil, err
	}
	return compactContact(contact), nil
}

func mcpUpdateContact(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermContactsWrite) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	id := mcp.IntArg(args, "id", 0)
	if id <= 0 {
		return nil, envelope.NewError(envelope.InputError, "id is required", nil)
	}
	contact, err := app.user.GetContactOrVisitor(id, "")
	if err != nil {
		return nil, err
	}
	if v := strings.TrimSpace(mcp.StrArg(args, "first_name")); v != "" {
		contact.FirstName = v
	}
	if v := strings.TrimSpace(mcp.StrArg(args, "last_name")); v != "" {
		contact.LastName = v
	}
	if v := strings.TrimSpace(mcp.StrArg(args, "email")); v != "" {
		if !stringutil.ValidEmail(v) {
			return nil, envelope.NewError(envelope.InputError, "invalid email", nil)
		}
		contact.Email = null.StringFrom(v)
	}
	if v := strings.TrimSpace(mcp.StrArg(args, "phone")); v != "" {
		contact.PhoneNumber = null.StringFrom(v)
	}
	if v := strings.TrimSpace(mcp.StrArg(args, "phone_code")); v != "" {
		contact.PhoneNumberCountryCode = null.StringFrom(v)
	}
	if v := strings.TrimSpace(mcp.StrArg(args, "country")); v != "" {
		contact.Country = null.StringFrom(v)
	}
	if !contact.Email.Valid || contact.Email.String == "" || contact.FirstName == "" {
		return nil, envelope.NewError(envelope.InputError, "contact must have first_name and email", nil)
	}
	if err := app.user.UpdateContact(id, contact); err != nil {
		return nil, err
	}
	updated, err := app.user.GetContactOrVisitor(id, "")
	if err != nil {
		return nil, err
	}
	return compactContact(updated), nil
}

func mcpGetContactNotes(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermContactNotesRead) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	id := mcp.IntArg(args, "id", 0)
	if id <= 0 {
		return nil, envelope.NewError(envelope.InputError, "id is required", nil)
	}
	return app.user.GetNotes(id)
}

func mcpAddContactNote(app *App, user umodels.User, args map[string]any) (any, error) {
	if !hasPerm(user, authzModels.PermContactNotesWrite) {
		return nil, envelope.NewError(envelope.PermissionError, "permission denied", nil)
	}
	id := mcp.IntArg(args, "id", 0)
	note := strings.TrimSpace(mcp.StrArg(args, "note"))
	if id <= 0 || note == "" {
		return nil, envelope.NewError(envelope.InputError, "id and note are required", nil)
	}
	created, err := app.user.CreateNote(id, user.ID, note)
	if err != nil {
		return nil, err
	}
	if full, err := app.user.GetNote(created.ID); err == nil {
		return full, nil
	}
	return created, nil
}

func mcpListStatuses(app *App) (any, error) {
	statuses, err := app.status.GetAll()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(statuses))
	for _, st := range statuses {
		out = append(out, map[string]any{"id": st.ID, "name": st.Name, "category": st.Category})
	}
	return out, nil
}

func mcpListPriorities(app *App) (any, error) {
	priorities, err := app.priority.GetAll()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(priorities))
	for _, p := range priorities {
		out = append(out, map[string]any{"id": p.ID, "name": p.Name})
	}
	return out, nil
}

func mcpListTags(app *App) (any, error) {
	tags, err := app.tag.GetAll()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(tags))
	for _, t := range tags {
		out = append(out, map[string]any{"id": t.ID, "name": t.Name})
	}
	return out, nil
}

func mcpListTeams(app *App) (any, error) {
	teams, err := app.team.GetAllCompact()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(teams))
	for _, t := range teams {
		item := map[string]any{"id": t.ID, "name": t.Name}
		if t.Emoji.Valid {
			item["emoji"] = t.Emoji.String
		}
		out = append(out, item)
	}
	return out, nil
}

func mcpListAgents(app *App) (any, error) {
	agents, err := app.user.GetAgentsCompact()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(agents))
	for _, a := range agents {
		out = append(out, map[string]any{
			"id":           a.ID,
			"first_name":   a.FirstName,
			"last_name":    a.LastName,
			"email":        a.Email.String,
			"availability": a.AvailabilityStatus,
			"enabled":      a.Enabled,
		})
	}
	return out, nil
}

func mcpListInboxes(app *App) (any, error) {
	inboxes, err := app.inbox.GetAll()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(inboxes))
	for _, in := range inboxes {
		out = append(out, map[string]any{
			"id":      in.ID,
			"name":    in.Name,
			"channel": in.Channel,
			"enabled": in.Enabled,
		})
	}
	return out, nil
}

func mcpListMacros(app *App) (any, error) {
	macros, err := app.macro.GetAll()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(macros))
	for _, m := range macros {
		item := map[string]any{
			"id":              m.ID,
			"name":            m.Name,
			"visibility":      m.Visibility,
			"message_content": m.MessageContent,
		}
		var actions any
		if err := json.Unmarshal(m.Actions, &actions); err == nil {
			item["actions"] = actions
		}
		out = append(out, item)
	}
	return out, nil
}

func compactContact(u umodels.User) map[string]any {
	return map[string]any{
		"id":                u.ID,
		"first_name":        u.FirstName,
		"last_name":         u.LastName,
		"email":             u.Email.String,
		"phone":             u.PhoneNumber.String,
		"phone_code":        u.PhoneNumberCountryCode.String,
		"country":           u.Country.String,
		"enabled":           u.Enabled,
		"external_id":       u.ExternalUserID.String,
		"custom_attributes": u.CustomAttributes,
	}
}
