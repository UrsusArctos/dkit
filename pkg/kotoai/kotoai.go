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
		// Public fields
		APIToken string
		// Private fields
		dext *dexternal.TDexternal
	}
)

func NewInstance(token string) *TKotOAI {
	koai := &TKotOAI{
		APIToken: token,
		dext:     dexternal.NewInstance(),
	}
	return koai
}

func (KAI *TKotOAI) formatURL(uriSUB string) string {
	return fmt.Sprintf(urlAPI, uriSUB)
}

func (KAI *TKotOAI) newJob(njMethod string, njURL string, njPayload any, njFile *dexternal.TPayloadFile) (njob *dexternal.TDexJob) {
	njob = new(dexternal.TDexJob)
	njob.Method = njMethod
	njob.URL = njURL
	njob.Headers = KAI.commonHeaders()
	njob.Payload = njPayload
	njob.PayloadFile = njFile
	return njob
}

func (KAI *TKotOAI) commonHeaders() dexternal.TDexJobHeaders {
	return dexternal.TDexJobHeaders{
		hhAuthorization: fmt.Sprintf("%s %s", hhBearer, KAI.APIToken),
		hhOpenAIBeta:    hhAssistantsV2,
	}
}

func (KAI *TKotOAI) commonProcessor(jobid dexternal.TDexJobID, outStruct any) error {
	if KAI.dext.Job(jobid).APICallError == nil {
		if KAI.dext.Job(jobid).RequestError == nil {
			if KAI.dext.Job(jobid).HTTPResponse.StatusCode == http.StatusOK {
				KAI.dext.Job(jobid).DecodeResponse(&outStruct)
				return nil
			} else {
				return errors.New(KAI.dext.Job(jobid).HTTPResponse.Status)
			}
		} else {
			return KAI.dext.Job(jobid).RequestError
		}
	} else {
		return KAI.dext.Job(jobid).APICallError
	}
}
