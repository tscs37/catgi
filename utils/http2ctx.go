package utils

import (
	"context"
	"net/http"
)

// PutHTTPIntoContext embeds the current http Request into the context
// with the key "http_ctx"
func PutHTTPIntoContext(r *http.Request, ctx context.Context) context.Context {
	return context.WithValue(ctx, "http_ctx", r)
}
