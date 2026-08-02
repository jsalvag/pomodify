package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/jsalvag/pomodify/internal/config"
	"github.com/jsalvag/pomodify/internal/session"
)

var (
	backgroundColor = color.NRGBA{R: 0x13, G: 0x14, B: 0x1a, A: 0xff}
	heroFillColor   = color.NRGBA{R: 0x1d, G: 0x1f, B: 0x28, A: 0xff}
	ringColor       = color.NRGBA{R: 0xff, G: 0x72, B: 0x5e, A: 0xff}
	workColor       = color.NRGBA{R: 0xff, G: 0x72, B: 0x5e, A: 0xff}
	restColor       = color.NRGBA{R: 0x53, G: 0xc4, B: 0xb8, A: 0xff}
	idleColor       = color.NRGBA{R: 0x72, G: 0x7a, B: 0x90, A: 0xff}
	errorColor      = color.NRGBA{R: 0xe5, G: 0x4b, B: 0x4b, A: 0xff}
)

type Callbacks struct {
	OnStart        func(settings config.Settings) error
	OnPause        func() error
	OnResume       func() error
	OnSkip         func() error
	OnStop         func() error
	OnSaveDefaults func(settings config.Settings) error
}

type MainWindow struct {
	window               fyne.Window
	callbacks            Callbacks
	snapshot             session.Snapshot
	workMinutesEntry     *widget.Entry
	restMinutesEntry     *widget.Entry
	workBlocksEntry      *widget.Entry
	automationCheck      *widget.Check
	statusValue          *widget.Label
	startButton          *widget.Button
	pauseResumeButton    *widget.Button
	skipButton           *widget.Button
	stopButton           *widget.Button
	saveDefaultsButton   *widget.Button
	timerValue           *canvas.Text
	timerCaption         *canvas.Text
	phaseHeadline        *canvas.Text
	progressBar          *widget.ProgressBar
	stateBadgeText       *canvas.Text
	stateBadgeBackground *canvas.Rectangle
	blockSummaryValue    *widget.Label
	spotifySummaryValue  *widget.Label
}

func NewMainWindow(application fyne.App, callbacks Callbacks, settings config.Settings) *MainWindow {
	window := application.NewWindow("pomodify")
	mainWindow := &MainWindow{
		window:               window,
		callbacks:            callbacks,
		snapshot:             session.Snapshot{State: session.StateIdle},
		workMinutesEntry:     widget.NewEntry(),
		restMinutesEntry:     widget.NewEntry(),
		workBlocksEntry:      widget.NewEntry(),
		automationCheck:      widget.NewCheck("Enable later", nil),
		statusValue:          widget.NewLabel("Ready to start a focus session."),
		timerValue:           canvas.NewText("25:00", color.White),
		timerCaption:         canvas.NewText("READY", color.NRGBA{R: 0xff, G: 0xbd, B: 0xb4, A: 0xff}),
		phaseHeadline:        canvas.NewText("Ready", color.White),
		progressBar:          widget.NewProgressBar(),
		stateBadgeText:       canvas.NewText("IDLE", color.White),
		stateBadgeBackground: canvas.NewRectangle(idleColor),
		blockSummaryValue:    widget.NewLabel("Not started"),
		spotifySummaryValue:  widget.NewLabel("Disabled"),
	}

	mainWindow.timerValue.TextSize = 68
	mainWindow.timerValue.Alignment = fyne.TextAlignCenter
	mainWindow.timerValue.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	mainWindow.timerCaption.TextSize = 14
	mainWindow.timerCaption.Alignment = fyne.TextAlignCenter
	mainWindow.timerCaption.TextStyle = fyne.TextStyle{Bold: true}
	mainWindow.phaseHeadline.TextSize = 18
	mainWindow.phaseHeadline.Alignment = fyne.TextAlignCenter
	mainWindow.phaseHeadline.TextStyle = fyne.TextStyle{Bold: true}
	mainWindow.stateBadgeText.TextSize = 13
	mainWindow.stateBadgeText.Alignment = fyne.TextAlignCenter
	mainWindow.stateBadgeText.TextStyle = fyne.TextStyle{Bold: true}
	mainWindow.statusValue.Wrapping = fyne.TextWrapWord
	mainWindow.workMinutesEntry.SetPlaceHolder("25")
	mainWindow.restMinutesEntry.SetPlaceHolder("5")
	mainWindow.workBlocksEntry.SetPlaceHolder("4")

	mainWindow.applySettings(settings)
	mainWindow.bindActions()
	mainWindow.window.SetContent(mainWindow.buildContent())
	mainWindow.window.Resize(fyne.NewSize(1080, 760))
	mainWindow.refreshForSnapshot(mainWindow.snapshot)
	return mainWindow
}

