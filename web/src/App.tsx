import { useEffect, useMemo, useRef, useState } from "react";
import { fetchOrganizationSchedule } from "./api";
import { googleCalendarConfigured, syncGoogleCalendar } from "./calendar";
import { downloadBoardJPEG, downloadHTML, downloadICS, downloadJSON } from "./export";
import { PublicationPreview } from "./Preview";
import { toNewsletterData } from "./transform";
import type { FutureMatchRecord, MatchType, NewsletterData, PastMatchRecord, Settings } from "./types";

const SETTINGS_KEY = "club-match-board:settings";
const DATA_KEY = "club-match-board:data";

function localISODate(date: Date): string {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 10);
}

const tomorrow = new Date();
tomorrow.setDate(tomorrow.getDate() + 1);

const defaultSettings: Settings = {
  organizationId: "225",
  extraTeamIds: "",
  boundaryDate: localISODate(tomorrow),
  pastDays: 7,
  futureDays: 14,
  apiBaseUrl: import.meta.env.VITE_NET_RESULTS_API_URL || "https://net-results-api.onrender.com",
  calendarName: "USTA Tennis",
};

function loadStored<T>(key: string, fallback: T): T {
  try {
    const value = localStorage.getItem(key);
    if (!value) return fallback;
    const parsed: unknown = JSON.parse(value);
    if (fallback && parsed && typeof fallback === "object" && typeof parsed === "object") {
      return { ...fallback, ...parsed } as T;
    }
    return parsed as T;
  } catch {
    return fallback;
  }
}

function isNewsletterData(value: unknown): value is NewsletterData {
  if (!value || typeof value !== "object") return false;
  const data = value as Partial<NewsletterData>;
  return typeof data.org_short_name === "string" && Array.isArray(data.past_matches) && Array.isArray(data.future_matches);
}

function outcomeState(match: PastMatchRecord): string {
  if (match.is_rained_out) return "rained";
  if (match.is_incomplete) return "incomplete";
  if (match.is_win) return "win";
  return "loss";
}

function blankPast(date: string): PastMatchRecord {
  return { date, gender_emoji: "👫", level: "", is_home: true, opponent: "", is_incomplete: true, outcome_text: "", match_type: "regular" };
}

function blankFuture(date: string): FutureMatchRecord {
  return { date, time: "18:00", gender_emoji: "👫", level: "", is_home: true, opponent: "", match_type: "regular" };
}

