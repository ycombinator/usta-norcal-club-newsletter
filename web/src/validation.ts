import type { FutureMatchRecord, NewsletterData, PastMatchRecord } from "./types";

export type PastMatchField = "date" | "gender" | "level" | "opponent" | "outcome" | "result";
export type FutureMatchField = "date" | "time" | "gender" | "level" | "opponent";

function isBlank(value?: string): boolean {
  return !value?.trim();
}

export function pastMatchIssues(match: PastMatchRecord): PastMatchField[] {
  const issues: PastMatchField[] = [];
  if (isBlank(match.date)) issues.push("date");
  if (isBlank(match.gender_emoji)) issues.push("gender");
  if (isBlank(match.level)) issues.push("level");
  if (isBlank(match.opponent)) issues.push("opponent");
  const isReviewedIncomplete = match.review_status === "to_be_completed"
    || (!match.review_status && match.footnote?.trim().toLowerCase() === "to be completed");
  if (match.is_incomplete && !isReviewedIncomplete) issues.push("outcome", "result");
  else if (!match.is_incomplete && !match.is_rained_out && isBlank(match.outcome_text)) issues.push("result");
  return issues;
}

export function futureMatchIssues(match: FutureMatchRecord): FutureMatchField[] {
  const issues: FutureMatchField[] = [];
  if (isBlank(match.date)) issues.push("date");
  if (isBlank(match.time)) issues.push("time");
  if (isBlank(match.gender_emoji)) issues.push("gender");
  if (isBlank(match.level)) issues.push("level");
  if (isBlank(match.opponent)) issues.push("opponent");
  return issues;
}

export function reviewMatchCount(data: NewsletterData): number {
  return data.past_matches.filter((match) => pastMatchIssues(match).length > 0).length
    + data.future_matches.filter((match) => futureMatchIssues(match).length > 0).length;
}
