import { forwardRef } from "react";
import type { FutureMatchRecord, NewsletterData, PastMatchRecord } from "./types";

function parseDate(date: string): Date {
  return new Date(`${date}T12:00:00`);
}

function dayLabel(date: string): string {
  return new Intl.DateTimeFormat("en-US", { weekday: "short", month: "numeric", day: "numeric" }).format(parseDate(date));
}

function timeLabel(time?: string): string {
  if (!time) return "TBD";
  const [hourValue, minute] = time.split(":").map(Number);
  const period = hourValue >= 12 ? "pm" : "am";
  const hour = hourValue % 12 || 12;
  return minute ? `${hour}:${String(minute).padStart(2, "0")}${period}` : `${hour}${period}`;
}

function tag(match: { match_type?: string }): string {
  if (match.match_type === "sectionals") return "Sectionals";
  if (match.match_type === "playoff") return "playoff";
  return "";
}

function matchTeam(match: PastMatchRecord | FutureMatchRecord) {
  return <span className="team-mark">{match.gender_emoji}{match.level}{match.superscript && <sup>{match.superscript}</sup>}</span>;
}

function mondayFor(date: string): string {
  const value = parseDate(date);
  const day = value.getDay();
  value.setDate(value.getDate() - ((day + 6) % 7));
  return value.toISOString().slice(0, 10);
}

function addDays(date: string, days: number): string {
  const value = parseDate(date);
  value.setDate(value.getDate() + days);
  return value.toISOString().slice(0, 10);
}

function RecentBoard({ data }: { data: NewsletterData }) {
  let previous = "";
  return (
    <article className="publication-card recent-board" id="recent-board">
      <header className="publication-heading">
        <p>🏆 🎾 {data.org_short_name} plays USTA league 🎾 🏆</p>
        <h2>Recent results</h2>
      </header>
      {data.past_matches.length ? (
        <table>
          <tbody>
            {data.past_matches.map((match, index) => {
              const showDate = previous !== match.date;
              previous = match.date;
              const weekend = [0, 6].includes(parseDate(match.date).getDay());
              return (
                <tr key={`${match.date}-${index}`}>
                  <th className={weekend ? "weekend" : ""}>{showDate ? dayLabel(match.date) : ""}</th>
                  <td>{matchTeam(match)}</td>
                  <td className={`result ${match.is_win ? "win" : "loss"}`}>
                    {match.is_rained_out ? "🌧️" : match.is_incomplete ? `${match.outcome_text || "pending"}*` : match.outcome_text}
                  </td>
                  <td>{match.is_home ? "🏠" : "🚗"}</td>
                  <td className="opponent">{match.opponent}</td>
                  <td>{tag(match) && <span className="match-tag">{tag(match)}</span>}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      ) : <p className="publication-empty">No recent results in this date window.</p>}
      {[...new Set(data.past_matches.map((match) => match.footnote).filter(Boolean))].map((note) => (
        <p className="footnote" key={note}>* {note}</p>
      ))}
    </article>
  );
}

function WeekBoard({ data, monday }: { data: NewsletterData; monday: string }) {
  const days = Array.from({ length: 7 }, (_, index) => addDays(monday, index));
  return (
    <article className="publication-card upcoming-board">
      <header className="publication-heading">
        <p>🏆 🎾 {data.org_short_name} plays USTA league 🎾 🏆</p>
        <h2>Upcoming matches</h2>
      </header>
      <div className="week-grid">
        {days.map((date) => {
          const matches = data.future_matches.filter((match) => match.date === date);
          const weekend = [0, 6].includes(parseDate(date).getDay());
          return (
            <section className="day-column" key={date}>
              <h3 className={weekend ? "weekend" : ""}>{dayLabel(date)}</h3>
              <div className="day-matches">
                {matches.map((match, index) => (
                  <div className="calendar-match" key={`${date}-${index}`}>
                    {tag(match) && <span className="match-tag">{tag(match)}</span>}
                    <p>{match.is_home ? "🏠" : "🚗"} <strong>{timeLabel(match.time)}</strong></p>
                    <p>{matchTeam(match)}</p>
                    <p className="opponent">{match.opponent}</p>
                  </div>
                ))}
              </div>
            </section>
          );
        })}
      </div>
    </article>
  );
}

export const PublicationPreview = forwardRef<HTMLDivElement, { data: NewsletterData }>(({ data }, ref) => {
  const weeks = [...new Set(data.future_matches.map((match) => mondayFor(match.date)))];
  return (
    <div className="publication" id="publication" ref={ref}>
      <RecentBoard data={data} />
      <div className="upcoming-collection" id="upcoming-board">
        {weeks.map((monday) => <WeekBoard data={data} monday={monday} key={monday} />)}
        {!weeks.length && (
          <article className="publication-card upcoming-board">
            <header className="publication-heading"><p>🏆 🎾 {data.org_short_name} plays USTA league 🎾 🏆</p><h2>Upcoming matches</h2></header>
            <p className="publication-empty">No upcoming matches in this date window.</p>
          </article>
        )}
      </div>
    </div>
  );
});
