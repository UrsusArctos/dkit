package kotobot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"

	"github.com/bogem/id3v2"
	tgtypes "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/maps"
)

const (
	apiURL      = "https://api.telegram.org/bot%s/%s"
	apiMIMEType = "application/json"
	apiWaitTime = 30
	// API Methods
	apiGetMe               = "getMe"
	apiGetUpdates          = "getUpdates"
	apiSendMessage         = "sendMessage"
	apiForwardMessage      = "forwardMessage"
	apiSendDocument        = "sendDocument"
	apiSendAudio           = "sendAudio"
	apiSendPhoto           = "sendPhoto"
	apiCreateForumTopic    = "createForumTopic"
	apiDeleteForumTopic    = "deleteForumTopic"
	apiAnswerCallbackQuery = "answerCallbackQuery"
	// API Constraints
	MaxMsgLength = 0x1000
	// field names
	fnDocument = "document"
	fnAudio    = "audio"
	fnPhto     = "photo"
	// internal tuning
	respoChanBacklog = 0x10
	// Parse modes
	PMPlainText  = ""
	PMMarkdown   = "Markdown"
	PMMarkdownV2 = "MarkdownV2"
	PMHTML       = "HTML"
)

type (
	// Imported structures
	TAPICallResult = tgtypes.APIResponse
	TUser          = tgtypes.User
	TChat          = tgtypes.Chat
	// TUpdate               = tgtypes.Update
	// TMessage              = tgtypes.Message
	TCallbackQuery        = tgtypes.CallbackQuery
	TInlineKeyboardMarkup = tgtypes.InlineKeyboardMarkup
	TInlineKeyboardButton = tgtypes.InlineKeyboardButton

	// Additional structures from TG Bot official API
	TForumTopic struct {
		MessageThreadID   int64  `json:"message_thread_id"`
		Name              string `json:"name"`
		IconColor         int64  `json:"icon_color"`
		IconCustomEmojiID string `json:"icon_custom_emoji_id,omitempty"`
	}

	// Internal structures
	TAttachment struct {
		LocalFileName  string
		FieldName      string
		Caption        string
		MIMEType       string
		RemoteFileName string
	}

	// Bot instance
	TKotoBot struct {
		APIToken string
		BotInfo  TUser
		// internal fields
		respoChan    chan TAsyncAPICallResult
		lastUpdateID int64
		// handlers
		ParseMode       string
		MessageHandler  TMessageHandler
		CallbackHandler TCallbackHandler
	}

	// API Call parameters
	TCallParams map[string]interface{}

	// Async API call result
	TAsyncAPICallResult struct {
		httpResponse *http.Response
		hresError    error
	}

	// Update handlers
	TMessageHandler  func(msg TMessage)
	TCallbackHandler func(cbq TCallbackQuery)
)

// Constructor
func NewInstance(Token string) (TKotoBot, error) {
	resp, posterr := http.Post(TKotoBot{APIToken: Token}.formatAPIURL(apiGetMe), apiMIMEType, nil)
	if posterr == nil {
		defer resp.Body.Close()
		var apires TAPICallResult
		decerr := json.NewDecoder(resp.Body).Decode(&apires)
		if decerr == nil {
			if apires.Ok {
				kb := TKotoBot{
					APIToken:       Token,
					respoChan:      make(chan TAsyncAPICallResult, respoChanBacklog),
					ParseMode:      PMPlainText,
					MessageHandler: nil,
				}
				unerr := json.Unmarshal(apires.Result, &kb.BotInfo)
				return kb, unerr
			}
		}
		return TKotoBot{}, decerr
	}
	return TKotoBot{}, posterr
}

// Helper formatter
func (kb TKotoBot) formatAPIURL(apimethod string) string {
	return fmt.Sprintf(apiURL, kb.APIToken, apimethod)
}

func RefmsgFromUID(uid int64) TMessage {
	return TMessage{Chat: &TChat{ID: uid}}
}

// Low level API calls (sync and async modes, JSON and FORM modes)
func (kb TKotoBot) apiSyncCallJSON(params TCallParams, apiMethod string) ([]byte, error) {
	paramsJSON, merr := json.Marshal(params)
	if merr == nil {
		apiResponse, rerr := http.Post(kb.formatAPIURL(apiMethod), apiMIMEType, bytes.NewBuffer(paramsJSON))
		if rerr == nil {
			defer apiResponse.Body.Close()
			rawResponse, ierr := io.ReadAll(apiResponse.Body)
			if ierr == nil {
				return rawResponse, nil
			}
			return nil, ierr
		}
		return nil, rerr
	}
	return nil, merr
}

