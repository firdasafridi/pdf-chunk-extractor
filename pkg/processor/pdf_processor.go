package processor

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/firdasafridi/pdf-chunk-extractor/pkg/config"
	"github.com/gen2brain/go-fitz"
)

// PDFProcessor handles PDF text extraction with OCR fallback
type PDFProcessor struct {
	config config.ChunkerConfig
}

// NewPDFProcessor creates a new PDF processor instance
func NewPDFProcessor(config config.ChunkerConfig) *PDFProcessor {
	return &PDFProcessor{
		config: config,
	}
}

// ExtractTextFromPDFPath extracts text from a PDF file path
func (p *PDFProcessor) ExtractTextFromPDFPath(pdfPath string) (string, error) {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	return p.extractTextFromDocument(doc)
}

// ExtractTextFromPDFBytes extracts text from PDF binary data
func (p *PDFProcessor) ExtractTextFromPDFBytes(data []byte) (string, error) {
	doc, err := fitz.NewFromMemory(data)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF from memory: %w", err)
	}
	defer doc.Close()

	return p.extractTextFromDocument(doc)
}

// ExtractTextFromPDFReader extracts text from PDF reader
func (p *PDFProcessor) ExtractTextFromPDFReader(reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF data: %w", err)
	}

	return p.ExtractTextFromPDFBytes(data)
}

// extractTextFromDocument extracts text from a fitz document
func (p *PDFProcessor) extractTextFromDocument(doc *fitz.Document) (string, error) {
	var result strings.Builder
	totalPages := doc.NumPage()

	for pageIndex := 0; pageIndex < totalPages; pageIndex++ {
		text, err := p.processPage(doc, pageIndex, totalPages)
		if err != nil {
			log.Printf("Warning: failed to process page %d: %v", pageIndex+1, err)
			continue
		}
		result.WriteString(text)
	}

	return result.String(), nil
}

// processPage extracts text from a single page
func (p *PDFProcessor) processPage(doc *fitz.Document, pageIndex, totalPages int) (string, error) {
	pageNum := pageIndex + 1

	// Try direct text extraction first
	text, err := doc.Text(pageIndex)
	if err != nil {
		log.Printf("Warning: failed to extract text from page %d: %v", pageNum, err)
	}

	// If no text found, use OCR
	if strings.TrimSpace(text) == "" {
		text = p.extractTextWithOCR(doc, pageIndex, pageNum)
	}

	// Add page separator
	separator := fmt.Sprintf("\n\n--- Page %d ---\n\n", pageNum)
	return separator + text, nil
}

// extractTextWithOCR uses OCR to extract text from a page image
func (p *PDFProcessor) extractTextWithOCR(doc *fitz.Document, pageIndex, pageNum int) string {
	// Render page as image
	img, err := doc.Image(pageIndex)
	if err != nil {
		log.Printf("Warning: failed to render page %d as image: %v", pageNum, err)
		return ""
	}

	// Save temporary image
	tempImagePath := fmt.Sprintf("temp_page_%d.png", pageIndex)
	if err := p.saveTemporaryImage(img, tempImagePath); err != nil {
		log.Printf("Warning: failed to save temp image: %v", err)
		return ""
	}
	defer os.Remove(tempImagePath)

	// Perform OCR
	ocrText, err := p.runTesseract(tempImagePath)
	if err != nil {
		log.Printf("Warning: OCR failed for page %d: %v", pageNum, err)
		return ""
	}

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

// runTesseract executes the tesseract OCR command
func (p *PDFProcessor) runTesseract(imagePath string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "eng+ind")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract command failed: %w", err)
	}

	return string(output), nil
}
