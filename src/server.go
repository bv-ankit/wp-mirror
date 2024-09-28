package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const (
	coreDir    = "/path/to/core/files"
	pluginsDir = "/path/to/plugins/files"
	themesDir  = "/path/to/themes/files"
)

func main() {
	// Initialize Redis connection
	err := InitRedis("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	r := gin.Default()

	// Core update check endpoint
	r.GET("/core-update-check/", handleCoreUpdateCheck)

	// Plugin info bulk endpoint
	r.POST("/plugin-info-bulk/", handlePluginInfoBulk)

	// Theme info bulk endpoint
	r.POST("/theme-info-bulk/", handleThemeInfoBulk)

	// Core download endpoint
	r.GET("/core/:version.zip", handleCoreDownload)

	// Plugin download endpoint
	r.GET("/plugins/:plugin-slug/:version.zip", handlePluginDownload)

	// Theme download endpoint
	r.GET("/themes/:theme-slug/:version.zip", handleThemeDownload)

	r.Run(":8080")
}

func handleCoreUpdateCheck(c *gin.Context) {
	versions, err := GetCoreVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve core versions"})
		return
	}

	if len(versions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No core versions available"})
		return
	}

	// Assuming the latest version is the first in the list
	latestVersion := versions[0]

	response := gin.H{
		"updates": []gin.H{
			{
				"version":         latestVersion.Version,
				"php_version":     latestVersion.PHPVersion,
				"mysql_version":   latestVersion.MySQLVersion,
				"new_bundled":     latestVersion.NewBundled,
				"partial_version": latestVersion.PartialVersion,
				"package":         latestVersion.Package,
				"current":         latestVersion.Current,
				"locale":          latestVersion.Locale,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

func handlePluginInfoBulk(c *gin.Context) {
	var requestBody map[string]string
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response := make(map[string]interface{})

	for pluginFile, pluginSlug := range requestBody {
		latestVersion, err := GetLatestPluginVersion(pluginFile)
		if err != nil {
			log.Printf("Error retrieving plugin info for %s: %v", pluginFile, err)
			continue
		}

		response[pluginFile] = gin.H{
			"slug":        pluginSlug,
			"new_version": latestVersion.NewVersion,
			"url":         latestVersion.URL,
			"package":     latestVersion.Package,
		}
	}

	c.JSON(http.StatusOK, response)
}

func handleThemeInfoBulk(c *gin.Context) {
	var requestBody []string
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response := make(map[string]interface{})

	for _, themeSlug := range requestBody {
		latestVersion, err := GetLatestThemeVersion(themeSlug)
		if err != nil {
			log.Printf("Error retrieving theme info for %s: %v", themeSlug, err)
			continue
		}

		response[themeSlug] = gin.H{
			"theme":       themeSlug,
			"new_version": latestVersion.NewVersion,
			"url":         latestVersion.URL,
			"package":     latestVersion.Package,
		}
	}

	c.JSON(http.StatusOK, response)
}

func handleCoreDownload(c *gin.Context) {
	version := c.Param("version")
	filename := fmt.Sprintf("wordpress-%s.zip", version)
	filepath := filepath.Join(coreDir, filename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filepath)
}

func handlePluginDownload(c *gin.Context) {
	pluginSlug := c.Param("plugin-slug")
	version := c.Param("version")
	filename := fmt.Sprintf("%s.%s.zip", pluginSlug, version)
	filepath := filepath.Join(pluginsDir, filename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filepath)
}

func handleThemeDownload(c *gin.Context) {
	themeSlug := c.Param("theme-slug")
	version := c.Param("version")
	filename := fmt.Sprintf("%s.%s.zip", themeSlug, version)
	filepath := filepath.Join(themesDir, filename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filepath)
}
