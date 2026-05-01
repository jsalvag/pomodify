package timer

import (
	"errors"
	"sync"
	"time"
)

var ErrAlreadyRunning = errors.New("timer already running")

type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type Clock interface {
	NewTicker(interval time.Duration) Ticker
}

type Engine struct {
	mu           sync.Mutex
	clock        Clock
	tickInterval time.Duration
	remaining    time.Duration
	running      bool
	paused       bool
	stopChannel  chan struct{}
	onTick       func(time.Duration)
	onDone       func()
}

type realClock struct{}

type realTicker struct {
	ticker *time.Ticker
}

func NewEngine() *Engine {
	return NewEngineWithClock(realClock{}, time.Second)
}

func NewEngineWithClock(clock Clock, tickInterval time.Duration) *Engine {
	if clock == nil {
		clock = realClock{}
	}

	if tickInterval <= 0 {
		tickInterval = time.Second
	}

	return &Engine{
		clock:        clock,
		tickInterval: tickInterval,
	}
}

func (realClock) NewTicker(interval time.Duration) Ticker {
	return realTicker{ticker: time.NewTicker(interval)}
}

func (t realTicker) Chan() <-chan time.Time {
	return t.ticker.C
}

func (t realTicker) Stop() {
	t.ticker.Stop()
}

func (e *Engine) Start(duration time.Duration, onTick func(time.Duration), onDone func()) error {
	if duration <= 0 {
		return errors.New("timer duration must be positive")
	}

	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return ErrAlreadyRunning
	}

	stopChannel := make(chan struct{})
	e.remaining = duration
	e.running = true
	e.paused = false
	e.stopChannel = stopChannel
	e.onTick = onTick
	e.onDone = onDone
	e.mu.Unlock()

	ticker := e.clock.NewTicker(e.tickInterval)
	go e.run(ticker, stopChannel)
	return nil
}

func (e *Engine) Pause() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running || e.paused {
		return false
	}

	e.paused = true
	return true
}

func (e *Engine) Resume() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running || !e.paused {
		return false
	}

	e.paused = false
	return true
}

func (e *Engine) Stop() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return false
	}

	close(e.stopChannel)
	e.stopChannel = nil
	e.remaining = 0
	e.running = false
	e.paused = false
	e.onTick = nil
	e.onDone = nil
	return true
}

func (e *Engine) Remaining() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.remaining
}

func (e *Engine) run(ticker Ticker, stopChannel chan struct{}) {
	defer ticker.Stop()

	for {
		select {
		case <-stopChannel:
			return
		case <-ticker.Chan():
			remaining, done, onTick, onDone := e.advance()
			if onTick != nil {
				onTick(remaining)
			}

			if done {
				if onDone != nil {
					onDone()
				}
				return
			}
		}
	}
}

func (e *Engine) advance() (time.Duration, bool, func(time.Duration), func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running || e.paused {
		return e.remaining, false, nil, nil
	}

	if e.remaining <= e.tickInterval {
		e.remaining = 0
		e.running = false
		e.paused = false
		e.stopChannel = nil
		return e.remaining, true, e.onTick, e.onDone
	}

	e.remaining -= e.tickInterval
	return e.remaining, false, e.onTick, nil
}
