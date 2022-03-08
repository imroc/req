package tests

import (
	"os"
	"path/filepath"
)

var testDataPath string

func init() {
	pwd, _ := os.Getwd()
	testDataPath = filepath.Join(pwd, ".testdata")
}

// GetTestFilePath return test file absolute path.
func GetTestFilePath(filename string) string {
	return filepath.Join(testDataPath, filename)
}
