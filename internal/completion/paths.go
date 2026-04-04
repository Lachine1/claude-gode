package completion

import (
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

// SuggestDirectories returns directory completions for a token
func SuggestDirectories(token string, cwd string) []SuggestionItem {
	dirPath, prefix := resolvePathToken(token, cwd)
	if dirPath == "" {
		return nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}

	var dirs []SuggestionItem
	count := 0
	for _, entry := range entries {
		if count >= 10 {
			break
		}
		if entry.IsDir() {
			name := entry.Name()
			if prefix != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
				continue
			}
			displayPath := buildDisplayPath(token, name) + "/"
			dirs = append(dirs, SuggestionItem{
				ID:          displayPath,
				DisplayText: displayPath,
				Tag:         "dir",
				Type:        SuggestionDirectory,
			})
			count++
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].DisplayText < dirs[j].DisplayText
	})

	return dirs
}

// SuggestPaths returns file and directory completions for a token
func SuggestPaths(token string, cwd string) []SuggestionItem {
	dirPath, prefix := resolvePathToken(token, cwd)
	if dirPath == "" {
		return nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}

	var results []SuggestionItem
	count := 0
	for _, entry := range entries {
		if count >= 100 {
			break
		}
		name := entry.Name()
		if prefix != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}

		displayPath := buildDisplayPath(token, name)
		itemType := SuggestionFile
		tag := "file"
		if entry.IsDir() {
			displayPath += "/"
			itemType = SuggestionDirectory
			tag = "dir"
		}

		results = append(results, SuggestionItem{
			ID:          displayPath,
			DisplayText: displayPath,
			Tag:         tag,
			Type:        itemType,
		})
		count++
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Type == SuggestionDirectory && results[j].Type != SuggestionDirectory {
			return true
		}
		if results[i].Type != SuggestionDirectory && results[j].Type == SuggestionDirectory {
			return false
		}
		return results[i].DisplayText < results[j].DisplayText
	})

	return results
}

func resolvePathToken(token string, cwd string) (dirPath string, prefix string) {
	token = strings.TrimPrefix(token, "@")

	if token == "" {
		return cwd, ""
	}

	if strings.HasPrefix(token, "~/") || token == "~" {
		usr, err := user.Current()
		if err != nil {
			return "", ""
		}
		home := usr.HomeDir
		rest := strings.TrimPrefix(token, "~")
		if rest == "" {
			return home, ""
		}
		dir, base := filepath.Split(strings.TrimPrefix(rest, "/"))
		return filepath.Join(home, dir), base
	}

	if strings.HasPrefix(token, "/") {
		dir, base := filepath.Split(token)
		return dir, base
	}

	dir, base := filepath.Split(token)
	if dir == "" {
		dir = "."
	}
	return filepath.Join(cwd, dir), base
}

func buildDisplayPath(token string, name string) string {
	token = strings.TrimPrefix(token, "@")
	dir := filepath.Dir(token)
	if dir == "." {
		return name
	}
	return filepath.Join(dir, name)
}
