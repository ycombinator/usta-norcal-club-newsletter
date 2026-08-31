import type {
  FutureMatchRecord,
  MatchType,
  NewsletterData,
  OrganizationScheduleResponse,
  OrganizationScheduleTeam,
  PastMatchRecord,
  ScheduleEntry,
} from "./types";
import orgNamesSource from "../../org_names.yaml?raw";

const organizationNames = new Map(
  orgNamesSource
    .split(/\r?\n/)
    .map((line) => line.match(/^([^#][^:]*):\s*(.+?)\s*$/))
    .filter((match): match is RegExpMatchArray => Boolean(match))
    .map((match) => [match[1].trim().toUpperCase(), match[2].trim()]),
);

const opponentPrefixes: Array<[RegExp, string]> = [
  [/^BAY CLUB COURTSIDE\b|^BCC\b/i, "Courtside"],
  [/^SUNNYVALE\b/i, "Sunnyvale TC"],
  [/^ALMADEN VALLEY\b|^AVAC\b/i, "AVAC"],
  [/^LOS GATOS\b/i, "Los Gatos"],
  [/^SILVER CREEK\b/i, "Silver Creek"],
  [/^VILLAGES\b/i, "Villages"],
  [/^MORGAN HILL\b/i, "Morgan Hill"],
];

export function organizationDisplayName(name: string): string {
  const normalized = name.trim();
  const mapped = organizationNames.get(normalized.toUpperCase());
  if (mapped) return mapped;
  if (normalized !== normalized.toUpperCase()) return normalized;
  const lowercaseWords = new Set(["and", "at", "of", "the"]);
  return normalized
    .toLowerCase()
    .split(/\s+/)
    .map((part, index) => index > 0 && lowercaseWords.has(part)
      ? part
      : part.replace(/(^|[-'])([a-z])/g, (_, prefix: string, letter: string) => prefix + letter.toUpperCase()))
    .join(" ");
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

function teamIdentity(team: OrganizationScheduleTeam): string {
  if (team.name?.trim()) return team.name.trim().toLowerCase();
  const code = team.code.replace(/\s*[[(].*$/, "").trim();
  const match = code.match(/((?:\d+(?:A?[MW]|MX)|C[MW])\d+(?:\.\d+)?\+?)[A-Z](-DT)?$/i);
  return match ? `${match[1].toLowerCase()}${match[2]?.toLowerCase() || ""}` : team.league.trim().toLowerCase();
}

// Match the CLI: a suffix distinguishes teams only when the organization has
// more than one team with the same full USTA team name. A lone "A" team needs
// no superscript. The code identity is a compatibility fallback for APIs that
// predate the team name field.
export function teamSuperscript(team: OrganizationScheduleTeam, teams: OrganizationScheduleTeam[]): string {
  const suffix = teamDisplay(team).suffix;
  if (!suffix) return "";
  const identity = teamIdentity(team);
  const hasSibling = teams.some((other) => other.id !== team.id && teamIdentity(other) === identity);
  return hasSibling ? suffix : "";
}

export function friendlyOpponent(raw: string): string {
  const value = raw.trim();
  for (const [pattern, replacement] of opponentPrefixes) {
    if (pattern.test(value)) return replacement;
  }
  const stripped = value
    .replace(/\s*[[(][^\])]*[\])]\s*$/i, "")
    .replace(/\s+\d+\s*$/, "")
    .replace(/\s+(?:(?:\d+)(?:A?[MW]|MX)|C[MW])\d+(?:\.\d+)?\+?[A-Z](?:-DT)?$/i, "")
    .trim();
  const mapped = organizationNames.get(stripped.toUpperCase());
  if (mapped) return mapped;
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

function outcome(entry: ScheduleEntry): Pick<PastMatchRecord, "is_win" | "is_incomplete" | "review_status" | "outcome_text" | "footnote"> {
  const result = entry.result.trim();
  if (/^won\b/i.test(result)) return { is_win: true, outcome_text: result.toLowerCase() };
  if (/^lost\b/i.test(result)) return { outcome_text: result.toLowerCase() };
  return {
    is_incomplete: true,
    review_status: "needs_review",
    outcome_text: "",
    footnote: entry.verification_pending ? "awaiting score verification" : "result needs review",
  };
}

function recordKey(team: OrganizationScheduleTeam, entry: ScheduleEntry): string {
  const pair = [team.id, entry.opponent_id].sort().join(":");
  return `${entry.date}|${entry.round}|${pair}`;
}

export function toNewsletterData(response: OrganizationScheduleResponse, boundaryDate: string): NewsletterData {
  const data: NewsletterData = {
    org_short_name: organizationDisplayName(response.organization.name),
    past_matches: [],
    future_matches: [],
  };
  const seen = new Set<string>();

  for (const team of response.teams) {
    const display = teamDisplay(team);
    const superscript = teamSuperscript(team, response.teams);
    for (const entry of team.schedule) {
      const key = recordKey(team, entry);
      if (seen.has(key)) continue;
      seen.add(key);
      const base = {
        date: entry.date,
        gender_emoji: display.gender,
        level: display.level,
        superscript: superscript || undefined,
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
