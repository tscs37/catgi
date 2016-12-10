package logger

import (
	"context"

	"github.com/Sirupsen/logrus"
)

func LogFromCtx(src string, ctx context.Context) logrus.FieldLogger {
	if ctx != nil {
		if val := ctx.Value("logger"); val != nil {
			if log, ok := val.(logrus.FieldLogger); ok {
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

func SetLoggingLevel(level string, ctx context.Context) {
	log, ok := ctx.Value("logger").(*logrus.Logger)
	if !ok {
		return
	}
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	log.Level = lvl
}
