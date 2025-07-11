package kotoai

import (
	"fmt"
	"net/http"
	"strings"

	oaimport "github.com/sashabaranov/go-openai"
)

type (
	TThreadObject          = oaimport.Thread
	TThreadDeleteResponse  = oaimport.ThreadDeleteResponse
	TMessagesList          = oaimport.MessagesList
	TThreadMessage         = oaimport.Message
	TMessageDeleteResponse = oaimport.MessageDeletionStatus
)

// THREADS

func (KAI *TKotOAI) CreateThread() (msgThd TThreadObject, err error) {
	djob := KAI.createJob(http.MethodPost, KAI.formatURL(uriThreads), nil, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &msgThd)
	}
	//
	return msgThd, err
}

func (KAI *TKotOAI) RetrieveThread(threadID string) (msgThd TThreadObject, err error) {
	djob := KAI.createJob(http.MethodGet, KAI.formatURL(fmt.Sprintf("%s/%s", uriThreads, strings.TrimSpace(threadID))), nil, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &msgThd)
	}
	//
	return msgThd, err
}

func (KAI *TKotOAI) DeleteThread(threadID string) bool {
	djob := KAI.createJob(http.MethodDelete, KAI.formatURL(fmt.Sprintf("%s/%s", uriThreads, strings.TrimSpace(threadID))), nil, nil)
	err := djob.Perform()
	var tdr TThreadDeleteResponse
	if err == nil {
		err = commonProcessor(djob, &tdr)
	}
	return (err == nil) && (tdr.Deleted)
}

// MESSAGES

func (KAI *TKotOAI) ListMessagesInThread(threadID string) (msgList TMessagesList, err error) {
	djob := KAI.createJob(http.MethodGet, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, strings.TrimSpace(threadID), uriMessages)), nil, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &msgList)
	}
	//
	return msgList, err
}

func (KAI *TKotOAI) CreateMessageInThread(threadID string, msgText string, fileID *string) (msgThdMessage TThreadMessage, err error) {
	msgThdReq := TThreadMessageRequest{
		Role: "user",
		Content: []TContentPart{
			{
				Type: "text",
				Text: msgText,
			},
		},
	}
	if fileID != nil {
		msgThdReq.Content = append(msgThdReq.Content, TContentPart{
			Type: "image_file",
			ImageFile: &TContentImageFile{
				FileID: *fileID,
				Detail: "auto",
			},
		})
	}
	djob := KAI.createJob(http.MethodPost, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, strings.TrimSpace(threadID), uriMessages)), msgThdReq, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &msgThdMessage)
	}
	//
	return msgThdMessage, err
}

func (KAI *TKotOAI) DeleteMessage(threadID string, messageID string) bool {
	djob := KAI.createJob(http.MethodDelete, KAI.formatURL(fmt.Sprintf("%s/%s/%s/%s", uriThreads, strings.TrimSpace(threadID), uriMessages, strings.TrimSpace(messageID))), nil, nil)
	err := djob.Perform()
	var mdr TMessageDeleteResponse
	if err == nil {
		err = commonProcessor(djob, &mdr)
	}
	return (err == nil) && (mdr.Deleted)
}
