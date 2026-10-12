package context
import (
 stdctx "context"
 "testing"
 "time"
 "github.com/hibiken/asynq/internal/base"
)
func TestUnlimitedCancellation(t *testing.T) {
 ctx,cancel:=New(stdctx.Background(),&base.TaskMessage{ID:"test",Retry:3},time.Time{})
 if _,ok:=ctx.Deadline();ok{t.Fatal("unlimited task acquired a deadline")}
 if id,ok:=GetTaskID(ctx);!ok||id!="test"{t.Fatal("task metadata lost")}
 cancel();if ctx.Err()!=stdctx.Canceled{t.Fatal("user cancellation lost")}
 parent,stop:=stdctx.WithCancel(stdctx.Background());child,done:=New(parent,&base.TaskMessage{},time.Time{});defer done();stop()
 if child.Err()!=stdctx.Canceled{t.Fatal("shutdown cancellation lost")}
}
