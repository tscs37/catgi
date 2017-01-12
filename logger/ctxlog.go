package logger

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"github.com/Sirupsen/logrus"
)

func LogFromCtx(src string, ctx context.Context) logrus.FieldLogger {
	if ctx != nil {
		if val := ctx.Value("logger"); val != nil {
			if log, ok := val.(logrus.FieldLogger); ok {
				if reqId, ok := ctx.Value("logger-req-id").(string); ok {
					return log.WithField("src", src).WithField("req-id", reqId)
				}
				return log.WithField("src", src)
			}
			logrus.WithField("src", "logger").
				Error("Context had an invalid logger")
		}
	}
	return logrus.New().WithField("src", src).WithField("no-ctx", "")
}

func NewLoggingContext() context.Context {
	return InjectLogToContext(context.Background())
}

func InjectLogToContext(ctx context.Context) context.Context {
	logs := logrus.New()
	return context.WithValue(ctx, "logger", logs)
}

func CreateRequestIDContext(ctx context.Context) context.Context {
	log := LogFromCtx("cr-req-id", ctx)
	sf, err := snowflakes.NewSnowflake()
	if err != nil {
		log.Error("Error on Generating Request ID: ", err)
		return ctx
	}
	return context.WithValue(ctx, "logger-req-id", sf)

}

func SetLoggingLevel(level string, ctx context.Context) context.Context {
	log, ok := ctx.Value("logger").(*logrus.Logger)
	if !ok {
		return ctx
	}
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	log.Level = lvl
	return context.WithValue(ctx, "logger", log)
}
