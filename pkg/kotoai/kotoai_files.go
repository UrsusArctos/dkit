package kotoai

import (
	"net/http"

	"github.com/UrsusArctos/dkit/pkg/dexternal"
	oaimport "github.com/sashabaranov/go-openai"
)

func (KAI *TKotOAI) FileUpload(purpose string, file dexternal.TPayloadFile) (upfile oaimport.File, err error) {
	mainPayload := map[string]string{
		"purpose": purpose,
	}
	dj := KAI.newJob(http.MethodPost, KAI.formatURL(uriFiles), mainPayload, &file)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	KAI.commonProcessor(jobid, &upfile)
	return upfile, err
}
