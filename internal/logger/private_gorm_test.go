package logger

import (
	"bytes"
	"context"
	"errors"
	gormlog "gorm.io/gorm/logger"
	"log"
	"strings"
	"testing"
	"time"
)

func TestPrivateSQLNeverFormatsContent(t *testing.T) {
	var b bytes.Buffer
	l := PrivateGORMLogger{gormlog.New(log.New(&b, "", 0), gormlog.Config{LogLevel: gormlog.Info})}
	l.Trace(WithPrivateContent(context.Background()), time.Now(), func() (string, int64) { t.Fatal("private SQL evaluated"); return "secret", 0 }, errors.New("failed"))
	l.Trace(context.Background(), time.Now(), func() (string, int64) { return "SELECT normal", 1 }, nil)
	if !strings.Contains(b.String(), "SELECT normal") {
		t.Fatal("normal logging changed")
	}
}
