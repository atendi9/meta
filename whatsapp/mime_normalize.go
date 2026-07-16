package whatsapp

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strings"
)

// Canonical MIME types accepted by the WhatsApp Cloud API for media uploads.
// These are the values NormalizeMediaMimeType resolves ambiguous inputs to.
const (
	mimeJPEG       = "image/jpeg"
	mimePNG        = "image/png"
	mimeWebP       = "image/webp"
	mimeMP4Video   = "video/mp4"
	mime3GPPVideo  = "video/3gpp"
	mimeAAC        = "audio/aac"
	mimeMP4Audio   = "audio/mp4"
	mimeAMR        = "audio/amr"
	mimeMPEGAudio  = "audio/mpeg"
	mimeOGGAudio   = "audio/ogg"
	mimePDF        = "application/pdf"
	mimeMSWord     = "application/msword"
	mimeDocx       = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	mimeMSExcel    = "application/vnd.ms-excel"
	mimeXlsx       = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	mimeMSPPoint   = "application/vnd.ms-powerpoint"
	mimePptx       = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	mimeTextPlain  = "text/plain"
	mimeZip        = "application/zip"
	mimeGenericBin = "application/octet-stream"
)

// extensionMimeTypes maps a lower-cased file extension (including the leading
// dot) to the canonical MIME type accepted by the WhatsApp Cloud API.
//
// Mobile clients and ffmpeg often produce ambiguous or platform-specific MIME
// types (audio/x-m4a, audio/m4a, application/octet-stream), so the file
// extension is the most reliable signal available for normalization.
var extensionMimeTypes = map[string]string{
	// Images.
	".jpg":  mimeJPEG,
	".jpeg": mimeJPEG,
	".jfif": mimeJPEG,
	".png":  mimePNG,
	".webp": mimeWebP,
	// Video.
	".mp4":  mimeMP4Video,
	".3gp":  mime3GPPVideo,
	".3gpp": mime3GPPVideo,
	// Audio. Meta treats .m4a/.aac containers as audio/mp4 or audio/aac;
	// .opus and .ogg are normalized to audio/ogg (the opus container).
	".m4a":  mimeMP4Audio,
	".aac":  mimeAAC,
	".mp3":  mimeMPEGAudio,
	".mpeg": mimeMPEGAudio,
	".mpga": mimeMPEGAudio,
	".opus": mimeOGGAudio,
	".ogg":  mimeOGGAudio,
	".oga":  mimeOGGAudio,
	".amr":  mimeAMR,
	// Documents.
	".pdf":  mimePDF,
	".doc":  mimeMSWord,
	".docx": mimeDocx,
	".xls":  mimeMSExcel,
	".xlsx": mimeXlsx,
	".ppt":  mimeMSPPoint,
	".pptx": mimePptx,
	".txt":  mimeTextPlain,
	// The WhatsApp Cloud API does not accept text/csv; CSV files are uploaded
	// as text/plain, which it does accept as a document.
	".csv": mimeTextPlain,
}

// isGenericMimeType reports whether a MIME type carries no useful information
// and should be resolved from the file extension or content sniffing.
func isGenericMimeType(mimeType string) bool {
	switch mimeType {
	case "", mimeGenericBin, "application/binary", "binary/octet-stream":
		return true
	default:
		return false
	}
}

// NormalizeMediaMimeType resolves a possibly generic, mobile-specific, or
// ffmpeg-mangled MIME type into a canonical MIME type accepted by the WhatsApp
// Cloud API. It combines three strategies, in order of reliability:
//
//  1. File extension of filePath (most reliable for mobile audio and OOXML
//     documents whose MIME type is frequently generic).
//  2. Magic-byte sniffing of content for generic/empty inputs.
//  3. Category-based correction of inputs that are valid in shape but use a
//     codec-specific or platform variant Meta does not accept.
//
// When no strategy yields a known-good type, the original (trimmed, lower-cased)
// MIME type is returned so the caller can still surface a meaningful error.
func NormalizeMediaMimeType(mimeType, filePath string, content []byte) string {
	original := strings.ToLower(strings.TrimSpace(mimeType))
	// Drop any codec/charset parameters, e.g. "audio/ogg; codecs=opus".
	base := original
	if idx := strings.IndexByte(base, ';'); idx >= 0 {
		base = strings.TrimSpace(base[:idx])
	}

	// Strategy 1: trust the file extension.
	ext := strings.ToLower(filepath.Ext(filePath))
	if canonical, ok := extensionMimeTypes[ext]; ok {
		return canonical
	}

	// Strategy 2: sniff content when the declared type is uninformative.
	if isGenericMimeType(base) {
		if sniffed := sniffMediaMimeType(content); sniffed != "" {
			return sniffed
		}
	}

	// Strategy 3: correct codec-specific or platform variants by category.
	if canonical := canonicalizeByCategory(base); canonical != "" {
		return canonical
	}

	if base == "" {
		return original
	}
	return base
}

