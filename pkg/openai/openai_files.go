package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	oaimport "github.com/sashabaranov/go-openai"
)

const (
	fieldNameFile    = "file"
	fieldNamePurpose = "purpose"
)

type (
	TDeletedFile struct {
		ID      string `json:"id,omitempty"`
		Object  string `json:"object,omitempty"`
		Deleted bool   `json:"deleted,omitempty"`
	}
)

func (oac TOpenAPIClient) UploadFile(localFileName string, mimeType string, purpose string) (*oaimport.File, error) {
	ATT := &TAttachedFile{
		LocalFileName:  localFileName,
		FieldName:      fieldNameFile,
		MIMEType:       mimeType,
		RemoteFileName: localFileName,
	}
	req := make(map[string]any)
	req[fieldNamePurpose] = purpose
	rawResp, err := oac.apiCallForm(eptFiles, req, ATT)
	if err == nil {
		var f oaimport.File
		err = json.Unmarshal(rawResp, &f)
		if err == nil {
			return &f, err
		}
	}
	//
	return nil, err
}

func (oac TOpenAPIClient) ListUploadedFiles() (*TFileList, error) {
	hreq, err := http.NewRequest(http.MethodGet, formatAPIURL(eptFiles), nil)
	if err == nil {
		// add headers
		hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, oac.APIToken))
		// perform request
		httpc := &http.Client{}
		resp, err := httpc.Do(hreq)
		if err == nil {
			defer resp.Body.Close()
			rawResp, err := io.ReadAll(resp.Body)
			if err == nil {
				var fileList TFileList
				err = json.Unmarshal(rawResp, &fileList)
				if err == nil {
					return &fileList, err
				}
			}
		}
	}
	//
	return nil, err
}

func (oac TOpenAPIClient) DeleteUploadedFile(fileID string) (bool, error) {
	hreq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", formatAPIURL(eptFiles), fileID), nil)
	if err == nil {
		// add headers
		hreq.Header.Add(keyAuth, fmt.Sprintf(valBearer, oac.APIToken))
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
