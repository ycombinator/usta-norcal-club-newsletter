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
  const organization = data.org_short_name.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_|_$/g, "");
  return `${organization}_usta_${date}_${suffix}.${extension}`;
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
