# PDF Chunk Extractor with AI-Powered Intelligent Chunking

A powerful Go application that extracts text from PDF files using OCR when needed, and creates intelligent, meaningful chunks using OpenAI's ChatGPT API.

## 🚀 Features

- **PDF Text Extraction**: Extracts text directly from PDF files
- **OCR Fallback**: Uses Tesseract OCR when direct text extraction fails
- **AI-Powered Chunking**: Creates intelligent chunks based on themes and content using ChatGPT
- **Multi-language Support**: Supports English and Indonesian text extraction
- **Organized Output**: Creates structured chunk directories with numbered files

## 📁 Directory Structure

```
pdf-chunk-extractor/
├── data/           # Place your PDF files here
├── output/         # Full extracted text files
├── chunk/          # AI-processed intelligent chunks
│   └── filename/   # Chunks for each PDF file
│       ├── chunk_1.txt
│       ├── chunk_2.txt
│       └── ...
├── main.go
├── go.mod
└── README.md
```

## 🛠️ Prerequisites

1. **Go** (version 1.16 or higher)
2. **Tesseract OCR** with Indonesian language support
3. **OpenAI API Key**

### Installing Tesseract

**macOS:**
```bash
brew install tesseract
brew install tesseract-lang  # For additional languages
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install tesseract-ocr
sudo apt install tesseract-ocr-ind  # Indonesian language pack
```

**Windows:**
Download from [Tesseract GitHub](https://github.com/UB-Mannheim/tesseract/wiki)

## 🔧 Setup

1. **Clone the repository:**
```bash
git clone <repository-url>
cd pdf-chunk-extractor
```

2. **Install Go dependencies:**
```bash
go mod tidy
```

3. **Set your OpenAI API key:**
```bash
export OPENAI_API_KEY="your-openai-api-key-here"
```

4. **Place PDF files in the data directory:**
```bash
mkdir -p data
# Copy your PDF files to the data/ directory
```

## 🚀 Usage

Run the application:
```bash
go run main.go
```

Or build and run:
```bash
go build -o pdf-chunk-extractor
./pdf-chunk-extractor
```

## 📊 Output

The application creates two types of output:

### 1. Full Text Files (`output/`)
- Complete extracted text from each PDF
- Includes page separators
- Useful for full document review

### 2. Intelligent Chunks (`chunk/filename/`)
- AI-processed chunks based on themes and content
- Each chunk is meaningful and self-contained
- Organized with clear headers and structure
- Perfect for analysis, search, or further processing

## 🧠 AI Chunking Process

The intelligent chunking works as follows:

1. **Text Extraction**: Extracts text from PDF (with OCR fallback)
2. **Initial Splitting**: Splits text into manageable chunks (~4000 characters)
3. **AI Processing**: Sends each chunk to ChatGPT for intelligent organization
4. **Structured Output**: Creates well-formatted chunks with:
   - Clear headers and sections
   - Logical grouping of related content
   - Preserved metadata (document codes, dates)
   - Improved readability

## 📝 Example Output

**Input PDF Content:**
```
--- Page 3 ---
Panen Kelapa Sawit
SOP/KCN-AGR/012/2023
Tujuan
1. Memastikan seluruh Tandan Buah Segar (TBS)...
```

**AI-Processed Chunk:**
```
# Palm Oil Harvesting Standard Operating Procedure

## Document Information
- **SOP Code**: SOP/KCN-AGR/012/2023
- **Document Type**: Standard Operating Procedure
- **Subject**: Palm Oil Harvesting (Panen Kelapa Sawit)

## Purpose (Tujuan)
1. Ensure all harvested Fresh Fruit Bunches (TBS) meet company quality standards
2. Ensure optimal transportation of bunches and loose fruits to Palm Oil Mill (PKS)

## Key Definitions
- **Panen**: Harvesting work of collecting ripe TBS and loose fruits
- **Seksi Panen**: Harvest area that must be completed in one day
- **Interval Panen**: Time between harvests in the same area
```

## ⚙️ Configuration

You can modify the following constants in `main.go`:

```go
const (
    DataDir    = "data"      // Input directory
    OutputDir  = "output"    // Full text output directory
    ChunkDir   = "chunk"     // Chunk output directory
    MaxChunkSize = 4000      // Maximum characters per AI chunk
)
```

## 🔍 Troubleshooting

### Common Issues

1. **"OPENAI_API_KEY environment variable is required"**
   - Set your OpenAI API key: `export OPENAI_API_KEY="your-key"`

2. **"tesseract command failed"**
   - Install Tesseract OCR
   - Ensure it's in your system PATH

3. **OCR quality issues**
   - Ensure good quality PDFs
   - Check if Tesseract language packs are installed

4. **API rate limits**
   - The application includes error handling for API failures
   - Falls back to original text if AI processing fails

## 📈 Performance Tips

- **Large PDFs**: The application processes large files efficiently by chunking them
- **API Costs**: Each chunk requires an API call, so monitor your OpenAI usage
- **Parallel Processing**: Consider running multiple instances for different files

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [go-fitz](https://github.com/gen2brain/go-fitz) for PDF processing
- [Tesseract OCR](https://github.com/tesseract-ocr/tesseract) for text recognition
- [OpenAI](https://openai.com/) for intelligent text processing 