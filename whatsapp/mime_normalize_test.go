package whatsapp

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

// Minimal magic-byte fixtures for each format. They carry only the bytes the
// sniffer inspects, which keeps the tests fast and dependency-free.
var (
	fixturePDF  = []byte("%PDF-1.7\n%minimal pdf body")
	fixtureJPEG = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	fixturePNG  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}
	fixtureWebP = []byte{'R', 'I', 'F', 'F', 0x1A, 0, 0, 0, 'W', 'E', 'B', 'P', 'V', 'P'}
	fixtureOgg  = []byte{'O', 'g', 'g', 'S', 0, 2, 0, 0, 0, 0, 0, 0}
	fixtureMP3  = []byte{'I', 'D', '3', 0x03, 0, 0, 0, 0, 0, 0}
	fixtureAMR  = []byte("#!AMR\n\x00\x00\x00")
)

// isoBaseMedia builds a minimal ISO Base Media File Format header with the
// given major brand, used to fixture MP4 / M4A / 3GP containers.
func isoBaseMedia(brand string) []byte {
	b := []byte{0, 0, 0, 0x18, 'f', 't', 'y', 'p'}
	b = append(b, []byte(brand)...)
	return append(b, make([]byte, 12)...)
}

// ooxmlZip builds a minimal ZIP container carrying the OOXML marker directory
// for the requested document family (word/, xl/ or ppt/).
func ooxmlZip(marker string) []byte {
	b := []byte{'P', 'K', 0x03, 0x04}
	b = append(b, make([]byte, 26)...)
	return append(b, []byte(marker)...)
}

