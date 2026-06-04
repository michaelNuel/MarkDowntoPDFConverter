package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/michaelNuel/markdownConverter/converter"
	"github.com/michaelNuel/markdownConverter/server"
	"log"
	"os"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: markdownconverter <command> [arguments]")
		fmt.Println("Commands: convert, serve")
		os.Exit(1)
	}

	//Look at the first argument to decide what command to run
	command := os.Args[1]

	switch command {
	case "convert":
		//Handle the convert command
		//os.Arg[2:] passes all flags  After the word "convert"
		runConvert(os.Args[2:])

	case "serve":
		//Handle Serve Command
		runServe(os.Args[2:])

	case "batch":
		//Handle Batch Command
		runBatch(os.Args[2:])

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func runConvert(args []string) {
	//Create an isolated flag set for this command
	cmd := flag.NewFlagSet("convert", flag.ExitOnError)
	inputPath := cmd.String("in", "", "Path to the markdown file (required)")
	outputPath := cmd.String("out", "output.pdf", "Path to save the PDF")

	// A. Add the theme flag here (defaulting to "modern")
	theme := cmd.String("theme", "modern", "CSS theme to apply (modern, github)")

	//Parse flags for the convert command
	cmd.Parse(args)

	if *inputPath == "" {
		log.Fatal("Error: Missing required flag -in")
	}

	//Read, convert, and save the PDF (our original logic)
	content, err := os.ReadFile(*inputPath)
	if err != nil {
		log.Fatalf("Error Reading file: %v", err)
	}

	// B. Pass the theme name to our HTML converter function
	htmlContent, err := converter.MarkdownToHTML(content, *theme)
	if err != nil {
		log.Fatalf("Conversion error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pdfBytes, err := converter.HTMLToPDF(ctx, htmlContent)
	if err != nil {
		log.Fatalf("PDF rendering error: %v", err)
	}

	err = os.WriteFile(*outputPath, pdfBytes, 0644)
	if err != nil {
		log.Fatalf("Error writing PDF: %v", err)
	}

	fmt.Printf("PDF saved to %s\n", *outputPath)
}

//Handle the 'serve' Command (Placeholder for now)

func runServe(args []string) {
	cmd := flag.NewFlagSet("serve", flag.ExitOnError)
	inputPath := cmd.String("in", "", "Path to the markdown file(required)")
	port := cmd.Int("port", 8000, "Local server port")

		// C. Add the theme flag here as well (defaulting to "modern")
	theme := cmd.String("theme", "modern", "CSS theme to apply (modern, github)")

	cmd.Parse(args)

	if *inputPath == "" {
		log.Fatal("Error: Missing required flag -in")
	}

	// CALL YOUR SERVER START FUNCTION HERE!
	server.Start(*inputPath, *port, *theme)

}

func runBatch(args []string){
	//Create a flag set specifically for the batch command 
	cmd := flag.NewFlagSet("batch", flag.ExitOnError)

	//Define the flags for this command 
	repos := cmd.String("repos", "", "Comma-seperated list of Github repositories (required)")
	outDir := cmd.String("outdir", "downloads", "Folder where the pdfs should be saved")
	workers := cmd.Int("workers", 3, "Number of parrallel workers (default is 3)")

	// 2. Parse the arguments
	cmd.Parse(args)

	//Validate that the user provide repos 
	if *repos == "" {
			log.Fatal("Error: Missing required flag -repos. Example: -repos 'username/repo1, username/repo2'")
	}
  
	//Split the comma seperated string into a raw list of strings 
	rawRepos := strings.Split(*repos, ",")
	var cleanRepos []string 

	//sanitze the inputs: remove any accidental spaces 
	for _,  r:= range rawRepos {
		clean := strings.TrimSpace(r)
		if clean != "" {
			//Append  the clean Repository path to our file 
			cleanRepos = append(cleanRepos, clean )
		}
	}

	//Validate we have at least one repository to process 
	if  len(cleanRepos) == 0 {
		log.Fatal("Error: no valid repository paths provided after cleaning. ")
	}

	//Call our RunBatchResponse function from the converter package 
	err := converter.RunBatchQueue(cleanRepos, *outDir, *workers)
	if err != nil {
		log.Fatalf("Batch queue failed: %v", err)
	}
}
