package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
)

type ctxKey struct{}

func InitLogger(isJSON bool) *slog.Logger {
	var handler slog.Handler

	logPath := os.Getenv("DAEMON_LOG_PATH")
	var writer io.Writer = os.Stdout

	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			writer = io.MultiWriter(os.Stdout, logFile)
		} else {
			log.Printf("failed to open log file %s: %v", logPath, err)
		}
	}

	switch isJSON {
	case true:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	case false:
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil{
		return slog.Default()
	}else{
		if logger, exist := ctx.Value(ctxKey{}).(*slog.Logger);exist{
			return logger
		}else{
			return slog.Default()
		}
	}
}

func WithFields(ctx context.Context, args ...any) context.Context {
	logger := FromContext(ctx).With(args...)
	return ToContext(ctx, logger)
}