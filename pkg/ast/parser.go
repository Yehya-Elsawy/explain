package ast

import (
	"strings"
	"unicode"
)

// Lex tokenizes a shell command string into a slice of Token.
func Lex(input string) ([]Token, error) {
	var tokens []Token
	runes := []rune(input)
	n := len(runes)
	i := 0

	for i < n {
		// Skip whitespace
		for i < n && unicode.IsSpace(runes[i]) {
			i++
		}
		if i >= n {
			break
		}

		startPos := i
		ch := runes[i]

		// Check multi-character operators
		if i+1 < n {
			two := string(runes[i : i+2])
			if two == "&&" {
				tokens = append(tokens, Token{Type: TokenAnd, Value: "&&", Pos: startPos})
				i += 2
				continue
			}
			if two == "||" {
				tokens = append(tokens, Token{Type: TokenOr, Value: "||", Pos: startPos})
				i += 2
				continue
			}
			if two == ">>" {
				tokens = append(tokens, Token{Type: TokenRedirectAppend, Value: ">>", Pos: startPos})
				i += 2
				continue
			}
			if two == "2>" {
				if i+3 < n && string(runes[i:i+4]) == "2>&1" {
					tokens = append(tokens, Token{Type: TokenRedirectErr, Value: "2>&1", Pos: startPos})
					i += 4
					continue
				}
				tokens = append(tokens, Token{Type: TokenRedirectErr, Value: "2>", Pos: startPos})
				i += 2
				continue
			}
			if two == "&>" {
				tokens = append(tokens, Token{Type: TokenRedirectOut, Value: "&>", Pos: startPos})
				i += 2
				continue
			}
		}

		// Single-character operators
		if ch == '|' {
			tokens = append(tokens, Token{Type: TokenPipe, Value: "|", Pos: startPos})
			i++
			continue
		}
		if ch == ';' {
			tokens = append(tokens, Token{Type: TokenSemi, Value: ";", Pos: startPos})
			i++
			continue
		}
		if ch == '>' {
			tokens = append(tokens, Token{Type: TokenRedirectOut, Value: ">", Pos: startPos})
			i++
			continue
		}
		if ch == '<' {
			tokens = append(tokens, Token{Type: TokenRedirectIn, Value: "<", Pos: startPos})
			i++
			continue
		}

		// Word parsing (handles quotes and escapes)
		var sb strings.Builder
		for i < n {
			c := runes[i]
			if unicode.IsSpace(c) {
				break
			}
			// Check if we hit an operator
			if c == '|' || c == ';' || c == '>' || c == '<' {
				break
			}
			if i+1 < n {
				two := string(runes[i : i+2])
				if two == "&&" || two == "||" || two == ">>" || two == "&>" {
					break
				}
			}

			if c == '\'' {
				// Single quoted string: read verbatim until closing single quote
				i++
				for i < n && runes[i] != '\'' {
					sb.WriteRune(runes[i])
					i++
				}
				if i < n && runes[i] == '\'' {
					i++
				}
			} else if c == '"' {
				// Double quoted string: handle backslash escapes
				i++
				for i < n && runes[i] != '"' {
					if runes[i] == '\\' && i+1 < n {
						i++
						sb.WriteRune(runes[i])
					} else {
						sb.WriteRune(runes[i])
					}
					i++
				}
				if i < n && runes[i] == '"' {
					i++
				}
			} else if c == '\\' {
				// Escaped character
				i++
				if i < n {
					sb.WriteRune(runes[i])
					i++
				}
			} else {
				sb.WriteRune(c)
				i++
			}
		}

		if sb.Len() > 0 {
			tokens = append(tokens, Token{Type: TokenWord, Value: sb.String(), Pos: startPos})
		}
	}

	return tokens, nil
}

// Known command wrappers/prefixes that run another command
var knownPrefixes = map[string]bool{
	"sudo":   true,
	"doas":   true,
	"nohup":  true,
	"time":   true,
	"nice":   true,
	"ionice": true,
	"env":    true,
	"exec":   true,
	"xargs":  true,
	"chroot": true,
}

// Parse converts a shell command string into a structured Pipeline.
func Parse(input string) (*Pipeline, error) {
	tokens, err := Lex(strings.TrimSpace(input))
	if err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		RawInput: input,
		Commands: []*SingleCommand{},
	}

	if len(tokens) == 0 {
		return pipeline, nil
	}

	currCmd := &SingleCommand{}
	i := 0

	for i < len(tokens) {
		tok := tokens[i]

		switch tok.Type {
		case TokenPipe:
			currCmd.PipedToNext = true
			pipeline.Commands = append(pipeline.Commands, currCmd)
			currCmd = &SingleCommand{}
			i++

		case TokenAnd:
			currCmd.ChainOp = "&&"
			pipeline.Commands = append(pipeline.Commands, currCmd)
			currCmd = &SingleCommand{}
			i++

		case TokenOr:
			currCmd.ChainOp = "||"
			pipeline.Commands = append(pipeline.Commands, currCmd)
			currCmd = &SingleCommand{}
			i++

		case TokenSemi:
			currCmd.ChainOp = ";"
			pipeline.Commands = append(pipeline.Commands, currCmd)
			currCmd = &SingleCommand{}
			i++

		case TokenRedirectOut, TokenRedirectAppend, TokenRedirectIn, TokenRedirectErr:
			target := ""
			if i+1 < len(tokens) && tokens[i+1].Type == TokenWord {
				target = tokens[i+1].Value
				i++
			}
			currCmd.Redirects = append(currCmd.Redirects, Redirect{
				Operator: tok.Value,
				Target:   target,
			})
			i++

		case TokenWord:
			val := tok.Value
			// If we don't have a command name yet:
			if currCmd.Name == "" {
				// Check if it's an environment variable assignment (e.g. VAR=val)
				if strings.Contains(val, "=") && !strings.HasPrefix(val, "-") {
					currCmd.EnvVars = append(currCmd.EnvVars, val)
				} else if knownPrefixes[val] {
					// It's a prefix like sudo or nohup
					currCmd.Prefixes = append(currCmd.Prefixes, val)
				} else {
					currCmd.Name = val
				}
			} else {
				currCmd.Args = append(currCmd.Args, val)
			}
			i++
		}
	}

	if currCmd.Name != "" || len(currCmd.Prefixes) > 0 || len(currCmd.Args) > 0 {
		pipeline.Commands = append(pipeline.Commands, currCmd)
	}

	return pipeline, nil
}
