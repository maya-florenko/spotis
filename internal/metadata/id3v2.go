package metadata

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Metadata contains track metadata to be written to the audio file
type Metadata struct {
	Title  string
	Artist string
	Album  string
	Cover  string // URL or path to cover image
}

// AddMetadata adds ID3v2 tags to MP3 audio data
func AddMetadata(audioData *bytes.Buffer, meta Metadata) (*bytes.Buffer, error) {
	// Create ID3v2 frames
	frames := &bytes.Buffer{}

	// Add title frame (TIT2)
	if meta.Title != "" {
		if err := writeTextFrame(frames, "TIT2", meta.Title); err != nil {
			return nil, fmt.Errorf("write title frame: %w", err)
		}
	}

	// Add artist frame (TPE1)
	if meta.Artist != "" {
		if err := writeTextFrame(frames, "TPE1", meta.Artist); err != nil {
			return nil, fmt.Errorf("write artist frame: %w", err)
		}
	}

	// Add album frame (TALB)
	if meta.Album != "" {
		if err := writeTextFrame(frames, "TALB", meta.Album); err != nil {
			return nil, fmt.Errorf("write album frame: %w", err)
		}
	}

	// Add cover image frame (APIC) if URL is provided
	if meta.Cover != "" {
		if err := writeCoverFrame(frames, meta.Cover); err != nil {
			// Don't fail if cover download fails, just skip it
			fmt.Printf("Warning: failed to add cover: %v\n", err)
		}
	}

	// Create ID3v2 header
	header := &bytes.Buffer{}
	header.WriteString("ID3")                     // ID3 identifier
	header.WriteByte(0x03)                        // Version (ID3v2.3.0)
	header.WriteByte(0x00)                        // Revision
	header.WriteByte(0x00)                        // Flags
	writeSynchsafe(header, uint32(frames.Len())) // Size

	// Combine header + frames + audio data
	result := &bytes.Buffer{}
	result.Write(header.Bytes())
	result.Write(frames.Bytes())
	result.Write(audioData.Bytes())

	return result, nil
}

// writeTextFrame writes a text information frame
func writeTextFrame(buf *bytes.Buffer, frameID, text string) error {
	// Frame header
	buf.WriteString(frameID) // Frame ID (4 bytes)

	// Frame data: encoding (1 byte) + text
	frameData := &bytes.Buffer{}
	frameData.WriteByte(0x03) // UTF-8 encoding
	frameData.WriteString(text)

	// Frame size (4 bytes, big-endian)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(frameData.Len()))
	buf.Write(size)

	// Frame flags (2 bytes)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)

	// Frame data
	buf.Write(frameData.Bytes())

	return nil
}

// writeCoverFrame writes an attached picture (APIC) frame
func writeCoverFrame(buf *bytes.Buffer, coverURL string) error {
	// Download cover image
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", coverURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download cover: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download cover: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read cover data: %w", err)
	}

	// Determine MIME type from response
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg" // Default to JPEG
	}

	// Frame header
	buf.WriteString("APIC") // Frame ID (4 bytes)

	// Frame data
	frameData := &bytes.Buffer{}
	frameData.WriteByte(0x00)            // Text encoding (ISO-8859-1)
	frameData.WriteString(mimeType)      // MIME type
	frameData.WriteByte(0x00)            // Null terminator
	frameData.WriteByte(0x03)            // Picture type (3 = Cover (front))
	frameData.WriteByte(0x00)            // Description (empty)
	frameData.Write(imageData)           // Picture data

	// Frame size (4 bytes, big-endian)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(frameData.Len()))
	buf.Write(size)

	// Frame flags (2 bytes)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)

	// Frame data
	buf.Write(frameData.Bytes())

	return nil
}

// writeSynchsafe writes a synchsafe integer (28-bit integer in 4 bytes)
// Used in ID3v2 tags where the most significant bit of each byte is always 0
func writeSynchsafe(buf *bytes.Buffer, value uint32) {
	buf.WriteByte(byte((value >> 21) & 0x7F))
	buf.WriteByte(byte((value >> 14) & 0x7F))
	buf.WriteByte(byte((value >> 7) & 0x7F))
	buf.WriteByte(byte(value & 0x7F))
}
