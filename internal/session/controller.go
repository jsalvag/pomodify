package session

import (
	"errors"
	"sync"
	"time"

	"github.com/jsalvag/pomodify/internal/timer"
)

type State string

const (
	StateIdle       State = "idle"
	StateWorkActive State = "work_active"
	StateWorkPaused State = "work_paused"
	StateRestActive State = "rest_active"
	StateRestPaused State = "rest_paused"
	StateCompleted  State = "completed"
	StateCancelled  State = "cancelled"
	StateError      State = "error"
)

type Snapshot struct {
	State             State
	PhaseIndex        int
	TotalPhases       int
	TotalWorkBlocks   int
	PhaseKind         PhaseKind
	BlockNumber       int
	PhaseDuration     time.Duration
	Remaining         time.Duration
	SpotifyAutomation bool
}

type Controller struct {
	mu                sync.RWMutex
	timer             *timer.Engine
	plan              Plan
	phaseIndex        int
	state             State
	remaining         time.Duration
	spotifyAutomation bool
	nextSubscriberID  int
	subscribers       map[int]chan Snapshot
}

func NewController(engine *timer.Engine) *Controller {
	return &Controller{
		timer:       engine,
		state:       StateIdle,
		subscribers: map[int]chan Snapshot{},
	}
}

func (c *Controller) Start(plan Plan, spotifyAutomation bool) error {
	if len(plan.Phases) == 0 {
		return errors.New("session plan cannot be empty")
	}

	phase := plan.Phases[0]

	c.mu.Lock()
	if c.isInProgressLocked() {
		c.mu.Unlock()
		return errors.New("a session is already in progress")
	}

	c.plan = plan
	c.phaseIndex = 0
	c.remaining = phase.Duration
	c.spotifyAutomation = spotifyAutomation
	c.state = stateForPhase(phase.Kind, false)
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	if err := c.timer.Start(phase.Duration, c.handleTick, c.handlePhaseDone); err != nil {
		c.setErrorState()
		return err
	}

	c.broadcast(snapshot)
	return nil
}

func (c *Controller) Pause() error {
	c.mu.Lock()
	if !c.isActiveLocked() {
		c.mu.Unlock()
		return errors.New("the session is not running")
	}

	if !c.timer.Pause() {
		c.mu.Unlock()
		return errors.New("the timer could not be paused")
	}

	phase, ok := c.currentPhaseLocked()
	if !ok {
		c.mu.Unlock()
		return errors.New("there is no active phase")
	}

	c.state = stateForPhase(phase.Kind, true)
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	c.broadcast(snapshot)
	return nil
}

func (c *Controller) Resume() error {
	c.mu.Lock()
	if !c.isPausedLocked() {
		c.mu.Unlock()
		return errors.New("the session is not paused")
	}

	if !c.timer.Resume() {
		c.mu.Unlock()
		return errors.New("the timer could not be resumed")
	}

	phase, ok := c.currentPhaseLocked()
	if !ok {
		c.mu.Unlock()
		return errors.New("there is no paused phase")
	}

	c.state = stateForPhase(phase.Kind, false)
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	c.broadcast(snapshot)
	return nil
}

func (c *Controller) Skip() error {
	c.mu.Lock()
	if !c.isInProgressLocked() {
		c.mu.Unlock()
		return errors.New("there is no active session")
	}

	c.timer.Stop()
	if c.phaseIndex >= len(c.plan.Phases)-1 {
		c.remaining = 0
		c.state = StateCompleted
		snapshot := c.snapshotLocked()
		c.mu.Unlock()
		c.broadcast(snapshot)
		return nil
	}

	c.phaseIndex++
	phase := c.plan.Phases[c.phaseIndex]
	c.remaining = phase.Duration
	c.state = stateForPhase(phase.Kind, false)
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	if err := c.timer.Start(phase.Duration, c.handleTick, c.handlePhaseDone); err != nil {
		c.setErrorState()
		return err
	}

	c.broadcast(snapshot)
	return nil
}

func (c *Controller) Stop() error {
	c.mu.Lock()
	if !c.isInProgressLocked() {
		c.mu.Unlock()
		return errors.New("there is no active session")
	}

	c.timer.Stop()
	c.remaining = 0
	c.state = StateCancelled
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	c.broadcast(snapshot)
	return nil
}

func (c *Controller) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshotLocked()
}

func (c *Controller) Subscribe() <-chan Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	channel := make(chan Snapshot, 1)
	id := c.nextSubscriberID
	c.nextSubscriberID++
	c.subscribers[id] = channel

	channel <- c.snapshotLocked()
	return channel
}

func (c *Controller) handleTick(remaining time.Duration) {
	c.mu.Lock()
	c.remaining = remaining
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	c.broadcast(snapshot)
}

func (c *Controller) handlePhaseDone() {
	c.mu.Lock()
	if c.phaseIndex >= len(c.plan.Phases)-1 {
		c.remaining = 0
		c.state = StateCompleted
		snapshot := c.snapshotLocked()
		c.mu.Unlock()
		c.broadcast(snapshot)
		return
	}

	c.phaseIndex++
	phase := c.plan.Phases[c.phaseIndex]
	c.remaining = phase.Duration
	c.state = stateForPhase(phase.Kind, false)
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	if err := c.timer.Start(phase.Duration, c.handleTick, c.handlePhaseDone); err != nil {
		c.setErrorState()
		return
	}

	c.broadcast(snapshot)
}

func (c *Controller) setErrorState() {
	c.mu.Lock()
	c.remaining = 0
	c.state = StateError
	snapshot := c.snapshotLocked()
	c.mu.Unlock()

	c.broadcast(snapshot)
}

func (c *Controller) isInProgressLocked() bool {
	return c.isActiveLocked() || c.isPausedLocked()
}

func (c *Controller) isActiveLocked() bool {
	return c.state == StateWorkActive || c.state == StateRestActive
}

func (c *Controller) isPausedLocked() bool {
	return c.state == StateWorkPaused || c.state == StateRestPaused
}

func (c *Controller) currentPhaseLocked() (Phase, bool) {
	if c.phaseIndex < 0 || c.phaseIndex >= len(c.plan.Phases) {
		return Phase{}, false
	}

	return c.plan.Phases[c.phaseIndex], true
}

func (c *Controller) snapshotLocked() Snapshot {
	snapshot := Snapshot{
		State:             c.state,
		PhaseIndex:        c.phaseIndex,
		TotalPhases:       len(c.plan.Phases),
		TotalWorkBlocks:   c.plan.WorkBlocks,
		Remaining:         c.remaining,
		SpotifyAutomation: c.spotifyAutomation,
	}

	if phase, ok := c.currentPhaseLocked(); ok {
		snapshot.PhaseKind = phase.Kind
		snapshot.BlockNumber = phase.BlockNumber
		snapshot.PhaseDuration = phase.Duration
	}

	return snapshot
}

func (c *Controller) broadcast(snapshot Snapshot) {
	c.mu.RLock()
	subscribers := make([]chan Snapshot, 0, len(c.subscribers))
	for _, subscriber := range c.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	c.mu.RUnlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- snapshot:
		default:
			select {
			case <-subscriber:
			default:
			}
			subscriber <- snapshot
		}
	}
}

func stateForPhase(kind PhaseKind, paused bool) State {
	if kind == RestPhase {
		if paused {
			return StateRestPaused
		}

		return StateRestActive
	}

	if paused {
		return StateWorkPaused
	}

	return StateWorkActive
}