func (w *MainWindow) Show() {
	w.window.Show()
}

func (w *MainWindow) ApplySnapshot(snapshot session.Snapshot) {
	fyne.Do(func() {
		w.snapshot = snapshot
		w.refreshForSnapshot(snapshot)
	})
}

func (w *MainWindow) applySettings(settings config.Settings) {
	normalized := config.Normalize(settings)
	w.workMinutesEntry.SetText(formatMinutesInput(normalized.DefaultWorkMinutes))
	w.restMinutesEntry.SetText(formatMinutesInput(normalized.DefaultRestMinutes))
	w.workBlocksEntry.SetText(strconv.Itoa(normalized.DefaultWorkBlocks))
	w.automationCheck.SetChecked(normalized.SpotifyAutomationEnabled)
}

func (w *MainWindow) bindActions() {
	w.startButton = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
		settings, err := w.readSettings()
		if err != nil {
			w.setStatus(err.Error())
			return
		}

		if w.callbacks.OnStart == nil {
			w.setStatus("Start callback is not configured.")
			return
		}

		if err := w.callbacks.OnStart(settings); err != nil {
			w.setStatus(err.Error())
			return
		}

		message := fmt.Sprintf("Focus session started: %s / %s for %d block(s).", formatMinutesInput(settings.DefaultWorkMinutes), formatMinutesInput(settings.DefaultRestMinutes), settings.DefaultWorkBlocks)
		if settings.SpotifyAutomationEnabled {
			message += " Spotify wiring is next, but this run already stores the intent."
		}

		w.setStatus(message)
	})
	w.startButton.Importance = widget.HighImportance

	w.pauseResumeButton = widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
		var err error
		if w.snapshot.State == session.StateWorkPaused || w.snapshot.State == session.StateRestPaused {
			if w.callbacks.OnResume == nil {
				w.setStatus("Resume callback is not configured.")
				return
			}

			err = w.callbacks.OnResume()
			if err == nil {
				w.setStatus("Session resumed.")
			}
		} else {
			if w.callbacks.OnPause == nil {
				w.setStatus("Pause callback is not configured.")
				return
			}

			err = w.callbacks.OnPause()
			if err == nil {
				w.setStatus("Session paused.")
			}
		}

		if err != nil {
			w.setStatus(err.Error())
		}
	})
	w.pauseResumeButton.Importance = widget.WarningImportance

	w.skipButton = widget.NewButtonWithIcon("Next", theme.MediaSkipNextIcon(), func() {
		if w.callbacks.OnSkip == nil {
			w.setStatus("Skip callback is not configured.")
			return
		}

		if err := w.callbacks.OnSkip(); err != nil {
			w.setStatus(err.Error())
			return
		}

		w.setStatus("Moved straight into the next phase.")
	})

	w.stopButton = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
		if w.callbacks.OnStop == nil {
			w.setStatus("Stop callback is not configured.")
			return
		}

		if err := w.callbacks.OnStop(); err != nil {
			w.setStatus(err.Error())
			return
		}

		w.setStatus("Session cancelled.")
	})
	w.stopButton.Importance = widget.DangerImportance

	w.saveDefaultsButton = widget.NewButtonWithIcon("Save Defaults", theme.DocumentSaveIcon(), func() {
		settings, err := w.readSettings()
		if err != nil {
			w.setStatus(err.Error())
			return
		}

		if w.callbacks.OnSaveDefaults == nil {
			w.setStatus("Save callback is not configured.")
			return
		}

		if err := w.callbacks.OnSaveDefaults(settings); err != nil {
			w.setStatus(err.Error())
			return
		}

		w.setStatus("Default rhythm saved locally.")
	})

	w.pauseResumeButton.Disable()
	w.skipButton.Disable()
	w.stopButton.Disable()
}

