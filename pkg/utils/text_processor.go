package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TextProcessor handles text chunking and formatting
type TextProcessor struct {
	maxChunkSize   int
	localChunkSize int
}

// NewTextProcessor creates a new text processor
func NewTextProcessor(maxChunkSize, localChunkSize int) *TextProcessor {
	return &TextProcessor{
		maxChunkSize:   maxChunkSize,
		localChunkSize: localChunkSize,
	}
}

// SplitTextIntoChunks splits text into manageable chunks for AI processing
func (t *TextProcessor) SplitTextIntoChunks(text string) []string {
	var chunks []string
	lines := strings.Split(text, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		currentChunk.WriteString(line + "\n")

		// If chunk is getting too large, split it
		if currentChunk.Len() > t.maxChunkSize {
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

// SplitTextIntoLocalChunks splits text into intelligent chunks based on natural breaks
func (t *TextProcessor) SplitTextIntoLocalChunks(text string) []string {
	var chunks []string
	var currentChunk strings.Builder

	// Split text into lines for processing
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this line is a natural break point
		if t.isNaturalBreak(trimmedLine, i, lines) {
			// If current chunk is getting large, save it and start new one
			if currentChunk.Len() > t.localChunkSize {
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
		if currentChunk.Len() > t.localChunkSize {
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

// IsNaturalBreak checks if a line represents a natural break point
func (t *TextProcessor) IsNaturalBreak(line string, lineIndex int, allLines []string) bool {
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

// isNaturalBreak is the internal version used by SplitTextIntoLocalChunks
func (t *TextProcessor) isNaturalBreak(line string, lineIndex int, allLines []string) bool {
	return t.IsNaturalBreak(line, lineIndex, allLines)
}

// FormatLocalChunk formats a chunk with headers and structure
func (t *TextProcessor) FormatLocalChunk(chunk string, chunkNum, totalChunks int) string {
	var formatted strings.Builder

	// Extract metadata
	metadata := t.extractMetadata(chunk)
	pageRange := t.extractPageRange(chunk)

	// Add comprehensive metadata header
	formatted.WriteString("# Document Chunk\n\n")

	// Metadata section
	formatted.WriteString("## Metadata\n")
	formatted.WriteString(fmt.Sprintf("- **Chunk Number**: %d of %d\n", chunkNum, totalChunks))

	if pageRange != "" {
		formatted.WriteString(fmt.Sprintf("- **Page Range**: %s\n", pageRange))
	}

	if metadata != "" {
		formatted.WriteString(metadata)
	}

	formatted.WriteString("\n")

	// Content section with clear formatting for future embedding
	formatted.WriteString("## Content\n\n")
	formatted.WriteString(t.cleanAndStructureContent(chunk))

	return formatted.String()
}

// ExtractMetadata extracts document metadata from the chunk
func (t *TextProcessor) ExtractMetadata(chunk string) string {
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

	// Look for document titles
	titlePattern := regexp.MustCompile(`(?m)^([A-Z][A-Za-z\s]{3,50})$`)
	if matches := titlePattern.FindAllString(chunk, -1); len(matches) > 0 {
		// Filter out common non-titles
		var titles []string
		for _, match := range matches {
			trimmed := strings.TrimSpace(match)
			if !strings.Contains(trimmed, "Page") && !strings.Contains(trimmed, "---") &&
				len(trimmed) > 5 && len(trimmed) < 100 {
				titles = append(titles, trimmed)
			}
		}
		if len(titles) > 0 {
			metadata.WriteString(fmt.Sprintf("- **Document Title**: %s\n", strings.Join(titles[:1], ", ")))
		}
	}

	return metadata.String()
}

// extractMetadata is the internal version used by FormatLocalChunk
func (t *TextProcessor) extractMetadata(chunk string) string {
	return t.ExtractMetadata(chunk)
}

// ExtractPageRange extracts page range from the chunk
func (t *TextProcessor) ExtractPageRange(chunk string) string {
	// Look for page separators like "--- Page X ---"
	pagePattern := regexp.MustCompile(`--- Page (\d+) ---`)
	matches := pagePattern.FindAllStringSubmatch(chunk, -1)

	if len(matches) == 0 {
		return ""
	}

	if len(matches) == 1 {
		// Single page
		return fmt.Sprintf("Page %s", matches[0][1])
	}

	// Multiple pages - get first and last
	firstPage := matches[0][1]
	lastPage := matches[len(matches)-1][1]

	if firstPage == lastPage {
		return fmt.Sprintf("Page %s", firstPage)
	}

	return fmt.Sprintf("Page %s–%s", firstPage, lastPage)
}

// extractPageRange is the internal version used by FormatLocalChunk
func (t *TextProcessor) extractPageRange(chunk string) string {
	return t.ExtractPageRange(chunk)
}

// CleanAndStructureContent cleans and structures the content for better embedding
func (t *TextProcessor) CleanAndStructureContent(chunk string) string {
	lines := strings.Split(chunk, "\n")
	var cleaned strings.Builder

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the beginning and end
		if trimmed == "" && (i == 0 || i == len(lines)-1) {
			continue
		}

		// Clean up page separators
		if strings.Contains(trimmed, "--- Page") {
			cleaned.WriteString(fmt.Sprintf("\n### Page %s\n\n",
				regexp.MustCompile(`Page (\d+)`).FindStringSubmatch(trimmed)[1]))
			continue
		}

		// Format headings
		if t.isHeading(trimmed) {
			cleaned.WriteString(fmt.Sprintf("\n### %s\n\n", trimmed))
			continue
		}

		// Format bullet points and numbered lists
		if strings.HasPrefix(trimmed, "•") || strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "*") {
			cleaned.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(trimmed[1:])))
			continue
		}

		// Format numbered lists
		if matched, _ := regexp.MatchString(`^\d+\.`, trimmed); matched {
			cleaned.WriteString(fmt.Sprintf("%s\n", trimmed))
			continue
		}

		// Regular text
		if trimmed != "" {
			cleaned.WriteString(trimmed + "\n")
		} else {
			cleaned.WriteString("\n")
		}
	}

	return strings.TrimSpace(cleaned.String())
}

// cleanAndStructureContent is the internal version used by FormatLocalChunk
func (t *TextProcessor) cleanAndStructureContent(chunk string) string {
	return t.CleanAndStructureContent(chunk)
}

// IsHeading checks if a line is a heading
func (t *TextProcessor) IsHeading(line string) bool {
	trimmed := strings.TrimSpace(line)

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

	// Check if it looks like a heading (short, ends with colon or period)
	if len(trimmed) < 100 && (strings.HasSuffix(trimmed, ":") || strings.HasSuffix(trimmed, ".")) {
		return true
	}

	return false
}

// isHeading is the internal version used by cleanAndStructureContent
func (t *TextProcessor) isHeading(line string) bool {
	return t.IsHeading(line)
}

// CreateLocalIntelligentChunk creates a local intelligent chunk (fallback for AI)
func (t *TextProcessor) CreateLocalIntelligentChunk(text string) string {
	chunks := t.SplitTextIntoLocalChunks(text)
	if len(chunks) == 0 {
		return text
	}

	// Return the first chunk (since this is called for individual chunks)
	return t.FormatLocalChunk(chunks[0], 1, 1)
}

// SaveJSONChunk saves a chunk as JSON file
func (t *TextProcessor) SaveJSONChunk(chunk interface{}, jsonDir, filename string, chunkIndex int) error {
	// Marshal to JSON
	jsonData, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create JSON directory for this file
	jsonFileDir := filepath.Join(jsonDir, strings.TrimSuffix(filename, filepath.Ext(filename)))
	if err := os.MkdirAll(jsonFileDir, 0755); err != nil {
		return fmt.Errorf("failed to create JSON directory: %w", err)
	}

	// Save JSON file
	jsonPath := filepath.Join(jsonFileDir, fmt.Sprintf("chunk_%d.json", chunkIndex))
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save JSON file: %w", err)
	}

	return nil
}
