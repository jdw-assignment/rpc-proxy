package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONRPCRequest struct {
	JSONRPC   string        `json:"jsonrpc"`
	RPCMethod string        `json:"method"`
	Params    []interface{} `json:"params,omitempty"`
	ID        int           `json:"id"`
}

var rpcURL = "https://polygon-rpc.com"

var httpClient = &http.Client{}

func SetRpcURL(url string) {
	rpcURL = url
}

func ProxyRPCRequest(request JSONRPCRequest) (*http.Response, error) {
	request.JSONRPC = "2.0"

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := httpClient.Post(rpcURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make POST request: %v", err)
	}

	return resp, nil
}
