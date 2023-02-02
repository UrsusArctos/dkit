package umintgbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"reflect"
)

const (
	apiBaseURL     = "https://api.telegram.org/bot"
	apiGetMe       = "getMe"
	apiGetUpdates  = "getUpdates"
	apiSendMessage = "sendMessage"
	apiSendAudio   = "sendAudio"
	// Misc
	apiWaitTime = 30
	apiMIMEType = "application/json"
	// Parse modes
	PMPlainText  = ""
	PMMarkdown   = "Markdown"
	PMMarkdownV2 = "MarkdownV2"
	PMHTML       = "HTML"
	// Buffered I/O
	responseBufSize = 0x100000
)

type (
	// Structure holding bot instance
	TGMinBotCore struct {
		APIToken     string
		LastUpdateID int64
		BotInfo      TBotInfo
		MSGHandler   TGMessageHandler
		MSGParseMode string
		// internals
		respoChan chan byte
	}

	// Common structure used to convey API call parameters
	JSONStruct map[string]interface{}

	// Message handler to call upon each incoming message
	TGMessageHandler func(msginfo TMessageInfo)

	// File attachment data
	AttachedFileData struct {
		LocalFile  string
		RemoteName string
		FieldName  string
		MimeType   string
		Caption    string
		Performer  string
		Title      string
	}
)

func NewInstance(BOTToken string) (tgbc TGMinBotCore) {
	resp, err := http.Post(apiBaseURL+BOTToken+"/"+apiGetMe, apiMIMEType, nil)
	if err == nil {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		err := decoder.Decode(&tgbc.BotInfo)
		if err == nil {
			if tgbc.BotInfo.Ok {
				tgbc.APIToken = BOTToken
				tgbc.LastUpdateID = 0
				tgbc.MSGParseMode = PMPlainText
				tgbc.MSGHandler = nil
				tgbc.respoChan = nil
			}
		}
	}
	return tgbc
}

