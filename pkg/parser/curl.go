package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// CurlCommand represents a parsed curl command
type CurlCommand struct {
	URL     string
	Headers map[string]string
	Cookies map[string]string
}

// ParseCurlCommand parses a curl-like command string
func ParseCurlCommand(cmdStr string) (*CurlCommand, error) {
	cmd := &CurlCommand{
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	// Clean up the command string by removing line continuations
	cmdStr = strings.ReplaceAll(cmdStr, "\\\n", " ")

	// Extract URL - support both single quotes and caret-quoted strings
	urlRegex := regexp.MustCompile(`curl\s+(?:'([^']+)'|\^"([^"]+)\^")`)
	urlMatches := urlRegex.FindStringSubmatch(cmdStr)
	if len(urlMatches) < 2 {
		return nil, fmt.Errorf("failed to extract URL from curl command")
	}

	// Get the URL from either the single-quoted or caret-quoted group
	if urlMatches[1] != "" {
		cmd.URL = urlMatches[1]
	} else if len(urlMatches) > 2 && urlMatches[2] != "" {
		cmd.URL = urlMatches[2]
	}

	// Extract headers - support both single quotes and caret-quoted strings
	headerRegex := regexp.MustCompile(`-H\s+(?:'([^']+)'|\^"([^"]+)\^")`)
	headerMatches := headerRegex.FindAllStringSubmatch(cmdStr, -1)
	for _, match := range headerMatches {
		var header string
		if match[1] != "" {
			header = match[1]
		} else if len(match) > 2 && match[2] != "" {
			header = match[2]
		} else {
			continue
		}

		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Skip cookie header as we'll handle it separately
		if strings.ToLower(key) == "cookie" {
			continue
		}

		cmd.Headers[key] = value
	}

	// Extract cookies - support both single quotes and caret-quoted strings
	cookieRegex := regexp.MustCompile(`-b\s+(?:'([^']+)'|\^"([^"]+)\^")`)
	cookieMatches := cookieRegex.FindStringSubmatch(cmdStr)
	if len(cookieMatches) >= 2 {
		var cookieStr string
		if cookieMatches[1] != "" {
			cookieStr = cookieMatches[1]
		} else if len(cookieMatches) > 2 && cookieMatches[2] != "" {
			cookieStr = cookieMatches[2]
		}

		// Parse cookies
		if cookieStr != "" {
			parts := strings.Split(cookieStr, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				kv := strings.SplitN(part, "=", 2)
				if len(kv) != 2 {
					continue
				}

				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				cmd.Cookies[key] = value
			}
		}
	}

	return cmd, nil
}

// String returns a string representation of the curl command
func (c *CurlCommand) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("URL: %s\n", c.URL))

	sb.WriteString("Headers:\n")
	for key, value := range c.Headers {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
	}

	sb.WriteString("Cookies:\n")
	for key, value := range c.Cookies {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
	}

	return sb.String()
}
