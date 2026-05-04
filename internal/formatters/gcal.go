package formatters

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type GCalFormatter struct {
	CredentialsFile string
	CalendarName    string
}

func (g *GCalFormatter) FormatUpcoming(data *PreparedData, cfg Config) error {
	ctx := context.Background()

	svc, err := newCalendarService(ctx, g.CredentialsFile, cfg.Reader, cfg.Writer)
	if err != nil {
		return fmt.Errorf("creating calendar service: %w", err)
	}

	calID, err := findCalendarByName(svc, g.CalendarName)
	if err != nil {
		return fmt.Errorf("finding calendar %q: %w", g.CalendarName, err)
	}

	slog.Info("using Google Calendar", "name", g.CalendarName, "id", calID)

	for _, m := range data.FutureMatches {
		if err := upsertEvent(ctx, svc, calID, m, data, cfg); err != nil {
			slog.Error("failed to upsert calendar event", "match_number", m.Number, "error", err)
		}
	}

	return nil
}

func findCalendarByName(svc *calendar.Service, name string) (string, error) {
	list, err := svc.CalendarList.List().Do()
	if err != nil {
		return "", fmt.Errorf("listing calendars: %w", err)
	}

	for _, item := range list.Items {
		if item.Summary == name {
			return item.Id, nil
		}
	}

	return "", fmt.Errorf("calendar %q not found", name)
}

func eventID(teamID, matchNumber int) string {
	return fmt.Sprintf("usta%dm%d", teamID, matchNumber)
}

func upsertEvent(ctx context.Context, svc *calendar.Service, calID string, m usta.Match, data *PreparedData, cfg Config) error {
	ourTeam, opponent, isHome := resolveTeams(m, data.Org)
	d := ourTeam.Display()
	opponent.LoadOrganization(ctx)

	opponentName := data.OrgNames.Resolve(cfg.Reader, cfg.Writer, opponent.Organization.Name)

	title := fmt.Sprintf("%s %s%s%s %s %s",
		locationEmoji(isHome),
		d.GenderEmoji(),
		d.Level,
		d.DaytimeEmoji(),
		locatorWord(isHome),
		opponentName,
	)

	var location string
	m.HomeTeam.LoadOrganization(ctx)
	m.HomeTeam.Organization.LoadAddress()
	if m.HomeTeam.Organization.Address != "" {
		location = m.HomeTeam.Organization.Address
	}

	start := m.Date
	if start.Hour() == 0 && start.Minute() == 0 {
		start = time.Date(start.Year(), start.Month(), start.Day(), 18, 0, 0, 0, start.Location())
	}
	end := start.Add(3 * time.Hour)

	event := &calendar.Event{
		Summary:  strings.TrimSpace(title),
		Location: location,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: "America/Los_Angeles",
		},
	}

	id := eventID(ourTeam.ID, m.Number)

	_, err := svc.Events.Get(calID, id).Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
			event.Id = id
			_, err = svc.Events.Insert(calID, event).Do()
			if err != nil {
				return fmt.Errorf("inserting event: %w", err)
			}
			slog.Info("created calendar event", "id", id, "title", event.Summary)
			return nil
		}
		return fmt.Errorf("checking existing event: %w", err)
	}

	_, err = svc.Events.Update(calID, id, event).Do()
	if err != nil {
		return fmt.Errorf("updating event: %w", err)
	}
	slog.Info("updated calendar event", "id", id, "title", event.Summary)
	return nil
}

func locatorWord(isHome bool) string {
	if isHome {
		return "vs."
	}
	return "@"
}
