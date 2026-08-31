import { toJpeg } from "html-to-image";
import type { NewsletterData } from "./types";

function saveBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

function datedFilename(data: NewsletterData, suffix: string, extension: string): string {
  const date = new Date().toISOString().slice(0, 10).replaceAll("-", "_");
  return `${data.org_short_name.toLowerCase()}_usta_${date}_${suffix}.${extension}`;
}

export function downloadJSON(data: NewsletterData): void {
  saveBlob(new Blob([JSON.stringify(data, null, 2)], { type: "application/json" }), "data.json");
}

export async function downloadBoardJPEG(element: HTMLElement, data: NewsletterData, suffix: string): Promise<void> {
  const dataUrl = await toJpeg(element, {
    quality: 0.95,
    pixelRatio: 2,
    backgroundColor: "#f7fbfd",
  });
  const response = await fetch(dataUrl);
  saveBlob(await response.blob(), datedFilename(data, suffix, "jpg"));
}

export function downloadHTML(data: NewsletterData, publication: HTMLElement): void {
  const css = Array.from(document.styleSheets)
    .flatMap((sheet) => {
      try {
        return Array.from(sheet.cssRules).map((rule) => rule.cssText);
      } catch {
        return [];
      }
    })
    .join("\n");
  const html = `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>${data.org_short_name} USTA newsletter</title><style>${css}</style></head><body><main class="standalone-publication">${publication.outerHTML}</main></body></html>`;
  saveBlob(new Blob([html], { type: "text/html" }), datedFilename(data, "newsletter", "html"));
}

function escapeICS(value: string): string {
  return value.replaceAll("\\", "\\\\").replaceAll(";", "\\;").replaceAll(",", "\\,").replaceAll("\n", "\\n");
}

function calendarStamp(date: string, time?: string): { start: string; end: string; allDay: boolean } {
  if (!time) {
    const start = date.replaceAll("-", "");
    const endDate = new Date(`${date}T12:00:00`);
    endDate.setDate(endDate.getDate() + 1);
    return { start, end: endDate.toISOString().slice(0, 10).replaceAll("-", ""), allDay: true };
  }
  const startDate = new Date(`${date}T${time}:00`);
  const endDate = new Date(startDate.getTime() + 3 * 60 * 60 * 1000);
  const format = (value: Date) => {
    const local = new Date(value.getTime() - value.getTimezoneOffset() * 60_000);
    return local.toISOString().slice(0, 19).replaceAll("-", "").replaceAll(":", "");
  };
  return { start: format(startDate), end: format(endDate), allDay: false };
}

export function downloadICS(data: NewsletterData): void {
  const events = data.future_matches.map((match, index) => {
    const stamp = calendarStamp(match.date, match.time);
    const summary = `${match.is_home ? "🏠" : "🚗"} ${match.gender_emoji}${match.level}${match.superscript || ""} ${match.is_home ? "vs." : "@"} ${match.opponent}`;
    const dateLines = stamp.allDay
      ? `DTSTART;VALUE=DATE:${stamp.start}\r\nDTEND;VALUE=DATE:${stamp.end}`
      : `DTSTART;TZID=America/Los_Angeles:${stamp.start}\r\nDTEND;TZID=America/Los_Angeles:${stamp.end}`;
    return `BEGIN:VEVENT\r\nUID:usta-${match.date}-${index}@club-match-board\r\nDTSTAMP:${new Date().toISOString().replaceAll(/[-:]/g, "").replace(/\.\d{3}/, "")}\r\n${dateLines}\r\nSUMMARY:${escapeICS(summary)}\r\n${match.location_note ? `LOCATION:${escapeICS(match.location_note)}\r\n` : ""}END:VEVENT`;
  });
  const calendar = `BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Club Match Board//USTA NorCal//EN\r\nCALSCALE:GREGORIAN\r\n${events.join("\r\n")}\r\nEND:VCALENDAR\r\n`;
  saveBlob(new Blob([calendar], { type: "text/calendar" }), datedFilename(data, "upcoming", "ics"));
}