func (tgbc TGMinBotCore) jsonRPC(params JSONStruct, apiMethod string) (rawresponse []byte, err error) {
	jsonparams, _ := json.Marshal(params)
	response, err := http.Post(apiBaseURL+tgbc.APIToken+"/"+apiMethod, apiMIMEType, bytes.NewBuffer(jsonparams))
	if err == nil {
		defer response.Body.Close()
		rawresponse, err := ioutil.ReadAll(response.Body)
		if err == nil {
			return rawresponse, err
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (tgbc *TGMinBotCore) jsonRPCWorkerThread(params JSONStruct, apiMethod string) {
	jsonparams, _ := json.Marshal(params)
	response, err := http.Post(apiBaseURL+tgbc.APIToken+"/"+apiMethod, apiMIMEType, bytes.NewBuffer(jsonparams))
	if err == nil {
		defer response.Body.Close()
		rawresponse, err := ioutil.ReadAll(response.Body)
		if (err == nil) && (tgbc.respoChan != nil) {
			for i := range rawresponse {
				tgbc.respoChan <- rawresponse[i]
			}
			close(tgbc.respoChan)
		}
	}
}

func (tgbc *TGMinBotCore) jsonRPCAsyncCommit(params JSONStruct, apiMethod string) bool {
	if tgbc.respoChan == nil {
		tgbc.respoChan = make(chan byte, responseBufSize)
		go tgbc.jsonRPCWorkerThread(params, apiMethod)
		return true
	}
	return false
}

func (tgbc *TGMinBotCore) jsonRPCAsyncResult() (rawresponse []byte) {
	if tgbc.respoChan != nil {
		select {
		case res, chanopen := <-tgbc.respoChan:
			rawresponse = append(rawresponse, res)
			if chanopen {
				// read the rest of it, if any
				for res = range tgbc.respoChan {
					rawresponse = append(rawresponse, res)
				}
			}
			tgbc.respoChan = nil
			return rawresponse
		default:
			// no data is available yet
		}
	}
	return nil
}

func (tgbc TGMinBotCore) formRPC(params JSONStruct, apiMethod string, attFile AttachedFileData) (rawresponse []byte, err error) {
	body := &bytes.Buffer{}
	// Set message parameters as multipart form data
	writer := multipart.NewWriter(body)
	for key, value := range params {
		vt := reflect.TypeOf(value)
		switch vt.Kind() {
		case reflect.String:
			writer.WriteField(key, value.(string))
		case reflect.Int64:
			writer.WriteField(key, fmt.Sprintf("%d", value.(int64)))
		}
	}
	// Attach a file
	afile, err := os.Open(attFile.LocalFile)
	if err == nil {
		defer afile.Close()
		aheader := make(textproto.MIMEHeader)
		aheader.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`, attFile.FieldName, attFile.RemoteName))
		aheader.Set("Content-Type", attFile.MimeType)
		afilepart, err := writer.CreatePart(aheader)
		if err == nil {
			io.Copy(afilepart, afile)
			// NB! writer.Close() is not deferred here because
			// the multipart closing boundary must be written before issuing HTTPS request
			writer.Close()
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
	req, err := http.NewRequest("POST", apiBaseURL+tgbc.APIToken+"/"+apiMethod, body)
	if err == nil {
		req.Header.Add("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
		hclient := &http.Client{}
		response, err := hclient.Do(req)
		if err == nil {
			defer response.Body.Close()
			rawresponse, err := ioutil.ReadAll(response.Body)
			if err == nil {
				return rawresponse, err
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (tgbc *TGMinBotCore) WatchForMessagesAsync() bool {
	APIReq := JSONStruct{"offset": tgbc.LastUpdateID + 1, "timeout": apiWaitTime}
	return tgbc.jsonRPCAsyncCommit(APIReq, apiGetUpdates)
}

func (tgbc *TGMinBotCore) LoadMessagesAsync() bool {
	jsonraw := tgbc.jsonRPCAsyncResult()
	if jsonraw != nil {
		decoder := json.NewDecoder(bytes.NewReader(jsonraw))
		decoder.UseNumber()
		var getMessageInfo TGetMessageInfo
		err := decoder.Decode(&getMessageInfo)
		if err == nil {
			if getMessageInfo.Ok {
				if len(getMessageInfo.Result) > 0 {
					for _, msgres := range getMessageInfo.Result {
						// Update LastUpdateID
						if msgres.UpdateID > tgbc.LastUpdateID {
							tgbc.LastUpdateID = msgres.UpdateID
						}
						// Call message handler if set
						if tgbc.MSGHandler != nil {
							tgbc.MSGHandler(msgres.Message)
						}
					}
				}
				return true
			}
		}
	}
	return false
}

func (tgbc TGMinBotCore) SendMessage_Text(msgtext string, chatid int64, replyto int64) (sentmsg TSentMessageInfo, err error) {
	APIReq := JSONStruct{"chat_id": chatid, "text": msgtext, "parse_mode": tgbc.MSGParseMode}
	if replyto != 0 {
		APIReq["reply_to_message_id"] = replyto
		APIReq["allow_sending_without_reply"] = true
	}
	rawr, err := tgbc.jsonRPC(APIReq, apiSendMessage)
	if err == nil {
		decoder := json.NewDecoder(bytes.NewReader(rawr))
		decoder.UseNumber()
		err := decoder.Decode(&sentmsg)
		if err == nil {
			if sentmsg.Ok {
				return sentmsg, nil
			}
		}
	}
	return sentmsg, err
}

func (tgbc TGMinBotCore) SendMessage_AsReply(msgtext string, quotedmsg TMessageInfo) (sentmsg TSentMessageInfo, err error) {
	return tgbc.SendMessage_Text(msgtext, quotedmsg.From.ID, quotedmsg.MessageID)
}

func (tgbc TGMinBotCore) SendMessage_Audio(audiofile AttachedFileData, chatid int64) (sentaudio TSentAudioMessageInfo, err error) {
	APIReq := JSONStruct{"chat_id": chatid,
		"caption":   audiofile.Caption,
		"performer": audiofile.Performer,
		"title":     audiofile.Title}
	audiofile.RemoteName = audiofile.Performer + " - " + audiofile.Title + ".mp3"
	audiofile.FieldName = "audio"
	audiofile.MimeType = "audio/mpeg"
	rawr, err := tgbc.formRPC(APIReq, apiSendAudio, audiofile)
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
