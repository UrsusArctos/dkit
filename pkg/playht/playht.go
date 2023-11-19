package playht

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type (
	TPlayHT struct {
		// Credentials
		userID    string
		secretKey string
	}

	TPHTTranscriptionRequest struct {
		Voice   string   `json:"voice"`
		Content []string `json:"content"`
		Speed   int64    `json:"speed"`
		Preset  string   `json:"preset"`
	}

	TPHTTranscriptionReply struct {
		Status          string `json:"status"`
		TranscriptionID string `json:"transcriptionId"`
		ContentLength   int64  `json:"contentLength"`
		WordCount       int64  `json:"wordCount"`
	}

	TPHArticleStatus struct {
		Voice         string  `json:"voice"`
		Converted     bool    `json:"converted"`
		AudioDuration float64 `json:"audioDuration"`
		AudioURL      string  `json:"audioUrl"`
		Message       string  `json:"message"`
	}
)

const (
	apiURL         = "https://play.ht/api/v1/%s"
	apiMIMEType    = "application/json"
	keyContentType = "Content-Type"
	valMimeType    = "application/json"
	keyUID         = "X-User-ID"
	keyAuth        = "Authorization"
	// Endpoints
	eptConvert       = "convert"
	eptArticleStatus = "articleStatus"
	// Voices
	voiceIDUKR = "uk-UA-Standard-A"
	voiceIDENG = "en-US-Standard-C"
	voiceIDRUS = "ru-RU-Standard-A"
)

// Helper formatter
func formatAPIURL(endpoint string) string {
	return fmt.Sprintf(apiURL, endpoint)
}

func NewInstance(uid string, skey string) TPlayHT {
	return TPlayHT{userID: uid, secretKey: skey}
}

func (ph TPlayHT) apiCallJSON(verb string, endPoint string, request any) ([]byte, error) {
	reqJSON, merr := json.Marshal(request)
	if request == nil {
		reqJSON = []byte{}
	}
	if merr == nil {
		hreq, herr := http.NewRequest(verb, formatAPIURL(endPoint), bytes.NewBuffer(reqJSON))
		if herr == nil {
			hreq.Header.Add(keyContentType, valMimeType)
			hreq.Header.Add(keyUID, ph.userID)
			hreq.Header.Add(keyAuth, ph.secretKey)
			httpc := &http.Client{}
			resp, rerr := httpc.Do(hreq)
			if rerr == nil {
				defer resp.Body.Close()
				rawResp, ierr := io.ReadAll(resp.Body)
				if ierr == nil {
					return rawResp, nil
				}
				return nil, ierr
			}
			return nil, rerr
		}
		return nil, herr
	}
	return nil, merr
}

func (ph TPlayHT) CreateNewTranscription(language string, textFragment string) (TPHTTranscriptionReply, error) {
	TXReq := TPHTTranscriptionRequest{
		Content: []string{textFragment},
		Speed:   1,
		Preset:  "low-latency"}
	//
	switch language {
	case "ukrainian":
		TXReq.Voice = voiceIDUKR
	case "russian":
		TXReq.Voice = voiceIDRUS
	default:
		TXReq.Voice = voiceIDENG
	}
	// Perform JSONRPC call
	rawResp, err := ph.apiCallJSON("POST", eptConvert, TXReq)
	if err == nil {
		var TXReply TPHTTranscriptionReply
		uerr := json.Unmarshal(rawResp, &TXReply)
		if uerr == nil {
			return TXReply, nil
		}
		return TPHTTranscriptionReply{}, uerr
	}
	return TPHTTranscriptionReply{}, err
}

func (ph TPlayHT) RetrieveArticleInfo(txid string) (TPHArticleStatus, error) {
	// Perform JSONRPC call
	rawResp, err := ph.apiCallJSON("GET", fmt.Sprintf("%s?transcriptionId=%s", eptArticleStatus, txid), nil)
	//defer rawResp.Body.Close()
	if err == nil {
		var ArtStat TPHArticleStatus
		uerr := json.Unmarshal(rawResp, &ArtStat)
		if uerr == nil {
			return ArtStat, nil
		}
		return TPHArticleStatus{}, uerr
	}
	return TPHArticleStatus{}, err
}

func (ph TPlayHT) DownloadTranscriptMP3(url string) ([]byte, error) {
	response, herr := http.Get(url)
	if herr == nil {
		defer response.Body.Close()
		rawResp, rerr := io.ReadAll(response.Body)
		if rerr == nil {
			return rawResp, nil
		}
		return nil, rerr
	}
	return nil, herr
}
