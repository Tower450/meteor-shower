package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Function to find Firefox profile folder dynamically
func findFirefoxProfile() ([]string, error) {
	var profiles []string

	switch runtime.GOOS {
	case "linux":
		// Uses `find` to scan /home for profile folders
		cmd := exec.Command("find", "/home", "-type", "d",
			"-name", "*.default-release", "-o", "-name", "*.default")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to find Firefox profiles (Linux): %v", err)
		}
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		profiles = append(profiles, lines...)

	case "darwin":
		// Scan /Users for Firefox profiles
		userDirs, err := os.ReadDir("/Users")
		if err != nil {
			return nil, fmt.Errorf("failed to read /Users: %v", err)
		}
		for _, user := range userDirs {
			if !user.IsDir() {
				continue
			}
			mozPath := filepath.Join("/Users", user.Name(), "Library", "Application Support", "Firefox", "Profiles")
			dirs, err := os.ReadDir(mozPath)
			if err != nil {
				continue // User might not have Firefox installed
			}
			for _, d := range dirs {
				if d.IsDir() && (strings.HasSuffix(d.Name(), ".default-release") || strings.HasSuffix(d.Name(), ".default")) {
					profiles = append(profiles, filepath.Join(mozPath, d.Name()))
				}
			}
		}

	case "windows":
		// Scan C:\Users\*\AppData\Roaming for Firefox profiles
		userDirs, err := os.ReadDir(`C:\Users`)
		if err != nil {
			return nil, fmt.Errorf("failed to read C:\\Users: %v", err)
		}
		for _, user := range userDirs {
			if !user.IsDir() {
				continue
			}
			mozPath := filepath.Join(`C:\Users`, user.Name(), `AppData\Roaming\Mozilla\Firefox\Profiles`)
			dirs, err := os.ReadDir(mozPath)
			if err != nil {
				continue
			}
			for _, d := range dirs {
				if d.IsDir() && (strings.HasSuffix(d.Name(), ".default-release") || strings.HasSuffix(d.Name(), ".default")) {
					profiles = append(profiles, filepath.Join(mozPath, d.Name()))
				}
			}
		}

	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Filter out empty results
	var nonEmpty []string
	for _, p := range profiles {
		if strings.TrimSpace(p) != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return nonEmpty, nil
}

// Function to query Firefox bookmarks from the copied SQLite DB
func extractFirefoxBookmarks(dbPath string) (bookmarks []Bookmark) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open Firefox database: %v\n", err)
	}
	defer db.Close()

	parentMap := make(map[int]string)
	rows, err := db.Query(`SELECT id, title FROM moz_bookmarks`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var title sql.NullString
		if err := rows.Scan(&id, &title); err != nil {
			log.Fatal(err)
		}
		parentMap[id] = title.String
	}
	rows.Close()

	query := `
	SELECT
	    moz_bookmarks.id as bookmark_id,
		moz_bookmarks.title AS bookmark_title,
		moz_places.url AS bookmark_url,
		moz_bookmarks.parent AS parent_id
	FROM
		moz_bookmarks
	JOIN
		moz_places ON moz_bookmarks.fk = moz_places.id
	WHERE
		moz_places.url IS NOT NULL
		AND moz_bookmarks.type = 1
	ORDER BY
		moz_bookmarks.dateAdded DESC;
	`

	rows, err = db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query Firefox bookmarks: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, title, url, parentId string
		err := rows.Scan(&id, &title, &url, &parentId)
		if err != nil {
			log.Fatalf("Failed to read Firefox bookmark row: %v\n", err)
		}
		if title == "" {
			title = "[No Title]"
		}

		parent, err := strconv.Atoi(parentId)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if *outputFlat {
			fmt.Printf("🌠 - %s, %s: %s  %s\n", id, title, url, parentId)
		}
		bookmarks = append(bookmarks, Bookmark{
			ID:     id,
			Name:   title,
			URL:    url,
			Parent: parentMap[parent],
		})
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error during Firefox row iteration: %v\n", err)
	}

	return bookmarks
}
