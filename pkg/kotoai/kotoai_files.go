package kotoai

import (
	"fmt"
	"net/http"
	"strings"

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
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	KAI.commonProcessor(jobid, &upfile)
	return upfile, err
}

func (KAI *TKotOAI) FileList() (fileList oaimport.FilesList, err error) {
	dj := KAI.newJob(http.MethodGet, KAI.formatURL(uriFiles), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	KAI.commonProcessor(jobid, &fileList)
	return fileList, err
}

func (KAI *TKotOAI) FileDelete(fileID string) bool {
	dj := KAI.newJob(http.MethodDelete, KAI.formatURL(fmt.Sprintf("%s/%s", uriFiles, strings.TrimSpace(fileID))), nil, nil)
	// Run job
	jobid := KAI.dext.RegisterJob(dj)
	defer KAI.dext.ClearJob(jobid)
	KAI.dext.StartJob(jobid)
	// Wait for all jobs to be done
	KAI.dext.WaitSyncForJobs()
	// Process results
	var fdr TFileDeleteResponse
	err := KAI.commonProcessor(jobid, &fdr)
	return (err == nil) && (fdr.Deleted)
}
