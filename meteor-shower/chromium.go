package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Function to extract Chromium bookmarks (JSON format)
func extractChromiumBookmarks(bookmarksPath string) {
	file, err := os.Open(bookmarksPath)
	if err != nil {
		log.Fatalf("Failed to open bookmarks file: %v\n", err)
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		log.Fatalf("Failed to decode bookmarks JSON: %v\n", err)
	}

	// Extract bookmarks from the different roots sections
	roots, ok := data["roots"].(map[string]interface{})
	if !ok {
		log.Fatalf("Failed to extract root sections\n")
	}

	// Define a function to recursively extract bookmarks
	var extractBookmarksFromSection func(interface{})
	extractBookmarksFromSection = func(section interface{}) {
		// Check if the section is a map with "children" key
		if sectionMap, ok := section.(map[string]interface{}); ok {
			if children, ok := sectionMap["children"].([]interface{}); ok {
				// Iterate through all the children in this section
				for _, item := range children {
					child := item.(map[string]interface{})
					id := child["id"].(string)
					name := child["name"].(string)

					// If it's a bookmark (i.e., it has a URL), print it
					if url, ok := child["url"].(string); ok {
						fmt.Printf("🌠 - %s %s: %s\n", id, name, url)
					} else {
						// If it's a folder, we recursively extract bookmarks from it
						extractBookmarksFromSection(child)
					}
				}
			}
		}
	}

	// Extract from all sections: bookmark_bar, other, etc.
	fmt.Println("🌐 Extracted Chromium Bookmarks:")
	for _, sectionName := range []string{"bookmark_bar", "other", "synced"} {
		if section, ok := roots[sectionName]; ok {
			extractBookmarksFromSection(section)
		}
	}
}
