package main

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"rinha-backend-2026-go/internal/ivf"
	"rinha-backend-2026-go/internal/vector"

	"github.com/valyala/fasthttp"
)

var (
	index   *ivf.Index
	norm    *vector.Norm
	mccRisk vector.MccRisk

	readyFlag uint32
)

func main() {
	listenAddr := envOrDefault("LISTEN_ADDR", "/run/sock/api.sock")
	normPath := envOrDefault("NORMALIZATION_PATH", "resources/normalization.json")
	mccPath := envOrDefault("MCC_RISK_PATH", "resources/mcc_risk.json")
	ivfPath := envOrDefault("IVF_PATH", "resources/ivf.bin")

	nProbeFast := envIntOrDefault("N_PROBE_FAST", 8)
	nProbeFull := envIntOrDefault("N_PROBE_FULL", 28)

	startLoad := time.Now()
	if err := loadResources(normPath, mccPath, ivfPath); err != nil {
		log.Fatalf("load resources: %v", err)
	}
	defer index.Close()
	log.Printf("resources loaded in %s", time.Since(startLoad).Round(time.Millisecond))
	atomic.StoreUint32(&readyFlag, 1)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			handler(ctx, nProbeFast, nProbeFull)
		},
		Name:                          "rinha-backend-2026",
		Concurrency:                   256,
		DisableHeaderNamesNormalizing: true,
		NoDefaultServerHeader:         true,
		NoDefaultContentType:          true,
		NoDefaultDate:                 true,
		ReadBufferSize:                2048,
		WriteBufferSize:               256,
		MaxRequestBodySize:            8 << 10,
		ReadTimeout:                   2 * time.Second,
		WriteTimeout:                  2 * time.Second,
		IdleTimeout:                   60 * time.Second,
	}

	listener, err := createListener(listenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("listening on %s", listenAddr)
	log.Fatal(server.Serve(listener))
}

func handler(ctx *fasthttp.RequestCtx, nProbeFast, nProbeFull int) {
	path := string(ctx.Path())
	if path == "/ready" {
		if !ctx.IsGet() {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			return
		}
		if atomic.LoadUint32(&readyFlag) == 1 {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		return
	}

	if path != "/fraud-score" {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	if !ctx.IsPost() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	var query [vector.Dim]float64
	if err := vector.FromPayload(ctx.PostBody(), norm, mccRisk, &query); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	fraudCount := index.FraudScore(query, nProbeFast, nProbeFull)
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(responses[fraudCount])
}

func loadResources(normPath, mccPath, ivfPath string) error {
	var err error

	norm, err = vector.LoadNorm(normPath)
	if err != nil {
		return err
	}

	mccRisk, err = vector.LoadMccRisk(mccPath)
	if err != nil {
		return err
	}

	index, err = ivf.Open(ivfPath)
	if err != nil {
		return err
	}
	index.PreTouch()
	return nil
}

func createListener(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, "/") {
		if err := os.MkdirAll(filepath.Dir(addr), 0o755); err != nil {
			return nil, err
		}
		_ = os.Remove(addr)
		listener, err := net.Listen("unix", addr)
		if err != nil {
			return nil, err
		}
		if err := os.Chmod(addr, 0o666); err != nil {
			listener.Close()
			return nil, err
		}
		return listener, nil
	}
	return net.Listen("tcp", addr)
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
