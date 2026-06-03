package server

import (
	"fmt"
	"github.com/michaelNuel/markdownConverter/converter"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// HTML template for our web upload form
var uploadHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Markdown to PDF Web Converter</title>
    <style>
        body { font-family: system-ui, sans-serif; background-color: #0f172a; color: #f8fafc; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .container { background-color: #1e293b; padding: 2.5rem; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.3); width: 100%; max-width: 450px; text-align: center; }
        h2 { margin-bottom: 1.5rem; }
        .form-group { margin-bottom: 1.5rem; text-align: left; }
        label { display: block; margin-bottom: 0.5rem; font-size: 0.9rem; color: #94a3b8; }
        input[type="file"], select { width: 100%; padding: 0.75rem; border-radius: 6px; border: 1px solid #334155; background-color: #0f172a; color: #f8fafc; box-sizing: border-box; }
        button { width: 100%; background-color: #3b82f6; color: white; border: none; padding: 0.75rem; border-radius: 6px; font-weight: 600; cursor: pointer; font-size: 1rem; }
        button:hover { background-color: #2563eb; }
    </style>
</head>
<body>
    <div class="container">
        <h2>MD to PDF Converter</h2>
        <form action="/upload" method="POST" enctype="multipart/form-data">
            <div class="form-group">
                <label for="markdownFile">Select Markdown File (.md)</label>
                <input type="file" id="markdownFile" name="markdownFile" accept=".md" required>
            </div>
            <div class="form-group">
                <label for="theme">Select Styling Theme</label>
                <select id="theme" name="theme">
                    <option value="modern">Modern</option>
                    <option value="github">GitHub</option>
                </select>
            </div>
            <button type="submit">Convert & Download PDF</button>
        </form>
    </div>
</body>
</html>`

//	Declare a channel. This is the communication pipe between
//
// our file watcher and our web server handler.
var reloadChan = make(chan bool)

// Start boots a simple HTTP web server on the specified port.
func Start(filePath string, port int, themeName string) {

	// 1. Start our file watcher in a background thread (Goroutine)
	// The 'go' keyword spins this off so it doesn't block ListenAndServe!
	go watchFile(filePath)
	//Register a handler for the route "/" URL Path
	// When a browser visits our server, this function runs.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Read the markdown file
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		//convert the markdown html using the convert package
		htmlContent, err := converter.MarkdownToHTML(content, themeName)
		if err != nil {
			http.Error(w, "Failed to parse markdown", http.StatusInternalServerError)
			return
		}
		// C. Write the HTML bytes directly to the browser
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		//Set headers to keep the network connection open
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Block (wait) here until the watcher sends a message down the reloadChan
		<-reloadChan

		// Once a message arrives, write "reload" to the browser
		fmt.Fprintf(w, "data: reload\n\n")

	})

	// Route to serve the file upload web form
	http.HandleFunc("/upload-page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(uploadHTML))
	})

	// Route to handle the uploaded file and return the PDF
	http.HandleFunc("/upload", handleFileUpload)

	// 2. Format the address string (e.g. "127.0.0.1:8080")
	// addr := fmt.Sprintf("127.0.0.1:%d", port)
	// fmt.Printf("\nWeb server running! open: http://%s\n", addr)
		// --- CLOUD DEPLOYMENT UPDATE ---
	// Check if the cloud provider gave us a "PORT" environment variable.
	// If it exists, we use it. If not, we fall back to our CLI port.
	var addr string
	cloudPort := os.Getenv("PORT")
	if cloudPort != "" {
		// In the cloud, bind to "0.0.0.0" (public internet)
		addr = ":" + cloudPort
		fmt.Printf("\nRunning in production mode on port %s\n", cloudPort)
	} else {
		// Locally, bind to "127.0.0.1" (localhost only)
		addr = fmt.Sprintf("127.0.0.1:%d", port)
		fmt.Printf("\nWeb server running! Open: http://%s\n", addr)
	}
	//Start the Server. this blocks and keeps running in the terminal
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func watchFile(filePath string) {
	var lastModTime time.Time

	for {
		//Get file statistics (contains size, modification time, etc)
		stat, err := os.Stat(filePath)
		if err == nil {
			// If this is the first time we check, set our baseline time
			if lastModTime.IsZero() {
				// 1. The program just booted.
				// 2. Save the file's current save-time as our starting baseline.
				lastModTime = stat.ModTime()
			} else if stat.ModTime().After(lastModTime) {
				// The file was modified since we last checked!
				lastModTime = stat.ModTime()

				// Print an alert to the terminal console
				fmt.Println("--> Markdown file modified! Re-rendering")

				// 3. Drop "true" into the channel to trigger the browser reload
				// We use a select block so the server doesn't freeze if no browser is open
				select {
				case reloadChan <- true:
				default:
				}
			}
		}

		// Sleep for 500 milliseconds to prevent using 100% of your CPU
		time.Sleep(500 * time.Millisecond)

	}
}

// handleFileUpload processes the POST request containing the uploaded file
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// A. Only allow POST requests (file submissions)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// B. Parse the uploaded data (limit memory to 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	//Extract the file from the form
	//markdownFile matches the name  attribute  in our html file input
	file, _, err := r.FormFile("markdownFile")
	if err != nil {
		http.Error(w, "Failed to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	//Extract the theme choice from the form
	theme := r.FormValue("theme")
	if theme == "" {
		theme = "modern" //Default fallback
	}

	//Read the Uploaded File contents into  memory bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to parse markdown", http.StatusInternalServerError)
		return
	}

	//Convert Markdown bytes -> Styled HTML
	htmlContent, err := converter.MarkdownToHTML(fileBytes, theme)
	if err != nil {
		http.Error(w, "Failed to parse markdown", http.StatusInternalServerError)
		return
	}

	//Convert HTML -> PDF using Chromedp (resuing context from request)
	pdfBytes, err := converter.HTMLToPDF(r.Context(), htmlContent)
	if err != nil {
		http.Error(w, fmt.Sprintf("PDF generation failed: %v", err),
			http.StatusInternalServerError)
		return
	}

	//  Set HTTP headers to tell the browser this is a PDF file download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"converted.pdf\"")
	// Write the PDF bytes directly to the browser download pipe
	w.Write(pdfBytes)
}
