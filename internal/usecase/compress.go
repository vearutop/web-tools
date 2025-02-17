package usecase

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/andybalholm/brotli"
	"github.com/swaggest/usecase"
)

// Compress converts a text to base64 of brotli.
func Compress(_ interface{}) usecase.Interactor {
	type req struct {
		Text string `contentType:"text/plain"`
	}

	type resp struct {
		Compressed string `contentType:"text/plain"`
		RawLen     int    `header:"X-Raw-Length"`
		CmpLen     int    `header:"X-Compressed-Length"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input req, output *resp) error {
		buf := bytes.NewBuffer(nil)
		w := brotli.NewWriterLevel(buf, brotli.BestCompression)

		_, err := w.Write([]byte(input.Text))
		if err != nil {
			return err
		}

		if err := w.Close(); err != nil {
			return err
		}

		output.Compressed = base64.StdEncoding.EncodeToString(buf.Bytes())
		output.RawLen = len(input.Text)
		output.CmpLen = len(output.Compressed)

		return nil
	})

	u.SetTags("Mock")
	u.SetDescription("Compresses text to base64(brotli) to use as /mock response body.")

	return u
}