export default function App() {
  const [settings, setSettings] = useState<Settings>(() => loadStored(SETTINGS_KEY, defaultSettings));
  const [data, setData] = useState<NewsletterData | null>(() => {
    const stored = loadStored<NewsletterData | null>(DATA_KEY, null);
    return stored && isNewsletterData(stored) ? stored : null;
  });
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [failures, setFailures] = useState<string[]>([]);
  const publicationRef = useRef<HTMLDivElement>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  useEffect(() => localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings)), [settings]);
  useEffect(() => {
    if (data) localStorage.setItem(DATA_KEY, JSON.stringify(data));
  }, [data]);

  const needsReview = useMemo(() => data?.past_matches.filter((match) => match.is_incomplete || match.is_rained_out).length || 0, [data]);

  const updateSetting = <K extends keyof Settings>(key: K, value: Settings[K]) => setSettings((current) => ({ ...current, [key]: value }));

  async function loadSchedule(force = false) {
    setBusy(true);
    setError("");
    setMessage(force ? "Refreshing directly from USTA…" : "Loading club schedules…");
    try {
      const response = await fetchOrganizationSchedule(settings, force);
      setData(toNewsletterData(response, settings.boundaryDate));
      setFailures(response.failures?.map((failure) => `Team ${failure.team_id}: ${failure.error}`) || []);
      setMessage(`Loaded ${response.teams.length} teams. Review the match board below.`);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "The schedule could not be loaded.");
      setMessage("");
    } finally {
      setBusy(false);
    }
  }

  function importData(file?: File) {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      try {
        const imported: unknown = JSON.parse(String(reader.result));
        if (!isNewsletterData(imported)) throw new Error("This file does not match the newsletter data format.");
        setData(imported);
        setError("");
        setMessage(`Loaded ${file.name}.`);
      } catch (importError) {
        setError(importError instanceof Error ? importError.message : "The data file could not be read.");
      }
    };
    reader.readAsText(file);
  }

  function updatePast(index: number, changes: Partial<PastMatchRecord>) {
    if (!data) return;
    const matches = [...data.past_matches];
    matches[index] = { ...matches[index], ...changes };
    setData({ ...data, past_matches: matches });
  }

  function updateFuture(index: number, changes: Partial<FutureMatchRecord>) {
    if (!data) return;
    const matches = [...data.future_matches];
    matches[index] = { ...matches[index], ...changes };
    setData({ ...data, future_matches: matches });
  }

  function setOutcome(index: number, value: string) {
    const current = data?.past_matches[index];
    if (!current) return;
    const score = current.outcome_text?.replace(/^(won|lost)\s+/i, "") || "";
    if (value === "rained") updatePast(index, { is_rained_out: true, is_incomplete: false, is_win: false, outcome_text: "" });
    if (value === "incomplete") updatePast(index, { is_rained_out: false, is_incomplete: true, is_win: false, outcome_text: score });
    if (value === "win") updatePast(index, { is_rained_out: false, is_incomplete: false, is_win: true, outcome_text: `won ${score || "3-0"}` });
    if (value === "loss") updatePast(index, { is_rained_out: false, is_incomplete: false, is_win: false, outcome_text: `lost ${score || "0-3"}` });
  }

  async function saveJPEG(id: string, suffix: string) {
    if (!data) return;
    const element = document.getElementById(id);
    if (!element) return;
    setBusy(true);
    try {
      await downloadBoardJPEG(element, data, suffix);
      setMessage(`Downloaded ${suffix} JPEG.`);
    } catch {
      setError("The browser could not capture that preview. Try downloading HTML instead.");
    } finally {
      setBusy(false);
    }
  }

  async function syncCalendar() {
    if (!data) return;
    setBusy(true);
    setError("");
    setMessage("Waiting for Google authorization…");
    try {
      const count = await syncGoogleCalendar(data, settings.calendarName);
      setMessage(`Synced ${count} upcoming matches to ${settings.calendarName}.`);
    } catch (syncError) {
      setError(syncError instanceof Error ? syncError.message : "Calendar sync failed.");
      setMessage("");
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <header className="site-header">
        <div className="court-lines" aria-hidden="true"><span /><span /><span /></div>
        <div className="header-copy">
          <p className="eyebrow">USTA NorCal club newsletter</p>
          <h1>Turn the week’s matches into a <em>club match board.</em></h1>
          <p>Load the official schedule, correct the details, and publish recent results and upcoming fixtures without installing the CLI.</p>
        </div>
        <div className="tennis-ball" aria-hidden="true">●</div>
      </header>

      <main>
        <section className="setup-panel" aria-labelledby="setup-title">
          <div className="section-intro">
            <p className="step-label">01 · Schedule</p>
            <h2 id="setup-title">Choose the match window</h2>
            <p>The boundary date is the first day shown as upcoming.</p>
          </div>
          <form className="settings-grid" onSubmit={(event) => { event.preventDefault(); void loadSchedule(false); }}>
            <label>Organization ID<input value={settings.organizationId} inputMode="numeric" required onChange={(event) => updateSetting("organizationId", event.target.value)} /></label>
            <label>Extra team IDs<input value={settings.extraTeamIds} placeholder="123, 456" onChange={(event) => updateSetting("extraTeamIds", event.target.value)} /></label>
            <label>Boundary date<input type="date" value={settings.boundaryDate} required onChange={(event) => updateSetting("boundaryDate", event.target.value)} /></label>
            <label>Past days<input type="number" min="1" max="60" value={settings.pastDays} onChange={(event) => updateSetting("pastDays", Number(event.target.value))} /></label>
            <label>Future days<input type="number" min="1" max="60" value={settings.futureDays} onChange={(event) => updateSetting("futureDays", Number(event.target.value))} /></label>
            <div className="form-actions">
              <button className="primary-action" disabled={busy} type="submit">{busy ? "Loading…" : "Load schedule"}</button>
              <button className="quiet-action" type="button" onClick={() => fileRef.current?.click()}>Import data.json</button>
              <input ref={fileRef} className="visually-hidden" type="file" accept="application/json,.json" onChange={(event) => importData(event.target.files?.[0])} />
            </div>
          </form>
          {(message || error) && <div className={`status-message ${error ? "error" : "success"}`} role="status">{error || message}</div>}
          {failures.length > 0 && <details className="failure-list"><summary>{failures.length} teams could not be loaded</summary><ul>{failures.map((failure) => <li key={failure}>{failure}</li>)}</ul></details>}
        </section>

        {data && (
          <>
            <section className="review-section" aria-labelledby="review-title">
              <div className="section-intro split-intro">
                <div><p className="step-label">02 · Review</p><h2 id="review-title">Make the board accurate</h2><p>Changes save in this browser and update the publication preview immediately.</p></div>
                <div className="review-count"><strong>{needsReview}</strong><span>results to check</span></div>
              </div>

              <div className="editor-block">
                <div className="editor-heading"><h3>Recent results</h3><button type="button" className="text-action" onClick={() => setData({ ...data, past_matches: [...data.past_matches, blankPast(settings.boundaryDate)] })}>+ Add result</button></div>
                <div className="table-scroll">
                  <table className="editor-table">
                    <thead><tr><th>Date</th><th>Team</th><th>Site</th><th>Opponent</th><th>Outcome</th><th>Score / note</th><th>Type</th><th><span className="visually-hidden">Remove</span></th></tr></thead>
                    <tbody>{data.past_matches.map((match, index) => (
                      <tr key={`${match.date}-${index}`} className={match.is_incomplete ? "needs-review" : ""}>
                        <td><input type="date" value={match.date} onChange={(event) => updatePast(index, { date: event.target.value })} /></td>
                        <td><div className="team-fields"><select aria-label="Gender" value={match.gender_emoji} onChange={(event) => updatePast(index, { gender_emoji: event.target.value })}><option>👭</option><option>👬</option><option>👫</option></select><input aria-label="Level" className="level-input" value={match.level} onChange={(event) => updatePast(index, { level: event.target.value })} /></div></td>
                        <td><select value={match.is_home ? "home" : "away"} onChange={(event) => updatePast(index, { is_home: event.target.value === "home" })}><option value="home">🏠 Home</option><option value="away">🚗 Away</option></select></td>
                        <td><input value={match.opponent} onChange={(event) => updatePast(index, { opponent: event.target.value })} /></td>
                        <td><select value={outcomeState(match)} onChange={(event) => setOutcome(index, event.target.value)}><option value="win">Won</option><option value="loss">Lost</option><option value="incomplete">Needs review</option><option value="rained">Rained out</option></select></td>
                        <td><input value={match.is_incomplete ? match.footnote || "" : match.outcome_text || ""} placeholder={match.is_incomplete ? "What remains?" : "won 3-0"} onChange={(event) => updatePast(index, match.is_incomplete ? { footnote: event.target.value } : { outcome_text: event.target.value })} /></td>
                        <td><select value={match.match_type || "regular"} onChange={(event) => updatePast(index, { match_type: event.target.value as MatchType })}><option value="regular">Regular</option><option value="playoff">Playoff</option><option value="sectionals">Sectionals</option></select></td>
                        <td><button className="remove-action" type="button" aria-label={`Remove ${match.opponent}`} onClick={() => setData({ ...data, past_matches: data.past_matches.filter((_, row) => row !== index) })}>×</button></td>
                      </tr>
                    ))}</tbody>
                  </table>
                </div>
              </div>

              <div className="editor-block">
                <div className="editor-heading"><h3>Upcoming matches</h3><button type="button" className="text-action" onClick={() => setData({ ...data, future_matches: [...data.future_matches, blankFuture(settings.boundaryDate)] })}>+ Add fixture</button></div>
                <div className="table-scroll">
                  <table className="editor-table">
                    <thead><tr><th>Date</th><th>Time</th><th>Team</th><th>Site</th><th>Opponent</th><th>Location note</th><th>Type</th><th><span className="visually-hidden">Remove</span></th></tr></thead>
                    <tbody>{data.future_matches.map((match, index) => (
                      <tr key={`${match.date}-${index}`}>
                        <td><input type="date" value={match.date} onChange={(event) => updateFuture(index, { date: event.target.value })} /></td>
                        <td><input type="time" value={match.time || ""} onChange={(event) => updateFuture(index, { time: event.target.value })} /></td>
                        <td><div className="team-fields"><select aria-label="Gender" value={match.gender_emoji} onChange={(event) => updateFuture(index, { gender_emoji: event.target.value })}><option>👭</option><option>👬</option><option>👫</option></select><input aria-label="Level" className="level-input" value={match.level} onChange={(event) => updateFuture(index, { level: event.target.value })} /></div></td>
                        <td><select value={match.is_home ? "home" : "away"} onChange={(event) => updateFuture(index, { is_home: event.target.value === "home" })}><option value="home">🏠 Home</option><option value="away">🚗 Away</option></select></td>
                        <td><input value={match.opponent} onChange={(event) => updateFuture(index, { opponent: event.target.value })} /></td>
                        <td><input value={match.location_note || ""} placeholder="Optional" onChange={(event) => updateFuture(index, { location_note: event.target.value })} /></td>
                        <td><select value={match.match_type || "regular"} onChange={(event) => updateFuture(index, { match_type: event.target.value as MatchType })}><option value="regular">Regular</option><option value="playoff">Playoff</option><option value="sectionals">Sectionals</option></select></td>
                        <td><button className="remove-action" type="button" aria-label={`Remove ${match.opponent}`} onClick={() => setData({ ...data, future_matches: data.future_matches.filter((_, row) => row !== index) })}>×</button></td>
                      </tr>
                    ))}</tbody>
                  </table>
                </div>
              </div>
            </section>

            <section className="publish-section" aria-labelledby="publish-title">
              <div className="section-intro split-intro">
                <div><p className="step-label">03 · Publish</p><h2 id="publish-title">Ready for the clubhouse</h2><p>Download an image for email, preserve the editable data, or add upcoming matches to a calendar.</p></div>
                <div className="publish-actions not-print">
                  <button type="button" onClick={() => downloadJSON(data)}>Data JSON</button>
                  <button type="button" onClick={() => publicationRef.current && downloadHTML(data, publicationRef.current)}>HTML</button>
                  <button type="button" onClick={() => void saveJPEG("recent-board", "recent")}>Recent JPEG</button>
                  <button type="button" onClick={() => void saveJPEG("upcoming-board", "upcoming")}>Upcoming JPEG</button>
                  <button type="button" onClick={() => downloadICS(data)}>Calendar file</button>
                  <button type="button" onClick={() => window.print()}>Print / PDF</button>
                </div>
              </div>
              <div className="calendar-sync not-print">
                <label>Google Calendar name<input value={settings.calendarName} onChange={(event) => updateSetting("calendarName", event.target.value)} /></label>
                <button type="button" disabled={busy || !googleCalendarConfigured()} onClick={() => void syncCalendar()}>Sync Google Calendar</button>
                {!googleCalendarConfigured() && <small>Set <code>VITE_GOOGLE_CLIENT_ID</code> during the Pages build to enable direct sync.</small>}
              </div>
              <PublicationPreview data={data} ref={publicationRef} />
            </section>
          </>
        )}
      </main>
      <footer><span>Club Match Board</span><span>Schedules from USTA NorCal via Net Results</span></footer>
    </>
  );
}
