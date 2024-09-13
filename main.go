package main

import (
	"archive/zip"
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config struct to hold configuration values from config.json
type Config struct {
	WatchDir    string `json:"watchDir"`
	ZipDir      string `json:"zipDir"`
	JsonFile    string `json:"jsonFile"`
	ZipFileName string `json:"zipFileName"`
	Zip         bool   `json:"zip"`
}

var (
	config           Config
	lastModifiedTime time.Time
)

func main() {
	// Load configuration from config.json
	loadConfig()

	// Check if watch directory and zip directory are the same
	if config.WatchDir == config.ZipDir {
		log.Fatal("Watch directory and zip directory cannot be the same.")
	}

	// Create a new file system watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Check if the event is a write or create event
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err != nil {
						log.Println("Error:", err)
						continue
					}
					// Check if the modification time is after the last modified time
					if info.ModTime().After(lastModifiedTime) {
						log.Println("Modified file:", event.Name)
						updateJSONFile()
						// Zip or copy folder contents based on the config
						if config.Zip {
							zipFolderContents(config.WatchDir, filepath.Join(config.ZipDir, config.ZipFileName))
						} else {
							copyFolderContents(config.WatchDir, config.ZipDir)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	// Add the watch directory to the watcher
	err = watcher.Add(config.WatchDir)
	if err != nil {
		log.Fatal(err)
	}

	// Walk through the watch directory and add all subdirectories to the watcher
	err = filepath.Walk(config.WatchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

// Load configuration from config.json
func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
}

// Update the JSON file with new UUIDs
func updateJSONFile() {
	filePath := filepath.Join(config.WatchDir, config.JsonFile)
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var jsonConfig struct {
		FormatVersion int `json:"format_version"`
		Header        struct {
			Description      string `json:"description"`
			Name             string `json:"name"`
			UUID             string `json:"uuid"`
			Version          []int  `json:"version"`
			MinEngineVersion []int  `json:"min_engine_version"`
		} `json:"header"`
		Modules []struct {
			Description string `json:"description"`
			Type        string `json:"type"`
			UUID        string `json:"uuid"`
			Version     []int  `json:"version"`
		} `json:"modules"`
	}

	err = json.Unmarshal(file, &jsonConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Generate new UUIDs for the header and modules
	jsonConfig.Header.UUID = uuid.New().String()
	for i := range jsonConfig.Modules {
		jsonConfig.Modules[i].UUID = uuid.New().String()
	}

	updatedFile, err := json.MarshalIndent(jsonConfig, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(filePath, updatedFile, 0644)
	if err != nil {
		log.Fatal(err)
	}

	lastModifiedTime = time.Now()
}

// Zip the contents of the source directory and save it to the target file
func zipFolderContents(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == source {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(source, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// Copy the contents of the source directory to the target directory
func copyFolderContents(source, target string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == source {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(target, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		targetFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, sourceFile)
		return err
	})
}
