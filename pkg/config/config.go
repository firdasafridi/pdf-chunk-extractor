package config

// ChunkerConfig holds configuration for the chunker
type ChunkerConfig struct {
	MaxChunkSize   int
	LocalChunkSize int
	OutputDir      string
	ChunkDir       string
	JSONDir        string
}

// DefaultConfig returns a default configuration
func DefaultConfig() ChunkerConfig {
	return ChunkerConfig{
		MaxChunkSize:   4000,
		LocalChunkSize: 3000,
		OutputDir:      "output",
		ChunkDir:       "chunk",
		JSONDir:        "json",
	}
}
