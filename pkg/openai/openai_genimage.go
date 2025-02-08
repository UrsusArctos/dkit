package openai

import "encoding/json"

func (oac TOpenAPIClient) GetGeneratedImage(prompt string, choicesWanted int) (TGeneratedImage, error) {
	// Compose completion request
	compReq := TImageRequest{
		Prompt: prompt,
		N:      choicesWanted,
		User:   projectName,
	}
	// Perform JSONRPC call
	rawResp, err := oac.apiCallJSON(eptImageGeneration, compReq)
	if err == nil {
		var GI TGeneratedImage
		uerr := json.Unmarshal(rawResp, &GI)
		if uerr == nil {
			return GI, nil
		}
		return TGeneratedImage{}, uerr
	}
	return TGeneratedImage{}, err
}
