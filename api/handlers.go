package main

import (
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/valyala/fasthttp"
)

var readyFlag uint32

func Router(ctx *fasthttp.RequestCtx) {
	// Minimal router to keep overhead low.
	path := string(ctx.Path())
	if path == "/ready" {
		Ready(ctx)
		return
	}
	if path == "/fraud-score" {
		FraudScore(ctx)
		return
	}
	ctx.Error("not found", fasthttp.StatusNotFound)
}

func Ready(ctx *fasthttp.RequestCtx) {
	if atomic.LoadUint32(&readyFlag) == 0 {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func FraudScore(ctx *fasthttp.RequestCtx) {
	// Validate method and body upfront.
	if string(ctx.Method()) != fasthttp.MethodPost {
		writeJSONError(ctx, fasthttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if atomic.LoadUint32(&readyFlag) == 0 {
		writeJSONError(ctx, fasthttp.StatusServiceUnavailable, "resources not loaded")
		return
	}

	body := ctx.PostBody()
	if len(body) == 0 {
		writeJSONError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	payload, err := parsePayload(body)
	if err != nil {
		writeJSONError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}

	// Run scoring pipeline (vectorization + nearest neighbors).
	approved, score, err := CalculateScore(payload)
	if err != nil {
		if errors.Is(err, ErrResourcesNotLoaded) {
			writeJSONError(ctx, fasthttp.StatusInternalServerError, "resources not loaded")
			return
		}
		writeJSONError(ctx, fasthttp.StatusBadRequest, "invalid payload")
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(buildScoreResponse(approved, score))
}

func buildScoreResponse(approved bool, score float64) []byte {
	// Manual JSON formatting avoids reflection and reduces allocs.
	buf := make([]byte, 0, 64)
	buf = append(buf, '{', '"', 'a', 'p', 'p', 'r', 'o', 'v', 'e', 'd', '"', ':')
	if approved {
		buf = append(buf, 't', 'r', 'u', 'e')
	} else {
		buf = append(buf, 'f', 'a', 'l', 's', 'e')
	}
	buf = append(buf, ',', '"', 'f', 'r', 'a', 'u', 'd', '_', 's', 'c', 'o', 'r', 'e', '"', ':')
	buf = strconv.AppendFloat(buf, score, 'f', -1, 64)
	buf = append(buf, '}')
	return buf
}

func writeJSONError(ctx *fasthttp.RequestCtx, status int, message string) {
	// Minimal JSON error payload for consistent client behavior.
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(buildErrorResponse(message))
}

func buildErrorResponse(message string) []byte {
	// Quote the message to keep JSON valid without extra allocations.
	buf := make([]byte, 0, len(message)+16)
	buf = append(buf, '{', '"', 'e', 'r', 'r', 'o', 'r', '"', ':')
	buf = strconv.AppendQuote(buf, message)
	buf = append(buf, '}')
	return buf
}

func SetReady(ready bool) {
	if ready {
		atomic.StoreUint32(&readyFlag, 1)
		return
	}
	atomic.StoreUint32(&readyFlag, 0)
}
