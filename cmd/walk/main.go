package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func matchInfoJSON(aPath string, anInfo os.FileInfo, anErr error) error {
	if anErr != nil {
		fmt.Printf("error %v at path %q\n", anErr, aPath)
		return anErr
	}
	if anInfo.Name() == "info.json" {
		fmt.Printf("  potential keyboard dir: %q\n", filepath.Dir(aPath))
	}
	return nil
}

func main() {
	myDir := "/Users/yanfa.li/projects/qmk_firmware/keyboards"
	err := filepath.Walk(myDir, matchInfoJSON)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", myDir, err)
	}
}
