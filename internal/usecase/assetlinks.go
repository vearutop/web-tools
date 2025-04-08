package usecase

import (
	"context"

	"github.com/swaggest/rest/response"
	"github.com/swaggest/usecase"
)

// AssetLinksJSON supports deept android app.
func AssetLinksJSON(_ any) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input struct{}, output *response.EmbeddedSetter) error {
		output.ResponseWriter().Header().Add("Content-Type", "application/json")

		if _, err := output.ResponseWriter().Write([]byte(`[{
  "relation": ["delegate_permission/common.handle_all_urls"],
  "target": {
    "namespace": "android_app",
    "package_name": "com.example.deept",
    "sha256_cert_fingerprints": [
		"BD:89:E0:41:4F:12:80:24:71:33:63:07:52:82:D4:29:27:04:EC:94:40:02:27:68:69:B9:FC:32:2F:84:48:E3",
		"80:B7:95:5A:62:42:01:32:F1:43:BC:E6:FB:2D:22:A7:6D:B7:21:B1:FF:B7:1F:07:68:57:33:4F:CE:40:4D:6B"
	]
  }
}]`)); err != nil {
			return err
		}

		return nil
	})

	u.SetTags("Universal Links")
	u.SetDescription("To support https://github.com/vearutop/deept")

	return u
}
