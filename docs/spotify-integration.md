# Spotify integration

## Integration objective

Automate Spotify playback during work intervals while keeping the application itself fully native in Go.

## Official integration model

The selected official model is:

- authenticate the user with Spotify using `Authorization Code with PKCE`
- call `Spotify Web API` endpoints from the Go app
- target the official Spotify desktop app or another active `Spotify Connect` device

This is the correct desktop strategy because Spotify does not expose a native Go desktop playback SDK.

## What this app will do

- connect a Spotify account
- list Spotify Connect devices
- select a target device
- search tracks, albums, artists, and playlists
- start or resume playback on work phases
- pause playback on rest phases
- configure shuffle and repeat mode
- inspect the current playback state for synchronization

## What this app will not do

- stream Spotify audio directly inside the Go process
- use Spotify content as an alarm tone
- alter or remix Spotify audio
- use browser UI code for playback

## Authentication model

### Flow

Use `Authorization Code with PKCE`.

High-level sequence:

1. Generate a code verifier with secure random bytes.
2. Derive the SHA-256 based code challenge.
3. Open the Spotify authorization URL in the system browser.
4. Listen on a loopback callback such as `http://127.0.0.1:<port>/callback`.
5. Exchange the authorization code for access and refresh tokens.
6. Store the refresh token in the OS keychain.
7. Refresh access tokens when they expire.

### Important auth constraints

- Desktop apps should use PKCE rather than storing a client secret.
- Loopback redirects must use `127.0.0.1` or `[::1]`.
- `localhost` is not allowed.

## Required scopes

Minimum practical scope set for the planned UX:

- `user-read-private`
- `user-read-playback-state`
- `user-modify-playback-state`
- `playlist-read-private`
- `playlist-read-collaborative`
- `user-library-read`

Possible future scopes only if needed later:

- `user-read-currently-playing`
- `user-top-read`

## Premium requirement

Playback control endpoints such as `Start/Resume Playback` are documented for Spotify Premium users.

The app should validate this early and show a clear message if the account is not eligible.

## Core endpoints

### Account and playback state

- `GET /me`
- `GET /me/player`
- `GET /me/player/devices`

### Playback control

- `PUT /me/player/play`
- `PUT /me/player/pause`
- `PUT /me/player/shuffle`
- `PUT /me/player/repeat`
- `PUT /me/player/volume`
- `POST /me/player/next`
- `POST /me/player/previous`

### Content discovery

- `GET /search`
- `GET /me/playlists`
- album, artist, playlist, and track detail endpoints as needed

## Device strategy

### Device rules

- Device ids are not guaranteed to be stable forever.
- The app should refresh devices regularly instead of trusting cached ids blindly.
- Devices can appear as `restricted`; restricted devices must not be targeted for control.

### Recommended behavior

1. Load the preferred device id from settings.
2. Refresh `/me/player/devices`.
3. If the preferred device is available and not restricted, use it.
4. Otherwise, prompt the user to choose an available device.

## Playback strategy by content type

### Track

Use `PUT /me/player/play` with `uris` containing a single track or ordered track list.

### Album or playlist

Use `context_uri` with optional `offset` and `position_ms`.

### Artist context

Spotify documents `context_uri` support for artist contexts in the playback endpoint.

## Mapping app settings to Spotify behavior

### Sequential playback

- `shuffle = false`
- `repeat = off` or `repeat = context`

### Single track in loop

- `uris` with one track
- `repeat = track`

### Context loop

- `context_uri`
- `repeat = context`

### Random playback

- `shuffle = true`

## Failure modes to handle

- user not logged in anymore
- expired access token
- refresh token revoked
- no active or available device
- selected device becomes restricted
- selected content unavailable in the user market
- rate limiting with `429`

## Policy constraints that affect the product

- playback integrations may not be commercial streaming products without Spotify approval
- Spotify content must remain in original form
- Spotify content must not be broadcast as a non-interactive service
- Spotify content must not be used as alarm tones

## Security recommendations

- never store the refresh token in plaintext config
- keep token refresh centralized in one module
- request the minimum scopes needed for the shipped UX
- separate secret storage from normal settings storage
