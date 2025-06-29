package service

import (
	"fmt"
	"github.com/ai-tutor-monorepo/services/speech-service/internal/config"
	"github.com/ai-tutor-monorepo/services/speech-service/pkg/audio"
	"github.com/sirupsen/logrus"
)

// AudioService handles audio processing operations
type AudioService struct {
	converter *audio.Converter
	config    *config.AudioConfig
	logger    *logrus.Logger
}

// NewAudioService creates a new audio service
func NewAudioService(cfg *config.AudioConfig, logger *logrus.Logger) *AudioService {
	converter := audio.NewConverter(cfg.SampleRate, cfg.Channels, cfg.BitDepth)
	
	// Check if ffmpeg is available
	if converter.CheckFFmpegAvailable() {
		logger.Infof("✅ ffmpeg detected - using real WebM conversion")
	} else {
		logger.Warnf("⚠️ ffmpeg not found - using simulated conversion (will produce noise)")
	}
	
	return &AudioService{
		converter: converter,
		config:    cfg,
		logger:    logger,
	}
}

// ConvertWebMToPCM converts WebM audio data to PCM format
func (s *AudioService) ConvertWebMToPCM(webmData []byte) ([]byte, error) {
	if len(webmData) == 0 {
		return nil, fmt.Errorf("empty WebM audio data")
	}
	
	s.logger.Debugf("Converting WebM audio data (%d bytes) to PCM", len(webmData))
	
	pcmData, err := s.converter.ConvertWebMToPCM(webmData)
	if err != nil {
		return nil, fmt.Errorf("WebM to PCM conversion failed: %v", err)
	}
	
	s.logger.Debugf("Converted to PCM: %d bytes", len(pcmData))
	return pcmData, nil
}

// ConvertPCMToWAV converts PCM data to WAV format
func (s *AudioService) ConvertPCMToWAV(pcmData []byte) ([]byte, error) {
	if len(pcmData) == 0 {
		return nil, fmt.Errorf("empty PCM audio data")
	}
	
	s.logger.Debugf("Converting PCM audio data (%d bytes) to WAV", len(pcmData))
	
	wavData, err := s.converter.ConvertPCMToWAV(pcmData)
	if err != nil {
		return nil, fmt.Errorf("PCM to WAV conversion failed: %v", err)
	}
	
	s.logger.Debugf("Converted to WAV: %d bytes", len(wavData))
	return wavData, nil
}

// ProcessAudioChunk processes a chunk of audio data
func (s *AudioService) ProcessAudioChunk(audioData []byte) ([]byte, error) {
	// Validate audio data
	if err := s.converter.ValidateAudioData(audioData); err != nil {
		return nil, fmt.Errorf("audio validation failed: %v", err)
	}
	
	// Convert WebM to PCM
	pcmData, err := s.ConvertWebMToPCM(audioData)
	if err != nil {
		return nil, err
	}
	
	// Convert PCM to WAV for ASR API compatibility
	wavData, err := s.ConvertPCMToWAV(pcmData)
	if err != nil {
		return nil, err
	}
	
	duration := s.converter.GetAudioDuration(pcmData)
	s.logger.Debugf("Processed audio chunk: %d ms duration", duration)
	
	return wavData, nil
}

// SplitAudioIntoChunks splits large audio data into smaller chunks
func (s *AudioService) SplitAudioIntoChunks(audioData []byte) [][]byte {
	return s.converter.ExtractAudioChunks(audioData, s.config.ChunkSize)
}

// GetAudioDuration returns the duration of audio data in milliseconds
func (s *AudioService) GetAudioDuration(audioData []byte) int {
	return s.converter.GetAudioDuration(audioData)
}

// ValidateAudioData validates audio data format and size
func (s *AudioService) ValidateAudioData(audioData []byte) error {
	return s.converter.ValidateAudioData(audioData)
}

// OptimizeAudioForASR optimizes audio data for ASR processing
func (s *AudioService) OptimizeAudioForASR(audioData []byte) ([]byte, error) {
	// Ensure audio meets ASR requirements (16kHz, mono, 16-bit)
	if err := s.ValidateAudioData(audioData); err != nil {
		return nil, err
	}
	
	// Convert to PCM first
	pcmData, err := s.ConvertWebMToPCM(audioData)
	if err != nil {
		return nil, err
	}
	
	// Resample if needed (though our target is already 16kHz)
	if s.config.SampleRate != 16000 {
		s.logger.Warnf("Resampling audio from %d Hz to 16000 Hz for ASR", s.config.SampleRate)
		pcmData, err = s.converter.ResampleAudio(pcmData, 16000)
		if err != nil {
			return nil, fmt.Errorf("resampling failed: %v", err)
		}
	}
	
	// Convert to WAV format for ASR API
	return s.ConvertPCMToWAV(pcmData)
}