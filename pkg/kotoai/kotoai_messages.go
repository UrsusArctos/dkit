package kotoai

import (
	"fmt"
	"net/http"
	"strings"

	oaimport "github.com/sashabaranov/go-openai"
)

func (KAI *TKotOAI) CreateThread() (msgThd oaimport.Thread, err error) {
	dj := KAI.newJob(http.MethodPost, KAI.formatURL(uriThreads), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &msgThd)
	return msgThd, err
}

func (KAI *TKotOAI) RetrieveThread(threadID string) (msgThd oaimport.Thread, err error) {
	dj := KAI.newJob(http.MethodGet, KAI.formatURL(fmt.Sprintf("%s/%s", uriThreads, strings.TrimSpace(threadID))), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &msgThd)
	return msgThd, err
}

func (KAI *TKotOAI) DeleteThread(threadID string) bool {
	dj := KAI.newJob(http.MethodDelete, KAI.formatURL(fmt.Sprintf("%s/%s", uriThreads, strings.TrimSpace(threadID))), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	var tdr oaimport.ThreadDeleteResponse
	err := KAI.commonProcessor(jobid, &tdr)
	return (err == nil) && (tdr.Deleted)
}

func (KAI *TKotOAI) ListMessagesInThread(threadID string) (msgList oaimport.MessagesList, err error) {
	dj := KAI.newJob(http.MethodGet, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, strings.TrimSpace(threadID), uriMessages)), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &msgList)
	return msgList, err
}

func (KAI *TKotOAI) CreateMessageInThread(threadID string, msgText string, fileID *string) (msgThdMessage oaimport.Message, err error) {
	assMReq := TThreadMessageRequest{
		Role: "user",
		Content: []TContentPart{
			{
				Type: "text",
				Text: msgText,
			},
		},
	}
	if fileID != nil {
		assMReq.Content = append(assMReq.Content, TContentPart{
			Type: "image_file",
			ImageFile: &TContentImageFile{
				FileID: *fileID,
				Detail: "auto",
			},
		})
	}
	dj := KAI.newJob(http.MethodPost, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, strings.TrimSpace(threadID), uriMessages)), assMReq, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &msgThdMessage)
	return msgThdMessage, err
}
