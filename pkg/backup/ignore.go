package backup

import (
	"os"
	"path/filepath"
	"strings"
)

func LoadIgnoreRules(extraIgnore []string, addGit bool) map[string]bool {
	rules := make(map[string]bool)

	if !addGit {
		rules[".git"] = true
	}

	if home, err := os.UserHomeDir(); err == nil {
		globalIgnorePath := filepath.Join(home, ".archiverignore")
		if content, err := os.ReadFile(globalIgnorePath); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					cleaned := strings.TrimRight(line, "/\\")
					if cleaned == ".git" && addGit {
						continue
					}
					rules[cleaned] = true
				}
			}
		}
	}

	for _, item := range extraIgnore {
		cleaned := strings.TrimSpace(item)
		cleaned = strings.TrimRight(cleaned, "/\\")
		if cleaned != "" {
			rules[cleaned] = true
		}
	}

	return rules
}
