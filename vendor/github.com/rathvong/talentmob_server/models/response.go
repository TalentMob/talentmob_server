package models

import (
	"encoding/json"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

// Base response for JSON data
type BaseResponse struct {
	Success     bool        `json:"success"`
	Info        string      `json:"info"`
	Result      interface{} `json:"result"`
	writer      rest.ResponseWriter
	writerV2    http.ResponseWriter
	wroteHeader bool
}

const (
	ErrorUnauthorized = "unauthorized"
)

// Initialise with response writer
func (r *BaseResponse) Init(w rest.ResponseWriter) {
	r.writer = w
}

func (r *BaseResponse) InitV2(w http.ResponseWriter) {
	r.writerV2 = w
}

// Return a failed JSON response
func (r *BaseResponse) SendError(info string) {
	r.Success = false
	r.Info = info

	if r.writerV2 == nil {
		r.writer.WriteJson(&r)
		return
	}

	r.writeJSON(&r)
}

func (r *BaseResponse) writeJSON(v interface{}) error {

	b, err := r.encodeJson(v)

	if err != nil {
		return err
	}

	return r.writeJSON(b)
}

func (r *BaseResponse) encodeJson(v interface{}) ([]byte, error) {

	b, err := json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *BaseResponse) writeHeader(code int) {
	if r.writerV2.Header().Get("Content-Type") == "" {
		// Per spec, UTF-8 is the default, and the charset parameter should not
		// be necessary. But some clients (eg: Chrome) think otherwise.
		// Since json.Marshal produces UTF-8, setting the charset parameter is a
		// safe option.
		r.writerV2.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	r.writerV2.WriteHeader(code)
	r.wroteHeader = true
}

// Provided in order to implement the http.ResponseWriter interface.
func (r *BaseResponse) write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.writeHeader(http.StatusOK)
	}
	return r.write(b)
}

// Return a successful JSON response
func (r *BaseResponse) SendSuccess(o interface{}) {
	r.Success = true
	r.Result = o

	if r.writerV2 == nil {
		r.writer.WriteJson(&r)
		return
	}

	r.writeJSON(&r)
}
