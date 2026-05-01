# pomodify

Desktop pomodoro timer built in Go with optional Spotify playback control during work sessions.

## Purpose

`pomodify` is a native desktop application for focused work sessions. It manages pomodoro cycles, desktop notifications, and optional Spotify playback automation without embedding a web frontend.

The product direction is intentionally strict:

- desktop-first application
- Go-only application code
- no HTML, CSS, or JavaScript UI layer
- Spotify used only as a playback source during work intervals
- no Spotify audio used as alarm or alert tones

## Product principles

- Use a native Go desktop UI, with `Fyne` as the current preferred toolkit.
- Control Spotify through the official `Spotify Web API` and `Spotify Connect`.
- Keep playback on the official Spotify desktop app or another active Spotify Connect device.
- Persist non-secret settings locally and store refresh tokens only in the OS keychain.
- Keep `main` as the protected release branch and `dev` as the development and CI branch.

## Why playback is controlled, not embedded

Spotify does not provide a native desktop SDK for Go that lets this app decode and play Spotify audio directly inside a pure Go desktop process. The official desktop-safe path is:

1. authenticate the user with `OAuth PKCE`
2. call Spotify Web API endpoints
3. control an active Spotify Connect device such as the official desktop client

This lets the app remain 100% Go while still supporting play, pause, resume, shuffle, repeat, content selection, and device targeting.

## Documentation map

- `docs/functional-spec.md` - functional scope, user journeys, and acceptance criteria
- `docs/technical-spec.md` - architecture, module design, storage, testing, and delivery plan
- `docs/spotify-integration.md` - Spotify auth, scopes, endpoints, device handling, and platform limits
- `docs/repository-and-ci.md` - repository policy, branch model, CI plan, and PR workflow
- `docs/sources.md` - official documentation used as the baseline for the project

## Initial architecture snapshot

- `Go` standard library for HTTP, timers, JSON, crypto, and local callback handling
- `Fyne` for the desktop UI, notifications, preferences, and app lifecycle hooks
- `go-keyring` for storing the Spotify refresh token in the OS credential store
- `Spotify Web API` for playback state, content search, device management, and playback commands
- `GitHub Actions` for CI on `dev` and PR validation before merging into `main`

## Repository workflow

- `main`: protected release branch, no direct commits
- `dev`: integration branch for active development and CI
- feature branches: branch from `dev`, merge back into `dev`
- release PRs: open from `dev` into `main`

## Current status

This repository currently contains the product and technical foundation needed to begin implementation.

The next implementation milestone is to scaffold the Go desktop app and wire the first Spotify authentication flow.
