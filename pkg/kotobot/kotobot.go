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
	apiGetMe          = "getMe"
	apiGetUpdates     = "getUpdates"
	apiSendMessage    = "sendMessage"
	apiForwardMessage = "forwardMessage"
	apiSendDocument   = "sendDocument"
	apiSendAudio      = "sendAudio"
	apiSendPhoto      = "sendPhoto"
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
	// Imported API
	TAPICallResult = tgtypes.APIResponse
	TUser          = tgtypes.User
	TUpdate        = tgtypes.Update
	TMessage       = tgtypes.Message

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
		ParseMode      string
		MessageHandler TMessageHandler
	}

	// API Call parameters
	TCallParams map[string]interface{}

	// Async API call result
	TAsyncAPICallResult struct {
		httpResponse *http.Response
		hresError    error
	}

	// Update handlers
	TMessageHandler func(msg TMessage)
)

// Constructor
func NewInstance(Token string) (kb TKotoBot, ierr error) {
	resp, err := http.Post(TKotoBot{APIToken: Token}.formatAPIURL(apiGetMe), apiMIMEType, nil)
	if err == nil {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		var apires TAPICallResult
		err = decoder.Decode(&apires)
		if err == nil {
			if apires.Ok {
				kb.APIToken = Token
				kb.respoChan = make(chan TAsyncAPICallResult, respoChanBacklog)
				kb.ParseMode = PMPlainText
				err = json.Unmarshal(apires.Result, &kb.BotInfo)
			}
		}
	}
	return kb, err
}

// Helper formatter
func (kb TKotoBot) formatAPIURL(apimethod string) string {
	return fmt.Sprintf(apiURL, kb.APIToken, apimethod)
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
			// relay http.Post's error to the caller
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
		if interr == nil {
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
	if (update.Message != nil) && (kb.MessageHandler != nil) {
		kb.MessageHandler(*update.Message)
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
			default:
				return nil, fmt.Errorf("unknown type hint")
			}
		}
		return nil, fmt.Errorf("call to API resulted in error")
	}
	return nil, fmt.Errorf("invalid input JSON")
}

// Sending
func (kb TKotoBot) SendMessage(mText string, mUseReply bool, mRefMsg TMessage) (TMessage, error) {
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
	// X. Send JSONRPC API request
	rawRes, apierr := kb.apiSyncCallJSON(params, apiSendMessage)
	if apierr == nil {
		imsg, merr := kb.interpretAPICallResult(rawRes, TMessage{})
		return imsg.(TMessage), merr
	}
	return TMessage{}, apierr
}

// Forwarding
func (kb TKotoBot) ForwardMessage(mToWhere int64, mRefMsg TMessage) (TMessage, error) {
	// Compose API request
	// 1. Common fields
	params := TCallParams{"from_chat_id": mRefMsg.Chat.ID, "message_id": mRefMsg.MessageID, "chat_id": mToWhere}
	// X. Send JSONRPC API request
	rawRes, apierr := kb.apiSyncCallJSON(params, apiForwardMessage)
	if apierr == nil {
		imsg, merr := kb.interpretAPICallResult(rawRes, TMessage{})
		return imsg.(TMessage), merr
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
		return imsg.(TMessage), merr
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

/*
func (tgbc TGMinBotCore) Send_File_Preloaded(file_id string, mReference TMessageInfo) (sentaudio TSentAudioMessageInfo, err error) {
	// Compose API Request
	APIReq := JSONStruct{"chat_id": mReference.Chat.ID, "document": file_id}
	// Send JSONRPC API request
	rawr, err := tgbc.jsonRPC(APIReq, apiSendDocument)
	// Decode result
	if err == nil {
		decoder := json.NewDecoder(bytes.NewReader(rawr))
		decoder.UseNumber()
		err := decoder.Decode(&sentaudio)
		if err == nil {
			if sentaudio.Ok {
				return sentaudio, nil
			}
		}
	}
	return sentaudio, err
}
*/
