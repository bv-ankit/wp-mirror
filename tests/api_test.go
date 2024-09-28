// tests/api_test.go
// go test ./tests -v

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	// Import your project's packages
	"your-project-path/src"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	// Setup your routes here
	r.GET("/core-update-check/", src.handleCoreUpdateCheck)
	r.POST("/plugin-info-bulk/", src.handlePluginInfoBulk)
	r.POST("/theme-info-bulk/", src.handleThemeInfoBulk)
	return r
}

func TestCoreUpdateCheck(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/core-update-check/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	updates, ok := response["updates"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, updates)

	update := updates[0].(map[string]interface{})
	assert.Contains(t, update, "version")
	assert.Contains(t, update, "php_version")
	assert.Contains(t, update, "mysql_version")
	assert.Contains(t, update, "package")
}

func TestPluginInfoBulk(t *testing.T) {
	router := setupRouter()

	requestBody := map[string]string{
		"contact-form-7/wp-contact-form-7.php": "contact-form-7",
		"akismet/akismet.php":                  "akismet",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/plugin-info-bulk/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	for pluginFile, _ := range requestBody {
		pluginInfo, ok := response[pluginFile].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, pluginInfo, "slug")
		assert.Contains(t, pluginInfo, "new_version")
		assert.Contains(t, pluginInfo, "url")
		assert.Contains(t, pluginInfo, "package")
	}
}

func TestThemeInfoBulk(t *testing.T) {
	router := setupRouter()

	requestBody := []string{"twentytwentythree", "twentytwentytwo"}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/theme-info-bulk/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	for _, themeSlug := range requestBody {
		themeInfo, ok := response[themeSlug].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, themeInfo, "theme")
		assert.Contains(t, themeInfo, "new_version")
		assert.Contains(t, themeInfo, "url")
		assert.Contains(t, themeInfo, "package")
	}
}
