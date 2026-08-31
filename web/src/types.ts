export type MatchType = "regular" | "playoff" | "sectionals";

export interface ScheduleEntry {
  round: string;
  date: string;
  time: string;
  opponent: string;
  opponent_id: string;
  home_away: "Home" | "Away" | "";
  result: string;
  played: boolean;
  verification_pending?: boolean;
  notes?: string;
}

export interface OrganizationScheduleTeam {
  id: string;
  code: string;
  league: string;
  organization_id?: string;
  organization_name?: string;
  extra?: boolean;
  schedule: ScheduleEntry[];
}

export interface OrganizationScheduleResponse {
  organization: { id: string; name: string };
  from: string;
  to: string;
  teams: OrganizationScheduleTeam[];
  failures?: Array<{ team_id: string; error: string }>;
  generated_at: string;
}

export interface PastMatchRecord {
  date: string;
  gender_emoji: string;
  level: string;
  superscript?: string;
  is_home: boolean;
  opponent: string;
  is_win?: boolean;
  is_rained_out?: boolean;
  is_incomplete?: boolean;
  outcome_text?: string;
  footnote?: string;
  match_type?: MatchType;
}

export interface FutureMatchRecord {
  date: string;
  time?: string;
  gender_emoji: string;
  level: string;
  superscript?: string;
  is_home: boolean;
  opponent: string;
  location_note?: string;
  match_type?: MatchType;
}

export interface NewsletterData {
  org_short_name: string;
  past_matches: PastMatchRecord[];
  future_matches: FutureMatchRecord[];
}

export interface Settings {
  organizationId: string;
  extraTeamIds: string;
  boundaryDate: string;
  pastDays: number;
  futureDays: number;
  apiBaseUrl: string;
  calendarName: string;
}
