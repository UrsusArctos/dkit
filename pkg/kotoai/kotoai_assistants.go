package kotoai

import (
	"fmt"
	"net/http"
	"strings"

	oaimport "github.com/sashabaranov/go-openai"
)

func (KAI *TKotOAI) ListAssistants() (assList oaimport.AssistantsList, err error) {
	dj := KAI.newJob(http.MethodGet, KAI.formatURL(uriAssistants), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &assList)
	return assList, err
}

func (KAI *TKotOAI) DeleteAssistant(assID string) bool {
	dj := KAI.newJob(http.MethodDelete, KAI.formatURL(fmt.Sprintf("%s/%s", uriAssistants, strings.TrimSpace(assID))), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	var adr oaimport.AssistantDeleteResponse
	err := KAI.commonProcessor(jobid, &adr)
	return (err == nil) && (adr.Deleted)
}

func (KAI *TKotOAI) CreateAssistant(assReq TAssistantRequest) (ass oaimport.Assistant, err error) {
	dj := KAI.newJob(http.MethodPost, KAI.formatURL(uriAssistants), assReq, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	err = KAI.commonProcessor(jobid, &ass)
	return ass, err
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
