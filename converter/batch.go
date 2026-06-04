package converter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DownloadRawREADME fetches the raw README.md file from GitHub.
// Since repos use both "main" and "master" branches, we attempt to download
// from "main" first. If that returns 404, we fall back to "master".

func DownloadRawREADME(repoPath string) ([]byte, error) {
	//Try the main branch first
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/README.md", repoPath)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	//if 404 (not found), try the master branch
	if resp.StatusCode == http.StatusNotFound {
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/master/README.md", repoPath)
		resp, err = http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	//If still not successfull, return  an error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned status: %s for %s", resp.Status, repoPath)
	}

	//Read all bytes  from  the HTTP response body
	return io.ReadAll(resp.Body)

}

//worker represents  a single barista. It listens to the jobs channel,
//process repo paths, and alerts the WaitGroup when the channel closes.
//
// the '<-chan string' syntax means this worker can ONLY read  from the jobs channel.

func worker(id int, jobs <-chan string, wg *sync.WaitGroup, outDir string) {
	// Defer calling Done() so the WaitGroup counter decrements when this worker stops
	defer wg.Done()

	// The loop will read from the channel one-by-one.
	// When the channel is closed, the loop automatically exits.
	for repo := range jobs {
		fmt.Printf("[Worker %d] Starting: %s\n", id, repo)

		//Download Markdown
		mdBytes, err := DownloadRawREADME(repo)
		if err != nil {
			fmt.Printf("[Worker %d] Error downloading %s: %v\n", id, repo, err)
			continue
		}
        

		// Convert Markdown -> HTML (We reuse our existing function!)
		htmlContent, err := MarkdownToHTML(mdBytes, "modern")
		if err != nil {
			fmt.Printf("[Worker %d] Error converting markdown %s: %v\n", id, repo, err)
			continue
		}

		//Render HTMl to PDF 
		pdfBytes, err := HTMLToPDF(context.Background(), htmlContent)
		if err != nil{
			fmt.Printf("[Worker %d] Error rendering PDF %s: %v\n", id, repo, err)
			continue
		}

		//save to disk 
		//We replace "/" with "_" in the filename (e.g. "gin-gonic/gin" becomes "gin-gonic_gin.pdf")
		fileName := strings.ReplaceAll(repo, "/", "_") + ".pdf"
		outputPath := filepath.Join(outDir, fileName)

		err = os.WriteFile(outputPath, pdfBytes, 0644)
		if err != nil {
			fmt.Printf("[Worker %d] Error saving PDF %s: %v\n", id, repo, err)
			continue
		}

		fmt.Printf("[Worker %d] Completed: %s -> %s\n", id, repo, fileName)
		
	}
}


// RunBatchQueue sets up the jobs queue and spawns the background workers
func RunBatchQueue(repos []string, outDir string, numWorkers int) error {
 //Ensure the output directory exists, 
 err := os.MkdirAll(outDir, 0755)
 if err != nil {
	return fmt.Errorf("failed to create output directory: %w", err)
 }

 //Intialise the jobs channel 
 jobs :=make(chan string, len(repos))	

 //Initialise the wait group 
 var wg sync.WaitGroup

 //Spawn the requested numbers of workers (goroutines)
 for i :=1; i<= numWorkers; i++ {
	wg.Add(1) //Write 1 on the counter board 
	go worker(i, jobs, &wg, outDir)
 }

 //Feed  the jobs  into the channel (the Cashier putting tickets on the rail )
 for _, repo := range repos {
	jobs <- repo
 }

 //close channel 
 //This tells the workers, "No more tickets are coming. Finish your current job and stop "
 close(jobs)

 //Freeze the main execution until all workers have finished 
 fmt.Printf("Queued %d repositories. Processing with %d workers....\n", len(repos), numWorkers)
 wg.Wait()

 fmt.Println("All batch jobs completed!")
 return nil
}