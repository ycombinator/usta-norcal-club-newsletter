import type {
  FutureMatchRecord,
  MatchType,
  NewsletterData,
  OrganizationScheduleResponse,
  OrganizationScheduleTeam,
  PastMatchRecord,
  ScheduleEntry,
} from "./types";

const opponentPrefixes: Array<[RegExp, string]> = [
  [/^BAY CLUB COURTSIDE\b|^BCC\b/i, "Courtside"],
  [/^SUNNYVALE\b/i, "Sunnyvale TC"],
  [/^ALMADEN VALLEY\b|^AVAC\b/i, "AVAC"],
  [/^LOS GATOS\b/i, "Los Gatos"],
  [/^SILVER CREEK\b/i, "Silver Creek"],
  [/^VILLAGES\b/i, "Villages"],
  [/^MORGAN HILL\b/i, "Morgan Hill"],
];

export function shortOrganizationName(name: string): string {
  const known: Record<string, string> = {
    "ALMADEN SWIM AND RACQUET CLUB": "ASRC",
    "ALMADEN VALLEY ATHLETIC CLUB": "AVAC",
    "BAY CLUB COURTSIDE": "Courtside",
  };
  const normalized = name.trim().toUpperCase();
  if (known[normalized]) return known[normalized];
  return normalized
    .split(/\s+/)
    .filter((part) => part !== "AND" && part !== "&")
    .map((part) => part[0])
    .join("");
}

export function teamDisplay(team: OrganizationScheduleTeam): { gender: string; level: string; suffix: string } {
  const source = `${team.league} ${team.code}`;
  const lower = source.toLowerCase();
  let gender = "👫";
  if (/women|womens|\bcw\d/.test(lower)) gender = "👭";
  else if (/men|mens|\bcm\d/.test(lower)) gender = "👬";

  const leagueLevel = team.league.match(/(\d+(?:\.\d+)?\+?)\s*$/);
  const codeLevel = team.code.match(/(\d+(?:\.\d+)?\+?)[A-Z](?:-DT)?(?:\s|\(|$)/i);
  const level = leagueLevel?.[1] || codeLevel?.[1] || "";
  const suffixPattern = level
    ? new RegExp(`${level.replace("+", "\\+")}([A-Z])(?:-DT)?(?:\\s|\\(|$)`, "i")
    : null;
  const suffix = suffixPattern?.exec(team.code)?.[1]?.toUpperCase() || "";
  return { gender, level, suffix };
}

export function friendlyOpponent(raw: string): string {
  const value = raw.trim();
  for (const [pattern, replacement] of opponentPrefixes) {
    if (pattern.test(value)) return replacement;
  }
  const stripped = value
    .replace(/\s+(?:18|40|55)(?:A?[MW]|MX)\d+(?:\.\d+)?\+?[A-Z](?:-DT)?(?:\s*\([^)]*\))?$/i, "")
    .replace(/\s+C[MW]\d+(?:\.\d+)?[A-Z](?:-DT)?(?:\s*\([^)]*\))?$/i, "")
    .replace(/\s*\([^)]*\)\s*$/, "")
    .trim();
  return stripped
    .toLowerCase()
    .replace(/(^|[\s/])([a-z])/g, (_, prefix: string, letter: string) => prefix + letter.toUpperCase()) || value;
}

export function matchType(round: string): MatchType {
  const value = round.toLowerCase();
  if (value.includes("sectional")) return "sectionals";
  if (value.includes("playoff")) return "playoff";
  return "regular";
}

function normalizeTime(value: string): string {
  const match = value.match(/(\d{1,2}):(\d{2})\s*([AP]M)/i);
  if (!match) return "";
  let hour = Number(match[1]);
  if (match[3].toUpperCase() === "PM" && hour !== 12) hour += 12;
  if (match[3].toUpperCase() === "AM" && hour === 12) hour = 0;
  return `${String(hour).padStart(2, "0")}:${match[2]}`;
}

function outcome(entry: ScheduleEntry): Pick<PastMatchRecord, "is_win" | "is_incomplete" | "outcome_text" | "footnote"> {
  const result = entry.result.trim();
  if (/^won\b/i.test(result)) return { is_win: true, outcome_text: result.toLowerCase() };
  if (/^lost\b/i.test(result)) return { outcome_text: result.toLowerCase() };
  return {
    is_incomplete: true,
    outcome_text: "",
    footnote: entry.verification_pending ? "awaiting score verification" : entry.notes?.trim() || "result needs review",
  };
}

function recordKey(team: OrganizationScheduleTeam, entry: ScheduleEntry): string {
  const pair = [team.id, entry.opponent_id].sort().join(":");
  return `${entry.date}|${entry.round}|${pair}`;
}

export function toNewsletterData(response: OrganizationScheduleResponse, boundaryDate: string): NewsletterData {
  const data: NewsletterData = {
    org_short_name: shortOrganizationName(response.organization.name),
    past_matches: [],
    future_matches: [],
  };
  const seen = new Set<string>();

  for (const team of response.teams) {
    const display = teamDisplay(team);
    for (const entry of team.schedule) {
      const key = recordKey(team, entry);
      if (seen.has(key)) continue;
      seen.add(key);
      const base = {
        date: entry.date,
        gender_emoji: display.gender,
        level: display.level,
        superscript: display.suffix || undefined,
        is_home: entry.home_away === "Home",
        opponent: friendlyOpponent(entry.opponent),
        match_type: matchType(entry.round),
      };
      if (entry.date < boundaryDate) {
        data.past_matches.push({ ...base, ...outcome(entry) });
      } else {
        const future: FutureMatchRecord = {
          ...base,
          time: normalizeTime(entry.time) || undefined,
          location_note: team.extra && entry.home_away === "Home" ? team.organization_name : undefined,
        };
        data.future_matches.push(future);
      }
    }
  }
  data.past_matches.sort((a, b) => a.date.localeCompare(b.date));
  data.future_matches.sort((a, b) => `${a.date} ${a.time || ""}`.localeCompare(`${b.date} ${b.time || ""}`));
  return data;
}