func TestNormalizeMediaMimeType_ByExtension(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		filePath string
		want     string
	}{
		// Images.
		{"jpg extension", "application/octet-stream", "photo.jpg", mimeJPEG},
		{"jpeg extension", "", "photo.JPEG", mimeJPEG},
		{"png extension", "application/octet-stream", "img.png", mimePNG},
		{"webp extension", "", "sticker.webp", mimeWebP},
		// Video.
		{"mp4 extension", "application/octet-stream", "clip.mp4", mimeMP4Video},
		{"3gp extension", "", "clip.3gp", mime3GPPVideo},
		{"3gpp extension", "", "clip.3gpp", mime3GPPVideo},
		// Audio: the core mobile bug.
		{"m4a extension", "application/octet-stream", "voice.m4a", mimeMP4Audio},
		{"opus extension", "application/octet-stream", "voice.opus", mimeOGGAudio},
		{"ogg extension", "application/octet-stream", "voice.ogg", mimeOGGAudio},
		{"aac extension", "", "voice.aac", mimeAAC},
		{"mp3 extension", "", "voice.mp3", mimeMPEGAudio},
		{"amr extension", "", "voice.amr", mimeAMR},
		// Documents.
		{"pdf extension", "application/octet-stream", "report.pdf", mimePDF},
		{"doc extension", "application/octet-stream", "report.doc", mimeMSWord},
		{"docx extension", "application/octet-stream", "report.docx", mimeDocx},
		{"xls extension", "application/octet-stream", "sheet.xls", mimeMSExcel},
		{"xlsx extension", "application/octet-stream", "sheet.xlsx", mimeXlsx},
		{"ppt extension", "application/octet-stream", "deck.ppt", mimeMSPPoint},
		{"pptx extension", "application/octet-stream", "deck.pptx", mimePptx},
		{"txt extension", "application/octet-stream", "notes.txt", mimeTextPlain},
		{"csv extension", "application/octet-stream", "data.csv", mimeCSV},
		// Extension wins even over a wrong declared type.
		{"extension overrides wrong mime", "image/gif", "real.pdf", mimePDF},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeMediaMimeType(tc.mimeType, tc.filePath, nil)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeMediaMimeType_ByContent(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    string
	}{
		{"pdf magic bytes", fixturePDF, mimePDF},
		{"jpeg magic bytes", fixtureJPEG, mimeJPEG},
		{"png magic bytes", fixturePNG, mimePNG},
		{"webp riff container", fixtureWebP, mimeWebP},
		{"ogg container", fixtureOgg, mimeOGGAudio},
		{"mp3 with id3 tag", fixtureMP3, mimeMPEGAudio},
		{"mp3 frame sync", []byte{0xFF, 0xFB, 0x90, 0x00}, mimeMPEGAudio},
		{"amr magic bytes", fixtureAMR, mimeAMR},
		{"mp4 ftyp brand", isoBaseMedia("isom"), mimeMP4Video},
		{"m4a ftyp brand", isoBaseMedia("M4A "), mimeMP4Audio},
		{"3gp ftyp brand", isoBaseMedia("3gp4"), mime3GPPVideo},
		{"docx zip container", ooxmlZip("word/document.xml"), mimeDocx},
		{"xlsx zip container", ooxmlZip("xl/workbook.xml"), mimeXlsx},
		{"pptx zip container", ooxmlZip("ppt/presentation.xml"), mimePptx},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// No extension, generic declared type: forces content sniffing.
			got := NormalizeMediaMimeType("application/octet-stream", "blob", tc.content)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeMediaMimeType_ByCategory(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     string
	}{
		// Mobile audio variants that the old allowlist rejected.
		{"audio x-m4a", "audio/x-m4a", mimeMP4Audio},
		{"audio m4a", "audio/m4a", mimeMP4Audio},
		{"audio opus", "audio/opus", mimeOGGAudio},
		{"audio webm", "audio/webm", mimeOGGAudio},
		{"audio 3gpp", "audio/3gpp", mimeAMR},
		{"audio x-aac", "audio/x-aac", mimeAAC},
		{"audio mp3 alias", "audio/mp3", mimeMPEGAudio},
		// Image variants.
		{"image jpg alias", "image/jpg", mimeJPEG},
		{"image x-png", "image/x-png", mimePNG},
		// Video variants.
		{"video 3gp alias", "video/3gp", mime3GPPVideo},
		{"video x-mp4", "video/x-mp4", mimeMP4Video},
		// Document variants.
		{"x-pdf alias", "application/x-pdf", mimePDF},
		{"csv alias", "application/csv", mimeCSV},
		// MIME parameters are stripped before matching.
		{"ogg with codec param", "audio/ogg; codecs=opus", mimeOGGAudio},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeMediaMimeType(tc.mimeType, "voice", nil)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeMediaMimeType_AlreadyCanonical(t *testing.T) {
	canonical := []string{
		mimeJPEG, mimePNG, mimeWebP, mimeMP4Video, mime3GPPVideo,
		mimeAAC, mimeMP4Audio, mimeAMR, mimeMPEGAudio, mimeOGGAudio,
		mimePDF, mimeTextPlain, mimeCSV,
	}
	for _, mt := range canonical {
		t.Run(mt, func(t *testing.T) {
			got := NormalizeMediaMimeType(mt, "file", nil)
			assert.True(t, IsValidMediaUploadType(got))
		})
	}
}

func TestNormalizeMediaMimeType_Unresolved(t *testing.T) {
	// No extension, no content, unknown declared type: the input is returned
	// untouched (trimmed, lower-cased) so the caller can report a real error.
	got := NormalizeMediaMimeType("  Application/JSON  ", "data", nil)
	assert.Equal(t, "application/json", got)
	assert.False(t, IsValidMediaUploadType(got))

	// Empty everything.
	assert.Equal(t, "", NormalizeMediaMimeType("", "", nil))
}

func TestNormalizeMediaMimeType_GenericWithUnknownContent(t *testing.T) {
	// Generic declared type with bytes that match no media format: the stdlib
	// fallback rejects them and the original generic type is returned.
	got := NormalizeMediaMimeType("application/octet-stream", "blob", []byte{0x00, 0x01, 0x02, 0x03})
	assert.False(t, IsValidMediaUploadType(got))
}

func TestNormalizeMediaMimeType_GenericWithPlainText(t *testing.T) {
	// Plain-text bytes under a generic declared type resolve to text/plain via
	// the stdlib sniffer, which the WhatsApp Cloud API accepts as a document.
	got := NormalizeMediaMimeType("application/octet-stream", "blob", []byte("just plain words"))
	assert.True(t, IsValidMediaUploadType(got))
}

func TestIsGenericMimeType(t *testing.T) {
	generic := []string{"", "application/octet-stream", "application/binary", "binary/octet-stream"}
	for _, mt := range generic {
		assert.True(t, isGenericMimeType(mt))
	}
	specific := []string{"image/png", "audio/ogg", "application/pdf"}
	for _, mt := range specific {
		assert.False(t, isGenericMimeType(mt))
	}
}

func TestSniffMediaMimeType_EmptyContent(t *testing.T) {
	assert.Equal(t, "", sniffMediaMimeType(nil))
	assert.Equal(t, "", sniffMediaMimeType([]byte{}))
}

func TestSniffMediaMimeType_LegacyOfficeDocument(t *testing.T) {
	cfb := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1, 0, 0}
	assert.Equal(t, mimeMSWord, sniffMediaMimeType(cfb))
}

func TestSniffZipContainer_PlainArchive(t *testing.T) {
	// A ZIP with no OOXML marker resolves to a plain archive type, which the
	// allowlist correctly rejects.
	plainZip := append([]byte{'P', 'K', 0x03, 0x04}, make([]byte, 30)...)
	assert.Equal(t, mimeZip, sniffZipContainer(plainZip))
	assert.False(t, IsValidMediaUploadType(sniffZipContainer(plainZip)))
}

func TestNormalizeMediaMimeType_FfmpegMangledAudio(t *testing.T) {
	// ffmpeg frequently labels transcoded opus output as audio/webm or leaves
	// it generic; both must resolve to a Meta-accepted audio type.
	assert.Equal(t, mimeOGGAudio, NormalizeMediaMimeType("audio/webm", "out", nil))
	assert.Equal(t, mimeOGGAudio, NormalizeMediaMimeType("application/octet-stream", "out", fixtureOgg))
}
