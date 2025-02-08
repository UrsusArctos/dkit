package openai

import (
	oaimport "github.com/sashabaranov/go-openai"
)

// This fixes certain structures

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

	// Assistants
	TAssistantRequest struct {
		Model          string                         `json:"model"`
		Name           string                         `json:"name,omitempty"`
		Description    string                         `json:"description,omitempty"`
		Instructions   string                         `json:"instructions,omitempty"`
		Tools          []oaimport.AssistantTool       `json:"-"`
		FileIDs        []string                       `json:"file_ids,omitempty"`
		Metadata       map[string]any                 `json:"metadata,omitempty"`
		ToolResources  oaimport.AssistantToolResource `json:"tool_resources,omitempty"`
		ResponseFormat TResponseFormat                `json:"response_format,omitempty"`
		Temperature    float32                        `json:"temperature,omitempty"`
		TopP           float32                        `json:"top_p,omitempty"`
	}

	TResponseFormat struct {
		Type string `json:"type,omitempty"`
	}

	// TAssistantsList struct {
	// 	Assistants []oaimport.Assistant `json:"data"`
	// 	LastID     string               `json:"last_id"`
	// 	FirstID    string               `json:"first_id"`
	// 	HasMore    bool                 `json:"has_more"`
	// }

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

	TFileList struct {
		Object  string          `json:"object,omitempty"`
		Files   []oaimport.File `json:"data,omitempty"`
		HasMore bool            `json:"has_more,omitempty"`
		FirstID string          `json:"first_id,omitempty"`
		LastID  string          `json:"last_id,omitempty"`
	}

	TThreadMessageRequest struct {
		Role    string         `json:"role"`
		Content []TContentPart `json:"content"`
	}

	TContentPart struct {
		Type      string             `json:"type"`
		Text      string             `json:"text,omitempty"`
		ImageFile *TContentImageFile `json:"image_file,omitempty"`
	}

	TContentImageFile struct {
		FileID string `json:"file_id,omitempty"`
		Detail string `json:"detail,omitempty"` // low, auto, high
	}
)

const (
	ChatRoleSystem    = oaimport.ChatMessageRoleSystem
	ChatRoleUser      = oaimport.ChatMessageRoleUser
	ChatRoleAssistant = oaimport.ChatMessageRoleAssistant
)