func (w *MainWindow) buildContent() fyne.CanvasObject {
	background := canvas.NewRectangle(backgroundColor)
	header := w.buildHeader()
	hero := w.buildHeroCard()
	details := w.buildDetails()

	body := container.NewVBox(
		header,
		hero,
		details,
	)

	scroller := container.NewVScroll(container.NewPadded(body))
	return container.NewStack(background, scroller)
}

func (w *MainWindow) buildHeader() fyne.CanvasObject {
	title := canvas.NewText("pomodify", color.White)
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewVBox(title)
}

func (w *MainWindow) buildHeroCard() fyne.CanvasObject {
	timerDial := w.buildTimerDial()
	progressColumn := container.NewVBox(
		w.buildStateBadge(),
		w.phaseHeadline,
		w.progressBar,
		container.NewGridWithColumns(2,
			w.newSummaryTile("Block", w.blockSummaryValue),
			w.newSummaryTile("Spotify", w.spotifySummaryValue),
		),
		w.buildControlRow(),
	)

	content := container.NewGridWithColumns(2, timerDial, progressColumn)
	return widget.NewCard("Focus Timer", "", content)
}

func (w *MainWindow) buildTimerDial() fyne.CanvasObject {
	outerGlow := canvas.NewCircle(color.NRGBA{R: 0x2c, G: 0x20, B: 0x24, A: 0xff})
	outerRing := canvas.NewCircle(color.Transparent)
	outerRing.StrokeColor = ringColor
	outerRing.StrokeWidth = 10
	innerDisc := canvas.NewCircle(heroFillColor)
	innerDisc.StrokeColor = color.NRGBA{R: 0x2a, G: 0x2d, B: 0x38, A: 0xff}
	innerDisc.StrokeWidth = 2

	centerContent := container.NewCenter(container.NewVBox(
		w.timerCaption,
		w.timerValue,
	))

	dial := container.NewStack(outerGlow, outerRing, innerDisc, centerContent)
	wrapped := container.New(layout.NewCenterLayout(), container.NewGridWrap(fyne.NewSize(340, 340), dial))
	return container.NewVBox(wrapped)
}

func (w *MainWindow) buildControlRow() fyne.CanvasObject {
	return container.NewHBox(
		layout.NewSpacer(),
		container.NewGridWrap(fyne.NewSize(150, 40), w.startButton),
		container.NewGridWrap(fyne.NewSize(110, 40), w.pauseResumeButton),
		container.NewGridWrap(fyne.NewSize(110, 40), w.skipButton),
		container.NewGridWrap(fyne.NewSize(110, 40), w.stopButton),
		layout.NewSpacer(),
	)
}

func (w *MainWindow) buildStateBadge() fyne.CanvasObject {
	w.stateBadgeBackground.CornerRadius = 14
	badge := container.NewStack(
		w.stateBadgeBackground,
		container.NewPadded(container.NewCenter(w.stateBadgeText)),
	)
	return container.NewGridWrap(fyne.NewSize(180, 42), badge)
}

func (w *MainWindow) buildSetupCard() fyne.CanvasObject {
	setupForm := &widget.Form{}
	setupForm.Append("Focus", w.workMinutesEntry)
	setupForm.Append("Break", w.restMinutesEntry)
	setupForm.Append("Rounds", w.workBlocksEntry)
	setupForm.Append("Spotify", w.automationCheck)

	return widget.NewCard(
		"Setup",
		"",
		container.NewVBox(setupForm, w.saveDefaultsButton),
	)
}

func (w *MainWindow) buildQuickStartCard() fyne.CanvasObject {
	classicButton := widget.NewButton("Classic 25/5 x4", func() {
		w.applyPreset(config.Settings{DefaultWorkMinutes: 25, DefaultRestMinutes: 5, DefaultWorkBlocks: 4, SpotifyAutomationEnabled: w.automationCheck.Checked}, "Classic 25/5 x4 loaded.")
	})
	deepButton := widget.NewButton("Deep 50/10 x3", func() {
		w.applyPreset(config.Settings{DefaultWorkMinutes: 50, DefaultRestMinutes: 10, DefaultWorkBlocks: 3, SpotifyAutomationEnabled: w.automationCheck.Checked}, "Deep focus 50/10 x3 loaded.")
	})
	testButton := widget.NewButton("Quick test 2/0.5 x2", func() {
		w.applyPreset(config.Settings{DefaultWorkMinutes: 2, DefaultRestMinutes: 0.5, DefaultWorkBlocks: 2, SpotifyAutomationEnabled: w.automationCheck.Checked}, "Quick test preset loaded for short verification runs.")
	})

	classicButton.Importance = widget.HighImportance
	testButton.Importance = widget.WarningImportance

	return widget.NewCard(
		"Presets",
		"",
		container.NewVBox(container.NewGridWithColumns(3, classicButton, deepButton, testButton)),
	)
}

