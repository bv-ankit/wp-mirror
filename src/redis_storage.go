package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"context"
	"errors"
)

var ctx = context.Background()

// Redis client
var rdb *redis.Client

// Structs for storing WordPress information
type CoreVersion struct {
	Version       string `json:"version"`
	PHPVersion    string `json:"php_version"`
	MySQLVersion  string `json:"mysql_version"`
	NewBundled    string `json:"new_bundled"`
	PartialVersion bool   `json:"partial_version"`
	Package       string `json:"package"`
	Current       string `json:"current"`
	Locale        string `json:"locale"`
}

type PluginVersion struct {
	Slug       string `json:"slug"`
	NewVersion string `json:"new_version"`
	URL        string `json:"url"`
	Package    string `json:"package"`
}

type ThemeVersion struct {
	Theme      string `json:"theme"`
	NewVersion string `json:"new_version"`
	URL        string `json:"url"`
	Package    string `json:"package"`
}

// InitRedis initializes the Redis connection
func InitRedis(addr string) error {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := rdb.Ping(ctx).Result()
	return err
}

// SetCoreVersions sets the list of core version information
func SetCoreVersions(versions []CoreVersion) error {
	for _, v := range versions {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = rdb.HSet(ctx, "core_versions", v.Version, jsonData).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetCoreVersions gets the list of core version information
func GetCoreVersions() ([]CoreVersion, error) {
	data, err := rdb.HGetAll(ctx, "core_versions").Result()
	if err != nil {
		return nil, err
	}

	versions := make([]CoreVersion, 0, len(data))
	for _, v := range data {
		var version CoreVersion
		err := json.Unmarshal([]byte(v), &version)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}

// SetPluginVersions sets the list of plugin version information for a given plugin file
func SetPluginVersions(pluginFile string, versions []PluginVersion) error {
	key := fmt.Sprintf("plugins:%s", pluginFile)
	for _, v := range versions {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = rdb.HSet(ctx, key, v.NewVersion, jsonData).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPluginVersions gets the list of plugin version information for a given plugin file
func GetPluginVersions(pluginFile string) ([]PluginVersion, error) {
	key := fmt.Sprintf("plugins:%s", pluginFile)
	data, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	versions := make([]PluginVersion, 0, len(data))
	for _, v := range data {
		var version PluginVersion
		err := json.Unmarshal([]byte(v), &version)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}

// SetThemeVersions sets the list of theme version information for a given theme slug
func SetThemeVersions(themeSlug string, versions []ThemeVersion) error {
	key := fmt.Sprintf("themes:%s", themeSlug)
	for _, v := range versions {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = rdb.HSet(ctx, key, v.NewVersion, jsonData).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetThemeVersions gets the list of theme version information for a given theme slug
func GetThemeVersions(themeSlug string) ([]ThemeVersion, error) {
	key := fmt.Sprintf("themes:%s", themeSlug)
	data, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	versions := make([]ThemeVersion, 0, len(data))
	for _, v := range data {
		var version ThemeVersion
		err := json.Unmarshal([]byte(v), &version)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}

// ListAllPluginFiles lists all stored plugin files
func ListAllPluginFiles() ([]string, error) {
	keys, err := rdb.Keys(ctx, "plugins:*").Result()
	if err != nil {
		return nil, err
	}

	pluginFiles := make([]string, len(keys))
	for i, key := range keys {
		pluginFiles[i] = key[8:] // Remove "plugins:" prefix
	}
	return pluginFiles, nil
}

// ListAllThemeSlugs lists all stored theme slugs
func ListAllThemeSlugs() ([]string, error) {
	keys, err := rdb.Keys(ctx, "themes:*").Result()
	if err != nil {
		return nil, err
	}

	themeSlugs := make([]string, len(keys))
	for i, key := range keys {
		themeSlugs[i] = key[7:] // Remove "themes:" prefix
	}
	return themeSlugs, nil
}

// GetLatestPluginVersion gets the latest plugin version information for a given plugin file
func GetLatestPluginVersion(pluginFile string) (*PluginVersion, error) {
	key := fmt.Sprintf("plugins:%s", pluginFile)
	versions, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, errors.New("no versions found for the plugin")
	}

	var latestVersion PluginVersion
	var latestVersionNumber string
	for version, data := range versions {
		if version > latestVersionNumber {
			latestVersionNumber = version
			err := json.Unmarshal([]byte(data), &latestVersion)
			if err != nil {
				return nil, err
			}
		}
	}
	return &latestVersion, nil
}

// GetLatestThemeVersion gets the latest theme version information for a given theme slug
func GetLatestThemeVersion(themeSlug string) (*ThemeVersion, error) {
	key := fmt.Sprintf("themes:%s", themeSlug)
	versions, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, errors.New("no versions found for the theme")
	}

	var latestVersion ThemeVersion
	var latestVersionNumber string
	for version, data := range versions {
		if version > latestVersionNumber {
			latestVersionNumber = version
			err := json.Unmarshal([]byte(data), &latestVersion)
			if err != nil {
				return nil, err
			}
		}
	}
	return &latestVersion, nil
}
