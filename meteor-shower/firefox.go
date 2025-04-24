package main

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// Function to find Firefox profile folder dynamically
func findFirefoxProfile() ([]string, error) {
	// Run the `find` command to search for profile directories containing places.sqlite
	cmd := exec.Command("find", "/home", "-name", "*.default-release", "-o", "-name", "*.default", "-type", "d")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to find Firefox profile: %v", err)
	}

	profilePaths := strings.Split(string(output), "\n")

	// Filter out any empty lines
	var nonEmptyPaths []string
	for _, path := range profilePaths {
		if path != "" {
			nonEmptyPaths = append(nonEmptyPaths, path)
		}
	}

	// Return the list of profile paths
	return nonEmptyPaths, nil
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
		// fmt.Printf("🌠 - %s, %s: %s  %s\n", id, title, url, parentId)
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
