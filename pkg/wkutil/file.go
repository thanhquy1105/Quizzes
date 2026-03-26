package wkutil

import (
	"os"
)

func WriteFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)

}
