package worker

import (
	"time"

	"github.com/hibiken/asynq"
)

// ServerConfig captures the configuration required to bootstrap an Asynq server.
type ServerConfig struct {
	RedisOpt       asynq.RedisClientOpt
	Concurrency    int
	Queues         map[string]int
	RetryDelayFunc asynq.RetryDelayFunc
}

// DefaultRetryDelay implements an exponential back-off retry strategy used across services.
func DefaultRetryDelay(n int, _ error, _ *asynq.Task) time.Duration {
	if n < 1 {
		n = 1
	}
	return time.Duration(n*n) * time.Second
}

// NewServer constructs a configured *asynq.Server with sensible defaults.
func NewServer(cfg ServerConfig) *asynq.Server {
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	queues := cfg.Queues
	if len(queues) == 0 {
		queues = map[string]int{"default": 1}
	}

	retryFn := cfg.RetryDelayFunc
	if retryFn == nil {
		retryFn = DefaultRetryDelay
	}

	return asynq.NewServer(cfg.RedisOpt, asynq.Config{
		Concurrency:    concurrency,
		Queues:         queues,
		RetryDelayFunc: retryFn,
	})
}

// NewClient builds an *asynq.Client using shared defaults so we can swap implementations centrally.
func NewClient(opt asynq.RedisClientOpt) *asynq.Client {
	return asynq.NewClient(opt)
}