func (kb TKotoBot) apiAsyncCallJSON_WorkerThread(params TCallParams, apiMethod string) {
	paramsJSON, merr := json.Marshal(params)
	if merr == nil {
		var R TAsyncAPICallResult
		R.httpResponse, R.hresError = http.Post(kb.formatAPIURL(apiMethod), apiMIMEType, bytes.NewBuffer(paramsJSON))
		kb.respoChan <- R
	}
}

func (kb TKotoBot) apiAsyncCallJSON_GetResult() (bool, []byte, error) {
	select {
	case R := <-kb.respoChan:
		if (R.httpResponse != nil) && (R.hresError == nil) {
			// valid http.Response received
			defer R.httpResponse.Body.Close()
			rawResponse, rerr := io.ReadAll(R.httpResponse.Body)
			return true, rawResponse, rerr
		} else {
			// relay Post's error to the caller
			return true, nil, R.hresError
		}
	default:
		// no response yet
		return false, nil, nil
	}
}

func (kb TKotoBot) apiSyncCallForm(params TCallParams, apiMethod string, attachment *TAttachment) ([]byte, error) {
	// Set message parameters as multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	for key, value := range params {
		vt := reflect.TypeOf(value)
		switch vt.Kind() {
		case reflect.String:
			writer.WriteField(key, value.(string))
		case reflect.Int64:
			writer.WriteField(key, fmt.Sprintf("%d", value.(int64)))
		}
	}
	// Attach a file, if any
	if attachment != nil {
		attfile, aerr := os.Open(attachment.LocalFileName)
		if aerr == nil {
			defer attfile.Close()
			attheader := make(textproto.MIMEHeader)
			attheader.Set("Content-Disposition",
				fmt.Sprintf(`form-data; name="%s"; filename="%s"`, attachment.FieldName, attachment.RemoteFileName))
			attheader.Set("Content-Type", attachment.MIMEType)
			afilepart, afperr := writer.CreatePart(attheader)
			if afperr == nil {
				io.Copy(afilepart, attfile)
				// NB! writer.Close() is not deferred here because
				// the multipart closing boundary must be written before issuing HTTPS request
				writer.Close()
			} else {
				return nil, afperr
			}
		} else {
			return nil, aerr
		}
	}
	// Make API request
	apireq, arqerr := http.NewRequest("POST", kb.formatAPIURL(apiMethod), body)
	if arqerr == nil {
		apireq.Header.Add("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
		hclient := &http.Client{}
		response, herr := hclient.Do(apireq)
		if herr == nil {
			defer response.Body.Close()
			rawResponse, rerr := io.ReadAll(response.Body)
			if rerr == nil {
				return rawResponse, rerr
			}
			return nil, rerr
		}
		return nil, herr
	}
	return nil, arqerr
}

func (kb TKotoBot) Updates_StartWatch() {
	params := TCallParams{"offset": kb.lastUpdateID + 1, "timeout": apiWaitTime}
	go kb.apiAsyncCallJSON_WorkerThread(params, apiGetUpdates)
}

func (kb *TKotoBot) Updates_ProcessAll() bool { // True means some updates were processed, false means none
	up, jsonraw, uerr := kb.apiAsyncCallJSON_GetResult()
	if up && (uerr == nil) {
		updi, interr := kb.interpretAPICallResult(jsonraw, TUpdate{})
		if (updi != nil) && (interr == nil) {
			update := updi.([]TUpdate)
			if len(update) > 0 {
				for upidx := range update {
					// Update lastUpdateID
					if int64(update[upidx].UpdateID) > kb.lastUpdateID {
						kb.lastUpdateID = int64(update[upidx].UpdateID)
					}
					// Dispatch update
					kb.dispatchUpdate(update[upidx])
				}
			}
		}
	} // TODO: Call error handler on uerr != nil
	return up
}

// Internal update dispatcher (since there are many possible types of update)
func (kb TKotoBot) dispatchUpdate(update TUpdate) {
	// Handle regular messages
	if (update.Message != nil) && (kb.MessageHandler != nil) {
		kb.MessageHandler(*update.Message)
	}
	// Handle callback queries
	if (update.CallbackQuery != nil) && (kb.CallbackHandler != nil) {
		kb.CallbackHandler(*update.CallbackQuery)
	}
}

