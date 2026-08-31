import { describe, expect, it } from "vitest";
import { friendlyOpponent, organizationDisplayName, teamDisplay, teamSuperscript, toNewsletterData } from "./transform";
import type { OrganizationScheduleResponse, OrganizationScheduleTeam } from "./types";

describe("teamDisplay", () => {
  it("parses combo and adult team identities", () => {
    expect(teamDisplay({ id: "1", code: "ALMADEN SR CW5.5A-DT", league: "2026 NorCal Combo Doubles Daytime Womens League 5.5", schedule: [] }))
      .toEqual({ gender: "👭", level: "5.5", suffix: "A" });
    expect(teamDisplay({ id: "2", code: "ALMADEN SR 40MX7.0B", league: "2026 Mixed 40 & Over 7.0", schedule: [] }))
      .toEqual({ gender: "👫", level: "7.0", suffix: "B" });
  });
});

describe("teamSuperscript", () => {
  const teamA: OrganizationScheduleTeam = { id: "1", code: "ALMADEN SR 40MX7.0A", league: "2026 Mixed 40 & Over 7.0", schedule: [] };
  const teamB: OrganizationScheduleTeam = { id: "2", code: "ALMADEN SR 40MX7.0B", league: "2026 Mixed 40 & Over 7.0", schedule: [] };

  it("omits the suffix when there is only one organization team in the league", () => {
    expect(teamSuperscript(teamA, [teamA])).toBe("");
  });

  it("uses suffixes to distinguish multiple organization teams in the same league", () => {
    expect(teamSuperscript(teamA, [teamA, teamB])).toBe("A");
    expect(teamSuperscript(teamB, [teamA, teamB])).toBe("B");
  });

  it("does not count a team in a different league as a sibling", () => {
    const other = { ...teamB, code: "ALMADEN SR 40MX8.0B", league: "2026 Mixed 40 & Over" };
    expect(teamSuperscript(teamA, [teamA, other])).toBe("");
  });

  it("uses the full USTA team name when the API provides it", () => {
    const namedA = { ...teamA, name: "2026 Adult 40 & Over Mixed 7.0" };
    const unrelatedA = { ...teamB, name: "2026 Adult 55 & Over Mixed 7.0" };
    expect(teamSuperscript(namedA, [namedA, unrelatedA])).toBe("");
    expect(teamSuperscript(namedA, [namedA, { ...teamB, name: namedA.name }])).toBe("A");
  });
});

describe("friendlyOpponent", () => {
  it("uses known club names and strips team codes", () => {
    expect(friendlyOpponent("BAY CLUB COURTSIDE 40MX7.0A")).toBe("Courtside");
    expect(friendlyOpponent("ROUND HILL CC 40MX7.0A")).toBe("Round Hill Cc");
  });

  it("resolves abbreviated team-code organizations through org_names", () => {
    expect(friendlyOpponent("Wallenberg Pk Cm7.5a 1")).toBe("Wallenberg");
    expect(friendlyOpponent("Imperial Tc 55mx6.0a [royals]")).toBe("Imperial TC");
    expect(friendlyOpponent("Bramhall Pk Cm6.5b [bbq Bros]")).toBe("Bramhall");
  });
});

describe("organizationDisplayName", () => {
  it("uses the shared organization-name mapping", () => {
    expect(organizationDisplayName("ALMADEN SWIM AND RACQUET CLUB")).toBe("ASRC");
  });

  it("normalizes an unmapped all-caps organization heading", () => {
    expect(organizationDisplayName("EXAMPLE TENNIS AND SWIM CLUB")).toBe("Example Tennis and Swim Club");
  });
});

describe("toNewsletterData", () => {
  it("splits, normalizes, and deduplicates organization matches", () => {
    const response: OrganizationScheduleResponse = {
      organization: { id: "225", name: "ALMADEN SWIM AND RACQUET CLUB" },
      from: "2026-08-24",
      to: "2026-09-07",
      generated_at: "2026-08-31T00:00:00Z",
      teams: [
        {
          id: "10", code: "ALMADEN SR 40MX7.0A", league: "2026 Mixed 40 & Over 7.0",
          schedule: [
            { round: "4", date: "2026-08-30", time: "7:00 PM", opponent: "BAY CLUB COURTSIDE 40MX7.0A", opponent_id: "20", home_away: "Home", result: "Won 3-0", played: true },
            { round: "5", date: "2026-08-31", time: "6:30 PM", opponent: "SILVER CREEK 40MX7.0A", opponent_id: "21", home_away: "Away", result: "", played: false, notes: "Bring refillable water bottles." },
            { round: "PlayOff", date: "2026-09-03", time: "6:30 PM", opponent: "SUNNYVALE MTC 40MX7.0A", opponent_id: "30", home_away: "Away", result: "", played: false },
          ],
        },
      ],
    };
    const data = toNewsletterData(response, "2026-09-01");
    expect(data.org_short_name).toBe("ASRC");
    expect(data.past_matches[0]).toMatchObject({ is_win: true, outcome_text: "won 3-0", opponent: "Courtside" });
    expect(data.past_matches[1]).toMatchObject({ is_incomplete: true, footnote: "result needs review", opponent: "Silver Creek" });
    expect(data.future_matches[0]).toMatchObject({ time: "18:30", match_type: "playoff", opponent: "Sunnyvale TC" });
  });
});
