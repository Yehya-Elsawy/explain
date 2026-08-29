package manparser

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// FlagInfo contains documentation discovered for a command-line option.
type FlagInfo struct {
	Description string
	TakesValue  bool
	ValueName   string
}

var manPageCache sync.Map

// StripOverstrike removes terminal bolding/underlining artifacts (e.g. "_\bc" or "c\bc").
func StripOverstrike(s string) string {
	var buf bytes.Buffer
	runes := []rune(s)
	n := len(runes)
	i := 0
	for i < n {
		if i+1 < n && runes[i+1] == '\b' {
			// Skip character and backspace, read the next one.
			i += 2
			continue
		}
		buf.WriteRune(runes[i])
		i++
	}
	return buf.String()
}

func loadManPage(cmdName string) string {
	if cached, ok := manPageCache.Load(cmdName); ok {
		return cached.(string)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "man", "-P", "cat", cmdName)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		manPageCache.Store(cmdName, "")
		return ""
	}

	clean := StripOverstrike(stdout.String())
	manPageCache.Store(cmdName, clean)
	return clean
}

// ExtractCommandSummary parses the NAME section of a command's man page.
func ExtractCommandSummary(cmdName string) string {
	return ParseCommandSummary(loadManPage(cmdName))
}

// ParseCommandSummary parses a command summary from already-rendered man text.
// It accepts the ASCII hyphen as well as the en dash, em dash, and minus sign
// commonly emitted by BSD, macOS, and localized man pages.
func ParseCommandSummary(manText string) string {
	lines := strings.Split(StripOverstrike(manText), "\n")
	inName := false
	var nameLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, "NAME") {
			inName = true
			continue
		}
		if !inName {
			continue
		}
		if trimmed == "" {
			continue
		}
		if isSectionHeading(line) {
			break
		}
		nameLines = append(nameLines, trimmed)
	}

	joined := strings.Join(nameLines, " ")
	separator := regexp.MustCompile(`\s+[-–—−]\s+`)
	if match := separator.FindStringIndex(joined); match != nil {
		return strings.TrimSpace(joined[match[1]:])
	}
	return ""
}

// ExtractFlagDescription parses the man page to find a description for a flag.
func ExtractFlagDescription(cmdName, flag string) string {
	return ExtractFlagInfo(cmdName, flag).Description
}

// ExtractFlagInfo finds a flag's description and whether it consumes a value.
func ExtractFlagInfo(cmdName, flag string) FlagInfo {
	return ParseFlagInfo(loadManPage(cmdName), flag)
}

// ParseFlagInfo parses flag metadata from already-rendered man text.
func ParseFlagInfo(manText, flag string) FlagInfo {
	clean := StripOverstrike(manText)
	lines := strings.Split(clean, "\n")

	for i, line := range lines {
		declaration, inlineDescription, indent, ok := splitOptionLine(line)
		if !ok || !containsOptionToken(declaration, flag) {
			continue
		}

		descriptionParts := []string{}
		if inlineDescription != "" {
			descriptionParts = append(descriptionParts, inlineDescription)
		}

		for j := i + 1; j < len(lines); j++ {
			nextLine := lines[j]
			trimmed := strings.TrimSpace(nextLine)
			if trimmed == "" {
				if len(descriptionParts) > 0 {
					break
				}
				continue
			}
			if isSectionHeading(nextLine) {
				break
			}
			if _, _, _, nextIsOption := splitOptionLine(nextLine); nextIsOption {
				break
			}
			if leadingIndent(nextLine) <= indent && len(descriptionParts) > 0 {
				break
			}
			descriptionParts = append(descriptionParts, trimmed)
		}

		valueName := inferValueName(declaration)
		if valueName == "" {
			valueName = inferValueFromSynopsis(clean, flag)
		}

		return FlagInfo{
			Description: compactDescription(strings.Join(descriptionParts, " ")),
			TakesValue:  valueName != "",
			ValueName:   valueName,
		}
	}

	// A synopsis can still tell us that a flag consumes a value even when the
	// rendered page has no dedicated options section.
	valueName := inferValueFromSynopsis(clean, flag)
	return FlagInfo{TakesValue: valueName != "", ValueName: valueName}
}

func splitOptionLine(line string) (declaration, description string, indent int, ok bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "-") {
		return "", "", 0, false
	}

	indent = leadingIndent(line)
	separator := regexp.MustCompile(`\s{2,}`)
	if match := separator.FindStringIndex(trimmed); match != nil {
		declaration = strings.TrimSpace(trimmed[:match[0]])
		description = strings.TrimSpace(trimmed[match[1]:])
	} else {
		declaration = trimmed
	}
	return declaration, description, indent, true
}

