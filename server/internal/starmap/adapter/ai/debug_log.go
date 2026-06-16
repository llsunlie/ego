package ai

import platformai "ego-server/internal/platform/ai"

func chatMessagesForLog(messages []platformai.ChatMessage) []map[string]string {
	result := make([]map[string]string, 0, len(messages))
	for _, message := range messages {
		result = append(result, map[string]string{
			"role":    message.Role,
			"content": message.Content,
		})
	}
	return result
}
