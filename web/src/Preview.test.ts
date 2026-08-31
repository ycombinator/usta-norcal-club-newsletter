import { describe, expect, it } from "vitest";
import { isLateMatch } from "./Preview";

describe("isLateMatch", () => {
  it("bottom-aligns matches beginning at 4 PM", () => {
    expect(isLateMatch("15:59")).toBe(false);
    expect(isLateMatch("16:00")).toBe(true);
    expect(isLateMatch("19:30")).toBe(true);
  });

  it("keeps matches without a known time in the daytime group", () => {
    expect(isLateMatch()).toBe(false);
    expect(isLateMatch("")).toBe(false);
  });
});
