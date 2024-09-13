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
	loadConfig()

	if config.WatchDir == config.ZipDir {
		log.Fatal("Watch directory and zip directory cannot be the same.")
	}

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
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err != nil {
						log.Println("Error:", err)
						continue
					}
					if info.ModTime().After(lastModifiedTime) {
						log.Println("Modified file:", event.Name)
						updateJSONFile()
						if config.Zip {
							err := zipFolderContents(config.WatchDir, filepath.Join(config.ZipDir, config.ZipFileName))
							if err != nil {
								return
							}
						} else {
							err := copyFolderContents(config.WatchDir, config.ZipDir)
							if err != nil {
								return
							}
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

	err = watcher.Add(config.WatchDir)
	if err != nil {
		log.Fatal(err)
	}

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
