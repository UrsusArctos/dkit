package kotoai

import (
	"net/http"

	"github.com/UrsusArctos/dkit/pkg/dexternal"
	oaimport "github.com/sashabaranov/go-openai"
)

type (
	TAudioResponse = oaimport.AudioResponse
)

func (kai *TKotOAI) CreateTranscript(file dexternal.TPayloadFile, modelAudio string) (aResp TAudioResponse, err error) {
	var transReq map[string]string = make(map[string]string)
	transReq["model"] = modelAudio
	transReq["response_format"] = "json"
	//
	djob := kai.createJob(http.MethodPost, kai.formatURL(uriAudioTranscriptions), transReq, &file)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &aResp)
	}
	return aResp, err
}
