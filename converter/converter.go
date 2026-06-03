package converter

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/yuin/goldmark"
	"os"
	"path/filepath"
	"github.com/michaelNuel/markdownConverter/themes"
)

// MarkdownToHTML takes raw markdown bytes and returns a basic HTML page.
// In Go, functions that start with a Capital letter are "Exported" (public).
// Functions starting with lowercase letters are private to the package.
func MarkdownToHTML(mdContent []byte, themeName string) (string, error) {
	var htmlBuf bytes.Buffer

	//Fetch the css content	dynamically from memory 
	css, err := themes.Get(themeName)
	if err != nil {
		return "", fmt.Errorf("failed to load theme: %w", err)
	}

	//parse markdown
	err = goldmark.Convert(mdContent, &htmlBuf)
	if err != nil {
		return "", fmt.Errorf("failed to convert markdown to html: %w", err)
	}

	// Wrap in a simple HTML document
	styledHTML := fmt.Sprintf(
		`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        %s
    </style>
</head>
<body>
    %s
	    <!-- Tiny Live Reload Script -->
    <script>
        // Open a live pipe to the Go server's events endpoint
        const events = new EventSource('/events');
        events.onmessage = function(event) {
            if (event.data === 'reload') {
                console.log("File change detected! Reloading page...");
                location.reload(); // Refresh the tab automatically!
            }
        };
    </script>
</body>
</html>`, css, htmlBuf.String(),
	)

	return styledHTML, nil

}

// HTMLToPDF launches headless Chrome, prints the HTML to PDF, and returns the raw bytes.

func HTMLToPDF(ctx context.Context, htmlContent string) ([]byte, error) {
	//Create a temp file
	tmpFile, err := os.CreateTemp("", "md2pdf-*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		return nil, fmt.Errorf("Failed to write temp file: %w", err)
	}

	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	targetURL := "file://" + filepath.ToSlash(absPath)
	// Spin up chromedp context
	cCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var pdfBytes []byte
	err = chromedp.Run(cCtx,
		chromedp.Navigate(targetURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBytes, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp rendering error: %w", err)
	}
	return pdfBytes, nil

}
