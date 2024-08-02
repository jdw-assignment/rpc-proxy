package handlers

import (
	"encoding/json"
	"github.com/joeldavidw/rpc-proxy/rpc"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"
	"strings"
)

var allowedMethods = map[string]bool{
	"eth_blockNumber":      true,
	"eth_getBlockByNumber": true,
}

const name = "github.com/joeldavidw/rpc/rpc_client"

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

func HandleRPCRequest(rpcClient func(rpc.JSONRPCRequest) (*http.Response, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "HandleRPCRequest")
		defer span.End()

		if r.Method != http.MethodPost {
			http.Error(w, "HTTP method not allowed: "+r.Method, http.StatusMethodNotAllowed)
			logger.InfoContext(ctx, "HTTP method not allowed", "httpMethod", r.Method)
			return
		}

		var req rpc.JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			logger.InfoContext(ctx, "Invalid request", "request", r.Body)
			return
		}

		if !allowedMethods[req.RPCMethod] {
			http.Error(w, "RPC method not allowed", http.StatusForbidden)
			logger.InfoContext(ctx, "RPC method not allowed", "method", req.RPCMethod)
			return
		}

		rpcResponse, err := rpcClient(req)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				http.Error(w, "Request timed out", http.StatusGatewayTimeout)
				logger.InfoContext(ctx, "Request timed out")
			} else {
				http.Error(w, "Failed to make RPC call: "+err.Error(), http.StatusInternalServerError)
				logger.InfoContext(ctx, "Failed to make RPC call", "error", err.Error())
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rpcResponse.StatusCode)
		if _, err := io.Copy(w, rpcResponse.Body); err != nil {
			http.Error(w, "Failed to copy response body", http.StatusInternalServerError)
			logger.InfoContext(ctx, "Failed to copy response body", "error", err.Error())
		}
		logger.InfoContext(ctx, "Proxied successfully", "method", req.RPCMethod)
	}
}
