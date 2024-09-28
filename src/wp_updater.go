package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	wpAPIURL          = "https://api.wordpress.org/core/version-check/1.7/"
	wpPluginsAPIURL   = "https://api.wordpress.org/plugins/info/1.2/?action=query_plugins&request[per_page]=100"
	wpThemesAPIURL    = "https://api.wordpress.org/themes/info/1.1/?action=query_themes&request[per_page]=100"
	updateInterval    = 1 * time.Hour
	lockKey           = "wp_updater_lock"
	lockDuration      = 65 * time.Minute
)

func runWPUpdater() {
	for {
		if acquireLock() {
			log.Println("Starting WordPress update job")
			updateWordPressInfo()
			releaseLock()
			log.Println("Finished WordPress update job")
		} else {
			log.Println("Another instance is already running. Skipping this run.")
		}
		time.Sleep(updateInterval)
	}
}

func acquireLock() bool {
	success, err := rdb.SetNX(ctx, lockKey, "locked", lockDuration).Result()
	if err != nil {
		log.Printf("Error acquiring lock: %v", err)
		return false
	}
	return success
}

func releaseLock() {
	_, err := rdb.Del(ctx, lockKey).Result()
	if err != nil {
		log.Printf("Error releasing lock: %v", err)
	}
}

func updateWordPressInfo() {
	updateCoreVersions()
	updatePlugins()
	updateThemes()
}

func updateCoreVersions() {
	log.Println("Updating WordPress core versions")
	resp, err := http.Get(wpAPIURL)
	if err != nil {
		log.Printf("Error fetching WordPress core versions: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading WordPress core versions response: %v", err)
		return
	}

	var coreData struct {
		Offers []CoreVersion `json:"offers"`
	}
	err = json.Unmarshal(body, &coreData)
	if err != nil {
		log.Printf("Error unmarshalling WordPress core versions: %v", err)
		return
	}

	existingVersions, err := GetCoreVersions()
	if err != nil {
		log.Printf("Error fetching existing core versions: %v", err)
		return
	}

	for _, newVersion := range coreData.Offers {
		found := false
		for _, existingVersion := range existingVersions {
			if newVersion.Version == existingVersion.Version {
				found = true
				break
			}
		}
		if !found {
			log.Printf("Adding new core version: %s", newVersion.Version)
			err = SetCoreVersions([]CoreVersion{newVersion})
			if err != nil {
				log.Printf("Error adding new core version: %v", err)
			}
		}
	}
}

func updatePlugins() {
	log.Println("Updating WordPress plugins")
	resp, err := http.Get(wpPluginsAPIURL)
	if err != nil {
		log.Printf("Error fetching WordPress plugins: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading WordPress plugins response: %v", err)
		return
	}

	var pluginData struct {
		Plugins []PluginVersion `json:"plugins"`
	}
	err = json.Unmarshal(body, &pluginData)
	if err != nil {
		log.Printf("Error unmarshalling WordPress plugins: %v", err)
		return
	}

	for _, plugin := range pluginData.Plugins {
		existingVersions, err := GetPluginVersions(plugin.Slug)
		if err != nil && err != redis.Nil {
			log.Printf("Error fetching existing plugin versions for %s: %v", plugin.Slug, err)
			continue
		}

		found := false
		for _, existingVersion := range existingVersions {
			if plugin.NewVersion == existingVersion.NewVersion {
				found = true
				break
			}
		}
		if !found {
			log.Printf("Adding new plugin version: %s %s", plugin.Slug, plugin.NewVersion)
			err = SetPluginVersions(plugin.Slug, []PluginVersion{plugin})
			if err != nil {
				log.Printf("Error adding new plugin version: %v", err)
			}
		}
	}
}

func updateThemes() {
	log.Println("Updating WordPress themes")
	resp, err := http.Get(wpThemesAPIURL)
	if err != nil {
		log.Printf("Error fetching WordPress themes: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading WordPress themes response: %v", err)
		return
	}

	var themeData struct {
		Themes []ThemeVersion `json:"themes"`
	}
	err = json.Unmarshal(body, &themeData)
	if err != nil {
		log.Printf("Error unmarshalling WordPress themes: %v", err)
		return
	}

	for _, theme := range themeData.Themes {
		existingVersions, err := GetThemeVersions(theme.Theme)
		if err != nil && err != redis.Nil {
			log.Printf("Error fetching existing theme versions for %s: %v", theme.Theme, err)
			continue
		}

		found := false
		for _, existingVersion := range existingVersions {
			if theme.NewVersion == existingVersion.NewVersion {
				found = true
				break
			}
		}
		if !found {
			log.Printf("Adding new theme version: %s %s", theme.Theme, theme.NewVersion)
			err = SetThemeVersions(theme.Theme, []ThemeVersion{theme})
			if err != nil {
				log.Printf("Error adding new theme version: %v", err)
			}
		}
	}
}

func main() {
	err := InitRedis("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	go runWPUpdater()

	// Keep the main goroutine running
	select {}
}
