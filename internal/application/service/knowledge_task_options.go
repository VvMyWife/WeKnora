package service

import (
	"time"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/hibiken/asynq"
)

func documentProcessTaskOptions(cfg *config.Config, extra ...asynq.Option) []asynq.Option {
	timeout := config.DocumentProcessTimeout(cfg)
	if timeout < 0 {
		// The local Asynq patch treats negative seconds as no deadline.
		timeout = -time.Second
	}
	opts := []asynq.Option{
		asynq.Queue(types.QueueDefault),
		asynq.Timeout(timeout),
		asynq.MaxRetry(3),
	}
	opts = append(opts, extra...)
	return opts
}

func knowledgePostProcessTaskOptions() []asynq.Option {
	return []asynq.Option{
		asynq.Queue(types.QueuePostProcess),
		asynq.MaxRetry(3),
		asynq.Timeout(30 * time.Minute),
	}
}
