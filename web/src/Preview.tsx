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

function WeekBoard({ data, days }: { data: NewsletterData; days: string[] }) {
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

interface PublicationPreviewProps {
  data: NewsletterData;
  futureStartDate: string;
  futureDays: number;
}

export const PublicationPreview = forwardRef<HTMLDivElement, PublicationPreviewProps>(({ data, futureStartDate, futureDays }, ref) => {
  const dayCount = Math.max(1, Math.floor(futureDays) || 1);
  const visibleDays = Array.from({ length: dayCount }, (_, index) => addDays(futureStartDate, index));
  const dateGroups: string[][] = [];
  for (let index = 0; index < visibleDays.length; index += 7) dateGroups.push(visibleDays.slice(index, index + 7));
  return (
    <div className="publication" id="publication" ref={ref}>
      <RecentBoard data={data} />
      <div className="upcoming-collection" id="upcoming-board">
        {dateGroups.map((days) => <WeekBoard data={data} days={days} key={days[0]} />)}
      </div>
    </div>
  );
});
