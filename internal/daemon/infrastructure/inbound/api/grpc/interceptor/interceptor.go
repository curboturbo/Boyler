package interceptor

import (
	"context"
	"time"
	
	"google.golang.org/grpc"
)

func ContextInterceptor(dur time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, cancel := context.WithTimeout(ctx, dur)
		defer cancel()
		return handler(ctx, req)
	}
}