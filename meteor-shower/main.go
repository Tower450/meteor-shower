package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	// TODO: make it generic for any chromium find the files and for all users!!!
	// Chrome/Brave bookmarks path (Linux default paths, adjust as necessary)
	chromeBookmarksPath := filepath.Join(os.Getenv("HOME"), ".config", "google-chrome", "Default", "Bookmarks")
	braveBookmarksPath := filepath.Join(os.Getenv("HOME"), ".config", "brave-browser", "Default", "Bookmarks")

	// Extract Chrome bookmarks if the file exists
	if fileExists(chromeBookmarksPath) {
		extractChromiumBookmarks(chromeBookmarksPath)
	}

	// Extract Brave bookmarks if the file exists
	if fileExists(braveBookmarksPath) {
		extractChromiumBookmarks(braveBookmarksPath)
	}

	// TODO: make a good way to dump fire or chrome or both... OR ALL OF THEM!~~~

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
			extractFirefoxBookmarks(tempDB)

			// Clean up temporary database file
			err = os.Remove(tempDB)
			if err != nil {
				log.Printf("Failed to remove temporary Firefox database: %v\n", err)
			}
		}
	}
}
