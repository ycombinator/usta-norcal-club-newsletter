import type { FutureMatchRecord, NewsletterData } from "./types";

declare global {
  interface Window {
    google?: {
      accounts: {
        oauth2: {
          initTokenClient(config: {
            client_id: string;
            scope: string;
            callback: (response: { access_token?: string; error?: string }) => void;
          }): { requestAccessToken(): void };
        };
      };
    };
  }
}

const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID as string | undefined;

export function googleCalendarConfigured(): boolean {
  return Boolean(googleClientId);
}

function requestToken(): Promise<string> {
  return new Promise((resolve, reject) => {
    if (!googleClientId) {
      reject(new Error("Google Calendar is not configured for this site."));
      return;
    }
    if (!window.google) {
      const script = document.createElement("script");
      script.src = "https://accounts.google.com/gsi/client";
      script.async = true;
      script.onload = () => requestToken().then(resolve, reject);
      script.onerror = () => reject(new Error("Google authorization could not be loaded."));
      document.head.append(script);
      return;
    }
    const client = window.google.accounts.oauth2.initTokenClient({
      client_id: googleClientId,
      scope: "https://www.googleapis.com/auth/calendar.events https://www.googleapis.com/auth/calendar.readonly",
      callback: (response) => response.access_token ? resolve(response.access_token) : reject(new Error(response.error || "Google authorization failed.")),
    });
    client.requestAccessToken();
  });
}

async function googleRequest<T>(token: string, url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null) as { error?: { message?: string } } | null;
    throw new Error(body?.error?.message || `Google Calendar request failed (${response.status})`);
  }
  return response.status === 204 ? undefined as T : response.json() as Promise<T>;
}

function eventId(match: FutureMatchRecord, index: number): string {
  const value = `${match.date}|${match.time}|${match.level}|${match.opponent}|${match.is_home}|${index}`;
  let hash = 2166136261;
  for (let i = 0; i < value.length; i++) {
    hash ^= value.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return `usta${(hash >>> 0).toString(16)}`;
}

function eventBody(match: FutureMatchRecord) {
  const startTime = match.time || "18:00";
  const start = new Date(`${match.date}T${startTime}:00`);
  const end = new Date(start.getTime() + 3 * 60 * 60 * 1000);
  return {
    summary: `${match.is_home ? "🏠" : "🚗"} ${match.gender_emoji}${match.level}${match.superscript || ""} ${match.is_home ? "vs." : "@"} ${match.opponent}`,
    location: match.location_note || undefined,
    start: { dateTime: start.toISOString(), timeZone: "America/Los_Angeles" },
    end: { dateTime: end.toISOString(), timeZone: "America/Los_Angeles" },
  };
}

export async function syncGoogleCalendar(data: NewsletterData, calendarName: string): Promise<number> {
  const token = await requestToken();
  const calendars = await googleRequest<{ items?: Array<{ id: string; summary: string }> }>(token, "https://www.googleapis.com/calendar/v3/users/me/calendarList");
  const calendar = calendars.items?.find((item) => item.summary === calendarName || item.id === calendarName);
  if (!calendar) throw new Error(`Google Calendar “${calendarName}” was not found.`);

  for (const [index, match] of data.future_matches.entries()) {
    const id = eventId(match, index);
    const eventUrl = `https://www.googleapis.com/calendar/v3/calendars/${encodeURIComponent(calendar.id)}/events/${id}`;
    const existing = await fetch(eventUrl, { headers: { Authorization: `Bearer ${token}` } });
    if (existing.ok) {
      await googleRequest(token, eventUrl, { method: "PUT", body: JSON.stringify(eventBody(match)) });
    } else if (existing.status === 404) {
      await googleRequest(token, `https://www.googleapis.com/calendar/v3/calendars/${encodeURIComponent(calendar.id)}/events`, {
        method: "POST",
        body: JSON.stringify({ id, ...eventBody(match) }),
      });
    } else {
      throw new Error(`Google Calendar lookup failed (${existing.status})`);
    }
  }
  return data.future_matches.length;
}
