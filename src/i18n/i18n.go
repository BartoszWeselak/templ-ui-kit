package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

type Translator struct {
	translations map[string]map[string]string
	mu           sync.RWMutex
	fallback     string
}

var (
	instance *Translator
	once     sync.Once
)

// GetTranslator returns singleton instance
func GetTranslator() *Translator {
	once.Do(func() {
		instance = &Translator{
			translations: make(map[string]map[string]string),
			fallback:     "en",
		}
		instance.loadTranslations()
	})
	return instance
}

// loadTranslations loads all translation files
func (t *Translator) loadTranslations() {
	languages := []string{"en", "pl", "de", "es", "fr"}

	for _, lang := range languages {
		filename := fmt.Sprintf("locales/%s.json", lang)
		data, err := localesFS.ReadFile(filename)
		if err != nil {
			continue
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			continue
		}

		t.mu.Lock()
		t.translations[lang] = translations
		t.mu.Unlock()
	}
}

// T translates a key for given language
func (t *Translator) T(lang, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try requested language
	if translations, ok := t.translations[lang]; ok {
		if text, ok := translations[key]; ok {
			return text
		}
	}

	// Try fallback language
	if translations, ok := t.translations[t.fallback]; ok {
		if text, ok := translations[key]; ok {
			return text
		}
	}

	// Return key if no translation found
	return key
}

// Tf translates with formatting
func (t *Translator) Tf(lang, key string, args ...interface{}) string {
	text := t.T(lang, key)
	return fmt.Sprintf(text, args...)
}

// GetAvailableLanguages returns list of available languages
func (t *Translator) GetAvailableLanguages() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	langs := make([]string, 0, len(t.translations))
	for lang := range t.translations {
		langs = append(langs, lang)
	}
	return langs
}

// SetFallback sets fallback language
func (t *Translator) SetFallback(lang string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fallback = lang
}
