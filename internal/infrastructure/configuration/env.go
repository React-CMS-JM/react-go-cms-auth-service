package configuration

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotEnv reads KEY=VALUE lines. Existing process environment wins, matching Quarkus.
func LoadDotEnv(path string) {
	var file *os.File
	var err error
	file, err = os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	var scanner *bufio.Scanner
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line string
		line = strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var key string
		var val string
		var ok bool
		key, val, ok = strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
