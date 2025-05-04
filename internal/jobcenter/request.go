package jobcenter

import (
	"encoding/json"
	"fmt"
)

// WRAPPER STRUCT

type ProcessResultType int

const (
	Put ProcessResultType = iota
	Get
	Delete
	Abort
	Error
)

type Request struct {
	// The request type (e.g., "put", "get", "delete", "abort")
	Type    ProcessResultType
	Request *PutRequest
	Get     *GetRequest
	Delete  *DeleteRequest
	Abort   *AbortRequest
	Error   error
	RawJson []byte // Keep raw JSON for debugging
}

// JSON PARSING FUNCTION

// parseJsonLine attempts to parse the jsonBytes into one of the known structures
func parseJsonLine(jsonBytes []byte) Request {
	// 1. Peek at the 'request' field first
	var baseReq Base
	err := json.Unmarshal(jsonBytes, &baseReq)
	if err != nil {
		// JSON syntax is invalid, we cannot even determine the request type
		return Request{
			Type:    Error,
			Error:   fmt.Errorf("invalid JSON syntax: %w", err),
			RawJson: jsonBytes, // Keep raw data for debugging
		}
	}

	// 2. Decide the target type based on the 'request' field
	switch baseReq.Request {
	case "put": // Example request type string
		var data PutRequest
		err = json.Unmarshal(jsonBytes, &data)
		if err != nil {
			// Valid JSON syntax, but doesn't match RequestData structure
			return Request{Type: Error, Error: fmt.Errorf("failed to parse as RequestData: %w", err), RawJson: jsonBytes}
		}
		return Request{Type: Put, Request: &data, RawJson: jsonBytes}

	case "get": // Example request type string
		var data GetRequest
		err = json.Unmarshal(jsonBytes, &data)
		if err != nil {
			return Request{Type: Error, Error: fmt.Errorf("failed to parse as GetData: %w", err), RawJson: jsonBytes}
		}
		return Request{Type: Get, Get: &data, RawJson: jsonBytes}

	case "delete": // Example request type string
		var data DeleteRequest
		err = json.Unmarshal(jsonBytes, &data)
		if err != nil {
			return Request{Type: Error, Error: fmt.Errorf("failed to parse as DeleteData: %w", err), RawJson: jsonBytes}
		}
		return Request{Type: Delete, Delete: &data, RawJson: jsonBytes}

	case "abort": // Example request type string
		var data AbortRequest
		err = json.Unmarshal(jsonBytes, &data)
		if err != nil {
			return Request{Type: Error, Error: fmt.Errorf("failed to parse as AbortData: %w", err), RawJson: jsonBytes}
		}
		return Request{Type: Abort, Abort: &data, RawJson: jsonBytes}

	default:
		// The 'request' field has a value we don't recognize
		return Request{
			Type:    Error, // Or ResultTypeUnknown if you want to differentiate
			Error:   fmt.Errorf("unknown request type: '%s'", baseReq.Request),
			RawJson: jsonBytes,
		}
	}
}

// JSON PARSING STRUCTS

type Job struct {
	Priority int         `json:"pri"`
	Id       int         `json:"id"`
	Content  interface{} `json:"job"`
	Queue    string      `json:"queue"`
}

// Base is used to peek at the 'request' field only
type Base struct {
	Request string `json:"request"`
}

// PutRequest represents the main JSON structure
type PutRequest struct {
	Request  string      `json:"request"` // Field names must be exported (start with uppercase)
	Queue    string      `json:"queue"`
	Job      interface{} `json:"job"` // Use the Job struct type for the nested object
	Priority int         `json:"pri"` // Map Go's "Priority" field to JSON's "pri" key
}

type PutResponse struct {
	Status string `json:"status"`
	Id     int    `json:"id"`
}

// GetRequest represents the JSON structure
type GetRequest struct {
	Request string   `json:"request"`
	Queues  []string `json:"queues"`

	// Use a pointer (*bool) for the optional "wait" field.
	// If "wait" is missing in the JSON, this field will be nil.
	// If "wait" is present (true or false), this field will point to the boolean value.
	// "omitempty" is good practice for marshaling: if Wait is nil, the key won't be included.
	Wait *bool `json:"wait,omitempty"`
}

type GetResponse struct {
	Status   string      `json:"status"`
	ID       *int        `json:"id,omitempty"`
	Job      interface{} `json:"job,omitempty"`
	Priority *int        `json:"pri,omitempty"`
	Queue    *string     `json:"queue,omitempty"`
}

// DeleteRequest represents the JSON structure for deleting a job
type DeleteRequest struct {
	Request string `json:"request"`
	Id      int    `json:"id"`
}

type DeleteResponse struct {
	Status string `json:"status"`
}

// AbortRequest represents the JSON structure for aborting a job
type AbortRequest struct {
	Request string `json:"request"`
	Id      int    `json:"id"`
}

type AbortResponse struct {
	Status string `json:"status"`
}
