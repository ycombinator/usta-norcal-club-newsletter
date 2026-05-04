package formatters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var tokenDir = filepath.Join(os.Getenv("HOME"), ".usta-norcal")
var tokenFile = filepath.Join(tokenDir, "token.json")

func newCalendarService(ctx context.Context, credentialsFile string, reader io.Reader, writer io.Writer) (*calendar.Service, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("reading credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	tok, err := loadToken()
	if err != nil {
		tok, err = getTokenFromWeb(ctx, config, reader, writer)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tok); err != nil {
			return nil, err
		}
	}

	return calendar.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, tok)))
}

func loadToken() (*oauth2.Token, error) {
	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tok oauth2.Token
	if err := json.NewDecoder(f).Decode(&tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func saveToken(tok *oauth2.Token) error {
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	f, err := os.Create(tokenFile)
	if err != nil {
		return fmt.Errorf("creating token file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(tok)
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config, reader io.Reader, writer io.Writer) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Fprintf(writer, "Open this URL in your browser and authorize the application:\n%s\n\nPaste the authorization code here: ", authURL)

	var code string
	if _, err := fmt.Fscan(reader, &code); err != nil {
		return nil, fmt.Errorf("reading authorization code: %w", err)
	}

	tok, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging authorization code: %w", err)
	}

	return tok, nil
}
