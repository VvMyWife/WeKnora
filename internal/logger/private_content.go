package logger

import (
	"context"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/sirupsen/logrus"
	"io"
)

type privateContentKey struct{}

// WithPrivateContent silences inherited verbose pipeline logs for the anonymous channel.
// The HTTP adapter logs only non-content request metrics using its original context.
func WithPrivateContent(ctx context.Context) context.Context {
	l := logrus.New()
	l.SetOutput(io.Discard)
	ctx = context.WithValue(ctx, privateContentKey{}, true)
	return context.WithValue(ctx, types.LoggerContextKey, logrus.NewEntry(l))
}
func PrivateContent(ctx context.Context) bool {
	v, _ := ctx.Value(privateContentKey{}).(bool)
	return v
}
