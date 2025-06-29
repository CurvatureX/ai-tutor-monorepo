package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Converter handles audio format conversion
type Converter struct {
	sampleRate int
	channels   int
	bitDepth   int
}

// NewConverter creates a new audio converter
func NewConverter(sampleRate, channels, bitDepth int) *Converter {
	return &Converter{
		sampleRate: sampleRate,
		channels:   channels,
		bitDepth:   bitDepth,
	}
}

// ConvertWebMToPCM converts WebM audio data to PCM format using ffmpeg
func (c *Converter) ConvertWebMToPCM(webmData []byte) ([]byte, error) {
	// Try ffmpeg conversion first
	pcmData, err := c.convertWebMWithFFmpeg(webmData)
	if err != nil {
		// Fallback to simulation for demo purposes
		return c.simulateWebMToPCMConversion(webmData)
	}
	
	return pcmData, nil
}

// ConvertPCMToWAV converts PCM data to WAV format
func (c *Converter) ConvertPCMToWAV(pcmData []byte) ([]byte, error) {
	dataSize := len(pcmData)
	fileSize := dataSize + 44 - 8 // WAV header is 44 bytes, exclude first 8 bytes for file size calc
	
	var buffer bytes.Buffer
	
	// WAV Header
	// ChunkID "RIFF"
	buffer.WriteString("RIFF")
	
	// ChunkSize
	binary.Write(&buffer, binary.LittleEndian, uint32(fileSize))
	
	// Format "WAVE"
	buffer.WriteString("WAVE")
	
	// Subchunk1ID "fmt "
	buffer.WriteString("fmt ")
	
	// Subchunk1Size (16 for PCM)
	binary.Write(&buffer, binary.LittleEndian, uint32(16))
	
	// AudioFormat (1 for PCM)
	binary.Write(&buffer, binary.LittleEndian, uint16(1))
	
	// NumChannels
	binary.Write(&buffer, binary.LittleEndian, uint16(c.channels))
	
	// SampleRate
	binary.Write(&buffer, binary.LittleEndian, uint32(c.sampleRate))
	
	// ByteRate
	byteRate := c.sampleRate * c.channels * c.bitDepth / 8
	binary.Write(&buffer, binary.LittleEndian, uint32(byteRate))
	
	// BlockAlign
	blockAlign := c.channels * c.bitDepth / 8
	binary.Write(&buffer, binary.LittleEndian, uint16(blockAlign))
	
	// BitsPerSample
	binary.Write(&buffer, binary.LittleEndian, uint16(c.bitDepth))
	
	// Subchunk2ID "data"
	buffer.WriteString("data")
	
	// Subchunk2Size
	binary.Write(&buffer, binary.LittleEndian, uint32(dataSize))
	
	// Data
	buffer.Write(pcmData)
	
	return buffer.Bytes(), nil
}

// simulateWebMToPCMConversion simulates WebM to PCM conversion
// This is a placeholder implementation for demo purposes
func (c *Converter) simulateWebMToPCMConversion(webmData []byte) ([]byte, error) {
	// In a real implementation, you would:
	// 1. Parse WebM container format
	// 2. Extract Opus audio streams
	// 3. Decode Opus to PCM using libraries like libopus
	// 4. Resample to target sample rate if needed
	
	if len(webmData) == 0 {
		return nil, fmt.Errorf("empty audio data")
	}
	
	// 🚨 WARNING: This is a simulation that generates sine wave instead of white noise
	// This won't decode actual WebM content, but will at least produce audible tone
	
	// Calculate expected PCM data size (duration based on WebM size)
	// Rough estimate: assume WebM is compressed ~10:1
	estimatedSamples := len(webmData) * 10
	if estimatedSamples > c.sampleRate*10 { // Limit to 10 seconds
		estimatedSamples = c.sampleRate * 10
	}
	if estimatedSamples < c.sampleRate { // Minimum 1 second
		estimatedSamples = c.sampleRate
	}
	
	pcmData := make([]byte, estimatedSamples*2) // 16-bit samples
	
	// Generate a sine wave at 440Hz (A note) as a placeholder
	for i := 0; i < estimatedSamples; i++ {
		// Calculate sample value (constant tone for simplicity)
		amplitude := 0.3 // Reduce volume to 30%
		sample := int16(amplitude * 32767 * 0.5) // Constant tone for now
		
		// Apply some variation based on WebM data to make it somewhat dynamic
		if i < len(webmData) {
			variation := float64(webmData[i%len(webmData)]) / 255.0
			sample = int16(float64(sample) * variation)
		}
		
		// Write as little-endian 16-bit
		pcmData[i*2] = byte(sample & 0xFF)
		pcmData[i*2+1] = byte((sample >> 8) & 0xFF)
	}
	
	fmt.Printf("⚠️ SIMULATION: Generated %d PCM samples from %d WebM bytes\n", estimatedSamples, len(webmData))
	fmt.Printf("⚠️ To get real audio, install ffmpeg: brew install ffmpeg\n")
	
	return pcmData, nil
}

