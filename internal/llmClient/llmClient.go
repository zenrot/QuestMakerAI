package llmClient

import (
	_ "embed"
)

const (
	Temperature  = 0.3
	MaxTokens    = "1500"
	SystemPrompt = "Ты умный помощник для создания квестов. Создай ровно %d вопросов."
)
