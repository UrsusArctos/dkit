package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"reflect"

	"golang.org/x/exp/slices"

	oaimport "github.com/sashabaranov/go-openai"
)

const (
	// API entry
	apiURL            = "https://api.openai.com/v1/%s"
	apiMIMEType       = "application/json"
	keyAuth           = "Authorization"
	keyContentType    = "Content-Type"
	keyOpenAIBeta     = "OpenAI-Beta"
	valBearer         = "Bearer %s"
	valMimeType       = "application/json"
	valBetaAssistants = "assistants=v2"
	// Endpoints
	eptModels              = "models"
	eptCompletions         = "completions"
	eptEmbeddings          = "embeddings"
	eptChatCompletions     = "chat/completions"
	eptImageGeneration     = "images/generations"
	eptAudioTranscriptions = "audio/transcriptions"
	eptAssistants          = "assistants"
	eptFiles               = "files"
	eptThreads             = "threads"
	eptMessages            = "messages"
	// Models
	DefaultModel       = "gpt-4o"
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
			hreq.Header.Add(keyOpenAIBeta, valBetaAssistants)
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
	// write down the request parameters first, if any
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

func (oac *TOpenAPIClient) SelectModel(modelId string) bool {
	oac.SelectedModel = slices.IndexFunc(oac.Models.Data, func(m TModel) bool { return m.ID == modelId })
	return oac.isModelSelected()
}

func (oac TOpenAPIClient) CreateAssistant() TAssistant {
	return TAssistant{ParentOAC: &oac, assObject: new(oaimport.Assistant)}
}
