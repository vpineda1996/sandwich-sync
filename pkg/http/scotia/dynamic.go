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
	"time"

	"github.com/rs/zerolog/log"
)

type cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HttpOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

func (c *cookie) ToHttpCookie() *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: c.HttpOnly,
		Secure:   c.Secure,
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

		err := s.authValidate(ctx)
		if err != nil && !errors.Is(err, ErrAuthRedirect) &&
			!errors.Is(err, ErrReadingConfigFile) && !errors.Is(err, ErrAuthTimeout) {
			return fmt.Errorf("failed to validate auth: %w", err)
		} else if err == nil {
			log.Info().Msg("Auth validated successfully")
			return nil
		}

		log.Info().Err(err).Msg("Auth redirect or not config present, trying to refresh token...")
		err = s.authCreate(ctx)
		if err != nil {
			return fmt.Errorf("failed to create auth: %w", err)
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
	sess, err := s.readSessionFile(ctx)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrReadingConfigFile, err)
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = s.validateHealthySession(ctx, nil)
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("timeout while validating session: %w", ErrAuthTimeout)
	}
	return err
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
