package whatsapp

import (
	"mime"
	"path/filepath"
	"strings"

	"github.com/atendi9/meta/uuid"
)

// MessageType represents the classification of a WhatsApp message.
type MessageType string

// NewMessageType evaluates the given contentType string and returns the corresponding [MessageType].
// If the content type indicates text, it returns [Text]. Otherwise, it maps the content type
// to [Audio], [Video], [Image], or [Sticker]. It defaults to [Document] if no match is found.
func NewMessageType(contentType string) MessageType {
	contentTypes := ContentTypes(
		Audio.String(),
		Video.String(),
		Image.String(),
		Sticker.String(),
	)

	if strings.HasPrefix(strings.ToLower(contentType), "text") {
		return Text
	}

	for k, t := range contentTypes {
		if k == contentType {
			return MessageType(t)
		}
	}
	return Document
}

// String returns the string representation of the [MessageType].
func (m MessageType) String() string {
	return string(m)
}

// audioDefaultMimeType defines the default MIME type used for audio messages.
const audioDefaultMimeType = "audio/ogg; codecs=opus"

// MimeTypeWithMessageType returns the default MIME type string for a specific [MessageType].
// It utilizes [mime.TypeByExtension] to determine the MIME type based on default extensions.
// If the [MessageType] does not match a known type, it returns DefaultContentType.
func MimeTypeWithMessageType(messageType MessageType) string {
	switch messageType {
	case Audio:
		return audioDefaultMimeType
	case Video:
		return mime.TypeByExtension(DefaultVideoExtension)
	case Image:
		return mime.TypeByExtension(DefaultImageExtension)
	case Sticker:
		return mime.TypeByExtension(DefaultStickerExtension)
	case Text:
		return "text/plain"
	}
	return DefaultContentType
}

// GenerateFileName creates a new unique file name by generating a UUID using [uuid.NewV7]
// and appending the extension from the original fileName.
func GenerateFileName(fileName string) string {
	name := func() string {
		id, _ := uuid.NewV7()
		return id.String()
	}
	return name() + filepath.Ext(fileName)
}

// Standard file extensions and default content types used within the package.
const (
	OGGExtension            = ".ogg"
	OpusExtension           = ".opus"
	AACExtension            = ".aac"
	DefaultAudioExtension   = OpusExtension
	DefaultVideoExtension   = MP4Extension
	DefaultImageExtension   = JPEGExtension
	DefaultStickerExtension = WebpExtension
	MP4Extension            = ".mp4"
	JPEGExtension           = ".jpeg"
	InvalidImgExtension     = ".jfif"
	ZipExtension            = ".zip"
	RarExtension            = ".rar"
	DocxExtension           = ".docx"
	XlsxExtension           = ".xlsx"
	PptxExtension           = ".pptx"
	WebpExtension           = ".webp"
	CSVExtension            = ".csv"
	VCFExtension            = ".vcf"
	TextExtension           = ".txt"
	DefaultContentType      = "application/octet-stream"
)

// Standard [MessageType] constants representing the supported WhatsApp message classifications.
const (
	Audio    MessageType = "AUDIO"
	Image    MessageType = "IMAGE"
	Video    MessageType = "VIDEO"
	Document MessageType = "DOCUMENT"
	Text     MessageType = "TEXT"
	Sticker  MessageType = "STICKER"
)

// ContentTypes returns a map associating standard MIME types with their
// corresponding [MessageType] string representations.
func ContentTypes(
	audioType,
	videoType,
	imageType,
	stickerType string,
) map[string]string {
	contentTypes := map[string]string{
		"audio/aac":              audioType,
		"application/aac":        audioType,
		"audio/mp3":              audioType,
		"audio/mpeg":             audioType,
		"application/mpeg":       audioType,
		"application/mp3":        audioType,
		"audio/ogg":              audioType,
		"application/ogg":        audioType,
		"video/mp4":              videoType,
		"application/mp4":        videoType,
		"video/wav":              videoType,
		"application/wav":        videoType,
		"video/x-msvideo":        videoType,
		"application/x-msvideo":  videoType,
		"video/x-matroska":       videoType,
		"application/x-matroska": videoType,
		"video/3gpp":             videoType,
		"application/3gpp":       videoType,
		"video/quicktime":        videoType,
		"application/quicktime":  videoType,
		"image/jpeg":             imageType,
		"application/jpeg":       imageType,
		"image/png":              imageType,
		"application/png":        imageType,
		"image/gif":              imageType,
		"application/gif":        imageType,
		"image/webp":             stickerType,
		"application/webp":       stickerType,
		"image/bmp":              imageType,
		"application/bmp":        imageType,
		"image/tiff":             imageType,
		"application/tiff":       imageType,
	}
	return contentTypes
}
