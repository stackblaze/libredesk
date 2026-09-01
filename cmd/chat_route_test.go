package main

import "testing"

func TestNormalizeLivechatIntent(t *testing.T) {
	tests := []struct {
		raw, message, want string
	}{
		{"sales", "hi", livechatIntentSales},
		{"support", "hi", livechatIntentSupport},
		{"SALES", "", livechatIntentSales},
		{"", "I need help with my account.", livechatIntentSupport},
		{"", "I'd like to talk to sales.", livechatIntentSales},
		{"", "Can you tell me about pricing and plans?", livechatIntentSales},
		{"", "I'd like to book a demo.", livechatIntentSales},
		{"", "hi", livechatIntentSales},
		{"other", "hi", livechatIntentSales},
	}
	for _, tt := range tests {
		if got := normalizeLivechatIntent(tt.raw, tt.message); got != tt.want {
			t.Fatalf("normalizeLivechatIntent(%q, %q) = %q, want %q", tt.raw, tt.message, got, tt.want)
		}
	}
}
