package main

import (
	"fmt"
	"log"
	"os"

	"github.com/firdasafridi/pdf-chunk-extractor/pkg/chunker"
	"github.com/firdasafridi/pdf-chunk-extractor/pkg/config"
	"github.com/firdasafridi/pdf-chunk-extractor/pkg/providers"
)

func main() {
	// Example 1: Initialize library with ChatGPT AI provider
	fmt.Println("=== Example 1: Using ChatGPT AI Provider ===")

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  OpenAI API key not found. Using local chunking.")
	}

	// Create AI provider
	var aiProvider chunker.AIProvider
	if apiKey != "" {
		aiProvider = providers.NewChatGPTProvider(apiKey)
		fmt.Printf("✅ Using AI provider: %s\n", aiProvider.GetName())
	} else {
		aiProvider = nil // Will use local chunking
		fmt.Println("✅ Using local intelligent chunking")
	}

	// Create configuration
	config := config.DefaultConfig()
	config.OutputDir = "output"
	config.ChunkDir = "chunks"
	config.JSONDir = "json"

	// Initialize chunker
	chunkerInstance := chunker.NewChunker(config, aiProvider)

	// Example 2: Process PDF file with token usage tracking
	fmt.Println("\n=== Example 2: Processing PDF File with Token Usage ===")

	// Input: PDF file path
	pdfPath := "data/13. Panen Kelapa Sawit.pdf"

	// Output: Both JSON array and files with token usage
	result, err := chunkerInstance.ChunkInputWithUsage(chunker.InputPDF, pdfPath, chunker.OutputBoth)
	if err != nil {
		log.Printf("Error processing PDF: %v", err)
	} else {
		fmt.Printf("✅ Successfully processed PDF. Created %d chunks.\n", len(result.Chunks))

		// Print token usage information
		if result.TokenUsage.TotalTokens > 0 {
			fmt.Printf("📊 Token Usage:\n")
			fmt.Printf("   - Prompt Tokens: %d\n", result.TokenUsage.PromptTokens)
			fmt.Printf("   - Completion Tokens: %d\n", result.TokenUsage.CompletionTokens)
			fmt.Printf("   - Total Tokens: %d\n", result.TokenUsage.TotalTokens)
		} else {
			fmt.Println("📊 No token usage (local processing)")
		}

		// Print first chunk info
		if len(result.Chunks) > 0 {
			fmt.Printf("First chunk: %s (Page Range: %s)\n",
				result.Chunks[0].Filename, result.Chunks[0].PageRange)
		}
	}

	// Example 3: Process TXT file with token usage
	fmt.Println("\n=== Example 3: Processing TXT File with Token Usage ===")

	// Input: TXT file path
	txtPath := "data/sample.txt"

	// Output: Only JSON array with token usage
	result, err = chunkerInstance.ChunkInputWithUsage(chunker.InputTXT, txtPath, chunker.OutputJSON)
	if err != nil {
		log.Printf("Error processing TXT: %v", err)
	} else {
		fmt.Printf("✅ Successfully processed TXT. Created %d chunks.\n", len(result.Chunks))

		// Print token usage information
		if result.TokenUsage.TotalTokens > 0 {
			fmt.Printf("📊 Token Usage:\n")
			fmt.Printf("   - Prompt Tokens: %d\n", result.TokenUsage.PromptTokens)
			fmt.Printf("   - Completion Tokens: %d\n", result.TokenUsage.CompletionTokens)
			fmt.Printf("   - Total Tokens: %d\n", result.TokenUsage.TotalTokens)
		} else {
			fmt.Println("📊 No token usage (local processing)")
		}
	}

	// Example 4: Process string content with token usage
	fmt.Println("\n=== Example 4: Processing String Content with Token Usage ===")

	// Input: String content
	textContent := `This is a sample document with multiple sections.

Section 1: Introduction
This is the introduction section of the document.

Section 2: Main Content
This is the main content section with important information.

Section 3: Conclusion
This concludes the document.`

	// Output: Only files with token usage
	result, err = chunkerInstance.ChunkInputWithUsage(chunker.InputString, textContent, chunker.OutputFile)
	if err != nil {
		log.Printf("Error processing string: %v", err)
	} else {
		fmt.Printf("✅ Successfully processed string. Created %d chunks.\n", len(result.Chunks))

		// Print token usage information
		if result.TokenUsage.TotalTokens > 0 {
			fmt.Printf("📊 Token Usage:\n")
			fmt.Printf("   - Prompt Tokens: %d\n", result.TokenUsage.PromptTokens)
			fmt.Printf("   - Completion Tokens: %d\n", result.TokenUsage.CompletionTokens)
			fmt.Printf("   - Total Tokens: %d\n", result.TokenUsage.TotalTokens)
		} else {
			fmt.Println("📊 No token usage (local processing)")
		}
	}

	// Example 5: Process PDF from binary data with token usage
	fmt.Println("\n=== Example 5: Processing PDF from Binary Data with Token Usage ===")

	// Read PDF file as binary
	pdfData, err := os.ReadFile("data/13. Panen Kelapa Sawit.pdf")
	if err != nil {
		log.Printf("Error reading PDF file: %v", err)
	} else {
		// Input: PDF binary data
		result, err = chunkerInstance.ChunkInputWithUsage(chunker.InputPDF, pdfData, chunker.OutputBoth)
		if err != nil {
			log.Printf("Error processing PDF binary: %v", err)
		} else {
			fmt.Printf("✅ Successfully processed PDF binary. Created %d chunks.\n", len(result.Chunks))

			// Print token usage information
			if result.TokenUsage.TotalTokens > 0 {
				fmt.Printf("📊 Token Usage:\n")
				fmt.Printf("   - Prompt Tokens: %d\n", result.TokenUsage.PromptTokens)
				fmt.Printf("   - Completion Tokens: %d\n", result.TokenUsage.CompletionTokens)
				fmt.Printf("   - Total Tokens: %d\n", result.TokenUsage.TotalTokens)
			} else {
				fmt.Println("📊 No token usage (local processing)")
			}
		}
	}

	// Example 6: Compare regular vs usage tracking methods
	fmt.Println("\n=== Example 6: Comparing Regular vs Usage Tracking Methods ===")

	// Regular method (no token usage)
	chunks, err := chunkerInstance.ChunkInput(chunker.InputString, "Simple text content", chunker.OutputJSON)
	if err != nil {
		log.Printf("Error with regular method: %v", err)
	} else {
		fmt.Printf("✅ Regular method: Created %d chunks\n", len(chunks))
	}

	// Usage tracking method
	result, err = chunkerInstance.ChunkInputWithUsage(chunker.InputString, "Simple text content", chunker.OutputJSON)
	if err != nil {
		log.Printf("Error with usage tracking method: %v", err)
	} else {
		fmt.Printf("✅ Usage tracking method: Created %d chunks\n", len(result.Chunks))
		if result.TokenUsage.TotalTokens > 0 {
			fmt.Printf("   Token usage: %d total tokens\n", result.TokenUsage.TotalTokens)
		}
	}

	fmt.Println("\n🎉 All examples completed!")
}

// Example function showing how to create custom AI provider
func createCustomAIProvider() chunker.AIProvider {
	// This is an example of how you could create a custom AI provider
	// that implements the AIProvider interface
	return &CustomAIProvider{}
}

// CustomAIProvider is an example of a custom AI provider
type CustomAIProvider struct{}

func (c *CustomAIProvider) ChunkText(text string) (string, error) {
	// Implement your custom AI logic here
	// This could be any AI service like Claude, Gemini, etc.
	return "Custom AI processed: " + text, nil
}

func (c *CustomAIProvider) GetName() string {
	return "CustomAI"
}
