package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"io"

	"github.com/gen2brain/go-fitz"
)

// Configuration constants
const (
	DataDir    = "data"
	OutputDir  = "output"
	ChunkDir   = "chunk"
	TempPrefix = "temp_page_"
	PageSep    = "\n\n--- Page %d ---\n\n"
)

// OpenAI API configuration
const (
	OpenAIAPIURL   = "https://api.openai.com/v1/chat/completions"
	MaxChunkSize   = 4000 // Maximum characters per chunk before sending to AI
	LocalChunkSize = 3000 // Maximum characters for local chunking
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
}

// PDFProcessor handles PDF text extraction with OCR fallback and intelligent chunking
type PDFProcessor struct {
	dataDir   string
	outputDir string
	chunkDir  string
	apiKey    string
	useAI     bool
}

// NewPDFProcessor creates a new PDF processor instance
func NewPDFProcessor(dataDir, outputDir, chunkDir string) *PDFProcessor {
	apiKey := os.Getenv("OPENAI_API_KEY")
	useAI := apiKey != ""

	if !useAI {
		log.Println("⚠️  OpenAI API key not found. Using local intelligent chunking.")
	}

	return &PDFProcessor{
		dataDir:   dataDir,
		outputDir: outputDir,
		chunkDir:  chunkDir,
		apiKey:    apiKey,
		useAI:     useAI,
	}
}

func main() {
	processor := NewPDFProcessor(DataDir, OutputDir, ChunkDir)

	if err := processor.ensureDirectories(); err != nil {
		log.Fatal("Failed to create directories:", err)
	}

	if err := processor.processAllPDFs(); err != nil {
		log.Fatal("Failed to process PDFs:", err)
	}
}

