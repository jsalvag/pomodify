package app

import (
	"fmt"
	"time"

	fyneapp "fyne.io/fyne/v2/app"

	"github.com/jsalvag/pomodify/internal/config"
	"github.com/jsalvag/pomodify/internal/notifications"
	"github.com/jsalvag/pomodify/internal/session"
	"github.com/jsalvag/pomodify/internal/timer"
	"github.com/jsalvag/pomodify/internal/ui"
)

const (
	appID    = "com.github.jsalvag.pomodify"
	appTitle = "pomodify"
)

func Run() error {
	application := fyneapp.NewWithID(appID)
	settingsStore := config.NewStore(application.Preferences())
	defaultSettings := settingsStore.Load()
	sessionController := session.NewController(timer.NewEngine())
	notifier := notifications.New(application)

	mainWindow := ui.NewMainWindow(application, ui.Callbacks{
		OnStart: func(settings config.Settings) error {
			plan, err := session.BuildPlan(
				time.Duration(settings.DefaultWorkMinutes*float64(time.Minute)),
				time.Duration(settings.DefaultRestMinutes*float64(time.Minute)),
				settings.DefaultWorkBlocks,
			)
			if err != nil {
				return err
			}

			return sessionController.Start(plan, settings.SpotifyAutomationEnabled)
		},
		OnPause:  sessionController.Pause,
		OnResume: sessionController.Resume,
		OnSkip:   sessionController.Skip,
		OnStop:   sessionController.Stop,
		OnSaveDefaults: func(settings config.Settings) error {
			settingsStore.Save(settings)
			return nil
		},
	}, defaultSettings)

	snapshotStream := sessionController.Subscribe()
	go func() {
		var previous session.Snapshot
		var hasPrevious bool

		for snapshot := range snapshotStream {
			mainWindow.ApplySnapshot(snapshot)
			if hasPrevious {
				publishNotifications(notifier, previous, snapshot)
			}

			previous = snapshot
			hasPrevious = true
		}
	}()

	mainWindow.Show()
	application.Run()
	return nil
}

func publishNotifications(notifier *notifications.Notifier, previous, current session.Snapshot) {
	if current.State == session.StateCompleted && previous.State != session.StateCompleted {
		notifier.Send("Session complete", "Your pomodoro session is finished.")
		return
	}

	if current.State == session.StateWorkActive && (previous.State != session.StateWorkActive || previous.PhaseIndex != current.PhaseIndex) {
		notifier.Send(
			"Work started",
			fmt.Sprintf("Block %d of %d is now running.", current.BlockNumber, current.TotalWorkBlocks),
		)
		return
	}

	if current.State == session.StateRestActive && (previous.State != session.StateRestActive || previous.PhaseIndex != current.PhaseIndex) {
		notifier.Send(
			"Rest started",
			fmt.Sprintf("Take a break after block %d.", current.BlockNumber),
		)
	}
}
