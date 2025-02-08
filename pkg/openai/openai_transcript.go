package openai

import (
	"bytes"
	"encoding/json"
	"mime"
	"path/filepath"
)

func (oac TOpenAPIClient) GetAudioTranscription(audioFile string) (TTranscriptResponse, error) {
	// Compose transcription request
	var transReq map[string]any = make(map[string]any)
	transReq["model"] = transcriptionModel
	transReq["response_format"] = "verbose_json"
	//
	fa := TAttachedFile{
		LocalFileName:  audioFile,
		FieldName:      "file",
		MIMEType:       mime.TypeByExtension(filepath.Ext(audioFile)),
		RemoteFileName: filepath.Base(audioFile),
	}
	//
	rawResp, err := oac.apiCallForm(eptAudioTranscriptions, transReq, &fa)
	if err == nil {
		respDecoder := json.NewDecoder(bytes.NewReader(rawResp))
		respDecoder.UseNumber()
		var TR TTranscriptResponse
		decerr := respDecoder.Decode(&TR)
		if decerr == nil {
			return TR, nil
		} else {
			return TTranscriptResponse{}, decerr
		}
	}
	return TTranscriptResponse{}, err
}
