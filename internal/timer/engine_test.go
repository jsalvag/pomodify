package timer

import (
	"testing"
	"time"
)

type fakeClock struct {
	ticker *fakeTicker
}

type fakeTicker struct {
	channel chan time.Time
}

func (c *fakeClock) NewTicker(time.Duration) Ticker {
	c.ticker = &fakeTicker{channel: make(chan time.Time, 8)}
	return c.ticker
}

func (t *fakeTicker) Chan() <-chan time.Time {
	return t.channel
}

func (t *fakeTicker) Stop() {}

func TestEngineCompletesCountdown(t *testing.T) {
	clock := &fakeClock{}
	engine := NewEngineWithClock(clock, time.Second)
	ticks := make(chan time.Duration, 4)
	done := make(chan struct{}, 1)

	if err := engine.Start(3*time.Second, func(remaining time.Duration) {
		ticks <- remaining
	}, func() {
		done <- struct{}{}
	}); err != nil {
		t.Fatalf("expected timer to start, got error: %v", err)
	}

	clock.ticker.channel <- time.Now()
	if remaining := receiveDuration(t, ticks); remaining != 2*time.Second {
		t.Fatalf("expected 2 seconds remaining, got %s", remaining)
	}

	clock.ticker.channel <- time.Now()
	if remaining := receiveDuration(t, ticks); remaining != time.Second {
		t.Fatalf("expected 1 second remaining, got %s", remaining)
	}

	clock.ticker.channel <- time.Now()
	if remaining := receiveDuration(t, ticks); remaining != 0 {
		t.Fatalf("expected countdown to reach zero, got %s", remaining)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected timer to complete")
	}
}

func TestEnginePausesAndResumes(t *testing.T) {
	clock := &fakeClock{}
	engine := NewEngineWithClock(clock, time.Second)
	ticks := make(chan time.Duration, 2)

	if err := engine.Start(2*time.Second, func(remaining time.Duration) {
		ticks <- remaining
	}, nil); err != nil {
		t.Fatalf("expected timer to start, got error: %v", err)
	}

	if !engine.Pause() {
		t.Fatal("expected pause to succeed")
	}

	clock.ticker.channel <- time.Now()
	select {
	case remaining := <-ticks:
		t.Fatalf("did not expect tick while paused, got %s", remaining)
	case <-time.After(20 * time.Millisecond):
	}

	if engine.Remaining() != 2*time.Second {
		t.Fatalf("expected paused timer to keep remaining duration, got %s", engine.Remaining())
	}

	if !engine.Resume() {
		t.Fatal("expected resume to succeed")
	}

	clock.ticker.channel <- time.Now()
	if remaining := receiveDuration(t, ticks); remaining != time.Second {
		t.Fatalf("expected 1 second remaining after resume, got %s", remaining)
	}
}

func receiveDuration(t *testing.T, values <-chan time.Duration) time.Duration {
	t.Helper()

	select {
	case value := <-values:
		return value
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for duration")
		return 0
	}
}
