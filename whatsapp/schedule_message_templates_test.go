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
}
