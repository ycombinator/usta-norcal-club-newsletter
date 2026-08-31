import { describe, expect, it } from "vitest";
import { futureMatchIssues, pastMatchIssues, reviewMatchCount } from "./validation";
import type { FutureMatchRecord, NewsletterData, PastMatchRecord } from "./types";

const completePast: PastMatchRecord = {
  date: "2026-08-29", gender_emoji: "👫", level: "7.0", is_home: true,
  opponent: "Courtside", is_win: true, outcome_text: "won 2-1",
};

const completeFuture: FutureMatchRecord = {
  date: "2026-09-01", time: "18:30", gender_emoji: "👫", level: "7.0",
  is_home: false, opponent: "Los Gatos",
};

describe("review validation", () => {
  it("accepts completed results and rainouts", () => {
    expect(pastMatchIssues(completePast)).toEqual([]);
    expect(pastMatchIssues({ ...completePast, is_win: false, is_rained_out: true, outcome_text: "" })).toEqual([]);
  });

  it("identifies the exact recent-result fields requiring review", () => {
    expect(pastMatchIssues({ ...completePast, is_incomplete: true })).toEqual(["outcome", "result"]);
    expect(pastMatchIssues({ ...completePast, outcome_text: "" })).toEqual(["result"]);
  });

  it("accepts a reviewed result that is marked to be completed", () => {
    expect(pastMatchIssues({ ...completePast, is_incomplete: true, review_status: "to_be_completed", outcome_text: "", footnote: "to be completed" })).toEqual([]);
  });

  it("identifies missing upcoming-match fields", () => {
    expect(futureMatchIssues({ ...completeFuture, time: "" })).toEqual(["time"]);
  });

  it("counts matches with issues once, regardless of issue count", () => {
    const data: NewsletterData = {
      org_short_name: "Club",
      past_matches: [completePast, { ...completePast, level: "", opponent: "" }],
      future_matches: [{ ...completeFuture, time: "" }],
    };
    expect(reviewMatchCount(data)).toBe(2);
  });
});
