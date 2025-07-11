package kotoai

import (
	"fmt"
	"net/http"
	"strings"

	oaimport "github.com/sashabaranov/go-openai"
)

type (
	TAssistantsList   = oaimport.AssistantsList
	TAssistantRequest = oaimport.AssistantRequest
	TAssistantObject  = oaimport.Assistant
	TRunRequest       = oaimport.RunRequest
	TRun              = oaimport.Run
	TRunList          = oaimport.RunList
)

// ASSISTANTS

func (KAI *TKotOAI) ListAssistants() (assList TAssistantsList, err error) {
	djob := KAI.createJob(http.MethodGet, KAI.formatURL(uriAssistants), nil, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &assList)
	}
	//
	return assList, err
}

func (KAI *TKotOAI) DeleteAssistant(assistantID string) (err error) {
	djob := KAI.createJob(http.MethodDelete, KAI.formatURL(fmt.Sprintf("%s/%s", uriAssistants, assistantID)), nil, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, nil)
	}
	//
	return err
}

func (KAI *TKotOAI) CreateAssistant(reqAss TAssistantRequest) (ass TAssistantObject, err error) {
	djob := KAI.createJob(http.MethodPost, KAI.formatURL(uriAssistants), reqAss, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &ass)
	}
	//
	return ass, err
}

// RUNS

func (KAI *TKotOAI) CreateRun(threadID string, runReq TRunRequest) (run TRun, err error) {
	djob := KAI.createJob(http.MethodPost, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, strings.TrimSpace(threadID), uriRuns)), runReq, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &run)
	}
	//
	return run, err
}

func (KAI *TKotOAI) ListRuns(threadID string) (runList TRunList, err error) {
	djob := KAI.createJob(http.MethodGet, KAI.formatURL(fmt.Sprintf("%s/%s/%s", uriThreads, strings.TrimSpace(threadID), uriRuns)), nil, nil)
	err = djob.Perform()
	if err == nil {
		err = commonProcessor(djob, &runList)
	}
	//
	return runList, err
}
