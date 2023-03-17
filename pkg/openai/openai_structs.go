package openai

import (
	oaimport "github.com/sashabaranov/go-openai"
)

type (
	// Models
	TModel = oaimport.Model

	TAIModels struct {
		Object string   `json:"object"`
		Data   []TModel `json:"data"`
	}

	// Permissions
	TPermission = oaimport.Permission

	// Completions
	TCompletionRequest  = oaimport.CompletionRequest
	TCompletionResponse = oaimport.CompletionResponse
	TCompletionChoices  = []oaimport.CompletionChoice

	// Chat
	TChatMessage            = oaimport.ChatCompletionMessage
	TChatMessages           = []TChatMessage
	TChatCompletionRequest  = oaimport.ChatCompletionRequest
	TChatCompletionResponse = oaimport.ChatCompletionResponse
	TChatCompletionChoices  = []oaimport.ChatCompletionChoice

	// Image generation
	TImageRequest   = oaimport.ImageRequest
	TGeneratedImage = oaimport.ImageResponse
)

const (
	ChatRoleSystem    = oaimport.ChatMessageRoleSystem
	ChatRoleUser      = oaimport.ChatMessageRoleUser
	ChatRoleAssistant = oaimport.ChatMessageRoleAssistant
)
