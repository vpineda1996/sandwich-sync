package scotia

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/vpnda/sandwich-sync/pkg/parser"
)

func (s *ScotiaClient) AuthenticateCurl(ctx context.Context) error {
	fmt.Println("Paste an authenticated curl command here once complete write EOF:")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 128*1024*1024), 128*1024*1024)
	var multilineInput strings.Builder

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "EOF" {
			break
		}

		// Handle multiline input
		if strings.HasSuffix(trimmedLine, "\\") {
			multilineInput.WriteString(trimmedLine[:len(trimmedLine)-1])
			multilineInput.WriteString(" ")
			continue
		} else {
			multilineInput.WriteString(line)
		}
	}

	input := multilineInput.String()

	if len(input) == 0 {
		return fmt.Errorf("no input provided")
	}

	return s.hydrateClientCookies(ctx, input)
}

func (s *ScotiaClient) hydrateClientCookies(ctx context.Context, input string) error {
	cmd, err := parser.ParseCurlCommand(input)
	if err != nil {
		return err
	}

	url, err := url.Parse("https://secure.scotiabank.com")
	if err != nil {
		return err
	}

	for key, value := range cmd.Cookies {
		s.authClient.Jar.SetCookies(url, []*http.Cookie{
			{
				Name:  key,
				Value: value,
			},
		})
	}

	return s.validateHealthySession(ctx, cmd.Headers)
}
