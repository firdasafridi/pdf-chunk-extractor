package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model     string          `json:"model"`
	Messages  []OpenAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

// OpenAIMessage represents a message in the OpenAI API
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response structure from OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChunkResult represents the result of chunking with token usage
type ChunkResult struct {
	Text       string     `json:"text"`
	TokenUsage TokenUsage `json:"token_usage"`
}

// ChatGPTProvider implements AIProvider for OpenAI's ChatGPT
type ChatGPTProvider struct {
	apiKey string
	model  string
	url    string
}

// NewChatGPTProvider creates a new ChatGPT provider
func NewChatGPTProvider(apiKey string) *ChatGPTProvider {
	return &ChatGPTProvider{
		apiKey: apiKey,
		model:  "gpt-3.5-turbo",
		url:    "https://api.openai.com/v1/chat/completions",
	}
}

// NewChatGPTProviderWithConfig creates a new ChatGPT provider with custom configuration
func NewChatGPTProviderWithConfig(apiKey, model, url string) *ChatGPTProvider {
	if url == "" {
		url = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &ChatGPTProvider{
		apiKey: apiKey,
		model:  model,
		url:    url,
	}
}

// ChunkText uses ChatGPT to create intelligent chunks
func (c *ChatGPTProvider) ChunkText(text string) (string, error) {
	result, err := c.ChunkTextWithUsage(text)
	if err != nil {
		return "", err
	}
	return result.Text, nil
}

// ChunkTextWithUsage uses ChatGPT to create intelligent chunks and returns token usage
func (c *ChatGPTProvider) ChunkTextWithUsage(text string) (*ChunkResult, error) {
	prompt := `You are an AI system optimizing document processing. If the chunking below fails or produces low-quality results, please gracefully degrade by returning the original text as fallback. Always include metadata like page numbers, chunk index, and document title in the output.

Your task is to chunk the provided text into meaningful, coherent sections based on themes, topics, or logical flow.

Please analyze the text and create a well-structured chunk that:
1. Groups related content together
2. Maintains logical flow and context
3. Includes relevant metadata when available (document codes, dates, etc.)
4. Preserves important formatting and structure
5. Makes the content easy to understand and navigate
6. Always includes page numbers, chunk index, and document title in the output
7. If chunking fails or produces poor results, return the original text with basic formatting

IMPORTANT: If you cannot create a meaningful chunk or the result would be worse than the original, simply return the original text with basic headers and metadata extraction.

Text to chunk:
` + text + `

Please return the chunked content with appropriate headers, sections, and formatting to make it clear and organized. If chunking is not beneficial, return the original text with basic structure.`

	request := OpenAIRequest{
		Model: c.model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are an AI system optimizing document processing with intelligent chunking capabilities. You excel at organizing and structuring text content for better readability and understanding. Always prioritize preserving meaning and context over aggressive restructuring. If chunking would degrade the content quality, gracefully fall back to the original text with basic formatting and metadata extraction.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 2000,
	}

	response, err := c.callAPI(request)
	if err != nil {
		return nil, fmt.Errorf("ChatGPT API call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from ChatGPT API")
	}

	return &ChunkResult{
		Text: response.Choices[0].Message.Content,
		TokenUsage: TokenUsage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}, nil
}

// GetName returns the provider name
func (c *ChatGPTProvider) GetName() string {
	return "ChatGPT"
}

// callAPI makes a request to the ChatGPT API
func (c *ChatGPTProvider) callAPI(request OpenAIRequest) (*OpenAIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
