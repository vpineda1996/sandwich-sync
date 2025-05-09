package scotia

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

type cookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

func (c *cookie) ToHttpCookie() *http.Cookie {
	return &http.Cookie{
		Name:  c.Name,
		Value: c.Value,
	}
}

type session struct {
	AuthSession struct {
		MultiUserCookie cookie `json:"multi_user_cookie"`
		UserRsid        string `json:"user_rsid"`
		AuthToken       string `json:"auth_token"`
	} `json:"auth_session"`
	ClientSession struct {
		SessionIdCookie cookie            `json:"session_id_cookie"`
		BypassAkami     map[string]cookie `json:"bypass_akamai"`
	} `json:"client_session"`
}

func (s *ScotiaClient) AuthenticateDynamic(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := s.authCreate(ctx)
		if err != nil {
			return fmt.Errorf("failed to create auth: %w", err)
		}

		err = s.authValidate(ctx)
		if err != nil {
			if !errors.Is(err, ErrAuthRedirect) {
				log.Info().Msg("Auth redirect, retrying...")
				continue
			}
			return fmt.Errorf("failed to validate auth: %w", err)
		} else {
			return nil
		}
	}
}

//go:embed scotia.py
var scotiaAuthPyScript string

func (s *ScotiaClient) authCreate(ctx context.Context) error {
	// Create a command with context
	cmd := exec.CommandContext(ctx, "python3", "-")
	cmd.Stdin = strings.NewReader(scotiaAuthPyScript)
	log.Info().Msg("Executing Scotia authentication script")

	// Run the command
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		fmt.Print(stderr.String())
		return fmt.Errorf("failed to execute Scotia authentication script: %w", err)
	}
	return nil
}

func (s *ScotiaClient) authValidate(ctx context.Context) error {
	// assume the python script runs
	sess, err := s.readSessionFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	log.Info().Msg("Read session file successfully")

	// Set the cookies in the authClient
	url, _ := url.Parse("https://secure.scotiabank.com/")
	s.authClient.Jar.SetCookies(url, []*http.Cookie{
		sess.ClientSession.SessionIdCookie.ToHttpCookie(),
	})

	for _, v := range sess.ClientSession.BypassAkami {
		s.authClient.Jar.SetCookies(url, []*http.Cookie{
			v.ToHttpCookie(),
		})
	}
	return s.validateHealthySession(ctx, nil)
}

func (s *ScotiaClient) readSessionFile(_ context.Context) (session, error) {
	// Read the session file
	sessionFile, err := os.ReadFile("scotia_session.json")
	if err != nil {
		return session{}, fmt.Errorf("failed to read session file: %w", err)
	}

	// Unmarshal the session file into a Session struct
	var sess session
	err = json.Unmarshal(sessionFile, &sess)
	if err != nil {
		return session{}, fmt.Errorf("failed to unmarshal session file: %w", err)
	}

	return sess, nil
}