func (w *MainWindow) buildFooterCard() fyne.CanvasObject {
	return widget.NewCard("Console", "", w.statusValue)
}

func (w *MainWindow) buildDetails() fyne.CanvasObject {
	accordion := widget.NewAccordion(
		widget.NewAccordionItem("Setup", w.buildSetupCard()),
		widget.NewAccordionItem("Presets", w.buildQuickStartCard()),
		widget.NewAccordionItem("Console", w.buildFooterCard()),
	)
	accordion.Open(0)
	return accordion
}

func (w *MainWindow) newSummaryTile(title string, value *widget.Label) fyne.CanvasObject {
	label := canvas.NewText(strings.ToUpper(title), color.NRGBA{R: 0xb7, G: 0xbd, B: 0xca, A: 0xff})
	label.TextSize = 11
	label.TextStyle = fyne.TextStyle{Bold: true}
	value.Wrapping = fyne.TextWrapWord
	return widget.NewCard("", "", container.NewVBox(label, value))
}

func (w *MainWindow) applyPreset(settings config.Settings, message string) {
	w.workMinutesEntry.SetText(formatMinutesInput(settings.DefaultWorkMinutes))
	w.restMinutesEntry.SetText(formatMinutesInput(settings.DefaultRestMinutes))
	w.workBlocksEntry.SetText(strconv.Itoa(settings.DefaultWorkBlocks))
	w.setStatus(message)
}

func (w *MainWindow) refreshForSnapshot(snapshot session.Snapshot) {
	w.timerValue.Text = formatDuration(snapshot.Remaining)
	w.timerValue.Color = stateAccentColor(snapshot)
	w.timerValue.Refresh()

	w.timerCaption.Text = strings.ToUpper(formatState(snapshot.State))
	w.timerCaption.Color = tintColor(stateAccentColor(snapshot), 0.75)
	w.timerCaption.Refresh()

	w.phaseHeadline.Text = formatHeroHeadline(snapshot)
	w.phaseHeadline.Refresh()
	w.blockSummaryValue.SetText(formatBlock(snapshot))
	w.spotifySummaryValue.SetText(formatAutomation(snapshot.SpotifyAutomation))

	w.stateBadgeText.Text = strings.ToUpper(formatBadgeText(snapshot))
	w.stateBadgeText.Refresh()
	w.stateBadgeBackground.FillColor = stateAccentColor(snapshot)
	w.stateBadgeBackground.Refresh()

	progress := phaseProgress(snapshot)
	w.progressBar.SetValue(progress)

	inProgress := snapshot.State == session.StateWorkActive ||
		snapshot.State == session.StateRestActive ||
		snapshot.State == session.StateWorkPaused ||
		snapshot.State == session.StateRestPaused
	paused := snapshot.State == session.StateWorkPaused || snapshot.State == session.StateRestPaused

	if inProgress {
		w.startButton.Disable()
		w.pauseResumeButton.Enable()
		w.skipButton.Enable()
		w.stopButton.Enable()
	} else {
		w.startButton.Enable()
		w.pauseResumeButton.Disable()
		w.skipButton.Disable()
		w.stopButton.Disable()
	}

	if paused {
		w.pauseResumeButton.SetText("Resume")
		w.pauseResumeButton.SetIcon(theme.MediaPlayIcon())
	} else {
		w.pauseResumeButton.SetText("Pause")
		w.pauseResumeButton.SetIcon(theme.MediaPauseIcon())
	}

	if snapshot.State == session.StateCompleted {
		w.setStatus("Session completed. Nice work.")
	}

	if snapshot.State == session.StateError {
		w.setStatus("The session entered an error state.")
	}

	if snapshot.State == session.StateCancelled {
		w.setStatus("Session cancelled.")
	}
}

