package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/src-d/enry.v1"
)

// Map to store the count of each language
var languageCount = make(map[string]int)
var totalFiles int

// List of paths to ignore
var ignoreList = []string{"node_modules", "vendor", "target"}
var unwantedLanguages = []string{"text", "markdown"}

// detectLanguage prints the detected language of the file
func detectLanguage(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return
	}

	language := enry.GetLanguage(filePath, content)
	// fmt.Printf("%s \t %s\n", filePath, language)

	// Update the language count
	if language != "" {
		languageCount[language]++
		totalFiles++
	}
}

// shouldIgnore checks if a path should be ignored
func shouldIgnore(path string) bool {
	// Ignore hidden files and directories
	if enry.IsDotFile(path) || enry.IsDocumentation(path) || enry.IsImage(path) {
		return true
	}
	for _, ignore := range ignoreList {
		if strings.Contains(path, ignore) {
			return true
		}
	}
	return false
}

// walkDirectory walks through the directory and detects the language of each file
func walkDirectory(root string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the path should be ignored
		if shouldIgnore(path) {
			if info.IsDir() {
				// Skip the directory
				return filepath.SkipDir
			}
			// Skip the file
			return nil
		}

		// Only process regular files
		if !info.IsDir() {
			detectLanguage(path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking through directory: %v", err)
	}
}

// printLanguagePercentages prints the percentage of each language
func printLanguagePercentages() {
	percentages := make(map[string]float64)
	for language, count := range languageCount {
		unwanted := false
		for _, unwantedLanguage := range unwantedLanguages {
			if strings.ToLower(language) == unwantedLanguage {
				unwanted = true
				break
			}
		}
		if unwanted {
				continue	
		}
		percentage := (float64(count) / float64(totalFiles))
		// if percentage < 0.2 {
		// 	continue
		// }
		percentages[language] = percentage
	}

	jsonOutput, err := json.MarshalIndent(percentages, "", "  ")
	if err != nil {
		log.Fatalf("Error creating JSON output: %v", err)
	}

	fmt.Println(string(jsonOutput))
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a file or directory path")
	}

	inputPath := os.Args[1]

	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalf("Failed to stat input: %v", err)
	}

	// Create a channel to signal completion
	done := make(chan bool)

	go func() {
		if fileInfo.IsDir() {
			// Input is a directory, walk through it
			walkDirectory(inputPath)
		} else {
			// Input is a file, detect language directly
			detectLanguage(inputPath)
		}

		// Print the percentages of each language
		printLanguagePercentages()

		// Signal completion
		done <- true
	}()

	// Set a timeout for the operation
	select {
	case <-done:
		// Operation completed within the timeout
	case <-time.After(3000 * time.Millisecond): // Wait for 0.3 seconds
		// Timeout reached, terminate the program
	}
}
