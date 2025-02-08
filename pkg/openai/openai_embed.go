package openai

import (
	"encoding/json"

	oaimport "github.com/sashabaranov/go-openai"
)

func (oac TOpenAPIClient) GetEmbeddings(input string) (emb []float32, err error) {
	// Here we ignore the model selected in the client and use specific embedding model explicitly
	embReq := TEmbeddingRequest{
		Input: input,
		Model: oaimport.SmallEmbedding3,
		User:  projectName,
	}
	// Perform JSONRPC call
	rawResp, err := oac.apiCallJSON(eptEmbeddings, embReq)
	if err == nil {
		var CR TEmbeddingResponse
		uerr := json.Unmarshal(rawResp, &CR)
		if uerr == nil {
			if len(CR.Data) > 0 {
				return CR.Data[0].Embedding, nil
			}
		}
		return nil, uerr
	}
	return nil, err
}
