package kotoai

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/UrsusArctos/dkit/pkg/dexternal"
)

const (
	// URLs and URIs
	urlAPI        = "https://api.openai.com/v1/%s"
	uriAssistants = "assistants"
	uriThreads    = "threads"
	uriMessages   = "messages"
	uriFiles      = "files"
	// HTTP headers
	hhAuthorization = "Authorization"
	hhBearer        = "Bearer"
	hhOpenAIBeta    = "OpenAI-Beta"
	hhAssistantsV2  = "assistants=v2"
)

type (
	TKotOAI struct {
		apiToken string
	}
)

func NewInstance(token string) *TKotOAI {
	return &TKotOAI{
		apiToken: token,
	}
}

func (KAI *TKotOAI) createJob(njMethod string, njURL string, njPayload any, njFile *dexternal.TPayloadFile) (njob *dexternal.TDexternalJob) {
	njob = new(dexternal.TDexternalJob)
	njob.Method = njMethod
	njob.URL = njURL
	njob.Headers = KAI.commonHeaders()
	njob.Payload = njPayload
	njob.PayloadFile = njFile
	return njob
}

func (KAI *TKotOAI) commonHeaders() dexternal.TDexJobHeaders {
	return dexternal.TDexJobHeaders{
		hhAuthorization: fmt.Sprintf("%s %s", hhBearer, KAI.apiToken),
		hhOpenAIBeta:    hhAssistantsV2,
	}
}

func (KAI *TKotOAI) formatURL(uriSUB string) string {
	return fmt.Sprintf(urlAPI, uriSUB)
}

func commonProcessor(djob *dexternal.TDexternalJob, outStruct any) error {
	if djob.RequestError == nil {
		if djob.HTTPResponse.StatusCode == http.StatusOK {
			djob.DecodeResponse(&outStruct)
			return nil
		} else {
			return errors.New(djob.HTTPResponse.Status)
		}
	} else {
		return djob.RequestError
	}
}
