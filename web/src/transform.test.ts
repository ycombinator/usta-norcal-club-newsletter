import { describe, expect, it } from "vitest";
import { friendlyOpponent, teamDisplay, toNewsletterData } from "./transform";
import type { OrganizationScheduleResponse } from "./types";

describe("teamDisplay", () => {
  it("parses combo and adult team identities", () => {
    expect(teamDisplay({ id: "1", code: "ALMADEN SR CW5.5A-DT", league: "2026 NorCal Combo Doubles Daytime Womens League 5.5", schedule: [] }))
      .toEqual({ gender: "👭", level: "5.5", suffix: "A" });
    expect(teamDisplay({ id: "2", code: "ALMADEN SR 40MX7.0B", league: "2026 Mixed 40 & Over 7.0", schedule: [] }))
      .toEqual({ gender: "👫", level: "7.0", suffix: "B" });
  });
});

describe("friendlyOpponent", () => {
  it("uses known club names and strips team codes", () => {
    expect(friendlyOpponent("BAY CLUB COURTSIDE 40MX7.0A")).toBe("Courtside");
    expect(friendlyOpponent("ROUND HILL CC 40MX7.0A")).toBe("Round Hill Cc");
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
            { round: "PlayOff", date: "2026-09-03", time: "6:30 PM", opponent: "SUNNYVALE MTC 40MX7.0A", opponent_id: "30", home_away: "Away", result: "", played: false },
          ],
        },
      ],
    };
    const data = toNewsletterData(response, "2026-09-01");
    expect(data.org_short_name).toBe("ASRC");
    expect(data.past_matches[0]).toMatchObject({ is_win: true, outcome_text: "won 3-0", opponent: "Courtside" });
    expect(data.future_matches[0]).toMatchObject({ time: "18:30", match_type: "playoff", opponent: "Sunnyvale TC" });
  });
});
