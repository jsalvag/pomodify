# Sources

This project documentation is based on the following official documentation and primary references.

## Go

- Go package reference: `https://pkg.go.dev/std`
- Go `net/http`: `https://pkg.go.dev/net/http`
- Go `encoding/json`: `https://pkg.go.dev/encoding/json`
- Go `crypto/rand`: `https://pkg.go.dev/crypto/rand`
- Go `crypto/sha256`: `https://pkg.go.dev/crypto/sha256`
- Go `encoding/base64`: `https://pkg.go.dev/encoding/base64`
- Go `time`: `https://pkg.go.dev/time`
- Go `context`: `https://pkg.go.dev/context`

## Fyne

- Fyne docs home: `https://docs.fyne.io/`
- Fyne app API: `https://docs.fyne.io/api/v2/fyne/app/`
- Fyne window API: `https://docs.fyne.io/api/v2/fyne/window/`
- Fyne preferences guide: `https://docs.fyne.io/explore/preferences/`
- Fyne notification API: `https://docs.fyne.io/api/v2/fyne/notification/`

## Spotify

- Authorization overview: `https://developer.spotify.com/documentation/web-api/concepts/authorization`
- PKCE flow: `https://developer.spotify.com/documentation/web-api/tutorials/code-pkce-flow`
- Redirect URI rules: `https://developer.spotify.com/documentation/web-api/concepts/redirect_uri`
- Scopes reference: `https://developer.spotify.com/documentation/web-api/concepts/scopes`
- Search API: `https://developer.spotify.com/documentation/web-api/reference/search`
- Playback state API: `https://developer.spotify.com/documentation/web-api/reference/get-information-about-the-users-current-playback`
- Available devices API: `https://developer.spotify.com/documentation/web-api/reference/get-a-users-available-devices`
- Start or resume playback API: `https://developer.spotify.com/documentation/web-api/reference/start-a-users-playback`
- Shuffle API: `https://developer.spotify.com/documentation/web-api/reference/toggle-shuffle-for-users-playback`
- Repeat API: `https://developer.spotify.com/documentation/web-api/reference/set-repeat-mode-on-users-playback`
- Web Playback SDK overview: `https://developer.spotify.com/documentation/web-playback-sdk`
- Developer policy: `https://developer.spotify.com/policy`
- Compliance tips: `https://developer.spotify.com/compliance-tips`

## GitHub

- `gh repo create`: `https://cli.github.com/manual/gh_repo_create`
- Protected branches: `https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches`
- Branch protection REST API: `https://docs.github.com/en/rest/branches/branch-protection?apiVersion=2022-11-28#update-branch-protection`
- GitHub Actions for Go: `https://docs.github.com/en/actions/use-cases-and-examples/building-and-testing/building-and-testing-go`

## Supporting library

- `go-keyring`: `https://github.com/zalando/go-keyring`

## Key conclusions derived from the sources

- A pure Go desktop app can control Spotify playback, but not officially stream Spotify audio inside the app without a web-based playback stack.
- Desktop auth should use `OAuth PKCE` with a loopback callback.
- `127.0.0.1` is allowed for redirect URIs; `localhost` is not.
- Spotify device control depends on Spotify Connect device availability and permissions.
- The repository should use `main` for protected releases and `dev` for active development and CI.
