package event

import (
	"context"

	"go.uber.org/zap"

	"github.com/evo-lee/evo-sonic/log"
)

// AsyncWorkerPool wraps a Bus and dispatches events to listeners
// via a bounded goroutine pool, preventing AI-heavy listeners from
// blocking the synchronous event bus.
type AsyncWorkerPool struct {
	bus      Bus
	queue    chan asyncTask
	workers  int
	stopCh   chan struct{}
}

type asyncTask struct {
	ctx      context.Context
	event    Event
	listener Listener
}

// NewAsyncWorkerPool creates a pool with the given worker count and queue depth,
// then subscribes an async wrapper for the specified event types.
func NewAsyncWorkerPool(bus Bus, workers, queueSize int, eventTypes ...string) *AsyncWorkerPool {
	p := &AsyncWorkerPool{
		bus:     bus,
		queue:   make(chan asyncTask, queueSize),
		workers: workers,
		stopCh:  make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		go p.run()
	}
	return p
}

// SubscribeAsync subscribes listener to eventType through the async pool.
func (p *AsyncWorkerPool) SubscribeAsync(eventType string, listener Listener) {
	p.bus.Subscribe(eventType, func(ctx context.Context, e Event) error {
		select {
		case p.queue <- asyncTask{ctx: ctx, event: e, listener: listener}:
		default:
			// Queue full — log and drop to avoid blocking the sync bus.
			log.Error("async worker queue full, dropping event", zap.String("type", e.EventType()))
		}
		return nil
	})
}

func (p *AsyncWorkerPool) run() {
	for {
		select {
		case task := <-p.queue:
			if err := task.listener(task.ctx, task.event); err != nil {
				log.Error("async listener error",
					zap.String("event", task.event.EventType()),
					zap.Error(err))
			}
		case <-p.stopCh:
			return
		}
	}
}

// Stop drains the queue and shuts down workers.
func (p *AsyncWorkerPool) Stop() {
	close(p.stopCh)
}
