---
title: Markdown Converter
emoji: 📝
colorFrom: blue
colorTo: indigo
sdk: docker
app_port: 8080
pinned: false
---

# Markdown to PDF Workspace & Web Uploader

A modular, high-fidelity Markdown-to-PDF converter CLI tool and live-preview web dashboard written in Go. This project is built using a clean, package-based architecture with zero heavy web framework dependencies.

## Key Features

1. **Multi-Command CLI**: Supports isolated command sets for direct compilation (`convert`) and workspace preview (`serve`).
2. **Dynamic Live Preview Server**: Serves a local website with Server-Sent Events (SSE) that auto-reloads your browser preview instantly when you save changes in your code editor.
3. **Web Upload Portal**: Provides a user-friendly drag-and-drop web uploader page (`/upload-page`) where anyone can upload a `.md` file, select a CSS theme, and instantly download a styled PDF.
4. **Embedded Styling Themes**: Uses Go's native `//go:embed` to package CSS themes (Modern, GitHub) directly inside the executable, avoiding relative path runtime errors.
5. **Headless Chrome Rendering**: Spawns a background Chromium instance using `chromedp` to print styled HTML documents into high-fidelity PDFs.

---

## Project Structure & Architecture

The codebase is organized into modular packages to ensure **Separation of Concerns** (SOC) and reusability:

```text
MarkdownConverter/
├── go.mod                     # Go module definitions & package tracking
├── main.go                    # Entrypoint & CLI Command Router (bootstrap)
├── Dockerfile                 # Multi-stage Docker deployment recipe
├── README.md                  # Hugging Face metadata & developer guide
├── converter/                 # core conversion package (Single Responsibility)
│   └── converter.go           # Markdown-to-HTML parser and Chromedp PDF renderer
├── server/                    # http server package
│   └── server.go              # Serves routes, handles file uploads, & runs SSE watcher
└── themes/                    # embedded styles package
    ├── themes.go              # Exposes the embedded filesystem (embed.FS)
    ├── modern.css             # HSL-variable modern professional theme
    └── github.css             # GitHub markdown replication theme
```

---

## Getting Started & Installation

### Prerequisites
* **Go**: Make sure Go (1.18 or higher) is installed on your computer.
* **Chrome/Chromium**: The PDF engine automates your default Google Chrome or Microsoft Edge browser.

### Setup Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/michaelNuel/MarkDowntoPDFConverter.git
   cd MarkDowntoPDFConverter
   ```
2. Download Go module dependencies:
   ```bash
   go mod download
   ```
3. Compile the code into a standalone executable:
   ```bash
   go build -o md2pdf.exe main.go
   ```

Now you have a portable `md2pdf.exe` binary in your directory!

---

## How to Use the CLI

### 1. Direct File-to-File Conversion
To instantly compile a markdown file directly to a PDF with a chosen theme:

```bash
# Convert using the default Modern theme
./md2pdf.exe convert -in readme.md -out output.pdf

# Convert using the GitHub theme
./md2pdf.exe convert -in readme.md -out output.pdf -theme github
```

### 2. Launch the Live Preview Web Workspace
To start the live-reload server and monitor your edits:

```bash
./md2pdf.exe serve -in readme.md -port 9000
```
1. Open **`http://localhost:9000`** in your browser.
2. Edit `readme.md` in your text editor (VS Code, Notepad, etc.) and hit **Save**.
3. The browser will instantly reload the page with your updates automatically!

### 3. Open the Web Uploader Tool
If you want to use the drag-and-drop file converter, start the server and navigate to:
👉 **`http://localhost:9000/upload-page`**

Select any markdown file, choose your theme, and click **Convert & Download PDF**.

---

## Deployment Guide (Docker & Cloud)

This app is designed to run in containerized environments (like **Hugging Face Spaces** or **Render**). 

The included `Dockerfile` uses a **multi-stage build**:
1. **Stage 1 (Builder)**: Compiles the Go code inside an alpine container.
2. **Stage 2 (Runner)**: Packages the compiled program inside a clean container and installs `chromium` and `ttf-freefont` so Chrome can render text in the generated PDFs in the cloud.

### Deploying to Hugging Face Spaces:
1. Create a new Space on Hugging Face.
2. Select **Docker** as the SDK and choose the **Blank** template.
3. Link your local project to Hugging Face and push:
   ```bash
   git remote add hf https://huggingface.co/spaces/YOUR_USERNAME/YOUR_SPACE_NAME
   git push -f hf main
   ```
   *(Note: Remember to use a Write Access Token as your password).*

---

## Go Backend Concepts Implemented

* **Pointers & Flagsets**: Using `flag.NewFlagSet` to create isolated subcommand flag contexts.
* **Static Embeds (`embed`)**: Packing static CSS files directly into the compiled executable.
* **Goroutines (`go` keyword)**: Running the file watcher in a background thread to prevent blocking the web server.
* **Channels (`chan bool` & `select`)**: Safe synchronization and communication between the file watcher and HTTP stream threads.
* **Error Wrapping (`%w`)**: Preserving low-level OS/Network errors inside custom error messages for better debugging.