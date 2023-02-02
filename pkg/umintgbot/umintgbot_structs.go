package umintgbot

type (
	// Structs returned by apiGetMe call
	TBotInfo struct {
		Ok     bool           `json:"ok"`
		Result TBotInfoResult `json:"result"`
	}

	TBotInfoResult struct {
		ID                      int64  `json:"id"`
		IsBot                   bool   `json:"is_bot"`
		FirstName               string `json:"first_name"`
		Username                string `json:"username"`
		CanJoinGroups           bool   `json:"can_join_groups"`
		CanReadAllGroupMessages bool   `json:"can_read_all_group_messages"`
		SupportsInlineQueries   bool   `json:"supports_inline_queries"`
	}

	// Messaging structures
	// Headers
	TGetMessageInfo struct {
		Ok     bool             `json:"ok"`
		Result []TMessageResult `json:"result"`
	}

	TSentMessageInfo struct {
		Ok     bool           `json:"ok"`
		Result TMessageResult `json:"result"`
	}

	TSentAudioMessageInfo struct {
		Ok     bool              `json:"ok"`
		Result TAudioMessageInfo `json:"result"`
	}

	// Subheaders
	TMessageResult struct {
		UpdateID int64        `json:"update_id"`
		Message  TMessageInfo `json:"message"`
	}

	TMessageInfo struct {
		MessageID int64         `json:"message_id"`
		From      TEndpointInfo `json:"from"`
		Chat      TEndpointInfo `json:"chat"`
		Date      int64         `json:"date"`
		Text      string        `json:"text"`
	}

	TAudioMessageInfo struct {
		MessageID       int64            `json:"message_id"`
		From            TEndpointInfo    `json:"from"`
		Chat            TEndpointInfo    `json:"chat"`
		Date            int64            `json:"date"`
		Audio           TAudio           `json:"audio"`
		Caption         string           `json:"caption"`
		CaptionEntities []TCaptionEntity `json:"caption_entities"`
	}

	// Generalized endpoint structure
	TEndpointInfo struct {
		ID           int64  `json:"id"`
		FirstName    string `json:"first_name"`
		Username     string `json:"username"`
		Type         string `json:"type"`
		LanguageCode string `json:"language_code,omitempty"`
		IsBot        bool   `json:"is_bot,omitempty"`
		IsPremium    bool   `json:"is_premium,omitempty"`
	}

	// File transfer related structures
	TAudio struct {
		Duration     int64  `json:"duration"`
		FileName     string `json:"file_name"`
		MIMEType     string `json:"mime_type"`
		Title        string `json:"title"`
		Performer    string `json:"performer"`
		FileID       string `json:"file_id"`
		FileUniqueID string `json:"file_unique_id"`
		FileSize     int64  `json:"file_size"`
	}

	TCaptionEntity struct {
		Offset int64  `json:"offset"`
		Length int64  `json:"length"`
		Type   string `json:"type"`
	}
)