func (kb TKotoBot) interpretAPICallResult(rawRes []byte, typeHint interface{}) (interface{}, error) {
	var apires TAPICallResult
	gerr := json.Unmarshal(rawRes, &apires)
	if gerr == nil {
		if apires.Ok {
			// Prepare decoder
			resdecoder := json.NewDecoder(bytes.NewReader(apires.Result))
			resdecoder.UseNumber()
			// Use type hint to form result
			switch reflect.TypeOf(typeHint) {
			case reflect.TypeOf(TMessage{}):
				var msg TMessage
				resdecoder.Decode(&msg)
				return msg, nil
			case reflect.TypeOf(TUpdate{}):
				var upds []TUpdate
				resdecoder.Decode(&upds)
				return upds, nil
			case reflect.TypeOf(TForumTopic{}):
				var fmtp TForumTopic
				resdecoder.Decode(&fmtp)
				return fmtp, nil
			default:
				return nil, fmt.Errorf("unknown type hint; error %d %s", apires.ErrorCode, apires.Description)
			}
		}
		return nil, fmt.Errorf("call to API resulted in error %d %s", apires.ErrorCode, apires.Description)
	}
	return nil, fmt.Errorf("invalid input JSON; error %d %s", apires.ErrorCode, apires.Description)
}

// Sending
func (kb TKotoBot) SendMessage(mText string, mUseReply bool, mRefMsg TMessage, inlineKB *TInlineKeyboardMarkup, threadID *int64) (TMessage, error) {
	// Compose API request
	// 1. Common fields
	params := TCallParams{"text": mText, "parse_mode": kb.ParseMode}
	// 2. Seems prudent to reply in private to private messages, and in groups to group messages
	params["chat_id"] = mRefMsg.Chat.ID
	// 3. Optional reply
	if mUseReply {
		params["reply_to_message_id"] = mRefMsg.MessageID
		params["allow_sending_without_reply"] = true
	}
	// 4. Optional inline keyboard
	if inlineKB != nil {
		params["reply_markup"] = *inlineKB
	}
	// 5. Optional thread ID
	if threadID != nil {
		params["message_thread_id"] = *threadID
	}
	// X. Send JSONRPC API request
	rawRes, apierr := kb.apiSyncCallJSON(params, apiSendMessage)
	if apierr == nil {
		imsg, merr := kb.interpretAPICallResult(rawRes, TMessage{})
		if imsg != nil {
			return imsg.(TMessage), merr
		}
		return TMessage{}, merr
	}
	return TMessage{}, apierr
}

// Forwarding
func (kb TKotoBot) ForwardMessage(mToWhere int64, mRefMsg TMessage, threadID *int64) (TMessage, error) {
	// Compose API request
	// 1. Common fields
	params := TCallParams{"from_chat_id": mRefMsg.Chat.ID, "message_id": mRefMsg.MessageID, "chat_id": mToWhere}
	// 2. Optional thread ID
	if threadID != nil {
		params["message_thread_id"] = *threadID
	}
	// X. Send JSONRPC API request
	rawRes, apierr := kb.apiSyncCallJSON(params, apiForwardMessage)
	if apierr == nil {
		imsg, merr := kb.interpretAPICallResult(rawRes, TMessage{})
		if imsg != nil {
			return imsg.(TMessage), merr
		}
		return TMessage{}, merr
	}
	return TMessage{}, apierr
}

// Sending files
func (kb TKotoBot) sendFileCommon(mToWhere int64, attFile TAttachment, addParams ...TCallParams) (TMessage, error) {
	// Fill common fields
	params := TCallParams{"chat_id": mToWhere, "caption": attFile.Caption}
	// Add document-specific fields
	if len(addParams) > 0 {
		for ap := range addParams {
			maps.Copy(params, addParams[ap])
		}
	}
	// X. Send X-WWWForm request
	apiMethod := map[string]string{fnDocument: apiSendDocument, fnAudio: apiSendAudio, fnPhto: apiSendPhoto}
	rawRes, apierr := kb.apiSyncCallForm(params, apiMethod[attFile.FieldName], &attFile)
	if apierr == nil {
		imsg, merr := kb.interpretAPICallResult(rawRes, TMessage{})
		if imsg != nil {
			return imsg.(TMessage), merr
		}
		return TMessage{}, merr
	}
	return TMessage{}, apierr
}

