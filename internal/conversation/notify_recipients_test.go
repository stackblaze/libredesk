package conversation

import (
	"testing"

	amodels "github.com/abhinavxd/libredesk/internal/automation/models"
)

func TestResolveNotifyExternalEmails(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name    string
		entries []string
		want    []string
	}{
		{
			name:    "email prefixed entries",
			entries: []string{"email:ops@example.com", "email:billing@example.com"},
			want:    []string{"ops@example.com", "billing@example.com"},
		},
		{
			name:    "bare email entries",
			entries: []string{"ops@example.com"},
			want:    []string{"ops@example.com"},
		},
		{
			name:    "internal recipients ignored",
			entries: []string{"assignee", "assigned_team", "team:5", "user:3"},
			want:    nil,
		},
		{
			name:    "mixed internal and external",
			entries: []string{"assignee", "email:ops@example.com", "team:5"},
			want:    []string{"ops@example.com"},
		},
		{
			name:    "duplicates deduplicated",
			entries: []string{"email:ops@example.com", "ops@example.com"},
			want:    []string{"ops@example.com"},
		},
		{
			name:    "known kinds that look like emails are not external",
			entries: []string{amodels.NotifyRecipientAssignee + ":12"},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.resolveNotifyExternalEmails(tt.entries)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}
