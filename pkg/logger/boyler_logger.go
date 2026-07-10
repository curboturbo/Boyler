package logger

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey struct{}

func InitLogger(isJSON bool) *slog.Logger{
	var handler slog.Handler
	switch isJSON{
	case true:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level: slog.LevelDebug,
		})
	case false:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level: slog.LevelDebug,
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