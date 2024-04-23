package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadProperties(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	properties := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			properties[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return properties, nil
}

func OverwriteProperties(filepath string, properties map[string]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer file.Close()

	for key, value := range properties {
		_, err := fmt.Fprintf(file, "%s=%s\n", key, value)
		if err != nil {
			return err
		}
	}

	return nil
}
