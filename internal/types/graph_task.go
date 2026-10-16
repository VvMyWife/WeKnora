package types

import "context"

type graphTaskKey struct{}

func WithGraphTask(ctx context.Context) context.Context {
	return context.WithValue(WithBackgroundTask(ctx), graphTaskKey{}, true)
}
func IsGraphTask(ctx context.Context) bool { v, _ := ctx.Value(graphTaskKey{}).(bool); return v }
