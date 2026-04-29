import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const layerOrder = ["shared", "entities", "features", "widgets", "pages", "app"] as const;

function layerRank(layer: string) {
  return layerOrder.indexOf(layer as (typeof layerOrder)[number]);
}

describe("renderer architecture boundaries", () => {
  it("keeps FSD dependency direction and electron boundary", () => {
    const root = path.resolve(import.meta.dirname, "..", "..");
    const files = walkFiles(root).filter(
      (file) => (file.endsWith(".ts") || file.endsWith(".tsx")) && !file.endsWith(".test.ts") && !file.endsWith(".test.tsx"),
    );

    for (const file of files) {
      const relative = path.relative(root, file).replace(/\\/g, "/");
      const content = fs.readFileSync(file, "utf8");
      const fileLayer = relative.split("/")[0] ?? "";
      const imports = [...content.matchAll(/from\s+["'](@\/[^"']+)["']/g)].map((match) => match[1]);

      for (const currentImport of imports) {
        const [, targetLayer] = currentImport.split("/");
        expect(targetLayer).not.toBe("electron");
        const currentRank = layerRank(fileLayer);
        const targetRank = layerRank(targetLayer ?? "");
        if (currentRank === -1 || targetRank === -1) {
          continue;
        }
        expect(currentRank).toBeGreaterThanOrEqual(targetRank);
      }

      expect(content.includes('@electron/main')).toBe(false);
      expect(content.includes('@electron/preload')).toBe(false);
    }
  });
});

function walkFiles(root: string): string[] {
  const result: string[] = [];
  for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
    const fullPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      result.push(...walkFiles(fullPath));
      continue;
    }
    result.push(fullPath);
  }
  return result;
}
