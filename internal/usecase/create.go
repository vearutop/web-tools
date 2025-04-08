package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/url"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/usecase"
)

type createMockInput struct {
	request.EmbeddedSetter
	LogsKey         string            `query:"logs_key" description:"Access log key to collect request data, put something random here"`
	R301            string            `query:"r301" description:"URL to redirect to with 302"`
	R302            string            `query:"r302" description:"URL to redirect to with 302"`
	Status          int               `query:"status" description:"Response status code, for non 301/302 cases, default 200"`
	RespBody        string            `contentType:"text/plain" description:"Response body" example:"<h1>Hello World</h1>"`
	RespContentType string            `query:"ct" description:"Response Content-Type, default text/plain"`
	RespHeaders     map[string]string `query:"headers" description:"Response headers, default empty"`
}

type createMockOutput struct {
	Text string `contentType:"text/plain"`
}

// CreateMock prepares a mock URL.
func CreateMock(_ any) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input createMockInput, output *createMockOutput) error {
		q := url.Values{}

		if input.LogsKey != "" {
			q.Set("logs_key", input.LogsKey)
		}

		if input.R301 != "" {
			q.Set("r301", input.R301)
		}

		if input.R302 != "" {
			q.Set("r302", input.R302)
		}

		if input.Status != 0 {
			q.Set("status", strconv.Itoa(input.Status))
		}

		if input.RespBody != "" {
			buf := bytes.NewBuffer(nil)
			w := brotli.NewWriterLevel(buf, brotli.BestCompression)

			_, err := w.Write([]byte(input.RespBody))
			if err != nil {
				return err
			}

			if err := w.Close(); err != nil {
				return err
			}

			q.Set("body", base64.StdEncoding.EncodeToString(buf.Bytes()))
		}

		if input.RespContentType != "" {
			q.Set("ct", input.RespContentType)
		}

		if input.RespHeaders != nil {
			for k, v := range input.RespHeaders {
				q.Set("headers["+k+"]", v)
			}
		}

		baseURL := ""

		if ref := input.Request().Referer(); ref != "" {
			r, err := url.Parse(ref)
			if err != nil {
				return err
			}

			baseURL = r.Scheme + "://" + r.Host
		}

		mu := baseURL + "/mock?" + q.Encode()

		output.Text = "Mock URL:\n" + mu + "\n"

		if input.LogsKey != "" {
			output.Text = output.Text + "\nLogs:\n" + baseURL + "/logs/" + input.LogsKey + "\n"
		}

		return nil
	})

	u.SetTags("Create Mock")
	u.SetDescription("Create a mock that serves pre-defined response based on query parameters.\n\nUse non-empty request body to define response body.\n\nMock is stateless, all response parameters are provided with request.")

	return u
}
