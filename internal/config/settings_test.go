package config

import "testing"

type testPreferences struct {
	boolValues   map[string]bool
	intValues    map[string]int
	stringValues map[string]string
}

func newTestPreferences() *testPreferences {
	return &testPreferences{
		boolValues:   map[string]bool{},
		intValues:    map[string]int{},
		stringValues: map[string]string{},
	}
}

func (p *testPreferences) BoolWithFallback(key string, fallback bool) bool {
	value, ok := p.boolValues[key]
	if !ok {
		return fallback
	}

	return value
}

func (p *testPreferences) IntWithFallback(key string, fallback int) int {
	value, ok := p.intValues[key]
	if !ok {
		return fallback
	}

	return value
}

func (p *testPreferences) SetBool(key string, value bool) {
	p.boolValues[key] = value
}

func (p *testPreferences) StringWithFallback(key string, fallback string) string {
	value, ok := p.stringValues[key]
	if !ok {
		return fallback
	}

	return value
}

func (p *testPreferences) SetInt(key string, value int) {
	p.intValues[key] = value
}

func (p *testPreferences) SetString(key string, value string) {
	p.stringValues[key] = value
}

func TestStoreLoadDefaults(t *testing.T) {
	store := NewStore(newTestPreferences())
	settings := store.Load()

	defaults := DefaultSettings()
	if settings != defaults {
		t.Fatalf("expected defaults %+v, got %+v", defaults, settings)
	}
}

func TestStoreSaveNormalizesValues(t *testing.T) {
	prefs := newTestPreferences()
	store := NewStore(prefs)

	store.Save(Settings{
		DefaultWorkMinutes:       0,
		DefaultRestMinutes:       -3,
		DefaultWorkBlocks:        0,
		SpotifyAutomationEnabled: true,
	})

	settings := store.Load()
	if settings.DefaultWorkMinutes != 25 {
		t.Fatalf("expected normalized work minutes, got %v", settings.DefaultWorkMinutes)
	}

	if settings.DefaultRestMinutes != 5 {
		t.Fatalf("expected normalized rest minutes, got %v", settings.DefaultRestMinutes)
	}

	if settings.DefaultWorkBlocks != 4 {
		t.Fatalf("expected normalized work blocks, got %d", settings.DefaultWorkBlocks)
	}

	if !settings.SpotifyAutomationEnabled {
		t.Fatal("expected spotify automation flag to persist")
	}
}

func TestStoreSavesDecimalMinutes(t *testing.T) {
	prefs := newTestPreferences()
	store := NewStore(prefs)

	store.Save(Settings{
		DefaultWorkMinutes:       2,
		DefaultRestMinutes:       0.5,
		DefaultWorkBlocks:        2,
		SpotifyAutomationEnabled: false,
	})

	settings := store.Load()
	if settings.DefaultRestMinutes != 0.5 {
		t.Fatalf("expected decimal rest minutes to round-trip, got %v", settings.DefaultRestMinutes)
	}
}

func TestStoreLoadsLegacyIntegerMinutes(t *testing.T) {
	prefs := newTestPreferences()
	prefs.intValues[defaultWorkMinutesKey] = 15
	prefs.intValues[defaultRestMinutesKey] = 3
	prefs.intValues[defaultWorkBlocksKey] = 2

	store := NewStore(prefs)
	settings := store.Load()

	if settings.DefaultWorkMinutes != 15 {
		t.Fatalf("expected legacy work minutes, got %v", settings.DefaultWorkMinutes)
	}

	if settings.DefaultRestMinutes != 3 {
		t.Fatalf("expected legacy rest minutes, got %v", settings.DefaultRestMinutes)
	}
}