func containsOptionToken(declaration, flag string) bool {
	for offset := 0; offset < len(declaration); {
		idx := strings.Index(declaration[offset:], flag)
		if idx == -1 {
			return false
		}
		idx += offset
		end := idx + len(flag)
		beforeOK := idx == 0 || isOptionBoundary(declaration[idx-1])
		afterOK := end == len(declaration) || isOptionBoundary(declaration[end]) || declaration[end] == '='
		if beforeOK && afterOK {
			return true
		}
		offset = idx + 1
	}
	return false
}

func isOptionBoundary(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == ',' || ch == '|' || ch == '/' || ch == '[' || ch == ']' || ch == '(' || ch == ')' || ch == '<' || ch == '>'
}

func inferValueName(declaration string) string {
	if match := regexp.MustCompile(`=\[?<?([A-Za-z][A-Za-z0-9_-]*)`).FindStringSubmatch(declaration); len(match) == 2 {
		return match[1]
	}
	if match := regexp.MustCompile(`<([A-Za-z][A-Za-z0-9_-]*)>`).FindStringSubmatch(declaration); len(match) == 2 {
		return match[1]
	}

	fields := strings.Fields(declaration)
	if len(fields) < 2 {
		return ""
	}
	candidate := strings.Trim(fields[len(fields)-1], "[]<>{}(),")
	candidate = strings.TrimSuffix(candidate, "...")
	if candidate == "" || strings.HasPrefix(candidate, "-") {
		return ""
	}
	return candidate
}

func inferValueFromSynopsis(manText, flag string) string {
	lines := strings.Split(manText, "\n")
	inSynopsis := false
	var synopsisLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, "SYNOPSIS") {
			inSynopsis = true
			continue
		}
		if !inSynopsis {
			continue
		}
		if trimmed == "" {
			continue
		}
		if isSectionHeading(line) {
			break
		}
		synopsisLines = append(synopsisLines, trimmed)
	}

	synopsis := strings.Join(synopsisLines, " ")
	for offset := 0; offset < len(synopsis); {
		idx := strings.Index(synopsis[offset:], flag)
		if idx == -1 {
			break
		}
		idx += offset
		end := idx + len(flag)
		beforeOK := idx == 0 || isOptionBoundary(synopsis[idx-1])
		afterOK := end == len(synopsis) || isOptionBoundary(synopsis[end]) || synopsis[end] == '='
		if beforeOK && afterOK {
			groupStart := strings.LastIndex(synopsis[:idx], "[")
			groupEndOffset := strings.Index(synopsis[end:], "]")
			if groupStart != -1 && groupEndOffset != -1 {
				group := synopsis[groupStart+1 : end+groupEndOffset]
				if containsOptionToken(group, flag) {
					if valueName := inferValueName(group); valueName != "" {
						return valueName
					}
				}
			}

			remainder := strings.TrimLeft(synopsis[end:], " =")
			if remainder != "" && !strings.ContainsRune("]|,)", rune(remainder[0])) && remainder[0] != '-' {
				candidate := strings.Fields(remainder)[0]
				candidate = strings.Trim(candidate, "[]<>{}(),")
				candidate = strings.TrimSuffix(candidate, "...")
				if candidate != "" && !strings.HasPrefix(candidate, "-") {
					return candidate
				}
			}
		}
		offset = idx + 1
	}
	return ""
}

func leadingIndent(line string) int {
	indent := 0
	for _, r := range line {
		if r == ' ' {
			indent++
			continue
		}
		if r == '\t' {
			indent += 4
			continue
		}
		break
	}
	return indent
}

func isSectionHeading(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || leadingIndent(line) != 0 {
		return false
	}
	hasLetter := false
	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				return false
			}
			continue
		}
		if !unicode.IsSpace(r) && !unicode.IsDigit(r) && r != '-' {
			return false
		}
	}
	return hasLetter
}

func compactDescription(description string) string {
	description = strings.Join(strings.Fields(description), " ")
	const maxLength = 240
	if len(description) <= maxLength {
		return description
	}
	if end := strings.LastIndex(description[:maxLength], ". "); end >= 80 {
		return description[:end+1]
	}
	if end := strings.LastIndex(description[:maxLength-3], " "); end != -1 {
		return description[:end] + "..."
	}
	return description[:maxLength-3] + "..."
}
