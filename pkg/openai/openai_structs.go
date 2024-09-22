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

	// Embeddings
	TEmbeddingRequest  = oaimport.EmbeddingRequest
	TEmbeddingResponse = oaimport.EmbeddingResponse

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

	// Transcription
	/*
		TTranscriptionRequest struct {
			model           string
			prompt          string
			response_format string
			temperature     float32
			language        string
		}
	*/
	TTranscriptResponse struct {
		Task     string               `json:"task"`
		Language string               `json:"language"`
		Duration float64              `json:"duration"`
		Segments []TTranscriptSegment `json:"segments"`
		Text     string               `json:"text"`
	}

	TTranscriptSegment struct {
		ID               int64   `json:"id"`
		Seek             int64   `json:"seek"`
		Start            float64 `json:"start"`
		End              float64 `json:"end"`
		Text             string  `json:"text"`
		Tokens           []int64 `json:"tokens"`
		Temperature      float64 `json:"temperature"`
		AvgLogprob       float64 `json:"avg_logprob"`
		CompressionRatio float64 `json:"compression_ratio"`
		NoSpeechProb     float64 `json:"no_speech_prob"`
		Transient        bool    `json:"transient"`
	}
)

const (
	ChatRoleSystem    = oaimport.ChatMessageRoleSystem
	ChatRoleUser      = oaimport.ChatMessageRoleUser
	ChatRoleAssistant = oaimport.ChatMessageRoleAssistant
)
