package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	platformai "ego-server/internal/platform/ai"
)

type jsonRepairOptions struct {
	Operation          string
	JSONMaxAttempts    int
	ChatRetryOptions   platformai.RetryOptions
	RequestLogMessage  string
	ResponseLogMessage string
	FailureLogMessage  string
	ExhaustLogMessage  string
	RepairInstruction  string
	LogAttrs           []any
}

func chatAndParseJSONWithRepair[T any](
	ctx context.Context,
	logger *slog.Logger,
	client *platformai.Client,
	baseMessages []platformai.ChatMessage,
	opts jsonRepairOptions,
	parse func(string) (T, error),
) (T, error) {
	resp, _, err := chatAndParseJSONWithRepairCount(ctx, logger, client, baseMessages, opts, parse)
	return resp, err
}

func chatAndParseJSONWithRepairCount[T any](
	ctx context.Context,
	logger *slog.Logger,
	client *platformai.Client,
	baseMessages []platformai.ChatMessage,
	opts jsonRepairOptions,
	parse func(string) (T, error),
) (T, int, error) {
	var zero T
	if client == nil {
		return zero, 0, fmt.Errorf("ai client is nil")
	}
	opts = normalizeJSONRepairOptions(opts)

	var (
		lastRaw string
		lastErr error
	)
	for attempt := 1; attempt <= opts.JSONMaxAttempts; attempt++ {
		messages := buildJSONRepairMessages(baseMessages, lastRaw, lastErr, opts.RepairInstruction)
		logger.DebugContext(ctx, opts.RequestLogMessage, appendLogAttrs(opts.LogAttrs,
			"attempt", attempt,
			"max_attempts", opts.JSONMaxAttempts,
			"messages", chatMessagesForLog(messages),
		)...)

		text, err := client.ChatWithRetry(ctx, messages, opts.ChatRetryOptions)
		if err != nil {
			return zero, attempt - 1, fmt.Errorf("chat: %w", err)
		}
		logger.DebugContext(ctx, opts.ResponseLogMessage, appendLogAttrs(opts.LogAttrs,
			"attempt", attempt,
			"raw_response", text,
		)...)

		resp, err := parse(text)
		if err == nil {
			if attempt > 1 {
				logger.InfoContext(ctx, "starmap json repair retry succeeded", appendLogAttrs(opts.LogAttrs,
					"operation", opts.Operation,
					"attempt", attempt,
				)...)
			}
			return resp, attempt, nil
		}

		lastRaw = text
		lastErr = err
		logger.WarnContext(ctx, opts.FailureLogMessage, appendLogAttrs(opts.LogAttrs,
			"attempt", attempt,
			"max_attempts", opts.JSONMaxAttempts,
			"error", err,
			"raw_response", text,
		)...)
	}

	logger.WarnContext(ctx, opts.ExhaustLogMessage, appendLogAttrs(opts.LogAttrs,
		"error", lastErr,
	)...)
	return zero, opts.JSONMaxAttempts, lastErr
}

func normalizeJSONRepairOptions(opts jsonRepairOptions) jsonRepairOptions {
	if opts.Operation == "" {
		opts.Operation = "starmap_json"
	}
	if opts.JSONMaxAttempts <= 0 {
		opts.JSONMaxAttempts = 2
	}
	if opts.ChatRetryOptions.MaxAttempts <= 0 {
		opts.ChatRetryOptions.MaxAttempts = 2
	}
	if strings.TrimSpace(opts.ChatRetryOptions.Operation) == "" {
		opts.ChatRetryOptions.Operation = opts.Operation
	}
	if opts.RequestLogMessage == "" {
		opts.RequestLogMessage = "starmap json ai request"
	}
	if opts.ResponseLogMessage == "" {
		opts.ResponseLogMessage = "starmap json ai response"
	}
	if opts.FailureLogMessage == "" {
		opts.FailureLogMessage = "starmap json validation failed"
	}
	if opts.ExhaustLogMessage == "" {
		opts.ExhaustLogMessage = "starmap json validation retry exhausted"
	}
	if opts.RepairInstruction == "" {
		opts.RepairInstruction = "请基于同一批输入重新生成。"
	}
	return opts
}

func buildJSONRepairMessages(baseMessages []platformai.ChatMessage, previousRaw string, previousErr error, instruction string) []platformai.ChatMessage {
	messages := append([]platformai.ChatMessage(nil), baseMessages...)
	if previousErr == nil {
		return messages
	}
	messages = append(messages,
		platformai.ChatMessage{Role: "assistant", Content: previousRaw},
		platformai.ChatMessage{Role: "user", Content: buildJSONRepairPrompt(previousRaw, previousErr, instruction)},
	)
	return messages
}

func buildJSONRepairPrompt(previousRaw string, previousErr error, instruction string) string {
	var b strings.Builder
	b.WriteString("你上一次返回的内容没有通过 JSON 校验。\n")
	fmt.Fprintf(&b, "失败原因：%v\n\n", previousErr)
	b.WriteString("上一次原始返回：\n")
	b.WriteString(previousRaw)
	b.WriteString("\n\n")
	b.WriteString(instruction)
	b.WriteString("\n\n通用要求：\n")
	b.WriteString("- 只返回一个严格合法的 JSON 对象。\n")
	b.WriteString("- 不要包含 markdown、解释文字或多余前后缀。\n")
	b.WriteString("- 不要重复 key，不要输出未转义的换行或破损字符串。\n")
	return b.String()
}

func appendLogAttrs(base []any, attrs ...any) []any {
	out := make([]any, 0, len(base)+len(attrs))
	out = append(out, base...)
	out = append(out, attrs...)
	return out
}
