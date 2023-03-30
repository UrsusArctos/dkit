package openai

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
	"strings"

	"golang.org/x/exp/slices"
)

const (
	// API entry
	apiURL         = "https://api.openai.com/v1/%s"
	apiMIMEType    = "application/json"
	keyAuth        = "Authorization"
	keyContentType = "Content-Type"
	valBearer      = "Bearer %s"
	valMimeType    = "application/json"
	// Endpoints
	eptModels              = "models"
	eptCompletions         = "completions"
	eptChatCompletions     = "chat/completions"
	eptImageGeneration     = "images/generations"
	eptAudioTranscriptions = "audio/transcriptions"
	// STT (transcription) models
	transcriptionModel = "whisper-1"
	// Project name
	projectName = "dkit-openai"
	// Chat
	defCapacity = 0x100
)

type (
	TAttachedFile struct {
		LocalFileName  string
		FieldName      string
		MIMEType       string
		RemoteFileName string
	}

	TOpenAPIClient struct {
		APIToken      string
		Models        TAIModels
		SelectedModel int
	}

	TChat struct {
		History TChatMessages
	}
)

// Helper formatter
func formatAPIURL(endpoint string) string {
	return fmt.Sprintf(apiURL, endpoint)
}

// TOpenAIClient
func NewInstance(Token string) (TOpenAPIClient, error) {
	req, err := http.NewRequest("GET", formatAPIURL(eptModels), nil)
	if err == nil {
		req.Header.Add(keyAuth, fmt.Sprintf(valBearer, Token))
		httpc := &http.Client{}
		resp, rerr := httpc.Do(req)
		if rerr == nil {
			defer resp.Body.Close()
			oac := TOpenAPIClient{APIToken: Token, SelectedModel: -1}
			jerr := json.NewDecoder(resp.Body).Decode(&oac.Models)
			if jerr == nil {
				return oac, nil
			}
			return TOpenAPIClient{}, jerr
		}
		return TOpenAPIClient{}, rerr
	}
	return TOpenAPIClient{}, err
}

func (oac *TOpenAPIClient) SelectModel(modelId string) bool {
	oac.SelectedModel = slices.IndexFunc(oac.Models.Data, func(m TModel) bool { return m.ID == modelId })
	return oac.isModelSelected()
}

func (oac TOpenAPIClient) isModelSelected() bool {
	return oac.SelectedModel != -1
}

func (oac TOpenAPIClient) apiCallJSON(endPoint string, request any) ([]byte, error) {
	reqJSON, merr := json.Marshal(request)
	if merr == nil {
		hreq, herr := http.NewRequest("POST", formatAPIURL(endPoint), bytes.NewBuffer(reqJSON))
		if herr == nil {
			hreq.Header.Add(keyContentType, valMimeType)
			hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, oac.APIToken))
			httpc := &http.Client{}
			resp, rerr := httpc.Do(hreq)
			if rerr == nil {
				defer resp.Body.Close()
				rawResp, ierr := io.ReadAll(resp.Body)
				if ierr == nil {
					return rawResp, nil
				}
				return nil, ierr
			}
			return nil, rerr
		}
		return nil, herr
	}
	return nil, merr
}

