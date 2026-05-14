package whatsapp

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/atendi9/meta/xhttp/xjson"
)

type SchedulingTemplateOptions struct {
	Lang           Lang
	TemplateName   string
	ScheduleLayout TimeLayout
	StartTime      time.Time
	Timezone       string
}

func SchedulingTemplate(
	h Header,
	opts SchedulingTemplateOptions,
) Message {
	loc, _ := time.LoadLocation(opts.Timezone)
	localTime := opts.StartTime.In(loc)
	m := MessageTemplate(h, xjson.JSON{
		"name":     opts.TemplateName,
		"language": xjson.JSON{"code": opts.Lang},
		"components": []xjson.JSON{
			{
				"type": "body",
				"parameters": []xjson.JSON{
					{"type": "text", "text": localTime.Format(opts.ScheduleLayout.String())},
				},
			},
		},
	})
	return m
}

// ScheduleConfirmationTemplateOptions holds the configuration fields required to build a schedule confirmation template.
// It includes details such as the event's start time ([time.Time]), current time ([time.Time]),
// a list of related URLs ([scheduleUrls]), and the target language ([Lang]).
type ScheduleConfirmationTemplateOptions struct {
	Name,
	EventTitle,
	CustomerName string
	StartTime, Now time.Time
	Urls           scheduleUrls
	Lang           Lang
}

// ScheduleConfirmationTemplate generates a localized confirmation [Message] based on the provided [Header]
// and [ScheduleConfirmationTemplateOptions]. It calculates the event hour, determines the correct
// grammatical gender for the event title, and constructs a JSON payload with the appropriate translations.
// It appends any provided [scheduleUrls] as text parameters within the template components.
func ScheduleConfirmationTemplate(
	h Header,
	opts ScheduleConfirmationTemplateOptions,
) Message {
	hour := opts.StartTime.Add(-3 * time.Hour).Format("15:04")
	dayDescription := "hoje"
	greeting := "Olá"

	gender := getGenderBasedOnGrammar(opts.EventTitle, opts.Lang)
	var eventRef MsgReference

	switch gender {
	case Masculine:
		eventRef = RefEventMasc
	case Feminine:
		eventRef = RefEventFem
	default:
		eventRef = RefEventNeutral
	}

	greetingFormat, ok := translations[LangMsgRef{opts.Lang, RefGreeting}]
	if !ok {
		greetingFormat = translations[LangMsgRef{PortugueseBrazil, RefGreeting}]
	}

	eventFormat, ok := translations[LangMsgRef{opts.Lang, eventRef}]
	if !ok {
		eventFormat = translations[LangMsgRef{PortugueseBrazil, eventRef}]
	}

	firstParam := fmt.Sprintf(greetingFormat, greeting, opts.CustomerName)
	secondParam := fmt.Sprintf(eventFormat, opts.EventTitle, dayDescription, hour)

	mt := MessageTemplate(h, xjson.JSON{
		"name":     opts.Name,
		"language": xjson.JSON{"code": opts.Lang},
		"components": []xjson.JSON{
			{
				"type": "body",
				"parameters": []xjson.JSON{
					{"type": "text", "text": firstParam},
					{"type": "text", "text": secondParam},
				},
			},
		},
	})

	for _, url := range opts.Urls {
		mt["template"].(xjson.JSON)["components"].([]xjson.JSON)[0]["parameters"] = append(
			mt["template"].(xjson.JSON)["components"].([]xjson.JSON)[0]["parameters"].([]xjson.JSON),
			xjson.JSON{
				"type": "text",
				"text": url,
			})
	}

	return mt
}

type MsgReference string

const (
	RefEventMasc    MsgReference = "event_masc"
	RefEventFem     MsgReference = "event_fem"
	RefEventNeutral MsgReference = "event_neutral"
	RefGreeting     MsgReference = "greeting"
)

// LangMsgRef represents the composite key for the translations map.
type LangMsgRef struct {
	Lang Lang
	Ref  MsgReference
}

