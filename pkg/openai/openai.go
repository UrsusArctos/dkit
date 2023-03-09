package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	eptModels          = "models"
	eptCompletions     = "completions"
	eptChatCompletions = "chat/completions"
	// Project name
	projectName = "dkit-openai"
	// Chat
	defCapacity = 0x100
)

type (
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

func (ch *TChat) SetupAssistant(traitPrompt string) {
	ch.recordNew(TChatMessage{Role: chatRoleSystem, Content: traitPrompt})
}

func (ch *TChat) Say(prompt string) {
	ch.recordNew(TChatMessage{Role: chatRoleUser, Content: prompt})
}

func (ch *TChat) PickAnswer(ccc TChatCompletionChoices, index ...int) {
	var realIndex int = 0
	if len(index) > 0 {
		realIndex = index[0]
	}
	for _, ccci := range ccc {
		if ccci.Index == realIndex {
			ch.recordNew(ccci.Message)
		}
	}
}
