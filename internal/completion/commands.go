package completion

import (
	"sort"
	"strings"
)

// SuggestCommands returns matching command suggestions for a query.
func (e *CompletionEngine) SuggestCommands(query string) []SuggestionItem {
	query = strings.TrimSpace(query)
	var results []SuggestionItem

	if query == "" {
		results = e.suggestRecentCommands()
		seen := make(map[string]bool)
		for _, r := range results {
			seen[r.ID] = true
		}
		sorted := make([]CommandInfo, len(e.commands))
		copy(sorted, e.commands)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})
		for _, cmd := range sorted {
			if len(results) >= 15 {
				break
			}
			if !seen[cmd.Name] {
				results = append(results, commandToItem(cmd))
				seen[cmd.Name] = true
			}
		}
		return results
	}

	type scored struct {
		cmd   CommandInfo
		score float64
	}
	var scoredCmds []scored

	for _, cmd := range e.commands {
		score := e.fuzzyScore(cmd, query)
		if score > 0 {
			scoredCmds = append(scoredCmds, scored{cmd, score})
		}
	}

	sort.SliceStable(scoredCmds, func(i, j int) bool {
		return scoredCmds[i].score > scoredCmds[j].score
	})

	for _, s := range scoredCmds {
		if len(results) >= 15 {
			break
		}
		results = append(results, commandToItem(s.cmd))
	}

	return results
}

func (e *CompletionEngine) RecordCommandUsage(cmdName string) {
	if e.recentCmds == nil {
		e.recentCmds = make(map[string]int)
	}
	e.recentCmds[cmdName]++
}

func (e *CompletionEngine) suggestRecentCommands() []SuggestionItem {
	type usage struct {
		name  string
		count int
	}
	var usages []usage
	for name, count := range e.recentCmds {
		usages = append(usages, usage{name, count})
	}
	sort.Slice(usages, func(i, j int) bool {
		return usages[i].count > usages[j].count
	})

	var results []SuggestionItem
	cmdMap := make(map[string]CommandInfo)
	for _, cmd := range e.commands {
		cmdMap[cmd.Name] = cmd
	}

	for _, u := range usages {
		if len(results) >= 15 {
			break
		}
		if cmd, ok := cmdMap[u.name]; ok {
			results = append(results, commandToItem(cmd))
		}
	}

	return results
}

func (e *CompletionEngine) fuzzyScore(cmd CommandInfo, query string) float64 {
	q := strings.ToLower(query)
	name := strings.ToLower(cmd.Name)

	if name == q {
		return 100.0
	}

	for _, alias := range cmd.Aliases {
		if strings.ToLower(alias) == q {
			return 90.0
		}
	}

	if strings.HasPrefix(name, q) {
		return 80.0 - float64(len(name))
	}

	for _, alias := range cmd.Aliases {
		a := strings.ToLower(alias)
		if strings.HasPrefix(a, q) {
			return 70.0 - float64(len(a))
		}
	}

	if fuzzyMatch(name, q) {
		return 0.5 * float64(len(q))
	}

	for _, alias := range cmd.Aliases {
		a := strings.ToLower(alias)
		if fuzzyMatch(a, q) {
			return 0.5 * float64(len(q))
		}
	}

	desc := strings.ToLower(cmd.Description)
	if strings.Contains(desc, q) {
		return 0.25
	}

	return 0
}

func fuzzyMatch(target, pattern string) bool {
	t := []rune(target)
	p := []rune(pattern)
	pi := 0
	for _, c := range t {
		if pi < len(p) && c == p[pi] {
			pi++
		}
	}
	return pi == len(p)
}

func commandToItem(cmd CommandInfo) SuggestionItem {
	return SuggestionItem{
		ID:          cmd.Name,
		DisplayText: "/" + cmd.Name,
		Tag:         "command",
		Description: cmd.Description,
		Type:        SuggestionCommand,
	}
}
