package whatsapp

import (
	"path/filepath"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNewMessageType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    MessageType
	}{
		{"Text type", "text/plain", Text},
		{"Uppercase Text type", "TEXT/HTML", Text},
		{"Audio type", "audio/ogg", Audio},
		{"Video type", "video/mp4", Video},
		{"Image type", "image/jpeg", Image},
		{"Sticker type", "image/webp", Sticker},
		{"Document fallback type", "application/pdf", Document},
		{"Unknown type", "unknown/type", Document},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMessageType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessageType_String(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"Audio", Audio, "AUDIO"},
		{"Image", Image, "IMAGE"},
		{"Video", Video, "VIDEO"},
		{"Document", Document, "DOCUMENT"},
		{"Text", Text, "TEXT"},
		{"Sticker", Sticker, "STICKER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msgType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMimeTypeWithMessageType(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"Audio", Audio, audioDefaultMimeType},
		{"Video", Video, "video/mp4"},
		{"Image", Image, "image/jpeg"},
		{"Sticker", Sticker, "image/webp"},
		{"Text", Text, "text/plain"},
		{"Document", Document, DefaultContentType},
		{"Unknown", MessageType("UNKNOWN"), DefaultContentType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MimeTypeWithMessageType(tt.msgType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateFileName(t *testing.T) {
	originalName := "report.pdf"
	result := GenerateFileName(originalName)

	expectedExt := filepath.Ext(originalName)
	resultExt := filepath.Ext(result)
	assert.Equal(t, expectedExt, resultExt)

	isDifferent := originalName != result
	assert.True(t, isDifferent)
}

func TestContentTypes(t *testing.T) {
	audio := Audio.String()
	video := Video.String()
	image := Image.String()
	sticker := Sticker.String()

	result := ContentTypes(audio, video, image, sticker)

	assert.LengthMap(t, 32, result)

	assert.Equal(t, audio, result["audio/ogg"])
	assert.Equal(t, video, result["video/mp4"])
	assert.Equal(t, image, result["image/jpeg"])
	assert.Equal(t, sticker, result["image/webp"])
}
