package handlers

import (
	"bytes"
	"github.com/joeldavidw/rpc-proxy/rpc"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var mockRPCClient = func(request rpc.JSONRPCRequest) (*http.Response, error) {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}

	switch request.RPCMethod {
	case "eth_blockNumber":
		response.Body = io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","result":"0x10d4f","id":1}`))
	case "eth_getBlockByNumber":
		response.Body = io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","result":{"number":"0x1b4"},"id":1}`))
	}

	return response, nil
}

func Test_ProxyRPCRequest_BlockedMethods(t *testing.T) {
	blockedMethods := []string{
		"web3_clientVersion", "web3_sha3", "net_version", "net_listening", "net_peerCount",
		"eth_protocolVersion", "eth_syncing", "eth_coinbase", "eth_chainId", "eth_mining",
		"eth_hashrate", "eth_gasPrice", "eth_accounts", "eth_getBalance", "eth_getStorageAt",
		"eth_getTransactionCount", "eth_getBlockTransactionCountByHash", "eth_getBlockTransactionCountByNumber",
		"eth_getUncleCountByBlockHash", "eth_getUncleCountByBlockNumber", "eth_getCode", "eth_sign",
		"eth_signTransaction", "eth_sendTransaction", "eth_sendRawTransaction", "eth_call", "eth_estimateGas",
		"eth_getBlockByHash", "eth_getTransactionByHash", "eth_getTransactionByBlockHashAndIndex",
		"eth_getTransactionByBlockNumberAndIndex", "eth_getTransactionReceipt", "eth_getUncleByBlockHashAndIndex",
		"eth_getUncleByBlockNumberAndIndex", "eth_newFilter", "eth_newBlockFilter", "eth_newPendingTransactionFilter",
		"eth_uninstallFilter", "eth_getFilterChanges", "eth_getFilterLogs", "eth_getLogs",
	}

	for _, method := range blockedMethods {
		body := `{"jsonrpc":"2.0","method":"` + method + `","params":[],"id":1}`
		req, err := http.NewRequest("POST", "/", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler := HandleRPCRequest(mockRPCClient)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403 but got %d for method %s", rec.Code, method)
		}
	}
}

func Test_ProxyRPCRequest_AllowedMethods(t *testing.T) {
	allowedMethods := []string{
		"eth_blockNumber",
		"eth_getBlockByNumber",
	}

	for _, method := range allowedMethods {
		body := `{"jsonrpc":"2.0","method":"` + method + `","params":[],"id":1}`
		req, err := http.NewRequest("POST", "/", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler := HandleRPCRequest(mockRPCClient)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 but got %d for method %s", rec.Code, method)
		}
	}
}

func Test_ProxyRPCRequest_Proxy(t *testing.T) {
	allowedMethods := []string{
		"eth_blockNumber",
		"eth_getBlockByNumber",
	}

	for _, method := range allowedMethods {
		t.Run(method, func(t *testing.T) {
			body := `{"jsonrpc":"2.0","method":"` + method + `","params":[],"id":1}`
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(body))
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handler := HandleRPCRequest(mockRPCClient)
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200 but got %d for method %s", rec.Code, method)
			}

			expectedBody := ""
			switch method {
			case "eth_blockNumber":
				expectedBody = `{"jsonrpc":"2.0","result":"0x10d4f","id":1}`
			case "eth_getBlockByNumber":
				expectedBody = `{"jsonrpc":"2.0","result":{"number":"0x1b4"},"id":1}`
			}

			respBody, err := io.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("Could not read response body: %v", err)
			}

			if strings.TrimSpace(string(respBody)) != expectedBody {
				t.Errorf("expected body %s but got %s for method %s", expectedBody, string(respBody), method)
			}
		})
	}
}
