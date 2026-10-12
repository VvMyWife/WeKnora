package asynq
import (
 "testing"
 "time"
 "github.com/hibiken/asynq/internal/base"
)
func TestUnlimitedDeadline(t *testing.T) {
 p:=&processor{}
 if d:=p.computeDeadline(&base.TaskMessage{Timeout:-1});!d.IsZero(){t.Fatalf("unexpected deadline %v",d)}
 want:=time.Now().Add(time.Hour).Unix()
 if d:=p.computeDeadline(&base.TaskMessage{Timeout:-1,Deadline:want});d.Unix()!=want{t.Fatal("explicit deadline lost")}
 opts,err:=composeOptions(Timeout(-time.Second));if err!=nil || opts.timeout!=-time.Second{t.Fatal("unlimited timeout lost in queue options")}
}
