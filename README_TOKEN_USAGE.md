# Token Usage Tracking - PDF Chunk Extractor Library

The library now includes comprehensive token usage tracking to help you monitor AI API costs and usage patterns.

## Overview

Token usage tracking provides detailed information about:
- **Prompt Tokens**: Tokens used in the input/prompt
- **Completion Tokens**: Tokens used in the AI response
- **Total Tokens**: Sum of prompt and completion tokens

This information is essential for:
- Cost monitoring and budgeting
- Usage analytics
- Performance optimization
- API quota management

## Usage

### Basic Token Usage Tracking

```go
package main

import (
    "github.com/firdasafridi/pdf-chunk-extractor/pkg/chunker"
    "github.com/firdasafridi/pdf-chunk-extractor/pkg/config"
    "github.com/firdasafridi/pdf-chunk-extractor/pkg/providers"
)

func main() {
    // Initialize with ChatGPT provider
    apiKey := os.Getenv("OPENAI_API_KEY")
    aiProvider := providers.NewChatGPTProvider(apiKey)
    config := config.DefaultConfig()
    chunkerInstance := chunker.NewChunker(config, aiProvider)

    // Process with token usage tracking
    result, err := chunkerInstance.ChunkInputWithUsage(
        chunker.InputPDF, 
        "document.pdf", 
        chunker.OutputJSON,
    )
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Access token usage information
    fmt.Printf("Total Tokens: %d\n", result.TokenUsage.TotalTokens)
    fmt.Printf("Prompt Tokens: %d\n", result.TokenUsage.PromptTokens)
    fmt.Printf("Completion Tokens: %d\n", result.TokenUsage.CompletionTokens)
    
    // Access chunks
    for _, chunk := range result.Chunks {
        fmt.Printf("Chunk %d: %s\n", chunk.ChunkIndex, chunk.Filename)
    }
}
```

### Token Usage Structure

```go
type TokenUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

type ChunkResult struct {
    Chunks     []ChunkData `json:"chunks"`
    TokenUsage TokenUsage  `json:"token_usage"`
}
```

### Methods Comparison

| Method | Returns | Token Usage |
|--------|---------|-------------|
| `ChunkInput()` | `[]ChunkData` | ❌ No |
| `ChunkInputWithUsage()` | `*ChunkResult` | ✅ Yes |

## Examples

### Example 1: Process PDF with Token Tracking

```go
// Process PDF and get token usage
result, err := chunkerInstance.ChunkInputWithUsage(
    chunker.InputPDF, 
    "document.pdf", 
    chunker.OutputBoth,
)

if err != nil {
    log.Fatal(err)
}

// Print results
fmt.Printf("Created %d chunks\n", len(result.Chunks))
fmt.Printf("Token Usage:\n")
fmt.Printf("  - Prompt: %d tokens\n", result.TokenUsage.PromptTokens)
fmt.Printf("  - Completion: %d tokens\n", result.TokenUsage.CompletionTokens)
fmt.Printf("  - Total: %d tokens\n", result.TokenUsage.TotalTokens)
```

### Example 2: Cost Calculation

```go
func calculateCost(result *chunker.ChunkResult) float64 {
    // OpenAI GPT-3.5-turbo pricing (as of 2024)
    const (
        inputCostPer1kTokens  = 0.0015  // $0.0015 per 1K input tokens
        outputCostPer1kTokens = 0.002   // $0.002 per 1K output tokens
    )
    
    inputCost := float64(result.TokenUsage.PromptTokens) / 1000 * inputCostPer1kTokens
    outputCost := float64(result.TokenUsage.CompletionTokens) / 1000 * outputCostPer1kTokens
    
    return inputCost + outputCost
}

// Usage
result, _ := chunkerInstance.ChunkInputWithUsage(chunker.InputPDF, "doc.pdf", chunker.OutputJSON)
cost := calculateCost(result)
fmt.Printf("Estimated cost: $%.4f\n", cost)
```

### Example 3: Batch Processing with Usage Tracking

```go
func processMultipleDocuments(chunkerInstance *chunker.Chunker, files []string) {
    var totalTokens int
    var totalChunks int
    
    for _, file := range files {
        result, err := chunkerInstance.ChunkInputWithUsage(
            chunker.InputPDF, 
            file, 
            chunker.OutputJSON,
        )
        
        if err != nil {
            log.Printf("Error processing %s: %v", file, err)
            continue
        }
        
        totalTokens += result.TokenUsage.TotalTokens
        totalChunks += len(result.Chunks)
        
        fmt.Printf("Processed %s: %d chunks, %d tokens\n", 
            file, len(result.Chunks), result.TokenUsage.TotalTokens)
    }
    
    fmt.Printf("Total: %d chunks, %d tokens\n", totalChunks, totalTokens)
}
```

