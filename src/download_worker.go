package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	publicFolder  = "./public"
	downloadQueue = "download_queue"
	maxWorkers    = 5
)

func DownloadWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Pop an item from the download queue
		result, err := rdb.BLPop(ctx, 0, downloadQueue).Result()
		if err != nil {
			fmt.Printf("Worker %d: Error popping from queue: %v\n", id, err)
			continue
		}

		var item DownloadItem
		err = json.Unmarshal([]byte(result[1]), &item)
		if err != nil {
			fmt.Printf("Worker %d: Error unmarshaling download item: %v\n", id, err)
			continue
		}

		// Download the file
		err = downloadFile(item)
		if err != nil {
			fmt.Printf("Worker %d: Error downloading file: %v\n", id, err)
			continue
		}

		// Update Redis with the new file information
		err = updateRedisInfo(item)
		if err != nil {
			fmt.Printf("Worker %d: Error updating Redis info: %v\n", id, err)
		}
	}
}

func downloadFile(item DownloadItem) error {
	fmt.Printf("Downloading %s version %s\n", item.Type, item.Version)

	resp, err := http.Get(item.URL)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the directory structure if it doesn't exist
	dir := filepath.Join(publicFolder, item.Type)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Create the file
	filename := filepath.Join(dir, fmt.Sprintf("%s-%s.zip", item.Type, item.Version))
	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	fmt.Printf("Downloaded %s to %s\n", item.URL, filename)
	return nil
}

func updateRedisInfo(item DownloadItem) error {
	switch item.Type {
	case "core":
		// Update core version info
		var coreVersion CoreVersion
		coreVersion.Version = item.Version
		coreVersion.Package = item.URL
		jsonData, err := json.Marshal(coreVersion)
		if err != nil {
			return fmt.Errorf("error marshaling core version: %w", err)
		}
		err = rdb.HSet(ctx, "core_versions", item.Version, jsonData).Err()
		if err != nil {
			return fmt.Errorf("error updating core version in Redis: %w", err)
		}
	case "plugin":
		// Update plugin version info
		var pluginVersion PluginVersion
		pluginVersion.NewVersion = item.Version
		pluginVersion.Package = item.URL
		jsonData, err := json.Marshal(pluginVersion)
		if err != nil {
			return fmt.Errorf("error marshaling plugin version: %w", err)
		}
		err = rdb.HSet(ctx, fmt.Sprintf("plugins:%s", item.Slug), item.Version, jsonData).Err()
		if err != nil {
			return fmt.Errorf("error updating plugin version in Redis: %w", err)
		}
	case "theme":
		// Update theme version info
		var themeVersion ThemeVersion
		themeVersion.NewVersion = item.Version
		themeVersion.Package = item.URL
		jsonData, err := json.Marshal(themeVersion)
		if err != nil {
			return fmt.Errorf("error marshaling theme version: %w", err)
		}
		err = rdb.HSet(ctx, fmt.Sprintf("themes:%s", item.Slug), item.Version, jsonData).Err()
		if err != nil {
			return fmt.Errorf("error updating theme version in Redis: %w", err)
		}
	default:
		return fmt.Errorf("unknown item type: %s", item.Type)
	}

	return nil
}

func StartDownloadWorkers() {
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go DownloadWorker(i, &wg)
	}
	wg.Wait()
}

func main() {
	// Initialize Redis connection
	err := InitRedis("localhost:6379")
	if err != nil {
		fmt.Printf("Error initializing Redis: %v\n", err)
		return
	}

	// Start the download workers
	go StartDownloadWorkers()

	// Start the background checker
	go BackgroundDownloadChecker()

	// Keep the main goroutine running
	select {}
}
