package kotoai

import (
	oaimport "github.com/sashabaranov/go-openai"
)

type (
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

	TFileDeleteResponse struct {
		ID      string `json:"id,omitempty"`
		Object  string `json:"object,omitempty"`
		Deleted bool   `json:"deleted,omitempty"`
	}

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
		Type string `json:"type"`
	}
)