### Example 4: Usage Analytics

```go
type UsageStats struct {
    TotalFiles    int
    TotalChunks   int
    TotalTokens   int
    AvgTokensPerChunk float64
    CostEstimate  float64
}

func analyzeUsage(results []*chunker.ChunkResult) UsageStats {
    var stats UsageStats
    
    for _, result := range results {
        stats.TotalFiles++
        stats.TotalChunks += len(result.Chunks)
        stats.TotalTokens += result.TokenUsage.TotalTokens
    }
    
    if stats.TotalChunks > 0 {
        stats.AvgTokensPerChunk = float64(stats.TotalTokens) / float64(stats.TotalChunks)
    }
    
    // Estimate cost (GPT-3.5-turbo pricing)
    stats.CostEstimate = float64(stats.TotalTokens) / 1000 * 0.002
    
    return stats
}
```

## Token Usage Scenarios

### When Tokens Are Tracked

- ✅ Using `ChunkInputWithUsage()` method
- ✅ AI provider supports usage tracking (ChatGPT)
- ✅ AI processing is successful

### When Tokens Are Not Tracked

- ❌ Using regular `ChunkInput()` method
- ❌ Local chunking (no AI provider)
- ❌ AI provider doesn't support usage tracking
- ❌ AI processing fails (falls back to local)

### Example Output

```json
{
  "chunks": [
    {
      "filename": "document.pdf",
      "chunk_index": 1,
      "page_range": "Page 1-3",
      "text": "# Document Chunk\n\n## Metadata\n..."
    }
  ],
  "token_usage": {
    "prompt_tokens": 245,
    "completion_tokens": 156,
    "total_tokens": 401
  }
}
```

## Cost Optimization Tips

### 1. Monitor Token Usage

```go
// Set up monitoring
func monitorTokenUsage(result *chunker.ChunkResult) {
    if result.TokenUsage.TotalTokens > 1000 {
        log.Printf("⚠️  High token usage: %d tokens", result.TokenUsage.TotalTokens)
    }
}
```

### 2. Optimize Chunk Sizes

```go
// Adjust configuration for cost optimization
config := config.DefaultConfig()
config.MaxChunkSize = 2000  // Smaller chunks = fewer tokens
config.LocalChunkSize = 1500
```

### 3. Use Local Fallback

```go
// When AI is expensive, use local processing
if shouldUseLocalProcessing() {
    // Use regular ChunkInput (no AI)
    chunks, _ := chunkerInstance.ChunkInput(chunker.InputPDF, "doc.pdf", chunker.OutputJSON)
} else {
    // Use AI with usage tracking
    result, _ := chunkerInstance.ChunkInputWithUsage(chunker.InputPDF, "doc.pdf", chunker.OutputJSON)
}
```

## Integration with Monitoring Systems

### Prometheus Metrics

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    tokensUsed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_tokens_used_total",
            Help: "Total AI tokens used",
        },
        []string{"provider", "operation"},
    )
)

func recordTokenUsage(result *chunker.ChunkResult) {
    tokensUsed.WithLabelValues("chatgpt", "chunking").Add(float64(result.TokenUsage.TotalTokens))
}
```

### Logging

```go
import "log"

func logTokenUsage(result *chunker.ChunkResult, filename string) {
    log.Printf("Token usage for %s: %d total tokens (prompt: %d, completion: %d)", 
        filename, 
        result.TokenUsage.TotalTokens,
        result.TokenUsage.PromptTokens,
        result.TokenUsage.CompletionTokens,
    )
}
```

## Best Practices

1. **Always use `ChunkInputWithUsage()`** when cost monitoring is important
2. **Set up alerts** for high token usage
3. **Monitor costs regularly** and adjust chunk sizes accordingly
4. **Use local processing** for non-critical documents
5. **Cache results** to avoid reprocessing the same content
6. **Batch process** documents to get better cost estimates

## Troubleshooting

### No Token Usage Data

If you're not getting token usage data:

1. Check if you're using `ChunkInputWithUsage()` method
2. Verify your AI provider supports usage tracking
3. Ensure your API key is valid
4. Check if AI processing is actually being used (not falling back to local)

### High Token Usage

If token usage is unexpectedly high:

1. Reduce `MaxChunkSize` in configuration
2. Use local processing for simple documents
3. Optimize your prompts (if custom AI providers)
4. Consider batching smaller documents together

The token usage tracking feature provides valuable insights for cost management and optimization of your AI-powered document processing pipeline. 