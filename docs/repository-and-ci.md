# Repository and CI policy

## Branch model

### `main`

- release branch
- protected branch
- no direct commits
- changes arrive only through pull requests

### `dev`

- active development branch
- integration branch for ongoing work
- primary branch for CI validation during implementation

### Feature branches

- branch from `dev`
- merge back into `dev`
- use short descriptive names once implementation starts

## Pull request model

- ongoing implementation PRs should normally target `dev`
- release or promotion PRs should target `main`
- every merge into `main` should come from an explicit PR

## Initial repository bootstrap flow

1. Create repository.
2. Ensure `main` exists.
3. Create `dev` from `main`.
4. Commit initial project foundation on `dev`.
5. Open first PR from `dev` to `main`.

## Protection policy for `main`

Desired settings:

- require pull request before merge
- require at least one approval
- apply restrictions to administrators as well
- block force pushes
- block deletions
- require conversation resolution

Optional later hardening:

- require CI status checks
- require linear history
- require signed commits

## CI plan

### Trigger strategy

- push to `dev`
- pull requests into `main`

### Initial CI commands

- `gofmt -l .`
- `go vet ./...`
- `go test ./...`

### Later CI additions

- packaging jobs for desktop builds
- artifact upload
- release tagging automation

## GitHub platform note

Branch protection enforcement for private repositories depends on the GitHub plan. On a personal free plan, protection rules for private repositories may not be available. If strict protection is required immediately, either:

- make the repository public, or
- use a paid GitHub plan that supports private repository branch protection

## Merge strategy recommendation

- keep commit history readable
- avoid direct pushes to `main`
- keep release promotions visible through PR history