// ensureDirectories creates the output and chunk directories if they don't exist
func (p *PDFProcessor) ensureDirectories() error {
	dirs := []string{p.outputDir, p.chunkDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// processAllPDFs processes all PDF files in the data directory
func (p *PDFProcessor) processAllPDFs() error {
	entries, err := os.ReadDir(p.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	processedCount := 0
	for _, entry := range entries {
		if p.isPDFFile(entry) {
			if err := p.processSinglePDF(entry.Name()); err != nil {
				log.Printf("Error processing %s: %v", entry.Name(), err)
			} else {
				processedCount++
				fmt.Printf("✓ Successfully processed: %s\n", entry.Name())
			}
		}
	}

	fmt.Printf("\n🎉 Processing complete! %d PDF files processed.\n", processedCount)
	return nil
}

// isPDFFile checks if the given entry is a PDF file
func (p *PDFProcessor) isPDFFile(entry os.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf")
}

// processSinglePDF processes a single PDF file
func (p *PDFProcessor) processSinglePDF(filename string) error {
	pdfPath := filepath.Join(p.dataDir, filename)
	outputPath := filepath.Join(p.outputDir, strings.TrimSuffix(filename, ".pdf")+".txt")

	fmt.Printf("\n📄 Processing: %s\n", filename)

	// Extract text from PDF
	if err := p.extractTextFromPDF(pdfPath, outputPath); err != nil {
		return err
	}

	// Create intelligent chunks
	chunkDir := filepath.Join(p.chunkDir, strings.TrimSuffix(filename, ".pdf"))
	if err := p.createIntelligentChunks(outputPath, chunkDir); err != nil {
		return err
	}

	return nil
}

// extractTextFromPDF extracts text from a PDF file with OCR fallback
func (p *PDFProcessor) extractTextFromPDF(pdfPath, outputPath string) error {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	totalPages := doc.NumPage()
	fmt.Printf("   📊 Total pages: %d\n", totalPages)

	for pageIndex := 0; pageIndex < totalPages; pageIndex++ {
		if err := p.processPage(doc, outputFile, pageIndex, totalPages); err != nil {
			log.Printf("   ⚠️  Warning: failed to process page %d: %v", pageIndex+1, err)
		}
	}

	return nil
}

// processPage extracts text from a single page
func (p *PDFProcessor) processPage(doc *fitz.Document, outputFile *os.File, pageIndex, totalPages int) error {
	pageNum := pageIndex + 1

	// Try direct text extraction first
	text, err := doc.Text(pageIndex)
	if err != nil {
		log.Printf("   ⚠️  Warning: failed to extract text from page %d: %v", pageNum, err)
	}

	// If no text found, use OCR
	if strings.TrimSpace(text) == "" {
		fmt.Printf("   🔍 Page %d: No text found, using OCR...\n", pageNum)
		text = p.extractTextWithOCR(doc, pageIndex, pageNum)
	} else {
		fmt.Printf("   ✅ Page %d: extracted %d characters\n", pageNum, len(strings.TrimSpace(text)))
	}

	// Write page separator and content
	p.writePageContent(outputFile, pageIndex, text)
	return nil
}

// extractTextWithOCR uses OCR to extract text from a page image
func (p *PDFProcessor) extractTextWithOCR(doc *fitz.Document, pageIndex, pageNum int) string {
	// Render page as image
	img, err := doc.Image(pageIndex)
	if err != nil {
		log.Printf("   ⚠️  Warning: failed to render page %d as image: %v", pageNum, err)
		return ""
	}

	// Save temporary image
	tempImagePath := fmt.Sprintf("%s%d.png", TempPrefix, pageIndex)
	if err := p.saveTemporaryImage(img, tempImagePath); err != nil {
		log.Printf("   ⚠️  Warning: failed to save temp image: %v", err)
		return ""
	}
	defer os.Remove(tempImagePath)

	// Perform OCR
	ocrText, err := p.runTesseract(tempImagePath)
	if err != nil {
		log.Printf("   ⚠️  Warning: OCR failed for page %d: %v", pageNum, err)
		return ""
	}

	fmt.Printf("   ✅ Page %d: OCR extracted %d characters\n", pageNum, len(strings.TrimSpace(ocrText)))
	return ocrText
}

// saveTemporaryImage saves an image to a temporary file
func (p *PDFProcessor) saveTemporaryImage(img image.Image, tempPath string) error {
	imgFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp image file: %w", err)
	}
	defer imgFile.Close()

	if err := png.Encode(imgFile, img); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

// writePageContent writes page content to the output file
func (p *PDFProcessor) writePageContent(outputFile *os.File, pageIndex int, text string) {
	pageNum := pageIndex + 1
	separator := fmt.Sprintf(PageSep, pageNum)

	outputFile.WriteString(separator)
	outputFile.WriteString(text)
}

// runTesseract executes the tesseract OCR command
func (p *PDFProcessor) runTesseract(imagePath string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "eng+ind")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract command failed: %w", err)
	}

	return string(output), nil
}

// createIntelligentChunks creates intelligent chunks using AI or local processing
func (p *PDFProcessor) createIntelligentChunks(textFilePath, chunkDir string) error {
	// Read the extracted text
	content, err := os.ReadFile(textFilePath)
	if err != nil {
		return fmt.Errorf("failed to read text file: %w", err)
	}

	text := string(content)
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text file is empty")
	}

	// Create chunk directory
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return fmt.Errorf("failed to create chunk directory: %w", err)
	}

	if p.useAI {
		fmt.Printf("   🧠 Creating AI-powered intelligent chunks...\n")
		return p.createAIChunks(text, chunkDir)
	} else {
		fmt.Printf("   🧠 Creating local intelligent chunks...\n")
		return p.createLocalChunks(text, chunkDir)
	}
}

