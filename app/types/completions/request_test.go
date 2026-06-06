package completions

import (
	"strings"
	"testing"
)

func TestBuildChatRequestStatefulConversationUsesLatestUserMessage(t *testing.T) {
	req := BuildChatRequest(&ApiReq{
		Model:           "auto",
		ConversationId:  "conv-1",
		ParentMessageId: "msg-1",
		Messages: []ApiMessage{
			{Role: "user", Content: "first question"},
			{Role: "assistant", Content: "first answer"},
			{Role: "user", Content: "second question"},
		},
	})

	if got := len(req.Messages); got != 1 {
		t.Fatalf("expected one upstream message, got %d", got)
	}
	if got := req.Messages[0].Author.Role; got != "user" {
		t.Fatalf("expected role user, got %q", got)
	}
	if got := req.Messages[0].Content.Parts[0]; got != "second question" {
		t.Fatalf("expected latest user content, got %q", got)
	}
}

func TestBuildChatRequestFullHistoryIsCollapsedForNewConversation(t *testing.T) {
	req := BuildChatRequest(&ApiReq{
		Model: "auto",
		Messages: []ApiMessage{
			{Role: "system", Content: "answer briefly"},
			{Role: "user", Content: "first question"},
			{Role: "assistant", Content: "first answer"},
			{Role: "user", Content: "second question"},
		},
	})

	if got := len(req.Messages); got != 1 {
		t.Fatalf("expected one collapsed upstream message, got %d", got)
	}
	if got := req.Messages[0].Author.Role; got != "user" {
		t.Fatalf("expected collapsed role user, got %q", got)
	}
	content := req.Messages[0].Content.Parts[0]
	for _, want := range []string{
		"Use the prior turns only as context",
		"system: answer briefly",
		"user: first question",
		"assistant: first answer",
		"user: second question",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected collapsed content to contain %q, got %q", want, content)
		}
	}
}

func TestBuildChatRequestSingleMessagePassesThrough(t *testing.T) {
	req := BuildChatRequest(&ApiReq{
		Model:    "auto",
		Messages: []ApiMessage{{Role: "user", Content: "ping"}},
	})

	if got := len(req.Messages); got != 1 {
		t.Fatalf("expected one upstream message, got %d", got)
	}
	if got := req.Messages[0].Author.Role; got != "user" {
		t.Fatalf("expected role user, got %q", got)
	}
	if got := req.Messages[0].Content.Parts[0]; got != "ping" {
		t.Fatalf("expected content ping, got %q", got)
	}
}
