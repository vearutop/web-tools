package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/bool64/cache"
	"github.com/bool64/ctxd"
	"github.com/bool64/logz"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

type mockInput struct {
	request.EmbeddedSetter
	MockData
}

// MockData defines mock behavior.
type MockData struct {
	LogsKey         string            `query:"logs_key" json:"-" description:"Access log key to collect request data, put something random here"`
	R301            string            `query:"r301" json:"r301,omitempty" description:"URL to redirect to with 302"`
	R302            string            `query:"r302"  json:"r302,omitempty" description:"URL to redirect to with 302"`
	Status          int               `query:"status" json:"status,omitempty" description:"Response status code, for non 301/302 cases, default 200"`
	RespBody        string            `query:"body" json:"respBody,omitempty" description:"Response body, text or base64(brotli) - see /compress to prepare body, default empty"`
	RespContentType string            `query:"ct" json:"ct,omitempty" description:"Response Content-Type, default text/plain"`
	RespHeaders     map[string]string `query:"headers" json:"respHeaders,omitempty" description:"Response headers, default empty"`
	rawCompressed   []byte
}

func (input *MockData) prepare() error {
	if input.RespContentType == "" {
		input.RespContentType = "text/plain"
	}

	if input.RespBody == "" && len(input.rawCompressed) == 0 {
		return nil
	}

	// Cached compressed.
	rawCompressed := input.rawCompressed

	// Base64 compressed from request.
	if rawCompressed == nil {
		if b64, err := base64.StdEncoding.DecodeString(input.RespBody); err == nil {
			rawCompressed = b64
		} else {
			// Not a compressed body.
			return nil
		}
	}

	// Not compressed from request.
	if rawCompressed == nil {
		return nil
	}

	// Unpacking compressed.
	b := bytes.NewReader(rawCompressed)
	r := brotli.NewReader(b)

	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	input.RespBody = string(raw)

	return nil
}

type mockDeps interface {
	AccessLogzObserver(ctx context.Context, key string) (*logz.Observer, error)
	CtxdLogger() ctxd.Logger
	ShortLinksStore() *cache.ShardedMapOf[MockData]
}

func serveMock(ctx context.Context, deps mockDeps, input MockData, r *http.Request, rw http.ResponseWriter) error {
	if err := input.prepare(); err != nil {
		return err
	}

	if input.LogsKey != "" {
		o, err := deps.AccessLogzObserver(ctx, input.LogsKey)
		if err != nil {
			return err
		}

		o.ObserveMessage("mock page served", ctxd.Tuples{
			"input", input,
			"header", r.Header,
			"requestUri", r.RequestURI,
		})
	}

	if input.R301 != "" {
		http.Redirect(rw, r, input.R301, http.StatusMovedPermanently)

		return nil
	}

	if input.R302 != "" {
		http.Redirect(rw, r, input.R302, http.StatusFound)

		return nil
	}

	rw.Header().Set("Content-Type", input.RespContentType)

	for k, v := range input.RespHeaders {
		rw.Header().Set(k, v)
	}

	if input.Status > 0 {
		rw.WriteHeader(input.Status)
	}

	if _, err := rw.Write([]byte(input.RespBody)); err != nil {
		deps.CtxdLogger().Error(ctx, "failed to write response", "error", err)
	}

	return nil
}

// Mock serves a request with a predefined response.
func Mock(deps mockDeps) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input mockInput, output *response.EmbeddedSetter) error {
		return serveMock(ctx, deps, input.MockData, input.Request(), output.ResponseWriter())
	})

	u.SetTags("Serve Mock")
	u.SetDescription("Mock serves pre-defined response based on query parameters.")

	return u
}

// ShortMock serves Mock from a shortened URL.
func ShortMock(deps mockDeps) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input request.EmbeddedSetter, output *response.EmbeddedSetter) error {
		key := []byte(strings.TrimPrefix(input.Request().URL.Path, "/"))

		md, err := deps.ShortLinksStore().Read(ctx, key)
		if err == nil {
			return serveMock(ctx, deps, md, input.Request(), output.ResponseWriter())
		}

		if errors.Is(err, cache.ErrNotFound) {
			return status.NotFound
		}

		return err
	})

	u.SetTags("Serve Mock")
	u.SetDescription("Mock serves pre-defined response from a shortened link.")

	return u
}
