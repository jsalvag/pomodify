package config

import "strconv"

const (
	defaultWorkMinutesKey       = "defaultWorkMinutes"
	defaultRestMinutesKey       = "defaultRestMinutes"
	defaultWorkBlocksKey        = "defaultWorkBlocks"
	spotifyAutomationEnabledKey = "spotifyAutomationEnabled"
)

type Preferences interface {
	BoolWithFallback(key string, fallback bool) bool
	IntWithFallback(key string, fallback int) int
	StringWithFallback(key string, fallback string) string
	SetBool(key string, value bool)
	SetInt(key string, value int)
	SetString(key string, value string)
}

type Settings struct {
	DefaultWorkMinutes       float64
	DefaultRestMinutes       float64
	DefaultWorkBlocks        int
	SpotifyAutomationEnabled bool
}

type Store struct {
	prefs Preferences
}

func DefaultSettings() Settings {
	return Settings{
		DefaultWorkMinutes:       25,
		DefaultRestMinutes:       5,
		DefaultWorkBlocks:        4,
		SpotifyAutomationEnabled: false,
	}
}

func Normalize(settings Settings) Settings {
	defaults := DefaultSettings()

	if settings.DefaultWorkMinutes <= 0 {
		settings.DefaultWorkMinutes = defaults.DefaultWorkMinutes
	}

	if settings.DefaultRestMinutes <= 0 {
		settings.DefaultRestMinutes = defaults.DefaultRestMinutes
	}

	if settings.DefaultWorkBlocks < 1 {
		settings.DefaultWorkBlocks = defaults.DefaultWorkBlocks
	}

	return settings
}

func NewStore(prefs Preferences) *Store {
	return &Store{prefs: prefs}
}

func (s *Store) Load() Settings {
	defaults := DefaultSettings()
	if s == nil || s.prefs == nil {
		return defaults
	}

	settings := Settings{
		DefaultWorkMinutes: loadMinutesPreference(s.prefs, defaultWorkMinutesKey, defaults.DefaultWorkMinutes),
		DefaultRestMinutes: loadMinutesPreference(s.prefs, defaultRestMinutesKey, defaults.DefaultRestMinutes),
		DefaultWorkBlocks:  s.prefs.IntWithFallback(defaultWorkBlocksKey, defaults.DefaultWorkBlocks),
		SpotifyAutomationEnabled: s.prefs.BoolWithFallback(
			spotifyAutomationEnabledKey,
			defaults.SpotifyAutomationEnabled,
		),
	}

	return Normalize(settings)
}

func (s *Store) Save(settings Settings) {
	if s == nil || s.prefs == nil {
		return
	}

	normalized := Normalize(settings)
	s.prefs.SetString(defaultWorkMinutesKey, strconv.FormatFloat(normalized.DefaultWorkMinutes, 'f', -1, 64))
	s.prefs.SetString(defaultRestMinutesKey, strconv.FormatFloat(normalized.DefaultRestMinutes, 'f', -1, 64))
	s.prefs.SetInt(defaultWorkBlocksKey, normalized.DefaultWorkBlocks)
	s.prefs.SetBool(spotifyAutomationEnabledKey, normalized.SpotifyAutomationEnabled)
}

func loadMinutesPreference(prefs Preferences, key string, fallback float64) float64 {
	value := prefs.StringWithFallback(key, "")
	if value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return parsed
		}
	}

	return float64(prefs.IntWithFallback(key, int(fallback)))
}
