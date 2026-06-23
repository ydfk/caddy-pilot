import { describe, expect, it } from "vitest";

import { mergeLogEntries } from "./logs";

describe("mergeLogEntries", () => {
  it("把最新一批日志插入顶部", () => {
    const current = [{ id: "1", message: "旧日志" }];
    const incoming = [
      { id: "3", message: "最新日志" },
      { id: "2", message: "较新日志" },
    ];
    expect(mergeLogEntries(current, incoming, false).map((entry) => entry.id)).toEqual([
      "3",
      "2",
      "1",
    ]);
  });
});
