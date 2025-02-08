package openai

import (
	"encoding/json"
	"strings"
)

type (
	TChat struct {
		History TChatMessages
	}
)

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
