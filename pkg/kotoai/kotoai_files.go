package kotoai

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/UrsusArctos/dkit/pkg/dexternal"
	oaimport "github.com/sashabaranov/go-openai"
)

type (
	TFilesList = oaimport.FilesList
	TFile      = oaimport.File
)

func (KAI *TKotOAI) ListFiles() (fileList TFilesList, err error) {
	djob := KAI.createJob(http.MethodGet, KAI.formatURL(uriFiles), nil, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &fileList)
	}
	//
	return fileList, err
}

func (KAI *TKotOAI) DeleteFile(fileID string) bool {
	djob := KAI.createJob(http.MethodDelete, KAI.formatURL(fmt.Sprintf("%s/%s", uriFiles, strings.TrimSpace(fileID))), nil, nil)
	err := djob.Perform()
	var fdr TFileDeleteResponse
	if err == nil {
		err = commonProcessor(djob, &fdr)
	}
	//
	return (err == nil) && (fdr.Deleted)
}

func (KAI *TKotOAI) UploadFile(purpose string, file dexternal.TPayloadFile) (upfile TFile, err error) {
	mainPayload := map[string]string{
		"purpose": purpose,
	}
	djob := KAI.createJob(http.MethodPost, KAI.formatURL(uriFiles), mainPayload, &file)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &upfile)
	}
	//
	return upfile, err
}
