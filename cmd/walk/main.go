package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeyboardMetadata model
type KeyboardMetadata struct {
	Name    string
	Keymaps []string
}

func main() {
	myDir := "/Users/yanfa.li/projects/qmk_firmware/keyboards"

	metas := map[string]KeyboardMetadata{}
	findMetadata := func(aPath string, anInfo os.FileInfo, anErr error) error {
		if anErr != nil {
			fmt.Printf("error %v at path %q\n", anErr, aPath)
			return anErr
		}

		if anInfo.IsDir() {
			if !strings.Contains(aPath, "keymaps") {
				if _, err := os.Stat(aPath + "/rules.mk"); !os.IsNotExist(err) {
					idx := strings.LastIndex(aPath, "keyboards/")
					fmt.Printf("%s\n", aPath[idx+len("keyboards/"):])
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
			fmt.Printf("  potential keyboard dir: %q\n", filepath.Dir(aPath))
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
}
