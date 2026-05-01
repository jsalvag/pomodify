# Functional specification

## Goal

Build a desktop pomodoro application that lets the user plan work and rest intervals, optionally automate Spotify playback during work blocks, and keep the full user experience inside a native Go desktop app.

## Target user

- Individual knowledge worker using a desktop computer
- Uses Spotify Premium and the official Spotify client
- Wants a focused work timer with optional music automation
- Does not want a browser-based desktop app

## Product scope

### In scope

- Create and run pomodoro cycles with configurable work and rest durations
- Support multi-block sessions such as `3 x 45m work` with `2 x 15m rest`
- Pause Spotify automatically when a work block ends
- Resume or start Spotify automatically when a work block starts, if enabled
- Let the user choose a playback source from Spotify:
  - track
  - album
  - artist context
  - playlist
- Let the user configure playback behavior:
  - one track only
  - repeat track
  - repeat context
  - sequential playback
  - shuffle on or off
- Let the user choose the Spotify target device
- Persist user preferences between sessions
- Show desktop notifications for state changes

### Out of scope for v1

- Streaming Spotify audio directly inside the Go process
- HTML, CSS, or JavaScript-based UI
- Team sync, cloud sync, or multi-user collaboration
- Calendar integration
- Task management beyond the timer session itself
- Spotify alarm tones or alert tones
- Monetized streaming features or commercial Spotify playback features

## Core concepts

### Session

A session is a planned sequence of work and rest blocks.

Example:

- work 45m
- rest 15m
- work 45m
- rest 15m
- work 45m

The final rest block is omitted by default unless a future feature explicitly adds a post-session cooldown.

### Playback automation

Playback automation is optional and controlled by a user toggle.

When enabled:

- entering a work block starts or resumes Spotify playback
- entering a rest block pauses playback
- ending the final work block pauses playback and ends the session

When disabled:

- the timer behaves normally and does not control Spotify

## Main user journeys

### First-time setup

1. User opens the app.
2. User configures default work and rest durations.
3. User chooses whether Spotify integration is enabled.
4. If Spotify integration is enabled, user connects their Spotify account.
5. User selects a default target device and optional default playback source.

### Start a focused session

1. User chooses the number of work blocks.
2. User adjusts work and rest durations if needed.
3. User enables or disables Spotify playback for this run.
4. User chooses the content source and playback mode.
5. User starts the session.
6. App switches between work and rest states until the cycle is complete.

### Recover from interrupted playback

1. App detects no active or controllable Spotify device.
2. App shows a clear message and keeps the timer running.
3. User can open Spotify, select a device, and retry playback control.

## Functional requirements

### Timer and scheduling

- User can configure default work duration in minutes.
- User can configure default rest duration in minutes.
- User can configure number of work blocks.
- The app must derive rest blocks as `workBlocks - 1` by default.
- User can start, pause, resume, skip, and stop a session.
- The UI must always show the current phase, remaining time, and current block index.

### Spotify account and device handling

- User can connect and disconnect a Spotify account.
- App must validate that the account is suitable for playback control.
- App must list available Spotify Connect devices.
- User can set a preferred target device.
- App must detect when a previously stored device is unavailable or restricted.

### Spotify content selection

- User can search Spotify content by keyword.
- User can browse and select playlists.
- User can browse and select albums.
- User can browse and select artists as playback context.
- User can select a single track.
- User can store a default playback choice.

### Playback modes

- User can enable sequential context playback.
- User can enable shuffle.
- User can set repeat mode to:
  - off
  - track
  - context
- User can start from a selected context or list of track URIs.

### Notifications and UX

- App must notify the user when work starts.
- App must notify the user when rest starts.
- App must notify the user when the full session ends.
- Notifications must not depend on Spotify audio.

## Non-functional requirements

- Must be a native desktop app written in Go.
- Must not require an embedded web UI for normal operation.
- Must keep secrets out of plaintext config files.
- Must degrade gracefully when Spotify is unavailable.
- Must make Spotify automation optional at runtime.

## Acceptance criteria for v1

- User can run a `3 x 45 / 2 x 15` session.
- User can enable Spotify playback automation for that session.
- User can authenticate with Spotify using a desktop-safe flow.
- User can select a playlist, album, artist context, or single track.
- Playback starts on work blocks and pauses on rest blocks.
- Session state survives normal app restarts when feasible.
- The app provides enough feedback to understand timer and Spotify state.
