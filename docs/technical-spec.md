# Technical specification

## Technical objective

Implement `pomodify` as a native Go desktop application with a pure Go UI, a deterministic timer engine, and Spotify control through official APIs.

## Chosen direction

### UI toolkit

- Preferred toolkit: `Fyne`
- Reasoning:
  - pure Go application model
  - cross-platform desktop support
  - native window lifecycle handling
  - built-in notifications and preference storage
  - no HTML, CSS, or JavaScript runtime required

### Spotify integration model

- Use `Spotify Web API`
- Authenticate with `Authorization Code with PKCE`
- Use the official Spotify desktop client or another active Spotify Connect device as the playback target
- Do not use `Web Playback SDK`

### Secret handling

- Store refresh token in the OS keychain using `go-keyring`
- Store non-secret app settings in local application storage

## Constraints from official docs

- Desktop apps should use `OAuth PKCE` because a client secret cannot be safely stored.
- Redirect URIs for local desktop callbacks must use loopback IPs such as `http://127.0.0.1:<port>/callback`; `localhost` is not allowed.
- Spotify playback control requires a valid authorized user and the right scopes.
- The user must have Spotify Premium for playback control scenarios such as `start/resume playback`.
- Some Spotify devices can be marked as `restricted`, in which case playback commands will fail.
- Spotify does not provide a native Go desktop playback SDK for streaming audio into the app.

## Proposed module layout

```text
cmd/pomodify/
internal/app/
internal/config/
internal/timer/
internal/session/
internal/spotify/auth/
internal/spotify/client/
internal/spotify/device/
internal/spotify/catalog/
internal/notifications/
internal/storage/
internal/ui/
docs/
```

## Responsibilities by module

### `cmd/pomodify`

- Application entry point
- Dependency wiring
- App metadata and startup sequence

### `internal/app`

- App bootstrap
- Lifecycle orchestration
- Startup recovery and shutdown behavior

### `internal/config`

- Read and write non-secret settings
- Manage default timer and playback preferences

### `internal/timer`

- Phase timing
- Pause, resume, skip, stop
- Remaining time calculations

### `internal/session`

- Build work/rest phase plans
- Hold current session state
- Emit state transitions for the UI and Spotify integration

### `internal/spotify/auth`

- PKCE code verifier and challenge generation
- Loopback callback server
- Token exchange and refresh flow

### `internal/spotify/client`

- Authorized HTTP client
- Spotify endpoint wrappers
- Error normalization and retry behavior for safe cases

### `internal/spotify/device`

- Device listing
- Device selection and validation
- Active device refresh

### `internal/spotify/catalog`

- Search and browse behavior for tracks, albums, artists, and playlists
- Convert UI choices into Spotify URIs and playback payloads

### `internal/notifications`

- Desktop notifications for work, rest, and session completion

### `internal/storage`

- Keychain access via `go-keyring`
- Persist refresh token and future secure secrets

### `internal/ui`

- Fyne windows and widgets
- Session controls, Spotify connection screen, and playback source selection

## Session state model

Recommended state machine:

- idle
- starting
- work_active
- work_paused
- rest_active
- rest_paused
- completed
- cancelled
- error

Each transition should emit an event that the Spotify controller can subscribe to.

## Playback automation flow

### Work phase start

1. Validate Spotify session.
2. Resolve target device.
3. Apply playback mode configuration.
4. Start or resume playback.
5. Update UI with current playback state.

### Rest phase start

1. Send pause command.
2. Confirm or infer paused state.
3. Notify the user.

### Session completion

1. Pause playback.
2. Mark session completed.
3. Send completion notification.

## Configuration model

### Non-secret settings

- default work duration
- default rest duration
- default work block count
- spotify automation enabled by default
- default device id
- default playback source metadata
- default shuffle and repeat mode

Suggested storage:

- config file under the OS application config directory
- optional simple values mirrored through `Fyne` preferences where convenient

### Secret settings

- Spotify refresh token
- future token metadata if needed

Suggested storage:

- OS keychain via `go-keyring`

## Concurrency model

- Use a single session coordinator as the source of truth.
- Keep timer progression in a controlled goroutine.
- Use channels or a lightweight event bus for UI and Spotify side effects.
- Avoid scattered timer ownership across widgets.

## Testing plan

### Unit tests

- session plan generation
- timer state transitions
- PKCE generator
- Spotify request builders
- config serialization

### Integration tests

- loopback auth callback handling
- token refresh behavior
- mocked Spotify API interactions

### Manual desktop verification

- initial account connection
- device selection and failure cases
- start, pause, resume, and stop session
- work to rest playback transitions
- app restart with saved preferences

## CI plan

On `dev` pushes and pull requests targeting `main`, CI should run at minimum:

- `gofmt -l .`
- `go vet ./...`
- `go test ./...`

Optional later additions:

- race detector on supported runners
- static analysis
- packaged desktop build artifacts

## Implementation phases

### Phase 1

- repository docs
- UI shell
- timer engine
- local settings

### Phase 2

- Spotify OAuth PKCE
- secure token storage
- device listing and health checks

### Phase 3

- search and playback selection
- automated playback transitions
- richer error handling

### Phase 4

- CI hardening
- packaging
- release preparation
