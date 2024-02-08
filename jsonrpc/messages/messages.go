package messages

import (
	"encoding/json"
	"errors"
	"fmt"
)

const version = "2.0"

var (
	ErrIsNotification = errors.New("request is a notification")
)

type JsonRpcRequest struct {
	Version string          `json:"jsonrpc"`
	ID      uint64          `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func (r JsonRpcRequest) IsNotification() bool {
	return r.ID == 0
}

func (r JsonRpcRequest) Validate() error {
	if r.Version != version {
		return fmt.Errorf("unsupported RPC version '%s'", r.Version)
	}

	if r.Method == "" {
		return fmt.Errorf("RPC method was not specified")
	}

	return nil
}

type JsonRpcResponse struct {
	Version string `json:"jsonrpc"`
	ID      uint64 `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func NewResponse(id uint64, result any) JsonRpcResponse {
	return JsonRpcResponse{
		Version: version,
		ID:      id,
		Result:  result,
	}
}

func NewErrorResponse(id uint64, err error) JsonRpcResponse {
	return JsonRpcResponse{
		Version: version,
		ID:      id,
		Error:   err.Error(),
	}
}