// canonicalizeByCategory maps codec-specific or platform-variant MIME types
// onto the canonical type Meta accepts for that media category. It returns an
// empty string when the input is not a recognized variant.
func canonicalizeByCategory(base string) string {
	switch base {
	// Audio variants emitted by mobile clients and transcoders.
	case "audio/x-m4a", "audio/m4a", "audio/mp4a-latm", "audio/x-mp4a":
		return mimeMP4Audio
	case "audio/opus", "audio/x-opus", "audio/webm", "audio/vorbis", "application/ogg":
		return mimeOGGAudio
	case "audio/3gpp", "audio/3gp", "audio/amr-wb", "audio/x-amr":
		return mimeAMR
	case "audio/mp3", "audio/x-mpeg", "audio/mpeg3", "audio/x-mp3", "application/mp3", "application/mpeg":
		return mimeMPEGAudio
	case "audio/x-aac", "audio/aacp", "application/aac":
		return mimeAAC
	// Image variants.
	case "image/jpg", "image/pjpeg", "application/jpeg":
		return mimeJPEG
	case "image/x-png", "application/png":
		return mimePNG
	case "image/x-webp", "application/webp":
		return mimeWebP
	// Video variants.
	case "video/3gp", "video/x-3gpp", "application/3gpp":
		return mime3GPPVideo
	case "video/x-mp4", "video/mpeg4", "application/mp4":
		return mimeMP4Video
	// Document variants.
	case "application/x-pdf":
		return mimePDF
	// CSV is not a Meta-accepted MIME type; route every CSV spelling to
	// text/plain, which the WhatsApp Cloud API accepts as a document.
	case "text/csv", "application/csv", "text/comma-separated-values":
		return mimeTextPlain
	default:
		return ""
	}
}

// sniffMediaMimeType inspects the leading magic bytes of content and returns a
// canonical WhatsApp-accepted MIME type, or an empty string when the format is
// not recognized. It is only consulted for inputs whose declared MIME type is
// generic, so a miss here is harmless.
func sniffMediaMimeType(content []byte) string {
	if len(content) == 0 {
		return ""
	}

	switch {
	case hasPrefix(content, []byte("%PDF-")):
		return mimePDF
	case hasPrefix(content, []byte("\xFF\xD8\xFF")):
		return mimeJPEG
	case hasPrefix(content, []byte("\x89PNG\r\n\x1a\n")):
		return mimePNG
	case isRIFFWebP(content):
		return mimeWebP
	case isOggContainer(content):
		return mimeOGGAudio
	case hasPrefix(content, []byte("ID3")), isMP3FrameSync(content):
		return mimeMPEGAudio
	case hasPrefix(content, []byte("#!AMR")):
		return mimeAMR
	case isISOBaseMedia(content):
		return sniffISOBaseMedia(content)
	case isZipContainer(content):
		return sniffZipContainer(content)
	case hasPrefix(content, []byte("\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1")):
		// Legacy Microsoft Compound File: doc/xls/ppt share this header.
		// Without deeper parsing it cannot be disambiguated, so report the
		// generic legacy Word type, the most common case for documents.
		return mimeMSWord
	default:
		return sniffWithStdlib(content)
	}
}