// createAIChunks creates chunks using OpenAI API
func (p *PDFProcessor) createAIChunks(text, chunkDir string) error {
	// Split text into manageable chunks for AI processing
	textChunks := p.splitTextIntoChunks(text)

	chunkIndex := 1
	for i, chunk := range textChunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		fmt.Printf("   📝 Processing chunk %d/%d (%d chars)\n", i+1, len(textChunks), len(chunk))

		// Get intelligent chunk from AI
		intelligentChunk, err := p.getIntelligentChunk(chunk)
		if err != nil {
			log.Printf("   ⚠️  Warning: AI chunking failed for chunk %d: %v", i+1, err)
			// Fallback to local chunking
			intelligentChunk = p.createLocalIntelligentChunk(chunk)
		}

		// Save chunk to file
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d.txt", chunkIndex))
		if err := os.WriteFile(chunkPath, []byte(intelligentChunk), 0644); err != nil {
			log.Printf("   ⚠️  Warning: failed to save chunk %d: %v", chunkIndex, err)
		} else {
			fmt.Printf("   ✅ Saved chunk_%d.txt (%d chars)\n", chunkIndex, len(intelligentChunk))
		}

		chunkIndex++
	}

	fmt.Printf("   🎯 Created %d AI-powered chunks in %s\n", chunkIndex-1, chunkDir)
	return nil
}

// createLocalChunks creates chunks using local intelligent processing
func (p *PDFProcessor) createLocalChunks(text, chunkDir string) error {
	chunks := p.splitTextIntoLocalChunks(text)

	chunkIndex := 1
	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		// Format the chunk with headers and structure
		formattedChunk := p.formatLocalChunk(chunk, i+1, len(chunks))

		// Save chunk to file
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d.txt", chunkIndex))
		if err := os.WriteFile(chunkPath, []byte(formattedChunk), 0644); err != nil {
			log.Printf("   ⚠️  Warning: failed to save chunk %d: %v", chunkIndex, err)
		} else {
			fmt.Printf("   ✅ Saved chunk_%d.txt (%d chars)\n", chunkIndex, len(formattedChunk))
		}

		chunkIndex++
	}

	fmt.Printf("   🎯 Created %d local intelligent chunks in %s\n", chunkIndex-1, chunkDir)
	return nil
}

