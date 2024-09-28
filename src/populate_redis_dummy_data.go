package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

func main() {
	// Initialize Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Populate Core Versions
	coreVersions := []CoreVersion{
		{
			Version:        "6.2.1",
			PHPVersion:     "5.6.20",
			MySQLVersion:   "5.0",
			NewBundled:     "6.1",
			PartialVersion: false,
			Package:        "https://wp-mirror.blogvault.net/release/wordpress-6.2.1.zip",
			Current:        "6.2.1",
			Locale:         "en_US",
		},
		{
			Version:        "6.1.3",
			PHPVersion:     "5.6.20",
			MySQLVersion:   "5.0",
			NewBundled:     "6.0",
			PartialVersion: false,
			Package:        "https://wp-mirror.blogvault.net/release/wordpress-6.1.3.zip",
			Current:        "6.1.3",
			Locale:         "en_US",
		},
	}

	for _, v := range coreVersions {
		jsonData, _ := json.Marshal(v)
		err := rdb.HSet(ctx, "core_versions", v.Version, jsonData).Err()
		if err != nil {
			log.Printf("Error setting core version %s: %v", v.Version, err)
		}
	}

	// Populate Plugin Versions
	plugins := map[string][]PluginVersion{
		"contact-form-7/wp-contact-form-7.php": {
			{
				Slug:       "contact-form-7",
				NewVersion: "5.7.2",
				URL:        "https://wordpress.org/plugins/contact-form-7/",
				Package:    "https://wp-mirror.blogvault.net/plugin/contact-form-7.5.7.2.zip",
			},
		},
		"akismet/akismet.php": {
			{
				Slug:       "akismet",
				NewVersion: "5.1",
				URL:        "https://wordpress.org/plugins/akismet/",
				Package:    "https://wp-mirror.blogvault.net/plugin/akismet.5.1.zip",
			},
		},
	}

	for pluginFile, versions := range plugins {
		key := fmt.Sprintf("plugins:%s", pluginFile)
		for _, v := range versions {
			jsonData, _ := json.Marshal(v)
			err := rdb.HSet(ctx, key, v.NewVersion, jsonData).Err()
			if err != nil {
				log.Printf("Error setting plugin version %s for %s: %v", v.NewVersion, pluginFile, err)
			}
		}
	}

	// Populate Theme Versions
	themes := map[string][]ThemeVersion{
		"twentytwentythree": {
			{
				Theme:      "twentytwentythree",
				NewVersion: "1.1",
				URL:        "https://wordpress.org/themes/twentytwentythree/",
				Package:    "https://wp-mirror.blogvault.net/theme/twentytwentythree.1.1.zip",
			},
		},
		"twentytwentytwo": {
			{
				Theme:      "twentytwentytwo",
				NewVersion: "1.4",
				URL:        "https://wordpress.org/themes/twentytwentytwo/",
				Package:    "https://wp-mirror.blogvault.net/theme/twentytwentytwo.1.4.zip",
			},
		},
	}

	for themeSlug, versions := range themes {
		key := fmt.Sprintf("themes:%s", themeSlug)
		for _, v := range versions {
			jsonData, _ := json.Marshal(v)
			err := rdb.HSet(ctx, key, v.NewVersion, jsonData).Err()
			if err != nil {
				log.Printf("Error setting theme version %s for %s: %v", v.NewVersion, themeSlug, err)
			}
		}
	}

	fmt.Println("Dummy data populated successfully.")
}
