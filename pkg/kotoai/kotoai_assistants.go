package kotoai

import (
	"fmt"
	"net/http"

	oaimport "github.com/sashabaranov/go-openai"
)

func (KAI *TKotOAI) ListAssistants() (assList oaimport.AssistantsList, err error) {
	dj := KAI.newJob(http.MethodGet, KAI.formatURL(uriAssistants), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &assList)
	return assList, err
}

func (KAI *TKotOAI) ListMessagesInThread(threadID string) (msgList oaimport.MessagesList, err error) {
	dj := KAI.newJob(http.MethodGet, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, threadID, uriMessages)), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
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
	dj := KAI.newJob(http.MethodPost, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, threadID, uriMessages)), assMReq, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &msgThdMessage)
	return msgThdMessage, err
}

/*
func (KAI *TKotOAI) PerformDummyFunction() {
	dj := KAI.newJob(http.MethodPost, KAI.formatURL(fmt.Sprintf("%s/%s", uriAssistants, "noop")), nil)
	dj.Payload = map[string]string{
		"purpose": "vision",
		"test":    "dummy",
	}
	dj.PayloadFile = &dexternal.TPayloadFile{
		FileNameLocal:  "assets/dummyfile.txt",
		FileNameRemote: "dummyfile_remote.txt",
		MIMEType:       "text/plain",
		FieldName:      "file",
	}
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	fmt.Println(KAI.dext.Job(jobid).GetRawResponseJSON())
	// err = KAI.commonProcessor(jobid, &msgList)
	// return msgList, err
}
*/
