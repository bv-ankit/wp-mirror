package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	publicFolder    = "./public"
	downloadQueue   = "download_queue"
	checkInterval   = 1 * time.Hour
)

type DownloadItem struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

func BackgroundDownloadChecker() {
	for {
		err := checkAndQueueDownloads()
		if err != nil {
			fmt.Printf("Error in background job: %v\n", err)
		}
		time.Sleep(checkInterval)
	}
}

func checkAndQueueDownloads() error {
	// Check core versions
	coreVersions, err := GetCoreVersions()
	if err != nil {
		return fmt.Errorf("error getting core versions: %w", err)
	}

	// Check plugin versions
	pluginFiles, err := ListAllPluginFiles()
	if err != nil {
		return fmt.Errorf("error listing plugin files: %w", err)
	}

	// Check theme versions
	themeSlugs, err := ListAllThemeSlugs()
	if err != nil {
		return fmt.Errorf("error listing theme slugs: %w", err)
	}

	var downloadItems []DownloadItem

	// Process core versions
	for _, core := range coreVersions {
		filename := fmt.Sprintf("wordpress-%s.zip", core.Version)
		if !fileExists(filepath.Join(publicFolder, filename)) {
			downloadItems = append(downloadItems, DownloadItem{
				Type:    "core",
				Version: core.Version,
				URL:     core.Package,
			})
		}
	}

	// Process plugin versions
	for _, pluginFile := range pluginFiles {
		latestVersion, err := GetLatestPluginVersion(pluginFile)
		if err != nil {
			fmt.Printf("Error getting latest version for plugin %s: %v\n", pluginFile, err)
			continue
		}

		filename := fmt.Sprintf("%s.%s.zip", pluginFile, latestVersion.NewVersion)
		if !fileExists(filepath.Join(publicFolder, filename)) {
			downloadItems = append(downloadItems, DownloadItem{
				Type:    "plugin",
				Version: latestVersion.NewVersion,
				URL:     latestVersion.Package,
			})
		}
	}

	// Process theme versions
	for _, themeSlug := range themeSlugs {
		latestVersion, err := GetLatestThemeVersion(themeSlug)
		if err != nil {
			fmt.Printf("Error getting latest version for theme %s: %v\n", themeSlug, err)
			continue
		}

		filename := fmt.Sprintf("%s.%s.zip", themeSlug, latestVersion.NewVersion)
		if !fileExists(filepath.Join(publicFolder, filename)) {
			downloadItems = append(downloadItems, DownloadItem{
				Type:    "theme",
				Version: latestVersion.NewVersion,
				URL:     latestVersion.Package,
			})
		}
	}

	// Store download items in Redis queue
	for _, item := range downloadItems {
		jsonData, err := json.Marshal(item)
		if err != nil {
			fmt.Printf("Error marshaling download item: %v\n", err)
			continue
		}

		err = rdb.RPush(ctx, downloadQueue, jsonData).Err()
		if err != nil {
			fmt.Printf("Error adding item to download queue: %v\n", err)
		}
	}

	fmt.Printf("Added %d items to download queue\n", len(downloadItems))
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func main() {
	// Initialize Redis connection
	err := InitRedis("localhost:6379")
	if err != nil {
		fmt.Printf("Error initializing Redis: %v\n", err)
		return
	}

	// Start the background job
	go BackgroundDownloadChecker()

	// Keep the main goroutine running
	select {}
}
