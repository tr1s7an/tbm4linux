package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Config struct {
	ExampleURL   string `json:"example_url"`
	Version      string `json:"version"`
	Architecture map[string]struct {
		Url       string            `json:"url"`
		AssetName string            `json:"asset_name"`
		Extract   bool              `json:"extract"`
		Bin       map[string]string `json:"bin"`
		Folder    map[string]string `json:"folder"`
	} `json:"architecture"`
	Checkver map[string]string `json:"checkver"`
}

var (
	bucketPath        = "bucket"
	cachePath         = "/tmp"
	homeDir, _        = os.UserHomeDir()
	binanryPath       = filepath.Join(homeDir, ".local/bin")
	folderPath        = filepath.Join(homeDir, ".local")
	extractScriptPath = "/usr/local/bin/extract.sh"

	arch = "x64"
)

const (
	CONCURRENCY = 10
)

func main() {
	checkFlag := flag.Bool("c", false, "Check for updates")
	installFlag := flag.Bool("i", false, "Install the binary")
	updateFlag := flag.Bool("u", false, "Check for updates and install")
	flag.Parse()
	args := flag.Args()

	var ids []string
	if len(args) == 0 {
		fmt.Println("Please provide at least one binary id")
		return
	}

	if args[0] == "*" {
		files, err := filepath.Glob(filepath.Join(bucketPath, "*.json"))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, file := range files {
			id := strings.TrimSuffix(filepath.Base(file), ".json")
			ids = append(ids, id)
		}
	} else {
		ids = args[:]
	}
	var final []string
	if *checkFlag || *updateFlag {
		sem := make(chan bool, CONCURRENCY)
		var mutex sync.Mutex
		for _, id := range ids {
			sem <- true
			go func(id string) {
				defer func() { <-sem }()
				configFile := filepath.Join(bucketPath, id+".json")
				if _, err := os.Stat(configFile); err == nil {
					config, formattedConfig := readConfig(configFile)
					oldVersion := config.Version
					newVersion := checkVersion(formattedConfig)
					if newVersion != oldVersion {
						fmt.Printf("[%s]: <%s> -> <%s>\n", id, oldVersion, newVersion)
						config.Version = newVersion
						updateConfig(config, configFile)
						mutex.Lock()
						final = append(final, id)
						mutex.Unlock()
					} else {
						fmt.Printf("[%s]: <%s>\n", id, oldVersion)
					}
				}
			}(id)
		}
		for i := 0; i < cap(sem); i++ {
			sem <- true
		}
	} else {
		final = ids[:]
	}

	if *updateFlag || *installFlag {
		fmt.Println("********Installing********")
		for _, id := range final {
			configFile := filepath.Join(bucketPath, id+".json")
			err := install(id, configFile)
			if err != nil {
				fmt.Printf("[%s]: error installing binary. %v\n", id, err)
				continue
			}
			fmt.Printf("[%s]: done\n", id)
		}
	}
}

func readConfig(configFile string) (config Config, formattedConfig Config) {
	content, _ := os.ReadFile(configFile)
	json.Unmarshal(content, &config)
	formattedContent := []byte(strings.ReplaceAll(string(content), "<version>", config.Version))
	json.Unmarshal(formattedContent, &formattedConfig)
	return config, formattedConfig
}

func updateConfig(config Config, configFile string) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	enc.Encode(config)
	os.WriteFile(configFile, buf.Bytes(), 0644)
}

func checkVersion(config Config) string {
	req, _ := http.NewRequest("GET", config.Checkver["url"], nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	re := regexp.MustCompile(config.Checkver["pattern"])
	match := re.FindStringSubmatch(html)
	if len(match) < 2 {
		return ""
	}
	newVersion := match[1]
	return newVersion
}

func install(id string, configFile string) error {
	_, formattedConfig := readConfig(configFile)
	assetName := formattedConfig.Architecture[arch].AssetName
	workDir := filepath.Join(cachePath, id)

	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		os.Mkdir(workDir, 0755)
	}
	cwd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(cwd)
	fmt.Printf("[%s]: downloading...", id)
	fileResponse, err := http.Get(formattedConfig.Architecture[arch].Url)
	if err != nil {
		return err
	}
	defer fileResponse.Body.Close()
	fileContent, _ := io.ReadAll(fileResponse.Body)
	os.WriteFile(filepath.Join(workDir, assetName), fileContent, 0644)
	fmt.Print("ok")

	if formattedConfig.Architecture[arch].Extract {
		fmt.Print("...extracting...")
		err := exec.Command(extractScriptPath, assetName).Run()
		if err != nil {
			fmt.Println("failed!")
			return err
		}
		fmt.Println("ok")
	} else {
		fmt.Println()
	}

	for k, v := range formattedConfig.Architecture[arch].Folder {
		dst := filepath.Join(folderPath, v)
		fmt.Printf("[%s]: %s -> %s\n", id, k, dst)
		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("[%s]: confirm that %s is going to be deleted(y/n): ", id, dst)
			var confirmation string
			fmt.Scanln(&confirmation)
			if strings.ToLower(confirmation) == "y" {
				os.RemoveAll(dst)
			} else {
				fmt.Printf("[%s]: %s not updated\n", id, dst)
				continue
			}
		}
		exec.Command("cp", "-r", k, dst).Run()

	}
	for k, v := range formattedConfig.Architecture[arch].Bin {
		dst := filepath.Join(binanryPath, v)
		fmt.Printf("[%s]: %s -> %s\n", id, k, dst)
		exec.Command("install", "-m", "0755", k, dst).Run()

	}
	return nil
}
