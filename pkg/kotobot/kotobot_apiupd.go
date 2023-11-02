package kotobot

import (
	tgtypes "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Message represents a message.
type (
	TMessage struct {
		MessageID                     int                                    `json:"message_id"`
		MessageThreadID               int                                    `json:"message_thread_id,omitempty"`
		From                          *tgtypes.User                          `json:"from,omitempty"`
		SenderChat                    *tgtypes.Chat                          `json:"sender_chat,omitempty"`
		Date                          int                                    `json:"date"`
		Chat                          *tgtypes.Chat                          `json:"chat"`
		ForwardFrom                   *tgtypes.User                          `json:"forward_from,omitempty"`
		ForwardFromChat               *tgtypes.Chat                          `json:"forward_from_chat,omitempty"`
		ForwardFromMessageID          int                                    `json:"forward_from_message_id,omitempty"`
		ForwardSignature              string                                 `json:"forward_signature,omitempty"`
		ForwardSenderName             string                                 `json:"forward_sender_name,omitempty"`
		ForwardDate                   int                                    `json:"forward_date,omitempty"`
		IsTopicMessage                bool                                   `json:"is_topic_message,omitempty"`
		IsAutomaticForward            bool                                   `json:"is_automatic_forward,omitempty"`
		ReplyToMessage                *TMessage                              `json:"reply_to_message,omitempty"`
		ViaBot                        *tgtypes.User                          `json:"via_bot,omitempty"`
		EditDate                      int                                    `json:"edit_date,omitempty"`
		HasProtectedContent           bool                                   `json:"has_protected_content,omitempty"`
		MediaGroupID                  string                                 `json:"media_group_id,omitempty"`
		AuthorSignature               string                                 `json:"author_signature,omitempty"`
		Text                          string                                 `json:"text,omitempty"`
		Entities                      []tgtypes.MessageEntity                `json:"entities,omitempty"`
		Animation                     *tgtypes.Animation                     `json:"animation,omitempty"`
		Audio                         *tgtypes.Audio                         `json:"audio,omitempty"`
		Document                      *tgtypes.Document                      `json:"document,omitempty"`
		Photo                         []tgtypes.PhotoSize                    `json:"photo,omitempty"`
		Sticker                       *tgtypes.Sticker                       `json:"sticker,omitempty"`
		Video                         *tgtypes.Video                         `json:"video,omitempty"`
		VideoNote                     *tgtypes.VideoNote                     `json:"video_note,omitempty"`
		Voice                         *tgtypes.Voice                         `json:"voice,omitempty"`
		Caption                       string                                 `json:"caption,omitempty"`
		CaptionEntities               []tgtypes.MessageEntity                `json:"caption_entities,omitempty"`
		Contact                       *tgtypes.Contact                       `json:"contact,omitempty"`
		Dice                          *tgtypes.Dice                          `json:"dice,omitempty"`
		Game                          *tgtypes.Game                          `json:"game,omitempty"`
		Poll                          *tgtypes.Poll                          `json:"poll,omitempty"`
		Venue                         *tgtypes.Venue                         `json:"venue,omitempty"`
		Location                      *tgtypes.Location                      `json:"location,omitempty"`
		NewChatMembers                []tgtypes.User                         `json:"new_chat_members,omitempty"`
		LeftChatMember                *tgtypes.User                          `json:"left_chat_member,omitempty"`
		NewChatTitle                  string                                 `json:"new_chat_title,omitempty"`
		NewChatPhoto                  []tgtypes.PhotoSize                    `json:"new_chat_photo,omitempty"`
		DeleteChatPhoto               bool                                   `json:"delete_chat_photo,omitempty"`
		GroupChatCreated              bool                                   `json:"group_chat_created,omitempty"`
		SuperGroupChatCreated         bool                                   `json:"supergroup_chat_created,omitempty"`
		ChannelChatCreated            bool                                   `json:"channel_chat_created,omitempty"`
		MessageAutoDeleteTimerChanged *tgtypes.MessageAutoDeleteTimerChanged `json:"message_auto_delete_timer_changed"`
		MigrateToChatID               int64                                  `json:"migrate_to_chat_id,omitempty"`
		MigrateFromChatID             int64                                  `json:"migrate_from_chat_id,omitempty"`
		PinnedMessage                 *tgtypes.Message                       `json:"pinned_message,omitempty"`
		Invoice                       *tgtypes.Invoice                       `json:"invoice,omitempty"`
		SuccessfulPayment             *tgtypes.SuccessfulPayment             `json:"successful_payment,omitempty"`
		ConnectedWebsite              string                                 `json:"connected_website,omitempty"`
		PassportData                  *tgtypes.PassportData                  `json:"passport_data,omitempty"`
		ProximityAlertTriggered       *tgtypes.ProximityAlertTriggered       `json:"proximity_alert_triggered"`
		ForumTopicCreated             *TForumTopicCreated                    `json:"forum_topic_created,omitempty"`
		VoiceChatScheduled            *tgtypes.VoiceChatScheduled            `json:"voice_chat_scheduled"`
		VoiceChatStarted              *tgtypes.VoiceChatStarted              `json:"voice_chat_started"`
		VoiceChatEnded                *tgtypes.VoiceChatEnded                `json:"voice_chat_ended"`
		VoiceChatParticipantsInvited  *tgtypes.VoiceChatParticipantsInvited  `json:"voice_chat_participants_invited"`
		ReplyMarkup                   *tgtypes.InlineKeyboardMarkup          `json:"reply_markup,omitempty"`
	}

	TForumTopicCreated struct {
		Name              string `json:"name"`
		IconColor         int    `json:"icon_color"`
		IconCustomEmijiID string `json:"icon_custom_emoji_id,omitempty"`
	}

	TUpdate struct {
		UpdateID           int                         `json:"update_id"`
		Message            *TMessage                   `json:"message,omitempty"`
		EditedMessage      *TMessage                   `json:"edited_message,omitempty"`
		ChannelPost        *TMessage                   `json:"channel_post,omitempty"`
		EditedChannelPost  *TMessage                   `json:"edited_channel_post,omitempty"`
		InlineQuery        *tgtypes.InlineQuery        `json:"inline_query,omitempty"`
		ChosenInlineResult *tgtypes.ChosenInlineResult `json:"chosen_inline_result,omitempty"`
		CallbackQuery      *tgtypes.CallbackQuery      `json:"callback_query,omitempty"`
		ShippingQuery      *tgtypes.ShippingQuery      `json:"shipping_query,omitempty"`
		PreCheckoutQuery   *tgtypes.PreCheckoutQuery   `json:"pre_checkout_query,omitempty"`
		Poll               *tgtypes.Poll               `json:"poll,omitempty"`
		PollAnswer         *tgtypes.PollAnswer         `json:"poll_answer,omitempty"`
		MyChatMember       *tgtypes.ChatMemberUpdated  `json:"my_chat_member"`
		ChatMember         *tgtypes.ChatMemberUpdated  `json:"chat_member"`
		ChatJoinRequest    *tgtypes.ChatJoinRequest    `json:"chat_join_request"`
	}
)