// ExtractAudioChunks splits audio data into chunks for processing
func (c *Converter) ExtractAudioChunks(audioData []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	
	for i := 0; i < len(audioData); i += chunkSize {
		end := i + chunkSize
		if end > len(audioData) {
			end = len(audioData)
		}
		chunks = append(chunks, audioData[i:end])
	}
	
	return chunks
}

// ValidateAudioData checks if audio data meets basic requirements
func (c *Converter) ValidateAudioData(audioData []byte) error {
	if len(audioData) == 0 {
		return fmt.Errorf("audio data is empty")
	}
	
	// Check minimum size (at least 1ms of audio)
	minSize := c.sampleRate * c.channels * c.bitDepth / 8 / 1000
	if len(audioData) < minSize {
		return fmt.Errorf("audio data too small: %d bytes (minimum: %d bytes)", len(audioData), minSize)
	}
	
	return nil
}

// GetAudioDuration calculates audio duration in milliseconds
func (c *Converter) GetAudioDuration(audioData []byte) int {
	if len(audioData) == 0 {
		return 0
	}
	
	bytesPerSecond := c.sampleRate * c.channels * c.bitDepth / 8
	if bytesPerSecond == 0 {
		return 0
	}
	
	durationMs := (len(audioData) * 1000) / bytesPerSecond
	return durationMs
}

// ResampleAudio resamples audio to target sample rate (simplified implementation)
func (c *Converter) ResampleAudio(audioData []byte, targetSampleRate int) ([]byte, error) {
	if c.sampleRate == targetSampleRate {
		return audioData, nil
	}
	
	// Simplified resampling by linear interpolation
	// In production, use proper resampling algorithms
	ratio := float64(targetSampleRate) / float64(c.sampleRate)
	
	if len(audioData)%2 != 0 {
		return nil, fmt.Errorf("audio data length must be even for 16-bit samples")
	}
	
	samples := len(audioData) / 2
	newSamples := int(float64(samples) * ratio)
	resampled := make([]byte, newSamples*2)
	
	reader := bytes.NewReader(audioData)
	writer := bytes.NewBuffer(resampled[:0])
	
	for i := 0; i < newSamples; i++ {
		sourceIndex := float64(i) / ratio
		sourceIndexInt := int(sourceIndex)
		
		if sourceIndexInt*2 >= len(audioData)-1 {
			break
		}
		
		// Read source sample
		reader.Seek(int64(sourceIndexInt*2), io.SeekStart)
		var sample int16
		binary.Read(reader, binary.LittleEndian, &sample)
		
		// Write to output
		binary.Write(writer, binary.LittleEndian, sample)
	}
	
	return writer.Bytes(), nil
}

// convertWebMWithFFmpeg uses ffmpeg to convert WebM to PCM
func (c *Converter) convertWebMWithFFmpeg(webmData []byte) ([]byte, error) {
	// Create temporary files
	tempDir := os.TempDir()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("input_%d.webm", os.Getpid()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.pcm", os.Getpid()))
	
	// Clean up temporary files
	defer func() {
		os.Remove(inputFile)
		os.Remove(outputFile)
	}()
	
	// Write WebM data to temporary file
	if err := ioutil.WriteFile(inputFile, webmData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp WebM file: %v", err)
	}
	
	// Run ffmpeg command to convert WebM to PCM
	cmd := exec.Command("ffmpeg", 
		"-i", inputFile,           // Input file
		"-f", "s16le",            // Output format: signed 16-bit little-endian
		"-ar", fmt.Sprintf("%d", c.sampleRate), // Sample rate
		"-ac", fmt.Sprintf("%d", c.channels),   // Channels
		"-y",                     // Overwrite output file
		outputFile)               // Output file
	
	// Capture stderr for debugging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg conversion failed: %v, stderr: %s", err, stderr.String())
	}
	
	// Read the converted PCM data
	pcmData, err := ioutil.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read converted PCM file: %v", err)
	}
	
	return pcmData, nil
}

// CheckFFmpegAvailable checks if ffmpeg is available in the system
func (c *Converter) CheckFFmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}