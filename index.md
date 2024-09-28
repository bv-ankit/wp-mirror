Here's the index of the Go project in markdown format:

# WordPress Update Server Project Index

## Files and Summaries

1. `src/download_checker.go`: Background job to check and queue new downloads
2. `src/download_worker.go`: Worker to download and process queued items
3. `src/redis_storage.go`: Redis storage operations for WordPress data
4. `src/server.go`: HTTP server handling API endpoints
5. `src/wp_updater.go`: Periodic updater for WordPress core, plugins, and themes

## Functions and I/O

### download_checker.go

- `BackgroundDownloadChecker()`: No input, no output. Runs continuously.
- `checkAndQueueDownloads()`: No input, returns error.
- `fileExists(filename string)`: Input: filename, Output: bool.
- `main()`: No input, no output. Program entry point.

### download_worker.go

- `DownloadWorker(id int, wg *sync.WaitGroup)`: Input: worker id and WaitGroup, no output.
- `downloadFile(item DownloadItem)`: Input: DownloadItem, returns error.
- `updateRedisInfo(item DownloadItem)`: Input: DownloadItem, returns error.
- `StartDownloadWorkers()`: No input, no output.
- `main()`: No input, no output. Program entry point.

### redis_storage.go

- `InitRedis(addr string)`: Input: Redis address, returns error.
- `SetCoreVersions(versions []CoreVersion)`: Input: CoreVersion slice, returns error.
- `GetCoreVersions()`: No input, returns CoreVersion slice and error.
- `SetPluginVersions(pluginFile string, versions []PluginVersion)`: Input: plugin file and PluginVersion slice, returns error.
- `GetPluginVersions(pluginFile string)`: Input: plugin file, returns PluginVersion slice and error.
- `SetThemeVersions(themeSlug string, versions []ThemeVersion)`: Input: theme slug and ThemeVersion slice, returns error.
- `GetThemeVersions(themeSlug string)`: Input: theme slug, returns ThemeVersion slice and error.
- `ListAllPluginFiles()`: No input, returns string slice and error.
- `ListAllThemeSlugs()`: No input, returns string slice and error.
- `GetLatestPluginVersion(pluginFile string)`: Input: plugin file, returns PluginVersion pointer and error.
- `GetLatestThemeVersion(themeSlug string)`: Input: theme slug, returns ThemeVersion pointer and error.

### server.go

- `main()`: No input, no output. Program entry point.
- `handleCoreUpdateCheck(c *gin.Context)`: Input: Gin context, no output.
- `handlePluginInfoBulk(c *gin.Context)`: Input: Gin context, no output.
- `handleThemeInfoBulk(c *gin.Context)`: Input: Gin context, no output.
- `handleCoreDownload(c *gin.Context)`: Input: Gin context, no output.
- `handlePluginDownload(c *gin.Context)`: Input: Gin context, no output.
- `handleThemeDownload(c *gin.Context)`: Input: Gin context, no output.

### wp_updater.go

- `runWPUpdater()`: No input, no output. Runs continuously.
- `acquireLock()`: No input, returns bool.
- `releaseLock()`: No input, no output.
- `updateWordPressInfo()`: No input, no output.
- `updateCoreVersions()`: No input, no output.
- `updatePlugins()`: No input, no output.
- `updateThemes()`: No input, no output.
- `main()`: No input, no output. Program entry point.

This index provides an overview of the project structure, file summaries, and function signatures with their inputs and outputs.