func (w *MainWindow) readSettings() (config.Settings, error) {
	workMinutes, err := parsePositiveMinutes(w.workMinutesEntry.Text, "work minutes")
	if err != nil {
		return config.Settings{}, err
	}

	restMinutes, err := parsePositiveMinutes(w.restMinutesEntry.Text, "rest minutes")
	if err != nil {
		return config.Settings{}, err
	}

	workBlocks, err := parsePositiveInt(w.workBlocksEntry.Text, "work blocks")
	if err != nil {
		return config.Settings{}, err
	}

	return config.Settings{
		DefaultWorkMinutes:       workMinutes,
		DefaultRestMinutes:       restMinutes,
		DefaultWorkBlocks:        workBlocks,
		SpotifyAutomationEnabled: w.automationCheck.Checked,
	}, nil
}

func (w *MainWindow) setStatus(message string) {
	w.statusValue.SetText(message)
}

func parsePositiveInt(raw, fieldName string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number", fieldName)
	}

	if value < 1 {
		return 0, fmt.Errorf("%s must be greater than zero", fieldName)
	}

	return value, nil
}

func parsePositiveMinutes(raw, fieldName string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", fieldName)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", fieldName)
	}

	return value, nil
}

func formatMinutesInput(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatState(state session.State) string {
	switch state {
	case session.StateWorkActive:
		return "Focus running"
	case session.StateWorkPaused:
		return "Focus paused"
	case session.StateRestActive:
		return "Break running"
	case session.StateRestPaused:
		return "Break paused"
	case session.StateCompleted:
		return "Completed"
	case session.StateCancelled:
		return "Cancelled"
	case session.StateError:
		return "Error"
	default:
		return "Idle"
	}
}

func formatBadgeText(snapshot session.Snapshot) string {
	switch snapshot.State {
	case session.StateWorkActive:
		return "Focus"
	case session.StateWorkPaused:
		return "Paused"
	case session.StateRestActive:
		return "Break"
	case session.StateRestPaused:
		return "Break paused"
	case session.StateCompleted:
		return "Done"
	case session.StateCancelled:
		return "Stopped"
	case session.StateError:
		return "Error"
	default:
		return "Ready"
	}
}

func formatHeroHeadline(snapshot session.Snapshot) string {
	if snapshot.BlockNumber == 0 {
		return "Ready"
	}

	if snapshot.PhaseKind == session.RestPhase {
		return fmt.Sprintf("Break - %d/%d", snapshot.BlockNumber, snapshot.TotalWorkBlocks)
	}

	return fmt.Sprintf("Focus - %d/%d", snapshot.BlockNumber, snapshot.TotalWorkBlocks)
}

func formatBlock(snapshot session.Snapshot) string {
	if snapshot.TotalWorkBlocks == 0 || snapshot.BlockNumber == 0 {
		return "Not started"
	}

	return fmt.Sprintf("%d of %d", snapshot.BlockNumber, snapshot.TotalWorkBlocks)
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}

	totalSeconds := int(duration.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func formatAutomation(enabled bool) string {
	if enabled {
		return "Prepared"
	}

	return "Disabled"
}

func phaseProgress(snapshot session.Snapshot) float64 {
	if snapshot.PhaseDuration <= 0 {
		return 0
	}

	progress := float64(snapshot.PhaseDuration-snapshot.Remaining) / float64(snapshot.PhaseDuration)
	if progress < 0 {
		return 0
	}

	if progress > 1 {
		return 1
	}

	return progress
}

func stateAccentColor(snapshot session.Snapshot) color.NRGBA {
	switch snapshot.State {
	case session.StateWorkActive, session.StateWorkPaused:
		return workColor
	case session.StateRestActive, session.StateRestPaused:
		return restColor
	case session.StateCompleted:
		return color.NRGBA{R: 0x78, G: 0xc6, B: 0x71, A: 0xff}
	case session.StateCancelled:
		return idleColor
	case session.StateError:
		return errorColor
	default:
		return idleColor
	}
}

func tintColor(base color.NRGBA, alpha float64) color.NRGBA {
	if alpha < 0 {
		alpha = 0
	}

	if alpha > 1 {
		alpha = 1
	}

	base.A = uint8(float64(base.A) * alpha)
	return base
}
