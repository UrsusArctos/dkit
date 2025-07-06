package dexternal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"reflect"
)

const (
	// Request headers
	hhContentType = "Content-Type"
	hhAppJSON     = "application/json"
)

type (
	TDexJobHeaders = map[string]string

	TPayloadFile struct {
		FileNameLocal  string
		FileNameRemote string
		MIMEType       string
		FieldName      string
	}

	TDexternalJob struct {
		// apiCall parameters
		Method  string // Refer to http.Method* constants
		URL     string // FQDN + URI
		Headers TDexJobHeaders
		// Payload
		Payload     any
		PayloadFile *TPayloadFile
		// Return values
		HTTPResponse    *http.Response // HTTP response object
		RequestError    error
		httpRawResponse []byte // HTTP response body as raw bytes
	}
)

func (DJob *TDexternalJob) Perform() error {
	var payloadJSON []byte = nil
	var payloadFile []byte = nil
	var err_pfile error = nil
	var requestBody *bytes.Buffer = &bytes.Buffer{}
	var requestHeader *http.Header = &http.Header{}
	var requestWriter *multipart.Writer = multipart.NewWriter(requestBody)

	// Check if we have a main payload
	if DJob.Payload != nil {
		// Check if the payload is a struct or a map
		switch reflect.ValueOf(DJob.Payload).Kind() {
		case reflect.Struct:
			{
				// This is a structure, so we should marshal it into JSON
				payloadJSON, _ = json.Marshal(DJob.Payload)
				requestBody = bytes.NewBuffer(payloadJSON)
				requestHeader.Set(hhContentType, hhAppJSON)
			}
		case reflect.Map:
			{
				// This is a map, so we should make it into multipart form data
				// For sake of simplicity, assume all values in the map are strings
				for k, v := range DJob.Payload.(map[string]string) {
					requestWriter.WriteField(k, v)
				}
				requestHeader.Set(hhContentType, requestWriter.FormDataContentType())
			}
		}
	}

	// Check if we have file payload also
	if DJob.PayloadFile != nil {
		payloadFile, err_pfile = os.ReadFile(DJob.PayloadFile.FileNameLocal)
		if (payloadFile != nil) && (err_pfile == nil) {
			pfheader := make(textproto.MIMEHeader)
			pfheader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, DJob.PayloadFile.FieldName, DJob.PayloadFile.FileNameRemote))
			pfheader.Set("Content-Type", DJob.PayloadFile.MIMEType)
			pfpart, _ := requestWriter.CreatePart(pfheader)
			pfpart.Write(payloadFile)
			requestHeader.Set(hhContentType, requestWriter.FormDataContentType())
		}
	}

	// Close the multipart writer (do not defer this, because multipart body needs to be closed before sending)
	if requestBody.Len() > 0 {
		requestWriter.Close()
	}

	// Create HTTP request
	httpRequest, err := http.NewRequest(DJob.Method, DJob.URL, requestBody)
	if err == nil {
		// Add HTTP headers specified in the Job structure
		for k, v := range DJob.Headers {
			httpRequest.Header.Add(k, v)
		}
		// Add global HTTP headers specified in the payload cases
		for k, v := range *requestHeader {
			httpRequest.Header[k] = v
		}
		// Create HTTP Client
		httpClient := &http.Client{}
		DJob.HTTPResponse, DJob.RequestError = httpClient.Do(httpRequest)
		if DJob.RequestError == nil {
			defer DJob.HTTPResponse.Body.Close()
			DJob.httpRawResponse, err = io.ReadAll(DJob.HTTPResponse.Body)
		}
	}
	//
	return err
}

// Response decoders

func (DJob *TDexternalJob) DecodeResponse(outStruct any) error {
	decoder := json.NewDecoder(bytes.NewReader(DJob.httpRawResponse))
	decoder.UseNumber()
	return decoder.Decode(&outStruct)
}

func (DJob *TDexternalJob) GetRawResponseJSON() string {
	return string(DJob.httpRawResponse)
}
