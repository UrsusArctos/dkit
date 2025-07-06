package kotoai

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
)