// sniffWithStdlib falls back to the standard library's content sniffer and
// keeps the result only when it is a media type the WhatsApp Cloud API
// accepts. Anything else (text/html, application/octet-stream, ...) is
// discarded so the caller can still surface a precise error.
func sniffWithStdlib(content []byte) string {
	detected := http.DetectContentType(content)
	if idx := strings.IndexByte(detected, ';'); idx >= 0 {
		detected = strings.TrimSpace(detected[:idx])
	}
	if canonical := canonicalizeByCategory(detected); canonical != "" {
		return canonical
	}
	if IsValidMediaUploadType(detected) {
		return detected
	}
	return ""
}

// imageMagicMimeType returns the canonical image MIME type carried by content's
// magic bytes, or an empty string when content is not one of the image formats
// the WhatsApp Cloud API accepts.
//
// Image signatures are exact, so a hit here is worth more than the file
// extension: browsers derive a file's declared type from its name, so a PNG
// saved as .jpeg is a routine mislabel, and Meta validates an uploaded image
// against its bytes rather than its name. Formats whose signature is ambiguous
// (legacy Compound File, ZIP) are deliberately absent, since only the extension
// can tell those apart.
func imageMagicMimeType(content []byte) string {
	switch {
	case hasPrefix(content, []byte("\xFF\xD8\xFF")):
		return mimeJPEG
	case hasPrefix(content, []byte("\x89PNG\r\n\x1a\n")):
		return mimePNG
	case isRIFFWebP(content):
		return mimeWebP
	default:
		return ""
	}
}

// hasPrefix reports whether b starts with the given prefix.
func hasPrefix(b, prefix []byte) bool {
	return len(b) >= len(prefix) && bytes.Equal(b[:len(prefix)], prefix)
}

// isRIFFWebP reports whether content is a RIFF container holding a WEBP payload.
func isRIFFWebP(content []byte) bool {
	return len(content) >= 12 &&
		bytes.Equal(content[0:4], []byte("RIFF")) &&
		bytes.Equal(content[8:12], []byte("WEBP"))
}

// isOggContainer reports whether content begins with the Ogg page marker,
// which covers both Ogg Vorbis and Ogg Opus audio.
func isOggContainer(content []byte) bool {
	return hasPrefix(content, []byte("OggS"))
}

// isMP3FrameSync reports whether content starts with an MPEG audio frame sync
// word (11 set bits), used to detect MP3 files lacking an ID3 tag.
func isMP3FrameSync(content []byte) bool {
	return len(content) >= 2 && content[0] == 0xFF && (content[1]&0xE0) == 0xE0
}

// isISOBaseMedia reports whether content is an ISO Base Media File Format
// container (MP4, M4A, 3GP), identified by an "ftyp" box at offset 4.
func isISOBaseMedia(content []byte) bool {
	return len(content) >= 12 && bytes.Equal(content[4:8], []byte("ftyp"))
}

// sniffISOBaseMedia resolves an ISO Base Media container into the canonical
// MIME type by inspecting its major brand. M4A audio brands map to audio/mp4,
// 3GP brands to video/3gpp, and everything else to video/mp4.
func sniffISOBaseMedia(content []byte) string {
	brand := string(content[8:12])
	switch {
	case strings.HasPrefix(brand, "M4A"), strings.HasPrefix(brand, "M4B"):
		return mimeMP4Audio
	case strings.HasPrefix(brand, "3gp"), strings.HasPrefix(brand, "3gs"):
		return mime3GPPVideo
	default:
		return mimeMP4Video
	}
}

// isZipContainer reports whether content begins with a ZIP local-file header.
// OOXML documents (docx/xlsx/pptx) are ZIP archives, as are plain .zip files.
func isZipContainer(content []byte) bool {
	return hasPrefix(content, []byte("PK\x03\x04")) ||
		hasPrefix(content, []byte("PK\x05\x06")) ||
		hasPrefix(content, []byte("PK\x07\x08"))
}

// sniffZipContainer disambiguates a ZIP-based container into the matching OOXML
// document MIME type by scanning for the marker entries OOXML packages always
// carry. A plain archive with no OOXML marker resolves to application/zip.
func sniffZipContainer(content []byte) string {
	switch {
	case bytes.Contains(content, []byte("word/")):
		return mimeDocx
	case bytes.Contains(content, []byte("xl/")):
		return mimeXlsx
	case bytes.Contains(content, []byte("ppt/")):
		return mimePptx
	default:
		return mimeZip
	}
}
