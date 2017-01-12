package main

import (
	"net/http"

	"git.timschuster.info/rls.moe/catgi/logger"
)

type handlerInjectLog struct {
	next http.Handler
}

func newHandlerInjectLog(nextHandler http.Handler) http.Handler {
	return &handlerInjectLog{
		next: nextHandler,
	}
}

func (h *handlerInjectLog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := logger.InjectLogToContext(r.Context())
	ctx = logger.CreateRequestIDContext(ctx)
	ctx = logger.SetLoggingLevel(curCfg.LogLevel, ctx)
	log := logger.LogFromCtx("httpLogInject", ctx)
	log.Debug("Starting new request")
	h.next.ServeHTTP(w, r.WithContext(ctx))
	log.Debug("Request Finished")
}