func (kb TKotoBot) SendDocument(mToWhere int64, fileName string, caption string) (TMessage, error) {
	_, fierr := os.Stat(fileName)
	if fierr == nil {
		fAtt := TAttachment{
			LocalFileName:  fileName,
			FieldName:      fnDocument,
			Caption:        caption,
			MIMEType:       mime.TypeByExtension(filepath.Ext(fileName)),
			RemoteFileName: filepath.Base(fileName),
		}
		return kb.sendFileCommon(mToWhere, fAtt)
	}
	return TMessage{}, fierr
}

func (kb TKotoBot) SendDocument_TGCloud(mToWhere int64, fileID string, caption string) (TMessage, error) {
	// Compose params
	params := TCallParams{"chat_id": mToWhere, fnDocument: fileID, "caption": caption}
	// X. Send JSONRPC API request
	rawRes, apierr := kb.apiSyncCallJSON(params, apiSendDocument)
	if apierr == nil {
		imsg, merr := kb.interpretAPICallResult(rawRes, TMessage{})
		if imsg != nil {
			return imsg.(TMessage), merr
		}
		return TMessage{}, merr
	}
	return TMessage{}, apierr
}

func (kb TKotoBot) SendPhoto(mToWhere int64, fileName string, caption string) (TMessage, error) {
	_, fierr := os.Stat(fileName)
	if fierr == nil {
		fAtt := TAttachment{
			LocalFileName:  fileName,
			FieldName:      fnPhto,
			Caption:        caption,
			MIMEType:       mime.TypeByExtension(filepath.Ext(fileName)),
			RemoteFileName: filepath.Base(fileName),
		}
		return kb.sendFileCommon(mToWhere, fAtt)
	}
	return TMessage{}, fierr
}

func (kb TKotoBot) SendAudio(mToWhere int64, fileName string, caption string) (TMessage, error) {
	_, fierr := os.Stat(fileName)
	if fierr == nil {
		// Bake audio-specific parameters
		mp3tags, _ := id3v2.Open(fileName, id3v2.Options{Parse: true})
		defer mp3tags.Close()
		tag_performer := mp3tags.Artist()
		if tag_performer == "" {
			tag_performer = "default"
		}
		tag_title := mp3tags.Title()
		if tag_title == "" {
			tag_title = "default"
		}
		audioparams := TCallParams{"performer": tag_performer, "title": tag_title}
		// Bake file attachment
		fAtt := TAttachment{
			LocalFileName:  fileName,
			FieldName:      fnAudio,
			Caption:        caption,
			MIMEType:       mime.TypeByExtension(filepath.Ext(fileName)),
			RemoteFileName: filepath.Base(fileName),
		}
		return kb.sendFileCommon(mToWhere, fAtt, audioparams)
	}
	return TMessage{}, fierr
}

func (kb TKotoBot) CreateForumTopic(mWhere int64, topicName string) (TForumTopic, error) {
	// Compose API request
	params := TCallParams{"chat_id": mWhere, "name": topicName}
	// Send JSONRPC API request
	rawRes, apierr := kb.apiSyncCallJSON(params, apiCreateForumTopic)
	if apierr == nil {
		iftop, merr := kb.interpretAPICallResult(rawRes, TForumTopic{})
		if iftop != nil {
			return iftop.(TForumTopic), merr
		}
		return TForumTopic{}, merr
	}
	return TForumTopic{}, apierr
}

func (kb TKotoBot) DeleteForumTopic(mWhere int64, MessageThreadID int64) (bool, error) {
	params := TCallParams{"chat_id": mWhere, "message_thread_id": MessageThreadID}
	rawRes, apierr := kb.apiSyncCallJSON(params, apiDeleteForumTopic)	
	if apierr == nil {
		var apires TAPIBoolResponse
		apierr = json.Unmarshal(rawRes, &apires)
		if apierr == nil {
			return apires.Result, apierr
		}
	}
	return false, apierr
}

type TAPIBoolResponse struct {
	Ok     bool `json:"ok"`
	Result bool `json:"result,omitempty"`
}

func (kb TKotoBot) AnswerCallbackQuery(cbq TCallbackQuery) (bool, error) {
	params := TCallParams{"callback_query_id": cbq.ID}
	rawRes, apierr := kb.apiSyncCallJSON(params, apiAnswerCallbackQuery)
	if apierr == nil {
		var apires TAPIBoolResponse
		apierr = json.Unmarshal(rawRes, &apires)
		if apierr == nil {
			return apires.Result, apierr
		}
	}
	return false, apierr
}
