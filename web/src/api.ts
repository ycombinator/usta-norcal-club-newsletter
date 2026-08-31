import type { OrganizationScheduleResponse, Settings } from "./types";

const apiBaseUrl = import.meta.env.VITE_NET_RESULTS_API_URL || (import.meta.env.DEV ? "/api" : "https://net-results-api.onrender.com");

function addDays(date: string, days: number): string {
  const value = new Date(`${date}T12:00:00`);
  value.setDate(value.getDate() + days);
  return value.toISOString().slice(0, 10);
}

export async function fetchOrganizationSchedule(settings: Settings, force = false): Promise<OrganizationScheduleResponse> {
  const base = apiBaseUrl.replace(/\/$/, "");
  const params = new URLSearchParams({
    from: addDays(settings.boundaryDate, -settings.pastDays),
    to: addDays(settings.boundaryDate, Math.max(0, settings.futureDays - 1)),
  });
  if (settings.extraTeamIds.trim()) {
    params.set("extra_team_ids", settings.extraTeamIds.replace(/\s+/g, ""));
  }
  const response = await fetch(`${base}/organizations/${encodeURIComponent(settings.organizationId)}/schedule?${params}`, {
    headers: force ? { "Cache-Control": "no-cache" } : undefined,
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null) as { error?: string } | null;
    throw new Error(body?.error || `Schedule request failed (${response.status})`);
  }
  return response.json() as Promise<OrganizationScheduleResponse>;
}
