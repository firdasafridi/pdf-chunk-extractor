# PDF Chunk Extractor - High Level Overview

## 🎯 Simple Process Flow

```mermaid
graph TD
    A[📁 PDF Files] --> B[🚀 PDF Processor]
    B --> C[📄 Extract Text + OCR]
    C --> D[📋 Full Text Output]
    C --> E[🧠 Intelligent Chunking]
    E --> F[📝 Formatted Chunks]
    E --> G[📄 JSON Chunks]
    
    D --> H[📂 output/]
    F --> I[📂 chunk/]
    G --> J[📂 json/]
    
    style A fill:#e3f2fd
    style H fill:#e8f5e8
    style I fill:#e8f5e8
    style J fill:#e8f5e8
```

## 📊 What You Get

```mermaid
graph LR
    subgraph "Input"
        A[📁 PDF Files in data/]
    end
    
    subgraph "Output"
        B[📋 Full Text<br/>output/document.txt]
        C[📝 Smart Chunks<br/>chunk/document/chunk_1.txt]
        D[📄 JSON Ready<br/>json/document/chunk_1.json]
    end
    
    A --> B
    A --> C
    A --> D
    
    style A fill:#e3f2fd
    style B fill:#e8f5e8
    style C fill:#e8f5e8
    style D fill:#e8f5e8
```

## 🔄 Processing Steps

```mermaid
flowchart TD
    Start([🎬 Start]) --> PDFs[📁 Read PDF Files]
    PDFs --> Extract[📄 Extract Text]
    Extract --> OCR{Need OCR?}
    OCR -->|Yes| OCRProcess[🔍 Use Tesseract OCR]
    OCR -->|No| Continue[📝 Use Direct Text]
    OCRProcess --> Continue
    Continue --> SaveFull[💾 Save Full Text]
    SaveFull --> Chunking{AI Available?}
    
    Chunking -->|Yes| AIChunk[🧠 AI-Powered Chunking]
    Chunking -->|No| LocalChunk[📝 Local Smart Chunking]
    
    AIChunk --> Output[📤 Create Outputs]
    LocalChunk --> Output
    
    Output --> TextFiles[📂 Text Files]
    Output --> JSONFiles[📄 JSON Files]
    
    style Start fill:#e3f2fd
    style TextFiles fill:#e8f5e8
    style JSONFiles fill:#e8f5e8
```

## 📁 Directory Structure

```mermaid
graph TD
    A[pdf-chunk-extractor/] --> B[📁 data/]
    A --> C[📁 output/]
    A --> D[📁 chunk/]
    A --> E[📁 json/]
    
    B --> B1[📄 document1.pdf]
    B --> B2[📄 document2.pdf]
    
    C --> C1[📋 document1.txt]
    C --> C2[📋 document2.txt]
    
    D --> D1[📁 document1/]
    D --> D2[📁 document2/]
    
    D1 --> D1a[📝 chunk_1.txt]
    D1 --> D1b[📝 chunk_2.txt]
    
    E --> E1[📁 document1/]
    E --> E2[📁 document2/]
    
    E1 --> E1a[📄 chunk_1.json]
    E1 --> E1b[📄 chunk_2.json]
    
    style A fill:#f3e5f5
    style B fill:#e3f2fd
    style C fill:#e8f5e8
    style D fill:#e8f5e8
    style E fill:#e8f5e8
```

## 🎯 Output Examples

### 📋 Full Text Output
```
--- Page 3 ---
Panen Kelapa Sawit
SOP/KCN-AGR/012/2023
Tujuan
1. Memastikan seluruh Tandan Buah Segar (TBS)...
```

### 📝 Smart Chunk Output
```
# Document Chunk

## Metadata
- **Chunk Number**: 1 of 3
- **Page Range**: Page 3–4
- **Document Code**: SOP/KCN-AGR/012/2023

## Content

### Panen Kelapa Sawit

### Tujuan

1. Memastikan seluruh Tandan Buah Segar (TBS)...
```

### 📄 JSON Output
```json
{
  "filename": "document.pdf",
  "chunk_index": 1,
  "page_range": "Page 3–4",
  "text": "### Panen Kelapa Sawit\n\n### Tujuan\n\n1. Memastikan seluruh Tandan Buah Segar (TBS)..."
}
```

## 🚀 How to Use

```mermaid
graph LR
    A[1. Put PDFs in data/] --> B[2. Set API Key]
    B --> C[3. Run: go run main.go]
    C --> D[4. Get Results!]
    
    style A fill:#e3f2fd
    style B fill:#fff3e0
    style C fill:#f3e5f5
    style D fill:#e8f5e8
```

## 🎨 Features

```mermaid
graph TD
    A[🎯 Features] --> B[📄 PDF Text Extraction]
    A --> C[🔍 OCR Fallback]
    A --> D[🧠 AI Chunking]
    A --> E[📝 Local Chunking]
    A --> F[📄 JSON Output]
    A --> G[🔍 Smart Breaks]
    A --> H[📊 Metadata Extraction]
    
    B --> B1[Direct text extraction]
    C --> C1[Tesseract OCR]
    D --> D1[OpenAI ChatGPT]
    E --> E1[Natural break detection]
    F --> F1[Vector DB ready]
    G --> G1[Headings, pages, lists]
    H --> H1[Document codes, dates]
    
    style A fill:#f3e5f5
    style B fill:#e3f2fd
    style C fill:#e3f2fd
    style D fill:#e3f2fd
    style E fill:#e3f2fd
    style F fill:#e3f2fd
    style G fill:#e3f2fd
    style H fill:#e3f2fd
```

## 🎉 Benefits

- **🔄 Dual Mode**: Works with or without AI
- **📊 Rich Output**: Text, chunks, and JSON
- **🔍 Smart Processing**: Respects document structure
- **📄 Vector Ready**: JSON perfect for databases
- **🎯 Easy to Use**: Simple command line interface
- **🛡️ Robust**: Handles errors gracefully 