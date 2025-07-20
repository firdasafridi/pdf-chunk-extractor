package chunker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/firdasafridi/pdf-chunk-extractor/pkg/config"
	"github.com/firdasafridi/pdf-chunk-extractor/pkg/processor"
	"github.com/firdasafridi/pdf-chunk-extractor/pkg/providers"
	"github.com/firdasafridi/pdf-chunk-extractor/pkg/utils"
)

// ChunkData represents a structured chunk for vector database embedding
type ChunkData struct {
	Filename   string `json:"filename"`
	ChunkIndex int    `json:"chunk_index"`
	PageRange  string `json:"page_range"`
	Text       string `json:"text"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChunkResult represents the result of chunking with token usage
type ChunkResult struct {
	Chunks     []ChunkData `json:"chunks"`
	TokenUsage TokenUsage  `json:"token_usage"`
}

// InputType represents the type of input data
type InputType int

const (
	InputPDF InputType = iota
	InputTXT
	InputString
)

// OutputType represents the type of output format
type OutputType int

const (
	OutputJSON OutputType = iota
	OutputFile
	OutputBoth
)

// AIProvider represents different AI providers for chunking
type AIProvider interface {
	ChunkText(text string) (string, error)
	GetName() string
}

// AIProviderWithUsage represents AI providers that can track token usage
type AIProviderWithUsage interface {
	AIProvider
	ChunkTextWithUsage(text string) (*providers.ChunkResult, error)
}

// Chunker is the main library interface
type Chunker struct {
	config        config.ChunkerConfig
	aiProvider    AIProvider
	pdfProcessor  *processor.PDFProcessor
	textProcessor *utils.TextProcessor
}

// NewChunker creates a new chunker instance
func NewChunker(config config.ChunkerConfig, aiProvider AIProvider) *Chunker {
	return &Chunker{
		config:        config,
		aiProvider:    aiProvider,
		pdfProcessor:  processor.NewPDFProcessor(config),
		textProcessor: utils.NewTextProcessor(config.MaxChunkSize, config.LocalChunkSize),
	}
}

// ChunkInput processes input data and returns chunks based on output type
func (c *Chunker) ChunkInput(inputType InputType, input interface{}, outputType OutputType) ([]ChunkData, error) {
	var text string
	var filename string

	// Process input based on type
	switch inputType {
	case InputPDF:
		text, filename = c.processPDFInput(input)
	case InputTXT:
		text, filename = c.processTXTInput(input)
	case InputString:
		text, filename = c.processStringInput(input)
	default:
		return nil, fmt.Errorf("unsupported input type: %v", inputType)
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("input text is empty")
	}

	// Create chunks
	chunks, err := c.createChunks(text, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunks: %w", err)
	}

	// Handle output based on type
	switch outputType {
	case OutputJSON:
		return chunks, nil
	case OutputFile:
		return chunks, c.saveChunksToFiles(chunks, filename)
	case OutputBoth:
		if err := c.saveChunksToFiles(chunks, filename); err != nil {
			return nil, fmt.Errorf("failed to save chunks to files: %w", err)
		}
		return chunks, nil
	default:
		return nil, fmt.Errorf("unsupported output type: %v", outputType)
	}
}

// ChunkInputWithUsage processes input data and returns chunks with token usage information
func (c *Chunker) ChunkInputWithUsage(inputType InputType, input interface{}, outputType OutputType) (*ChunkResult, error) {
	var text string
	var filename string

	// Process input based on type
	switch inputType {
	case InputPDF:
		text, filename = c.processPDFInput(input)
	case InputTXT:
		text, filename = c.processTXTInput(input)
	case InputString:
		text, filename = c.processStringInput(input)
	default:
		return nil, fmt.Errorf("unsupported input type: %v", inputType)
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("input text is empty")
	}

	// Create chunks with usage tracking
	chunks, tokenUsage, err := c.createChunksWithUsage(text, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunks: %w", err)
	}

	// Handle output based on type
	switch outputType {
	case OutputJSON:
		return &ChunkResult{Chunks: chunks, TokenUsage: tokenUsage}, nil
	case OutputFile:
		if err := c.saveChunksToFiles(chunks, filename); err != nil {
			return nil, fmt.Errorf("failed to save chunks to files: %w", err)
		}
		return &ChunkResult{Chunks: chunks, TokenUsage: tokenUsage}, nil
	case OutputBoth:
		if err := c.saveChunksToFiles(chunks, filename); err != nil {
			return nil, fmt.Errorf("failed to save chunks to files: %w", err)
		}
		return &ChunkResult{Chunks: chunks, TokenUsage: tokenUsage}, nil
	default:
		return nil, fmt.Errorf("unsupported output type: %v", outputType)
	}
}

// processPDFInput handles PDF input (file path or binary data)
func (c *Chunker) processPDFInput(input interface{}) (string, string) {
	switch v := input.(type) {
	case string:
		// File path
		filename := filepath.Base(v)
		text, err := c.pdfProcessor.ExtractTextFromPDFPath(v)
		if err != nil {
			return "", filename
		}
		return text, filename
	case []byte:
		// Binary data
		filename := "input.pdf"
		text, err := c.pdfProcessor.ExtractTextFromPDFBytes(v)
		if err != nil {
			return "", filename
		}
		return text, filename
	case io.Reader:
		// Reader
		filename := "input.pdf"
		text, err := c.pdfProcessor.ExtractTextFromPDFReader(v)
		if err != nil {
			return "", filename
		}
		return text, filename
	default:
		return "", "unknown.pdf"
	}
}

// processTXTInput handles TXT input (file path or string content)
func (c *Chunker) processTXTInput(input interface{}) (string, string) {
	switch v := input.(type) {
	case string:
		// Check if it's a file path
		if _, err := os.Stat(v); err == nil {
			// File path
			filename := filepath.Base(v)
			content, err := os.ReadFile(v)
			if err != nil {
				return "", filename
			}
			return string(content), filename
		} else {
			// String content
			return v, "input.txt"
		}
	case []byte:
		// Binary data
		return string(v), "input.txt"
	case io.Reader:
		// Reader
		filename := "input.txt"
		content, err := io.ReadAll(v)
		if err != nil {
			return "", filename
		}
		return string(content), filename
	default:
		return "", "unknown.txt"
	}
}

// processStringInput handles string input
func (c *Chunker) processStringInput(input interface{}) (string, string) {
	switch v := input.(type) {
	case string:
		return v, "input.txt"
	case []byte:
		return string(v), "input.txt"
	default:
		return "", "unknown.txt"
	}
}

// createChunks creates intelligent chunks using AI or local processing
func (c *Chunker) createChunks(text, filename string) ([]ChunkData, error) {
	if c.aiProvider != nil {
		return c.createAIChunks(text, filename)
	} else {
		return c.createLocalChunks(text, filename)
	}
}

// createChunksWithUsage creates intelligent chunks with token usage tracking
func (c *Chunker) createChunksWithUsage(text, filename string) ([]ChunkData, TokenUsage, error) {
	if c.aiProvider != nil {
		return c.createAIChunksWithUsage(text, filename)
	} else {
		chunks, err := c.createLocalChunks(text, filename)
		return chunks, TokenUsage{}, err
	}
}

// createAIChunks creates chunks using AI provider
func (c *Chunker) createAIChunks(text, filename string) ([]ChunkData, error) {
	// Split text into manageable chunks for AI processing
	textChunks := c.textProcessor.SplitTextIntoChunks(text)
	var chunks []ChunkData

	for i, chunk := range textChunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		// Get intelligent chunk from AI
		intelligentChunk, err := c.aiProvider.ChunkText(chunk)
		if err != nil {
			// Fallback to local chunking
			intelligentChunk = c.textProcessor.CreateLocalIntelligentChunk(chunk)
		}

		// Create chunk data
		chunkData := ChunkData{
			Filename:   filename,
			ChunkIndex: i + 1,
			PageRange:  c.textProcessor.ExtractPageRange(chunk),
			Text:       intelligentChunk,
		}

		chunks = append(chunks, chunkData)
	}

	return chunks, nil
}

// createAIChunksWithUsage creates chunks using AI provider with token usage tracking
func (c *Chunker) createAIChunksWithUsage(text, filename string) ([]ChunkData, TokenUsage, error) {
	// Split text into manageable chunks for AI processing
	textChunks := c.textProcessor.SplitTextIntoChunks(text)
	var chunks []ChunkData
	var totalTokenUsage TokenUsage

	// Check if AI provider supports usage tracking
	aiProviderWithUsage, ok := c.aiProvider.(AIProviderWithUsage)
	if !ok {
		// Fallback to regular AI chunking
		chunks, err := c.createAIChunks(text, filename)
		return chunks, TokenUsage{}, err
	}

	for i, chunk := range textChunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		// Get intelligent chunk from AI with usage tracking
		result, err := aiProviderWithUsage.ChunkTextWithUsage(chunk)
		if err != nil {
			// Fallback to local chunking
			intelligentChunk := c.textProcessor.CreateLocalIntelligentChunk(chunk)
			chunkData := ChunkData{
				Filename:   filename,
				ChunkIndex: i + 1,
				PageRange:  c.textProcessor.ExtractPageRange(chunk),
				Text:       intelligentChunk,
			}
			chunks = append(chunks, chunkData)
		} else {
			// Add token usage to total
			totalTokenUsage.PromptTokens += result.TokenUsage.PromptTokens
			totalTokenUsage.CompletionTokens += result.TokenUsage.CompletionTokens
			totalTokenUsage.TotalTokens += result.TokenUsage.TotalTokens

			// Create chunk data
			chunkData := ChunkData{
				Filename:   filename,
				ChunkIndex: i + 1,
				PageRange:  c.textProcessor.ExtractPageRange(chunk),
				Text:       result.Text,
			}

			chunks = append(chunks, chunkData)
		}
	}

	return chunks, totalTokenUsage, nil
}

// createLocalChunks creates chunks using local intelligent processing
func (c *Chunker) createLocalChunks(text, filename string) ([]ChunkData, error) {
	chunks := c.textProcessor.SplitTextIntoLocalChunks(text)
	var chunkData []ChunkData

	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		// Format the chunk with headers and structure
		formattedChunk := c.textProcessor.FormatLocalChunk(chunk, i+1, len(chunks))

		// Create chunk data
		data := ChunkData{
			Filename:   filename,
			ChunkIndex: i + 1,
			PageRange:  c.textProcessor.ExtractPageRange(chunk),
			Text:       formattedChunk,
		}

		chunkData = append(chunkData, data)
	}

	return chunkData, nil
}

// saveChunksToFiles saves chunks to files
func (c *Chunker) saveChunksToFiles(chunks []ChunkData, filename string) error {
	// Ensure directories exist
	if err := c.ensureDirectories(); err != nil {
		return err
	}

	// Create chunk directory for this file
	chunkDir := filepath.Join(c.config.ChunkDir, strings.TrimSuffix(filename, filepath.Ext(filename)))
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return fmt.Errorf("failed to create chunk directory: %w", err)
	}

	// Save each chunk
	for _, chunk := range chunks {
		// Save text chunk
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d.txt", chunk.ChunkIndex))
		if err := os.WriteFile(chunkPath, []byte(chunk.Text), 0644); err != nil {
			return fmt.Errorf("failed to save chunk %d: %w", chunk.ChunkIndex, err)
		}

		// Save JSON chunk
		if err := c.saveJSONChunk(chunk); err != nil {
			return fmt.Errorf("failed to save JSON chunk %d: %w", chunk.ChunkIndex, err)
		}
	}

	return nil
}

// ensureDirectories creates the output and chunk directories if they don't exist
func (c *Chunker) ensureDirectories() error {
	dirs := []string{c.config.OutputDir, c.config.ChunkDir, c.config.JSONDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// saveJSONChunk creates a JSON object for vector database embedding
func (c *Chunker) saveJSONChunk(chunk ChunkData) error {
	return c.textProcessor.SaveJSONChunk(chunk, c.config.JSONDir, chunk.Filename, chunk.ChunkIndex)
}
