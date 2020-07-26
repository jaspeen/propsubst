package main

import (
	"bufio"
	"io/ioutil"
	"strings"
)

// ReadPropertiesFile reads java-like properties file
func ReadPropertiesFile(filename string, res map[string]string) error {

	if len(filename) == 0 {
		return nil
	}

	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return ReadProperties(string(text), res)
}

// ReadProperties reads java-like properties from text
func ReadProperties(text string, res map[string]string) error {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				res[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
