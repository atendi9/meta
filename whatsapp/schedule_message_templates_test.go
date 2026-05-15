package whatsapp

import (
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestScheduleTemplates(t *testing.T) {

	t.Run("MsgReferenceConstants", func(t *testing.T) {
		assert.Equal(t, MsgReference("event_masc"), RefEventMasc)
		assert.Equal(t, MsgReference("event_fem"), RefEventFem)
		assert.Equal(t, MsgReference("event_neutral"), RefEventNeutral)
		assert.Equal(t, MsgReference("greeting"), RefGreeting)
	})

	t.Run("GenderConstants", func(t *testing.T) {
		assert.Equal(t, Gender("M"), Masculine)
		assert.Equal(t, Gender("F"), Feminine)
		assert.Equal(t, Gender("N"), Neutral)
	})

	t.Run("TimeLayoutConstants", func(t *testing.T) {
		assert.Equal(t, TimeLayout("01/02/2006 at 03:04 PM"), TimeLayoutEN)
		assert.Equal(t, TimeLayout("02/01/2006 a las 15:04"), TimeLayoutES)
		assert.Equal(t, TimeLayout("02/01/2006 at 15:04"), TimeLayoutENGB)
		assert.Equal(t, TimeLayout("02/01/2006 às 15:04"), TimeLayoutPT)
		assert.Equal(t, TimeLayout("02/01/2006 à 15:04"), TimeLayoutFR)
		assert.Equal(t, TimeLayout("02.01.2006 um 15:04"), TimeLayoutDE)
		assert.Equal(t, TimeLayout("2006年01月02日 15:04"), TimeLayoutZH)
		assert.Equal(t, TimeLayout("02/01/2006 في 15:04"), TimeLayoutAR)
		assert.Equal(t, TimeLayout("02/01/2006 15:04"), TimeLayoutHI)
		assert.Equal(t, TimeLayout("2006年01月02日 15:04"), TimeLayoutJA)
	})

	t.Run("TimeLayout_String", func(t *testing.T) {
		assert.Equal(t, "02/01/2006 às 15:04", TimeLayoutPT.String())
		assert.Equal(t, "01/02/2006 at 03:04 PM", TimeLayoutEN.String())
	})

	t.Run("NewTimeLayout", func(t *testing.T) {
		layoutPT, err := NewTimeLayout(PortugueseBrazil)
		assert.NoError(t, err)
		assert.Equal(t, TimeLayoutPT, layoutPT)

		layoutEN, err := NewTimeLayout(English)
		assert.NoError(t, err)
		assert.Equal(t, TimeLayoutEN, layoutEN)

		_, errInvalid := NewTimeLayout(Lang("unsupported_language"))
		assert.Error(t, errInvalid)
	})

	t.Run("ScheduleUrls", func(t *testing.T) {
		cancelURL := "https://example.com/cancel"
		rescheduleURL := "https://example.com/reschedule"

		urls := ScheduleUrls(cancelURL, rescheduleURL)

		assert.Equal(t, cancelURL, urls[0])
		assert.Equal(t, rescheduleURL, urls[1])
	})

	t.Run("getGenderBasedOnGrammar", func(t *testing.T) {
		masculineGender := getGenderBasedOnGrammar("Agendamento Médico", PortugueseBrazil)
		assert.Equal(t, Masculine, masculineGender)

		feminineGender := getGenderBasedOnGrammar("Reunião de Alinhamento", PortugueseBrazil)
		assert.Equal(t, Feminine, feminineGender)

		neutralGender := getGenderBasedOnGrammar("Palestra", PortugueseBrazil)
		assert.Equal(t, Neutral, neutralGender)

		unknownLangGender := getGenderBasedOnGrammar("Evento", Lang("unsupported_language"))
		assert.Equal(t, Neutral, unknownLangGender)
	})

	t.Run("SchedulingTemplate", func(t *testing.T) {
		opts := SchedulingTemplateOptions{
			Lang:           PortugueseBrazil,
			TemplateName:   "template_test",
			ScheduleLayout: TimeLayoutPT,
			StartTime:      time.Now(),
			Timezone:       "America/Sao_Paulo",
		}

		h := MessageHeader("55819999999", "template")

		_ = SchedulingTemplate(h, opts)
	})

	t.Run("ScheduleConfirmationTemplate", func(t *testing.T) {
		urls := ScheduleUrls("https://example.com/cancel", "https://example.com/reschedule")

		opts := ScheduleConfirmationTemplateOptions{
			Name:         "confirm_test",
			EventTitle:   "Agendamento",
			CustomerName: "John Doe",
			StartTime:    time.Now(),
			Now:          time.Now(),
			Urls:         urls,
			Lang:         PortugueseBrazil,
		}

		h := MessageHeader("55819999999", "template")
		_ = ScheduleConfirmationTemplate(h, opts)

		urlSlice := []string{urls[0], urls[1]}
		assert.LengthSlice(t, 2, urlSlice)
	})

	t.Run("getLocalizedGreeting", func(t *testing.T) {
		cases := []struct {
			lang Lang
			want string
		}{
			{English, "Hello"},
			{EnglishUK, "Hello"},
			{Spanish, "Hola"},
			{French, "Bonjour"},
			{German, "Hallo"},
			{ChineseSimplified, "你好"},
			{Arabic, "مرحباً"},
			{Hindi, "नमस्ते"},
			{Japanese, "こんにちは"},
			{PortugueseBrazil, "Olá"},
			{Lang("xx"), "Olá"},
		}
		for _, c := range cases {
			assert.Equal(t, c.want, getLocalizedGreeting(c.lang))
		}
	})

	t.Run("NewTimeLayout all languages", func(t *testing.T) {
		cases := []struct {
			lang Lang
			want TimeLayout
		}{
			{English, TimeLayoutEN},
			{Spanish, TimeLayoutES},
			{EnglishUK, TimeLayoutENGB},
			{PortugueseBrazil, TimeLayoutPT},
			{French, TimeLayoutFR},
			{German, TimeLayoutDE},
			{ChineseSimplified, TimeLayoutZH},
			{Arabic, TimeLayoutAR},
			{Hindi, TimeLayoutHI},
			{Japanese, TimeLayoutJA},
		}
		for _, c := range cases {
			layout, err := NewTimeLayout(c.lang)
			assert.NoError(t, err)
			assert.Equal(t, c.want, layout)
		}
	})

	t.Run("getLocalizedFormatLayouts", func(t *testing.T) {
		cases := []struct {
			lang     Lang
			wantDate string
			wantTime string
		}{
			{English, "01/02/2006", "03:04 PM"},
			{ChineseSimplified, "2006年01月02日", "15:04"},
			{Japanese, "2006年01月02日", "15:04"},
			{German, "02.01.2006", "15:04"},
			{PortugueseBrazil, "02/01/2006", "15:04"},
			{Lang("xx"), "02/01/2006", "15:04"},
		}
		for _, c := range cases {
			d, tm := getLocalizedFormatLayouts(c.lang)
			assert.Equal(t, c.wantDate, d)
			assert.Equal(t, c.wantTime, tm)
		}
	})

	t.Run("getLocalizedRelativeDay today", func(t *testing.T) {
		cases := []struct {
			lang Lang
			want string
		}{
			{English, "today"},
			{EnglishUK, "today"},
			{Spanish, "hoy"},
			{French, "aujourd'hui"},
			{German, "heute"},
			{ChineseSimplified, "今天"},
			{Arabic, "اليوم"},
			{Hindi, "आज"},
			{Japanese, "今日"},
			{PortugueseBrazil, "hoje"},
			{Lang("xx"), "hoje"},
		}
		for _, c := range cases {
			assert.Equal(t, c.want, getLocalizedRelativeDay(c.lang, true))
		}
	})

	t.Run("getLocalizedRelativeDay tomorrow", func(t *testing.T) {
		cases := []struct {
			lang Lang
			want string
		}{
			{English, "tomorrow"},
			{EnglishUK, "tomorrow"},
			{Spanish, "mañana"},
			{French, "demain"},
			{German, "morgen"},
			{ChineseSimplified, "明天"},
			{Arabic, "غدًا"},
			{Hindi, "कल"},
			{Japanese, "明日"},
			{PortugueseBrazil, "amanhã"},
			{Lang("xx"), "amanhã"},
		}
		for _, c := range cases {
			assert.Equal(t, c.want, getLocalizedRelativeDay(c.lang, false))
		}
	})

	t.Run("DescribeDate relative days", func(t *testing.T) {
		now := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)

		today, _ := DescribeDate(now, now, PortugueseBrazil)
		assert.Equal(t, "hoje", today)

		tomorrowDate := now.Add(24 * time.Hour)
		tomorrow, _ := DescribeDate(tomorrowDate, now, PortugueseBrazil)
		assert.Equal(t, "amanhã", tomorrow)

		futureDate := now.Add(72 * time.Hour)
		future, hour := DescribeDate(futureDate, now, PortugueseBrazil)
		assert.Equal(t, "18/05/2026", future)
		assert.Equal(t, "10:00", hour)
	})

	t.Run("ScheduleConfirmationTemplate feminine and neutral", func(t *testing.T) {
		h := MessageHeader("55819999999", "template")

		femOpts := ScheduleConfirmationTemplateOptions{
			Name:         "confirm_test",
			EventTitle:   "Consulta",
			CustomerName: "Jane Doe",
			StartTime:    time.Now(),
			Now:          time.Now(),
			Lang:         PortugueseBrazil,
		}
		femMsg := ScheduleConfirmationTemplate(h, femOpts)
		assert.NotNil(t, femMsg)

		neutralOpts := ScheduleConfirmationTemplateOptions{
			Name:         "confirm_test",
			EventTitle:   "Palestra",
			CustomerName: "Jane Doe",
			StartTime:    time.Now(),
			Now:          time.Now(),
			Lang:         PortugueseBrazil,
		}
		neutralMsg := ScheduleConfirmationTemplate(h, neutralOpts)
		assert.NotNil(t, neutralMsg)
	})

	t.Run("ScheduleConfirmationTemplate fallback language", func(t *testing.T) {
		h := MessageHeader("55819999999", "template")

		opts := ScheduleConfirmationTemplateOptions{
			Name:         "confirm_test",
			EventTitle:   "Evento",
			CustomerName: "John Doe",
			StartTime:    time.Now(),
			Now:          time.Now(),
			Lang:         Lang("unsupported_language"),
		}
		msg := ScheduleConfirmationTemplate(h, opts)
		assert.NotNil(t, msg)
	})

	t.Run("NormalizeDate", func(t *testing.T) {
		date := time.Date(2026, 5, 15, 13, 45, 30, 0, time.UTC)
		normalized := NormalizeDate(date)
		assert.Equal(t, 0, normalized.Hour())
		assert.Equal(t, 0, normalized.Minute())
		assert.Equal(t, 0, normalized.Second())
		assert.Equal(t, 15, normalized.Day())
	})
}
