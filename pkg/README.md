# PDF Chunk Extractor Library

A Go library for intelligent document chunking with AI-powered text processing. Supports PDF, TXT, and string inputs with flexible output formats.

## Features

- **Multiple Input Types**: PDF files, TXT files, and string content
- **AI-Powered Chunking**: Integration with ChatGPT and extensible AI provider interface
- **Local Fallback**: Intelligent local chunking when AI is unavailable
- **OCR Support**: Automatic OCR for PDFs with no extractable text
- **Flexible Output**: JSON arrays, files, or both
- **Metadata Extraction**: Automatic extraction of document codes, dates, and titles
- **Page Range Detection**: Automatic page range identification
- **Extensible**: Easy to add new AI providers

## Installation

```bash
go get github.com/firdasafridi/pdf-chunk-extractor
```

## Quick Start

### 1. Initialize the Library

```go
package main

import (
    "github.com/firdasafridi/pdf-chunk-extractor/pkg/chunker"
    "github.com/firdasafridi/pdf-chunk-extractor/pkg/config"
    "github.com/firdasafridi/pdf-chunk-extractor/pkg/providers"
)

func main() {
    // Create configuration
    config := config.DefaultConfig()
    
    // Create AI provider (optional)
    apiKey := os.Getenv("OPENAI_API_KEY")
    var aiProvider chunker.AIProvider
    if apiKey != "" {
        aiProvider = providers.NewChatGPTProvider(apiKey)
    }
    
    // Initialize chunker
    chunkerInstance := chunker.NewChunker(config, aiProvider)
}
```

### 2. Process Different Input Types

#### PDF File
```go
// Process PDF file and get JSON array
chunks, err := chunkerInstance.ChunkInput(
    chunker.InputPDF, 
    "path/to/document.pdf", 
    chunker.OutputJSON,
)
```

#### TXT File
```go
// Process TXT file and save to files
chunks, err := chunkerInstance.ChunkInput(
    chunker.InputTXT, 
    "path/to/document.txt", 
    chunker.OutputFile,
)
```

#### String Content
```go
// Process string content and get both JSON and files
textContent := "Your document content here..."
chunks, err := chunkerInstance.ChunkInput(
    chunker.InputString, 
    textContent, 
    chunker.OutputBoth,
)
```

#### PDF Binary Data
```go
// Process PDF from binary data
pdfData, _ := os.ReadFile("document.pdf")
chunks, err := chunkerInstance.ChunkInput(
    chunker.InputPDF, 
    pdfData, 
    chunker.OutputJSON,
)
```

## Configuration

```go
config := config.ChunkerConfig{
    MaxChunkSize:   4000,  // Max characters per AI chunk
    LocalChunkSize: 3000,  // Max characters per local chunk
    OutputDir:      "output",
    ChunkDir:       "chunks",
    JSONDir:        "json",
}
```

## Output Types

### OutputJSON
Returns an array of `ChunkData` structs:

```go
type ChunkData struct {
    Filename   string `json:"filename"`
    ChunkIndex int    `json:"chunk_index"`
    PageRange  string `json:"page_range"`
    Text       string `json:"text"`
}
```

### OutputFile
Saves chunks as text files and JSON files in the configured directories.

### OutputBoth
Returns the JSON array and saves files.

## AI Providers

### ChatGPT Provider
```go
// Basic usage
aiProvider := providers.NewChatGPTProvider("your-api-key")

// With custom configuration
aiProvider := providers.NewChatGPTProviderWithConfig(
    "your-api-key",
    "gpt-4",  // model
    "https://api.openai.com/v1/chat/completions", // URL
)
```

### Custom AI Provider
Implement the `AIProvider` interface:

```go
type CustomAIProvider struct{}

func (c *CustomAIProvider) ChunkText(text string) (string, error) {
    // Your AI logic here
    return processedText, nil
}

func (c *CustomAIProvider) GetName() string {
    return "CustomAI"
}
```

## Input Types

### Supported Input Formats

| Input Type | Supported Formats |
|------------|-------------------|
| `InputPDF` | File path (string), Binary data ([]byte), Reader (io.Reader) |
| `InputTXT` | File path (string), String content (string), Binary data ([]byte), Reader (io.Reader) |
| `InputString` | String content (string), Binary data ([]byte) |

### Examples

```go
// PDF from file path
chunkerInstance.ChunkInput(chunker.InputPDF, "document.pdf", chunker.OutputJSON)

// PDF from binary data
pdfData, _ := os.ReadFile("document.pdf")
chunkerInstance.ChunkInput(chunker.InputPDF, pdfData, chunker.OutputJSON)

// PDF from reader
file, _ := os.Open("document.pdf")
chunkerInstance.ChunkInput(chunker.InputPDF, file, chunker.OutputJSON)

// TXT from file path
chunkerInstance.ChunkInput(chunker.InputTXT, "document.txt", chunker.OutputJSON)

// TXT from string content
chunkerInstance.ChunkInput(chunker.InputTXT, "Your text content", chunker.OutputJSON)

// String content
chunkerInstance.ChunkInput(chunker.InputString, "Your string content", chunker.OutputJSON)
```

## Features

### Intelligent Chunking
- **AI-Powered**: Uses ChatGPT to create meaningful chunks based on content structure
- **Local Fallback**: Intelligent local chunking when AI is unavailable
- **Natural Breaks**: Detects headings, sections, and logical break points
- **Metadata Preservation**: Extracts and preserves document metadata

### PDF Processing
- **Text Extraction**: Direct text extraction from PDFs
- **OCR Fallback**: Automatic OCR for PDFs with no extractable text
- **Page Detection**: Automatic page range identification
- **Multi-language Support**: OCR supports English and Indonesian

### Output Formatting
- **Structured Content**: Clean, formatted output with headers and sections
- **Metadata Extraction**: Document codes, dates, titles, and page ranges
- **Vector Database Ready**: JSON format optimized for vector database embedding

## Error Handling

The library provides comprehensive error handling:

```go
chunks, err := chunkerInstance.ChunkInput(chunker.InputPDF, "document.pdf", chunker.OutputJSON)
if err != nil {
    log.Printf("Error processing document: %v", err)
    return
}
```

## Dependencies

- `github.com/gen2brain/go-fitz`: PDF processing
- `tesseract`: OCR processing (must be installed on system)

## Requirements

- Go 1.24.4 or higher
- Tesseract OCR installed on the system
- OpenAI API key (optional, for AI-powered chunking)

## Examples

See the `examples/` directory for complete usage examples:

- `basic_usage.go`: Basic library usage with different input types
- Custom AI provider implementation
- Error handling examples

## License

This library is licensed under the same license as the main project. 