var translations = map[LangMsgRef]string{
	{PortugueseBrazil, RefGreeting}:     "%s %s",
	{PortugueseBrazil, RefEventMasc}:    "O %s está marcado para %s às %s",
	{PortugueseBrazil, RefEventFem}:     "A %s está marcada para %s às %s",
	{PortugueseBrazil, RefEventNeutral}: "O/A %s está marcado(a) para %s às %s",

	{English, RefGreeting}:     "%s %s",
	{English, RefEventMasc}:    "The %s is scheduled for %s at %s",
	{English, RefEventFem}:     "The %s is scheduled for %s at %s",
	{English, RefEventNeutral}: "The %s is scheduled for %s at %s",

	{EnglishUK, RefGreeting}:     "%s %s",
	{EnglishUK, RefEventMasc}:    "The %s is scheduled for %s at %s",
	{EnglishUK, RefEventFem}:     "The %s is scheduled for %s at %s",
	{EnglishUK, RefEventNeutral}: "The %s is scheduled for %s at %s",

	{Spanish, RefGreeting}:     "%s %s",
	{Spanish, RefEventMasc}:    "El %s está programado para el %s a las %s",
	{Spanish, RefEventFem}:     "La %s está programada para el %s a las %s",
	{Spanish, RefEventNeutral}: "El/La %s está programado(a) para el %s a las %s",

	{French, RefGreeting}:     "%s %s",
	{French, RefEventMasc}:    "Le %s est prévu pour le %s à %s",
	{French, RefEventFem}:     "La %s est prévue pour le %s à %s",
	{French, RefEventNeutral}: "Le/La %s est prévu(e) pour le %s à %s",

	{German, RefGreeting}:     "%s %s",
	{German, RefEventMasc}:    "Das %s ist für %s um %s geplant",
	{German, RefEventFem}:     "Die %s ist für %s um %s geplant",
	{German, RefEventNeutral}: "Der/Die/Das %s ist für %s um %s geplant",

	{ChineseSimplified, RefGreeting}:     "%s %s",
	{ChineseSimplified, RefEventMasc}:    "%s 已安排在 %s %s",
	{ChineseSimplified, RefEventFem}:     "%s 已安排在 %s %s",
	{ChineseSimplified, RefEventNeutral}: "%s 已安排在 %s %s",

	{Arabic, RefGreeting}:     "%s %s",
	{Arabic, RefEventMasc}:    "تمت جدولة %s ليوم %s في %s",
	{Arabic, RefEventFem}:     "تمت جدولة %s ليوم %s في %s",
	{Arabic, RefEventNeutral}: "تمت جدولة %s ليوم %s في %s",

	{Hindi, RefGreeting}:     "%s %s",
	{Hindi, RefEventMasc}:    "%s %s को %s बजे निर्धारित है",
	{Hindi, RefEventFem}:     "%s %s को %s बजे निर्धारित है",
	{Hindi, RefEventNeutral}: "%s %s को %s बजे निर्धारित है",

	{Japanese, RefGreeting}:     "%[2]s様、%[1]s",
	{Japanese, RefEventMasc}:    "%s は %s の %s に予定されています",
	{Japanese, RefEventFem}:     "%s は %s の %s に予定されています",
	{Japanese, RefEventNeutral}: "%s は %s の %s に予定されています",
}

// scheduleUrls holds the URLs for canceling or rescheduling an event.
type scheduleUrls [2]string

// ScheduleUrls creates and returns a new [scheduleUrls].
func ScheduleUrls(cancel, reschedule string) scheduleUrls {
	return scheduleUrls{cancel, reschedule}
}

// Gender represents the grammatical gender of a word.
type Gender string

const (
	Masculine Gender = "M"
	Feminine  Gender = "F"
	Neutral   Gender = "N"
)

// genderKeywords holds masculine and feminine keywords for a specific language.
type genderKeywords struct {
	Masculine []string
	Feminine  []string
}

