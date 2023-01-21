// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package i18n

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	defaultLang    = language.AmericanEnglish
	defaultDict    = enUS
	defaultPrinter = message.NewPrinter(defaultLang)
)

var locales = map[language.Tag]map[string]*Translation{
	defaultLang:                  defaultDict,
	language.BrazilianPortuguese: ptBR,
	language.SimplifiedChinese:   zhCN,
	language.Polish:              pl,
	language.German:              de,
	language.Arabic:              ar,
}

func init() {
	for lang, translations := range locales {
		for key, tran := range translations {
			if err := message.SetString(lang, key, tran.Value); err != nil {
				panic(fmt.Sprintf("message.SetString(%s): %v", lang, err))
			}
		}
	}
}

// AvailableLangs returns a list of languages with available translation.
func AvailableLangs() []language.Tag {
	var langs []language.Tag
	for lang := range locales {
		langs = append(langs, lang)
	}
	return langs
}

type Translation struct {
	Value string
	// stale is used to indicate that a translation has changed, is only
	// partially translated, or just needs review, and should be updated. This
	// is useful when it's better than falling back to english, but it allows
	// these translations to be identified programmatically.
	Stale bool
	// Docs contains information concerning the translated value.
	Docs string
}

type Translator struct {
	pkg     string
	lang    language.Tag
	printer *message.Printer

	dict  map[string]*Translation
	dicts map[language.Tag]map[string]*Translation
}

// NewPackageTranslator returns a new translator. It will default to en-US is
// the provided lang does not have translation available.
func NewPackageTranslator(pkg string, lang language.Tag) *Translator {
	if pkg == "" {
		pkg = "i18n"
	}

	dict, ok := locales[lang]
	if !ok {
		lang = defaultLang
		dict = defaultDict
	}

	t := &Translator{
		pkg:     pkg,
		lang:    lang,
		dict:    dict,
		dicts:   locales,
		printer: message.NewPrinter(lang),
	}
	return t
}

func (t *Translator) Format(key string, args ...interface{}) string {
	trans, found := t.dict[key]
	if !found {
		trans, ok := defaultDict[key]
		if !ok {
			return fmt.Sprintf("translation key %s does not exist", key)
		}
		return defaultPrinter.Sprintf(trans.Value, args...)
	}
	return t.printer.Sprintf(trans.Value, args...)
}

// SetLanguage updates the language of this translator, It will be a no-op if
// there is not translation for the provided language.
func (t *Translator) SetLanguage(lang language.Tag) {
	dict, ok := t.dicts[lang]
	if !ok {
		return
	}
	t.lang = lang
	t.dict = dict
	t.printer = message.NewPrinter(lang)
}

func (t *Translator) Lang() language.Tag {
	return t.lang
}

func (t *Translator) Translate(key string) string {
	trans, ok := t.dict[key]
	if !ok {
		trans, ok := defaultDict[key]
		if !ok {
			return key
		}
		return trans.Value
	}
	return trans.Value
}

// Langs returns a list of supported language translations.
func (t *Translator) Langs() []language.Tag {
	langs := make([]language.Tag, 0, len(t.dicts))
	for lang := range t.dicts {
		langs = append(langs, lang)
	}
	return langs
}

// NotificationTopicSubject returns the formate for a notification subject.
func NotificationTopicSubject(topic string) string {
	return fmt.Sprintf("%s_Subject", topic)
}

// NotificationTopicSubject returns the formate for a notification template.
func NotificationTopicTemplate(topic string) string {
	return fmt.Sprintf("%s_Template", topic)
}

// CheckTranslations is used to report missing and stale translations.
func CheckTranslations() (missing, stale map[language.Tag][]string) {
	missing = make(map[language.Tag][]string)
	stale = make(map[language.Tag][]string)

	for lang, translations := range locales {
		if lang == defaultLang {
			continue
		}
		var missingKeys, staleKeys []string
		for key := range defaultDict {
			t, found := translations[key]
			if !found {
				missingKeys = append(missingKeys, key)
			} else if t.Stale {
				staleKeys = append(staleKeys, key)
			}
		}
		if len(missingKeys) > 0 {
			missing[lang] = missingKeys
		}
		if len(staleKeys) > 0 {
			stale[lang] = staleKeys
		}
	}

	return missing, stale
}
