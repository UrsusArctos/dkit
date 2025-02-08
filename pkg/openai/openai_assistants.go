package openai

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	oaimport "github.com/sashabaranov/go-openai"
)

const (
	defaultTemperature    = 0.1
	defaultResponseFormat = "json_object"
	// custom errors
	errNoModel = "model is not selected"
)

type (
	TAssistant struct {
		ParentOAC *TOpenAPIClient
		assObject *oaimport.Assistant
		assThread *oaimport.Thread
	}
)

func (ass *TAssistant) RegisterNewAssistant(sName, sDescription string, sbInstructions *strings.Builder) error {
	// Assert the model is selected
	if !ass.ParentOAC.isModelSelected() {
		return errors.New(errNoModel)
	}
	// Compose assistant creation request
	assCReq := TAssistantRequest{
		Model:          ass.ParentOAC.Models.Data[ass.ParentOAC.SelectedModel].ID,
		Name:           sName,
		Description:    sDescription,
		Instructions:   sbInstructions.String(),
		ResponseFormat: TResponseFormat{Type: defaultResponseFormat},
		Temperature:    defaultTemperature,
	}
	// Perform JSON call
	rawResp, err := ass.ParentOAC.apiCallJSON(eptAssistants, assCReq)
	if err == nil {
		err = json.Unmarshal(rawResp, ass.assObject)
	}
	//
	return err
}

func (ass TAssistant) ListAvailableAssistants() ([]oaimport.Assistant, error) {
	hreq, err := http.NewRequest(http.MethodGet, formatAPIURL(eptAssistants), nil)
	if err == nil {
		// add headers
		hreq.Header.Add(keyContentType, valMimeType)
		hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, ass.ParentOAC.APIToken))
		hreq.Header.Add(keyOpenAIBeta, valBetaAssistants)
		// perform request
		httpc := &http.Client{}
		resp, err := httpc.Do(hreq)
		if err == nil {
			defer resp.Body.Close()
			rawResp, err := io.ReadAll(resp.Body)
			if err == nil {
				var assList oaimport.AssistantsList
				err = json.Unmarshal(rawResp, &assList)
				if err == nil {
					return assList.Assistants, err
				}
			}
		}
	}
	//
	return nil, err
}

func (ass *TAssistant) LoadAssistant(assID string) error {
	hreq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", formatAPIURL(eptAssistants), assID), nil)
	if err == nil {
		// add headers
		// hreq.Header.Add(keyContentType, valMimeType)
		hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, ass.ParentOAC.APIToken))
		hreq.Header.Add(keyOpenAIBeta, valBetaAssistants)
		// perform request
		httpc := &http.Client{}
		resp, err := httpc.Do(hreq)
		if err == nil {
			defer resp.Body.Close()
			rawResp, err := io.ReadAll(resp.Body)
			if err == nil {
				err = json.Unmarshal(rawResp, &ass.assObject)
				return err
			}
		}
	}
	//
	return err
}

func (ass *TAssistant) CreateThread() error {
	// Perform JSON call
	rawResp, err := ass.ParentOAC.apiCallJSON(eptThreads, nil)
	if err == nil {
		ass.assThread = new(oaimport.Thread)
		err = json.Unmarshal(rawResp, ass.assThread)
		fmt.Printf("th: %+v\n", ass.assThread)
	}
	//
	return err
}

func (ass *TAssistant) RetrieveThread(threadID string) error {
	hreq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", formatAPIURL(eptThreads), threadID), nil)
	if err == nil {
		// add headers
		// hreq.Header.Add(keyContentType, valMimeType)
		hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, ass.ParentOAC.APIToken))
		hreq.Header.Add(keyOpenAIBeta, valBetaAssistants)
		// perform request
		httpc := &http.Client{}
		resp, err := httpc.Do(hreq)
		if err == nil {
			defer resp.Body.Close()
			rawResp, err := io.ReadAll(resp.Body)
			if err == nil {
				ass.assThread = new(oaimport.Thread)
				err = json.Unmarshal(rawResp, &ass.assThread)
				fmt.Printf("RETR THREAD: %+v\n", ass.assThread)
				return err
			}
		}
	}
	//
	return err
}

func (ass *TAssistant) DeleteThread() (bool, error) {
	hreq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", formatAPIURL(eptThreads), ass.assThread.ID), nil)
	if err == nil {
		// add headers
		hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, ass.ParentOAC.APIToken))
		hreq.Header.Add(keyOpenAIBeta, valBetaAssistants)
		// perform request
		httpc := &http.Client{}
		resp, err := httpc.Do(hreq)
		if err == nil {
			defer resp.Body.Close()
			rawResp, err := io.ReadAll(resp.Body)
			if err == nil {
				var delFile TDeletedFile
				err = json.Unmarshal(rawResp, &delFile)
				if err == nil {
					return delFile.Deleted, err
				}
			}
		}
	}
	//
	return false, err
}

func (ass *TAssistant) CreateMessageInThread(threadID string, msgText string, attImgFile string) (msgObject *oaimport.Message, err error) {
	// Compose assistant creation request
	assMReq := TThreadMessageRequest{
		Role: "user",
		Content: []TContentPart{
			{
				Type: "text",
				Text: msgText,
			},
			{
				Type: "image_file",
				ImageFile: &TContentImageFile{
					FileID: attImgFile,
					Detail: "auto",
				},
			},
		},
	}
	// Perform JSON call
	var rawResp []byte
	rawResp, err = ass.ParentOAC.apiCallJSON(fmt.Sprintf("%s/%s/%s", eptThreads, ass.assThread.ID, eptMessages), assMReq)
	if err == nil {
		fmt.Println(string(rawResp))
		msgObject = new(oaimport.Message)
		err = json.Unmarshal(rawResp, msgObject)
		if err == nil {
			return msgObject, err
		}
	}
	//
	return nil, err
}

func (ass *TAssistant) DeleteMessageInThread(messageID string) (bool, error) {
	hreq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s/%s/%s", formatAPIURL(eptThreads), ass.assThread.ID, eptMessages, messageID), nil)
	if err == nil {
		// add headers
		hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, ass.ParentOAC.APIToken))
		hreq.Header.Add(keyOpenAIBeta, valBetaAssistants)
		// perform request
		httpc := &http.Client{}
		resp, err := httpc.Do(hreq)
		if err == nil {
			defer resp.Body.Close()
			rawResp, err := io.ReadAll(resp.Body)
			if err == nil {
				fmt.Println(string(rawResp))
				var delFile TDeletedFile
				err = json.Unmarshal(rawResp, &delFile)
				if err == nil {
					return delFile.Deleted, err
				}
			}
		}
	}
	//
	return false, err
}
