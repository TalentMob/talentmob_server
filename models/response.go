package models

import "github.com/ant0ine/go-json-rest/rest"

// Base response for JSON data
type BaseResponse struct {
	Success bool        `json:"success"`
	Info    string      `json:"info"`
	Result  interface{} `json:"result"`
	writer  rest.ResponseWriter
}

const (
	ErrorUnauthorized = "unauthorized"
)

// Initialise with response writer
func (r *BaseResponse) Init(w rest.ResponseWriter) {
	r.writer = w
}

// Return a failed JSON response
func (r *BaseResponse) SendError(info string) {
	r.Success = false
	r.Info = info
	r.writer.WriteJson(&r)
}

// Return a successful JSON response
func (r *BaseResponse) SendSuccess(o interface{}) {
	r.Success = true
	r.Result = o
	r.writer.WriteJson(&r)
}