var languageKeywords = map[Lang]genderKeywords{
	PortugueseBrazil: {
		Masculine: []string{"agendamento", "evento", "compromisso"},
		Feminine:  []string{"consulta", "reunião", "sessão", "entrevista"},
	},
	Spanish: {
		Masculine: []string{"evento", "compromiso", "turno", "encuentro"},
		Feminine:  []string{"consulta", "reunión", "sesión", "entrevista", "cita"},
	},
	French: {
		Masculine: []string{"événement", "rendez-vous", "engagement", "evenement"},
		Feminine:  []string{"consultation", "réunion", "session", "entrevue", "rencontre", "reunion"},
	},
	German: {
		Masculine: []string{"termin"},
		Feminine:  []string{"sitzung", "besprechung", "konferenz"},
	},
}

// getGenderBasedOnGrammar determines the [Gender] of the given title based on the provided [Lang].
func getGenderBasedOnGrammar(title string, lang Lang) Gender {
	keywords, ok := languageKeywords[lang]
	if !ok {
		return Neutral
	}

	titleStr := strings.ToLower(title)

	if slices.ContainsFunc(keywords.Masculine, func(k string) bool {
		return strings.Contains(titleStr, k)
	}) {
		return Masculine
	}

	if slices.ContainsFunc(keywords.Feminine, func(k string) bool {
		return strings.Contains(titleStr, k)
	}) {
		return Feminine
	}

	return Neutral
}

// TimeLayout represents a date and time format string based on the Go reference time.
type TimeLayout string

const (
	// TimeLayoutEN is the layout for [English].
	TimeLayoutEN TimeLayout = "01/02/2006 at 03:04 PM"

	// TimeLayoutES is the layout for [Spanish].
	TimeLayoutES TimeLayout = "02/01/2006 a las 15:04"

	// TimeLayoutENGB is the layout for [EnglishUK].
	TimeLayoutENGB TimeLayout = "02/01/2006 at 15:04"

	// TimeLayoutPT is the layout for [PortugueseBrazil].
	TimeLayoutPT TimeLayout = "02/01/2006 às 15:04"

	// TimeLayoutFR is the layout for [French].
	TimeLayoutFR TimeLayout = "02/01/2006 à 15:04"

	// TimeLayoutDE is the layout for [German].
	TimeLayoutDE TimeLayout = "02.01.2006 um 15:04"

	// TimeLayoutZH is the layout for [ChineseSimplified].
	TimeLayoutZH TimeLayout = "2006年01月02日 15:04"

	// TimeLayoutAR is the layout for [Arabic].
	TimeLayoutAR TimeLayout = "02/01/2006 في 15:04"

	// TimeLayoutHI is the layout for [Hindi].
	TimeLayoutHI TimeLayout = "02/01/2006 15:04"

	// TimeLayoutJA is the layout for [Japanese].
	TimeLayoutJA TimeLayout = "2006年01月02日 15:04"
)

// String returns the string representation of the [TimeLayout].
func (s TimeLayout) String() string {
	return string(s)
}

// ErrUnsupportedLanguage is returned when the provided [Lang] has no mapped [TimeLayout].
var ErrUnsupportedLanguage = errors.New("unsupported language for time layout")

// NewTimeLayout returns the corresponding [TimeLayout] for the provided [Lang].
func NewTimeLayout(lang Lang) (TimeLayout, error) {
	switch lang {
	case English:
		return TimeLayoutEN, nil
	case Spanish:
		return TimeLayoutES, nil
	case EnglishUK:
		return TimeLayoutENGB, nil
	case PortugueseBrazil:
		return TimeLayoutPT, nil
	case French:
		return TimeLayoutFR, nil
	case German:
		return TimeLayoutDE, nil
	case ChineseSimplified:
		return TimeLayoutZH, nil
	case Arabic:
		return TimeLayoutAR, nil
	case Hindi:
		return TimeLayoutHI, nil
	case Japanese:
		return TimeLayoutJA, nil
	default:
		return "", ErrUnsupportedLanguage
	}
}
