package main

import (
	"os"
	"strings"
)

// Read node_name from node.toml configuration file
func getNodeName() string {
	// Try common paths for node.toml
	paths := []string{
		"/root/.monad/config/node.toml",
		"../monad-bft/config/node.toml",
		"./config/node.toml",
	}

	var content []byte
	var err error

	for _, path := range paths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "Monad Node"
	}

	// Simple TOML parsing for node_name
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "node_name") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				// Remove quotes
				name = strings.Trim(name, `"`)
				name = strings.Trim(name, `'`)
				return name
			}
		}
	}

	return "Monad Node"
}
