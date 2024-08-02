package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joeldavidw/rpc-proxy/handlers"
	"github.com/joeldavidw/rpc-proxy/logging"
	"github.com/joeldavidw/rpc-proxy/rpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Setup OpenTelemetry SDK
	otelShutdown, err := logging.SetupOTelSDK(ctx)
	if err != nil {
		log.Fatalf("failed to set up OpenTelemetry SDK: %v", err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	rpcURL := getEnv("RPC_URL", "https://polygon-rpc.com")
	rpc.SetRpcURL(rpcURL)

	serverPort := getEnv("PORT", "8080")

	srv := &http.Server{
		Addr:         ":" + serverPort,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}

	log.Printf("Starting server on %s", srv.Addr)
	log.Printf("RPC URL: %s", rpcURL)

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		log.Printf("Server error: %v", err)
	case <-ctx.Done():
		stop()
	}

	if err = srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handleFunc("GET /health", handlers.HealthCheck)
	handleFunc("POST /rpc", handlers.HandleRPCRequest(rpc.ProxyRPCRequest))

	return otelhttp.NewHandler(mux, "/")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
