package chrome

import (
	"context"
	"log"
)

var contextErrorLoggerKey = "error-logger" //nolint:gochecknoglobals

func errorLogger(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value(&contextErrorLoggerKey).(*log.Logger)
	if !ok {
		panic("no error logger")
	}

	return log.New(logger.Writer(), logger.Prefix(), logger.Flags())
}

func withErrorLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, &contextErrorLoggerKey, log.New(logger.Writer(), logger.Prefix(), logger.Flags()))
}

var contextInfoLoggerKey = "info-logger" //nolint:gochecknoglobals

func infoLogger(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value(&contextInfoLoggerKey).(*log.Logger)
	if !ok {
		panic("no info logger")
	}

	return log.New(logger.Writer(), logger.Prefix(), logger.Flags())
}

func withInfoLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, &contextInfoLoggerKey, log.New(logger.Writer(), logger.Prefix(), logger.Flags()))
}
