package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver using a blank identifier
)

func printBanner() {
	banner := `
 ███▄ ▄███▓▓█████▄▄▄█████▓▓█████  ▒█████   ██▀███       ██████  ██░ ██  ▒█████   █     █░▓█████  ██▀███  
▓██▒▀█▀ ██▒▓█   ▀▓  ██▒ ▓▒▓█   ▀ ▒██▒  ██▒▓██ ▒ ██▒   ▒██    ▒ ▓██░ ██▒▒██▒  ██▒▓█░ █ ░█░▓█   ▀ ▓██ ▒ ██▒
▓██    ▓██░▒███  ▒ ▓██░ ▒░▒███   ▒██░  ██▒▓██ ░▄█ ▒   ░ ▓██▄   ▒██▀▀██░▒██░  ██▒▒█░ █ ░█ ▒███   ▓██ ░▄█ ▒
▒██    ▒██ ▒▓█  ▄░ ▓██▓ ░ ▒▓█  ▄ ▒██   ██░▒██▀▀█▄       ▒   ██▒░▓█ ░██ ▒██   ██░░█░ █ ░█ ▒▓█  ▄ ▒██▀▀█▄  
▒██▒   ░██▒░▒████▒ ▒██▒ ░ ░▒████▒░ ████▓▒░░██▓ ▒██▒   ▒██████▒▒░▓█▒░██▓░ ████▓▒░░░██▒██▓ ░▒████▒░██▓ ▒██▒
░ ▒░   ░  ░░░ ▒░ ░ ▒ ░░   ░░ ▒░ ░░ ▒░▒░▒░ ░ ▒▓ ░▒▓░   ▒ ▒▓▒ ▒ ░ ▒ ░░▒░▒░ ▒░▒░▒░ ░ ▓░▒ ▒  ░░ ▒░ ░░ ▒▓ ░▒▓░
░  ░      ░ ░ ░  ░   ░     ░ ░  ░  ░ ▒ ▒░   ░▒ ░ ▒░   ░ ░▒  ░ ░ ▒ ░▒░ ░  ░ ▒ ▒░   ▒ ░ ░   ░ ░  ░  ░▒ ░ ▒░
░      ░      ░    ░         ░   ░ ░ ░ ▒    ░░   ░    ░  ░  ░   ░  ░░ ░░ ░ ░ ▒    ░   ░     ░     ░░   ░ 
       ░      ░  ░           ░  ░    ░ ░     ░              ░   ░  ░  ░    ░ ░      ░       ░  ░   ░     
                                                                                                         
`
	fmt.Println(banner)
	time.Sleep(1 * time.Second)
}

func main() {
	printBanner()

	/** CHROME **/
	chromiumBookmarkFiles, err := findBookmarkFiles()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	for _, path := range chromiumBookmarkFiles {
		fmt.Printf("Extracting for profile: %s", path)
		// Extract Chrome bookmarks if the file exists
		if fileExists(path) {
			// extractChromiumBookmarks(chromeBookmarksPath) => RAW
			bookmarks, err := parseBookmarks(path)
			if err != nil {
				log.Fatalf("", err)
			}
			// outputBookmarksJSON(bookmarks) => OUTPUT JSON

			// Build the tree: key = parent, value = children
			tree := make(map[string][]Bookmark)

			for _, b := range bookmarks {
				tree[b.Parent] = append(tree[b.Parent], b)
			}

			// Sort the bookmarks for each parent to ensure consistent output order
			for _, children := range tree {
				sort.SliceStable(children, func(i, j int) bool {
					return children[i].Name < children[j].Name
				})
			}

			// Recursively print starting from top
			fmt.Println("|- bookmarks")
			for k, v := range tree {
				fmt.Printf("📁 %q has %d children\n", k, len(v))
				printBookmarkTree(tree, k, 1)
			}
		}
	}

	// TODO: make a good way to dump fire or chrome or both... OR ALL OF THEM!~~~

	/** FIREFOX **/

	// Firefox profile path
	firefoxProfilePaths, err := findFirefoxProfile()
	if err != nil {
		log.Fatalf("Error finding Firefox profile: %v\n", err)
	}

	for _, foxFile := range firefoxProfilePaths {
		firefoxDBPath := filepath.Join(foxFile, "places.sqlite")

		// Verify if the Firefox database exists
		if fileExists(firefoxDBPath) {
			// Create a temporary copy of the Firefox database
			tempDB := "/tmp/temp_places.sqlite"
			err := copyFile(firefoxDBPath, tempDB)
			if err != nil {
				log.Fatalf("Failed to copy Firefox database: %v\n", err)
			}

			// Extract Firefox bookmarks
			fmt.Printf("🦊🔥 Extracted Firefox Bookmarks %s:\n", firefoxDBPath)
			extractFirefoxBookmarks(tempDB)

			// Clean up temporary database file
			err = os.Remove(tempDB)
			if err != nil {
				log.Printf("Failed to remove temporary Firefox database: %v\n", err)
			}
		}
	}
}

func printBookmarkTree(tree map[string][]Bookmark, parent string, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	// First, print folders (those with no URL)
	for _, b := range tree[parent] {
		if b.URL == "" {
			fmt.Printf("%s|-📁 %s\n", indent, b.Name)
			printBookmarkTree(tree, b.Name, depth+1) // Recurse into subfolders
		}
	}

	// Then, print actual bookmarks (those with URLs)
	for _, b := range tree[parent] {
		if b.URL != "" {
			fmt.Printf("%s|-🌠 %s → %s\n", indent, b.Name, b.URL)
		}
	}
}

func outputBookmarksJSON(bookmarks []Bookmark) error {
	jsonData, err := json.MarshalIndent(bookmarks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %v", err)
	}

	fmt.Println(string(jsonData))
	return nil
}
