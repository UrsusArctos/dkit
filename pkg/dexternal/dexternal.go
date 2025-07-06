package dexternal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"reflect"
	"sync"
	"time"
)

const (
	jobBackLog = 0x10
	// job statuses
	JobStatusUndefined = 0x00
	JobStatusReady     = 0x01
	JobStatusRunning   = 0x02
	JobStatusDone      = 0x03
	// Payload headers
	hhContentType = "Content-Type"
	hhAppJSON     = "application/json"
)

type (
	TDexJobHeaders           = map[string]string
	TDexJobID                = uint64
	TDexJobHeap              = map[TDexJobID]*TDexJob
	TDexJobChan              = chan TDexJobID
	TDexJobStatus            = uint8
	TDexJobCompletionHandler = func(DJob *TDexJob)

	TDexJob struct {
		// Public fields
		// apiCall parameters
		Method  string // Refer to http.Method* constants
		URL     string // FQDN + URI
		Headers TDexJobHeaders
		// Payload
		Payload     any
		PayloadFile *TPayloadFile
		// apiCall outcome
		APICallError error
		RequestError error
		HTTPResponse *http.Response
		// Private fields
		status          TDexJobStatus
		mut             sync.Mutex
		httpRawResponse []byte
	}

	TDexternal struct {
		// Public fields
		CompletionHandler TDexJobCompletionHandler
		// Private fields
		rng     *rand.Rand
		job     TDexJobHeap
		jobChan TDexJobChan
		wgroup  sync.WaitGroup
	}

	TPayloadFile struct {
		FileNameLocal  string
		FileNameRemote string
		MIMEType       string
		FieldName      string
	}
)

// TDexternal constructor

func NewInstance() *TDexternal {
	dext := &TDexternal{
		CompletionHandler: nil,
		rng:               rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)),
		job:               make(TDexJobHeap),
		jobChan:           make(TDexJobChan, jobBackLog),
	}
	go dext.monitorWorker()
	return dext
}

// Job Management

func (DEX *TDexternal) RegisterJob(newjob *TDexJob) TDexJobID {
	id := DEX.rng.Uint64()
	newjob.status = JobStatusReady
	DEX.job[id] = newjob
	return id
}

func (DEX *TDexternal) isJobRegistered(jobid TDexJobID) bool {
	_, exists := DEX.job[jobid]
	return exists
}

func (DEX *TDexternal) StartJob(jobid TDexJobID) {
	if DEX.isJobRegistered(jobid) && (DEX.job[jobid].status == JobStatusReady) {
		DEX.wgroup.Add(1)
		go DEX.jobWorker(jobid)
	}
}

func (DEX *TDexternal) WaitSyncForJobs() {
	DEX.wgroup.Wait()
}

func (DEX *TDexternal) Job(jobid TDexJobID) *TDexJob {
	return DEX.job[jobid]
}

func (DEX *TDexternal) jobWorker(jobid TDexJobID) {
	defer DEX.wgroup.Done()
	// Lock job
	DEX.job[jobid].Lock()
	DEX.job[jobid].status = JobStatusRunning
	// Make HTTP query
	DEX.job[jobid].APICallError = DEX.job[jobid].apiCallEx()
	// Release the job
	DEX.job[jobid].Unlock()
	// Worker is done, raw response received
	DEX.jobChan <- jobid
	// job is still marked as running
}

func (DEX *TDexternal) monitorWorker() {
	for {
		select {
		case jobid := <-DEX.jobChan:
			{
				// Lock job
				DEX.job[jobid].Lock()
				DEX.job[jobid].status = JobStatusDone
				// call completion handler
				if DEX.CompletionHandler != nil {
					DEX.CompletionHandler(DEX.job[jobid])
				}
				// Unlock job
				// This is POTENTIALLY DANGEROUS and may cause deadlock because it prevents any further job from starting
				// until completion handler is done.
				DEX.job[jobid].Unlock()
			}
		default:
			{
				// do nothing, just wait
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// TDexJob access synchronization

func (DJob *TDexJob) Lock() {
	DJob.mut.Lock()
}

func (DJob *TDexJob) Unlock() {
	DJob.mut.Unlock()
}

// Actual job functions for the worker goroutine

func (DJob *TDexJob) apiCallEx() error {
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

	// Close the multipart writer (do not defer, because multipart body needs to be closed before sending)
	if requestBody.Len() > 0 {
		requestWriter.Close()
	}
	// Create HTTP request
	// fmt.Println(requestBody.String())
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

func (DJob *TDexJob) DecodeResponse(outStruct any) error {
	decoder := json.NewDecoder(bytes.NewReader(DJob.httpRawResponse))
	decoder.UseNumber()
	return decoder.Decode(&outStruct)
}

func (DJob *TDexJob) GetRawResponseJSON() string {
	return string(DJob.httpRawResponse)
}
