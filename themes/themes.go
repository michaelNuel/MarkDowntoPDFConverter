package themes 

import (
	"embed"
	"fmt"
)

// The '//go:embed' compiler directive tells Go to read all files ending in '.css' 
// in this folder at compile time and save them inside the 'cssFiles' filesystem.
//
//go:embed *.css

var cssFiles embed.FS

//Get reads  the css  contents of the requested theme from memeory 
func Get(themeName string) (string, error ) {
	filename := themeName + ".css"

	//Read the file directly  from compiled memeory (no disk reads!)
	data, err := cssFiles.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("theme %q not found: %w", themeName, err)
	}

	return string(data), nil
} 