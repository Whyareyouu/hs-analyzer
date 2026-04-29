import { describe, expect, it } from "vitest";
import fs from "node:fs";
import path from "node:path";

describe("design tokens", () => {
  it("defines critical overlay and dashboard tokens", () => {
    const tokensPath = path.resolve(import.meta.dirname, "tokens.css");
    const content = fs.readFileSync(tokensPath, "utf8");

    for (const token of [
      "--overlay-card-row-width",
      "--overlay-card-row-height",
      "--overlay-panel-width",
      "--dashboard-sidebar-width",
      "--dashboard-winrate-ring-size",
    ]) {
      expect(content).toContain(token);
    }
  });
});
