package completions

import (
	"fmt"
	"strings"

	"chat2api/app/types/chat"

	"github.com/google/uuid"
)

func BuildChatRequest(apiReq *ApiReq) *chat.Request {
	messages := buildChatMessages(apiReq)
	parentMessageId := strings.TrimSpace(apiReq.ParentMessageId)
	if parentMessageId == "" {
		parentMessageId = uuid.New().String()
	}

	return &chat.Request{
		Action:                     "next",
		Messages:                   messages,
		ConversationId:             strings.TrimSpace(apiReq.ConversationId),
		ParentMessageId:            parentMessageId,
		Model:                      normalizeModel(apiReq.Model),
		Timezone:                   "Asia/Shanghai",
		TimeZoneOffsetMin:          -480,
		Suggestions:                make([]string, 0),
		SupportedEncodings:         make([]string, 0),
		SystemHints:                make([]string, 0),
		HistoryAndTrainingDisabled: true,
		ForceUseSse:                true,
		FaceUseSse:                 false,
		ForceParagen:               false,
		ForceParagenModelSlug:      "",
		ForceRateLimit:             false,
		ResetRateLimits:            false,
		VariantPurpose:             "comparison_implicit",
		ConversationMode: chat.ConversationMode{
			Kind: "primary_assistant",
		},
		WebsocketRequestId: uuid.New().String(),
		ClientContextualInfo: chat.ClientContextualInfo{
			IsDarkMode:      false,
			TimeSinceLoaded: 120,
			PageHeight:      900,
			PageWidth:       1400,
			PixelRatio:      2,
			ScreenHeight:    1440,
			ScreenWidth:     2560,
		},
	}
}

func buildChatMessages(apiReq *ApiReq) []chat.Message {
	if len(apiReq.Messages) == 0 {
		return nil
	}
	if isStatefulConversation(apiReq) {
		return []chat.Message{newChatMessage("user", latestUserContent(apiReq.Messages))}
	}
	if len(apiReq.Messages) == 1 {
		msg := apiReq.Messages[0]
		return []chat.Message{newChatMessage(msg.Role, msg.Content)}
	}
	return []chat.Message{newChatMessage("user", renderOpenAIHistory(apiReq.Messages))}
}

func isStatefulConversation(apiReq *ApiReq) bool {
	return strings.TrimSpace(apiReq.ConversationId) != "" || strings.TrimSpace(apiReq.ParentMessageId) != ""
}

func latestUserContent(messages []ApiMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			return messages[i].Content
		}
	}
	return messages[len(messages)-1].Content
}

func renderOpenAIHistory(messages []ApiMessage) string {
	var b strings.Builder
	b.WriteString("You are continuing an OpenAI-compatible chat conversation. Use the prior turns only as context. Answer only the final user message; do not repeat or answer earlier turns again.\n\n")
	b.WriteString("[conversation]\n")
	for _, msg := range messages {
		role := strings.TrimSpace(msg.Role)
		if role == "" {
			role = "user"
		}
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", role, content))
	}
	return strings.TrimSpace(b.String())
}

func newChatMessage(role string, content string) chat.Message {
	role = strings.TrimSpace(role)
	if role == "" {
		role = "user"
	}
	return chat.Message{
		Id: uuid.New().String(),
		Author: chat.Author{
			Role: role,
		},
		Content: chat.Content{
			ContentType: "text",
			Parts:       []string{content},
		},
	}
}

func normalizeModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return "auto"
	}
	return model
}