// splitTextIntoLocalChunks splits text into intelligent chunks based on natural breaks
func (p *PDFProcessor) splitTextIntoLocalChunks(text string) []string {
	var chunks []string
	var currentChunk strings.Builder

	// Split text into lines for processing
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this line is a natural break point
		if p.isNaturalBreak(trimmedLine, i, lines) {
			// If current chunk is getting large, save it and start new one
			if currentChunk.Len() > LocalChunkSize {
				chunk := strings.TrimSpace(currentChunk.String())
				if chunk != "" {
					chunks = append(chunks, chunk)
				}
				currentChunk.Reset()
			}
		}

		// Add the line to current chunk
		currentChunk.WriteString(line + "\n")

		// If chunk is getting too large, force a break
		if currentChunk.Len() > LocalChunkSize {
			chunk := strings.TrimSpace(currentChunk.String())
			if chunk != "" {
				chunks = append(chunks, chunk)
			}
			currentChunk.Reset()
		}
	}

	// Add remaining content
	if currentChunk.Len() > 0 {
		chunk := strings.TrimSpace(currentChunk.String())
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// isNaturalBreak checks if a line represents a natural break point
func (p *PDFProcessor) isNaturalBreak(line string, lineIndex int, allLines []string) bool {
	trimmed := strings.TrimSpace(line)

	// Empty lines are natural breaks
	if trimmed == "" {
		return true
	}

	// Check for various heading patterns
	headingPatterns := []string{
		`^Bab\s+\d+`,         // Bab 1, Bab 2, etc.
		`^Pasal\s+\d+`,       // Pasal 1, Pasal 2, etc.
		`^Chapter\s+\d+`,     // Chapter 1, Chapter 2, etc.
		`^Section\s+\d+`,     // Section 1, Section 2, etc.
		`^Artikel\s+\d+`,     // Artikel 1, Artikel 2, etc.
		`^BAB\s+\d+`,         // BAB 1, BAB 2, etc.
		`^PASAL\s+\d+`,       // PASAL 1, PASAL 2, etc.
		`^\d+\.\s+[A-Z]`,     // 1. Title, 2. Title, etc.
		`^[A-Z][A-Z\s]{3,}$`, // ALL CAPS HEADINGS
		`^[A-Z][a-z\s]{3,}$`, // Title Case Headings
	}

	for _, pattern := range headingPatterns {
		if matched, _ := regexp.MatchString(pattern, trimmed); matched {
			return true
		}
	}

	// Check for bullet points or numbered lists
	if strings.HasPrefix(trimmed, "•") || strings.HasPrefix(trimmed, "-") ||
		strings.HasPrefix(trimmed, "*") {
		return true
	}

	// Check for numbered lists
	if matched, _ := regexp.MatchString(`^\d+\.`, trimmed); matched {
		return true
	}

	// Check for page separators
	if strings.Contains(trimmed, "--- Page") {
		return true
	}

	// Check if previous line was empty and this line looks like a heading
	if lineIndex > 0 && strings.TrimSpace(allLines[lineIndex-1]) == "" {
		if len(trimmed) < 100 && (strings.ToUpper(trimmed) == trimmed ||
			strings.HasSuffix(trimmed, ":") || strings.HasSuffix(trimmed, ".")) {
			return true
		}
	}

	return false
}

// formatLocalChunk formats a chunk with headers and structure
func (p *PDFProcessor) formatLocalChunk(chunk string, chunkNum, totalChunks int) string {
	var formatted strings.Builder

	// Add chunk header
	formatted.WriteString(fmt.Sprintf("# Chunk %d of %d\n\n", chunkNum, totalChunks))

	// Extract and format document metadata if present
	metadata := p.extractMetadata(chunk)
	if metadata != "" {
		formatted.WriteString("## Document Information\n")
		formatted.WriteString(metadata + "\n\n")
	}

	// Format the main content
	formatted.WriteString("## Content\n\n")
	formatted.WriteString(chunk)

	return formatted.String()
}

// extractMetadata extracts document metadata from the chunk
func (p *PDFProcessor) extractMetadata(chunk string) string {
	var metadata strings.Builder

	// Look for document codes
	docCodePattern := regexp.MustCompile(`(SOP|KCN|AGR|KEP|PER|UU|PP|PMK)[/-][A-Z0-9/]+`)
	if matches := docCodePattern.FindAllString(chunk, -1); len(matches) > 0 {
		metadata.WriteString(fmt.Sprintf("- **Document Code**: %s\n", strings.Join(matches, ", ")))
	}

	// Look for dates
	datePattern := regexp.MustCompile(`(\d{1,2}\s+[-–]\s+[A-Za-z]+\s+[-–]\s+\d{4})`)
	if matches := datePattern.FindAllString(chunk, -1); len(matches) > 0 {
		metadata.WriteString(fmt.Sprintf("- **Date**: %s\n", strings.Join(matches, ", ")))
	}

	// Look for page numbers
	pagePattern := regexp.MustCompile(`Page\s+(\d+)`)
	if matches := pagePattern.FindAllString(chunk, -1); len(matches) > 0 {
		metadata.WriteString(fmt.Sprintf("- **Pages**: %s\n", strings.Join(matches, ", ")))
	}

	return metadata.String()
}

// createLocalIntelligentChunk creates a local intelligent chunk (fallback for AI)
func (p *PDFProcessor) createLocalIntelligentChunk(text string) string {
	chunks := p.splitTextIntoLocalChunks(text)
	if len(chunks) == 0 {
		return text
	}

	// Return the first chunk (since this is called for individual chunks)
	return p.formatLocalChunk(chunks[0], 1, 1)
}

// splitTextIntoChunks splits text into manageable chunks for AI processing
func (p *PDFProcessor) splitTextIntoChunks(text string) []string {
	var chunks []string
	lines := strings.Split(text, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		currentChunk.WriteString(line + "\n")

		// If chunk is getting too large, split it
		if currentChunk.Len() > MaxChunkSize {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}
	}

	// Add remaining content
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// getIntelligentChunk uses OpenAI API to create intelligent chunks
func (p *PDFProcessor) getIntelligentChunk(text string) (string, error) {
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
		Model: "gpt-3.5-turbo",
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

	response, err := p.callOpenAIAPI(request)
	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return response.Choices[0].Message.Content, nil
}

// callOpenAIAPI makes a request to the OpenAI API
func (p *PDFProcessor) callOpenAIAPI(request OpenAIRequest) (*OpenAIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", OpenAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
