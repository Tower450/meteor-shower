package main

import (
	"fmt"
	"io"
	"os"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Function to copy a file to a temporary location
func copyFile(sourcePath, tempPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}
	return nil
}
