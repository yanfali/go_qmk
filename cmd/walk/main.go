package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeyboardMetadata model
type KeyboardMetadata struct {
	Name    string
	Keymaps []string
	Config  map[string]string
	Rules   map[string]string
}

func main() {
	myDir := "/Users/yanfa.li/projects/qmk_firmware/keyboards"

	metas := map[string]KeyboardMetadata{}
	keebSet := map[string]bool{}
	findMetadata := func(aPath string, anInfo os.FileInfo, anErr error) error {
		if anErr != nil {
			fmt.Printf("error %v at path %q\n", anErr, aPath)
			return anErr
		}

		if anInfo.IsDir() {
			if !strings.Contains(aPath, "keymaps") {
				if _, err := os.Stat(aPath + "/rules.mk"); !os.IsNotExist(err) {
					idx := strings.LastIndex(aPath, "keyboards/")
					newPath := fmt.Sprintf("%s", aPath[idx+len("keyboards/"):])
					keebSet[newPath] = true

					prefixPath := filepath.Dir(newPath)

					if keebSet[prefixPath] {
						delete(keebSet, prefixPath)
					}
				}
			}
		}

		return nil
	}

	matchInfoJSON := func(aPath string, anInfo os.FileInfo, anErr error) error {
		if anErr != nil {
			fmt.Printf("error %v at path %q\n", anErr, aPath)
			return anErr
		}

		keyboards := []string{}
		if !anInfo.IsDir() && anInfo.Name() == "info.json" {
			//fmt.Printf("  potential keyboard dir: %q\n", filepath.Dir(aPath))
			keyboards = append(keyboards, aPath)
		}

		for _, keebPath := range keyboards {
			baseDir := filepath.Dir(keebPath)
			keyboardName := filepath.Base(keebPath)
			metas[keyboardName] = KeyboardMetadata{Name: keyboardName}
			err := filepath.Walk(baseDir, findMetadata)

			if err != nil {
				fmt.Printf("error walking the path %q: %v\n", baseDir, err)
			}
		}

		return nil
	}

	err := filepath.Walk(myDir, matchInfoJSON)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", myDir, err)
	}

	for key := range keebSet {
		fmt.Printf("%v\n", key)
		rulesPath := filepath.Join(myDir, key, "rules.mk")
		fmt.Println(rulesPath)
		f, err := os.Open(rulesPath)
		if err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "#") || len(line) == 0 {
					continue
				}
				tokens := strings.Split(line, "=")
				if len(tokens) > 1 {
					key := strings.TrimSpace(tokens[0])
					value := strings.TrimSpace(tokens[1])
					if len(strings.Split(value, "#")) > 1 {
						value = strings.TrimSpace(strings.Split(value, "#")[0])
					}
					fmt.Printf("%q=%q\n", key, value)
				}

			}
		}
	}

}
