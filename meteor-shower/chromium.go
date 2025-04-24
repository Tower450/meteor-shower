package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Function to extract Chromium bookmarks (STDOUT format)
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

func findBookmarkFiles() ([]string, error) {
	cmd := exec.Command("find", "/home", "-type", "f", "-path", "*/.config/*/Default/Bookmarks")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run find: %v", err)
	}

	// Split by line
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}

func extractBookmarks(path string) ([]Bookmark, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	roots := jsonData["roots"].(map[string]interface{})
	var allBookmarks []Bookmark
	for _, root := range []string{"bookmark_bar", "other", "synced"} {
		if node, ok := roots[root].(map[string]interface{}); ok {
			allBookmarks = append(allBookmarks, parseBookmarks(node, root)...)
		}
	}
	return allBookmarks, nil
}

// parseBookmarks recursively walks children and tracks parent folder name
func parseBookmarks(data map[string]interface{}, parent string) []Bookmark {
	var results []Bookmark

	if children, ok := data["children"].([]interface{}); ok {
		for _, child := range children {
			childMap := child.(map[string]interface{})
			typ := childMap["type"].(string)
			name := childMap["name"].(string)

			if typ == "url" {
				results = append(results, Bookmark{
					Name:   name,
					URL:    childMap["url"].(string),
					Parent: parent,
				})
			} else if typ == "folder" {
				results = append(results, parseBookmarks(childMap, name)...)
			}
		}
	}
	return results
}
