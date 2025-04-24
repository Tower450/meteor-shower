package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	var paths []string

	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("find", "/home", "-type", "f", "-path", "*/.config/*/Default/Bookmarks")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("linux find failed: %v", err)
		}
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		paths = append(paths, lines...)

	case "darwin":
		userDirs, err := os.ReadDir("/Users")
		if err != nil {
			return nil, fmt.Errorf("failed to read /Users: %v", err)
		}
		for _, user := range userDirs {
			if !user.IsDir() {
				continue
			}
			home := filepath.Join("/Users", user.Name())
			base := filepath.Join(home, "Library", "Application Support")
			candidates := []string{
				"Google/Chrome/Default/Bookmarks",
				"BraveSoftware/Brave-Browser/Default/Bookmarks",
			}
			for _, c := range candidates {
				full := filepath.Join(base, c)
				if _, err := os.Stat(full); err == nil {
					paths = append(paths, full)
				}
			}
		}

	case "windows":
		userDirs, err := os.ReadDir(`C:\Users`)
		if err != nil {
			return nil, fmt.Errorf("failed to read C:\\Users: %v", err)
		}
		for _, user := range userDirs {
			fmt.Println(user.Name())
			if !user.IsDir() {
				continue
			}
			home := filepath.Join(`C:\Users`, user.Name())
			base := filepath.Join(home, `AppData\Local`)
			candidates := []string{
				"Chromium\\User Data\\Default\\Bookmarks",
				"Google\\Chrome\\User Data\\Default\\Bookmarks",
				"BraveSoftware\\Brave-Browser\\User Data\\Default\\Bookmarks",
				"Microsoft\\Edge\\User Data\\Default\\Bookmarks",
				"Opera Software\\Opera Stable\\Bookmarks",
				"Vivaldi\\User Data\\Default\\Bookmarks",
			}
			for _, c := range candidates {
				full := filepath.Join(base, c)
				if _, err := os.Stat(full); err == nil {
					paths = append(paths, full)
				}
			}
		}

	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return paths, nil
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