func (oac TOpenAPIClient) apiCallForm(endPoint string, request map[string]any, fileAttached *TAttachedFile) ([]byte, error) {
	// Init multipart request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	// write down the request parameters first
	for key, value := range request {
		vt := reflect.TypeOf(value)
		switch vt.Kind() {
		case reflect.String:
			wrerr := writer.WriteField(key, value.(string))
			if wrerr != nil {
				return nil, wrerr
			}
		case reflect.Int64:
			wrerr := writer.WriteField(key, fmt.Sprintf("%d", value.(int64)))
			if wrerr != nil {
				return nil, wrerr
			}
		}
	}
	// Attach a file, if any
	if fileAttached != nil {
		attfile, aerr := os.Open(fileAttached.LocalFileName)
		if aerr == nil {
			defer attfile.Close()
			attheader := make(textproto.MIMEHeader)
			attheader.Set("Content-Disposition",
				fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fileAttached.FieldName, fileAttached.RemoteFileName))
			attheader.Set("Content-Type", fileAttached.MIMEType)
			afilepart, afperr := writer.CreatePart(attheader)
			if afperr == nil {
				_, copyerr := io.Copy(afilepart, attfile)
				if copyerr != nil {
					return nil, copyerr
				}
				writer.Close()
			} else {
				return nil, afperr
			}
		} else {
			return nil, aerr
		}
	}
	// Make API request
	apireq, arqerr := http.NewRequest("POST", formatAPIURL(endPoint), body)
	if arqerr == nil {
		apireq.Header.Add(keyContentType, "multipart/form-data; boundary="+writer.Boundary())
		apireq.Header.Add(keyAuth, fmt.Sprintf(valBearer, oac.APIToken))
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

func (oac TOpenAPIClient) GetTextCompletion(prompt string, choicesWanted int) (TCompletionChoices, error) {
	// Assert the model is selected
	if !oac.isModelSelected() {
		return nil, nil
	}
	// Compose completion request
	compReq := TCompletionRequest{
		Model:  oac.Models.Data[oac.SelectedModel].ID,
		Prompt: prompt,
		N:      choicesWanted,
		User:   projectName,
	}
	// Perform JSONRPC call
	rawResp, err := oac.apiCallJSON(eptCompletions, compReq)
	if err == nil {
		var CR TCompletionResponse
		uerr := json.Unmarshal(rawResp, &CR)
		if uerr == nil {
			return CR.Choices, nil
		}
		return nil, uerr
	}
	return nil, err
}

func (oac TOpenAPIClient) GetChatCompletion(chat TChat, choicesWanted int) (TChatCompletionChoices, error) {
	// Assert the model is selected
	if !oac.isModelSelected() {
		return TChatCompletionChoices{}, nil
	}
	// Compose completion request
	compReq := TChatCompletionRequest{
		Model:    oac.Models.Data[oac.SelectedModel].ID,
		Messages: chat.History,
		N:        choicesWanted,
		User:     projectName,
	}
	// Perform JSONRPC call
	rawResp, err := oac.apiCallJSON(eptChatCompletions, compReq)
	if err == nil {
		var CR TChatCompletionResponse
		uerr := json.Unmarshal(rawResp, &CR)
		if uerr == nil {
			return CR.Choices, nil
		}
		return TChatCompletionChoices{}, uerr
	}
	return TChatCompletionChoices{}, err
}

func (oac TOpenAPIClient) GetGeneratedImage(prompt string, choicesWanted int) (TGeneratedImage, error) {
	// Compose completion request
	compReq := TImageRequest{
		Prompt: prompt,
		N:      choicesWanted,
		User:   projectName,
	}
	// Perform JSONRPC call
	rawResp, err := oac.apiCallJSON(eptImageGeneration, compReq)
	if err == nil {
		var GI TGeneratedImage
		uerr := json.Unmarshal(rawResp, &GI)
		if uerr == nil {
			return GI, nil
		}
		return TGeneratedImage{}, uerr
	}
	return TGeneratedImage{}, err
}

func (oac TOpenAPIClient) GetAudioTranscription(audioFile string) (TTranscriptResponse, error) {
	// Compose transcription request
	var transReq map[string]any = make(map[string]any)
	transReq["model"] = transcriptionModel
	transReq["response_format"] = "verbose_json"
	//
	fa := TAttachedFile{
		LocalFileName:  audioFile,
		FieldName:      "file",
		MIMEType:       mime.TypeByExtension(filepath.Ext(audioFile)),
		RemoteFileName: filepath.Base(audioFile),
	}
	//
	rawResp, err := oac.apiCallForm(eptAudioTranscriptions, transReq, &fa)
	if err == nil {
		respDecoder := json.NewDecoder(bytes.NewReader(rawResp))
		respDecoder.UseNumber()
		var TR TTranscriptResponse
		decerr := respDecoder.Decode(&TR)
		if decerr == nil {
			return TR, nil
		} else {
			return TTranscriptResponse{}, decerr
		}
	}
	return TTranscriptResponse{}, err
}

// TChat
func NewChat() TChat {
	return TChat{History: make(TChatMessages, 0, defCapacity)}
}

func (ch *TChat) Clear() {
	ch.History = ch.History[:0]
}

func (ch *TChat) recordNew(newMessage TChatMessage) {
	ch.History = append(ch.History, TChatMessage{Role: newMessage.Role, Content: strings.Replace(newMessage.Content, "\n", "", 2)})
}

func (ch *TChat) RecordMessage(role string, message string) {
	ch.recordNew(TChatMessage{Role: role, Content: message})
}

func (ch *TChat) PickAnswer(ccc TChatCompletionChoices, index ...int) TChatMessage {
	var realIndex int = 0
	if len(index) > 0 {
		realIndex = index[0]
	}
	for _, ccci := range ccc {
		if ccci.Index == realIndex {
			ch.recordNew(ccci.Message)
			return ccci.Message
		}
	}
	return TChatMessage{}
}
