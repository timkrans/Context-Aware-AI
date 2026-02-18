package loadenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadEnv(filePath string) error {
	if filePath == "" {
		filePath = ".env"
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open .env file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Skipping invalid line: %s\n", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		err := os.Setenv(key, value)
		if err != nil {
			fmt.Printf("Error setting environment variable: %s=%s\n", key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	return nil
}
