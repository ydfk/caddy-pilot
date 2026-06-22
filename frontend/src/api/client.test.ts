import { beforeEach, describe, expect, it, vi } from "vitest";

import { apiRequest } from "./client";

describe("apiRequest", () => {
  beforeEach(() => vi.restoreAllMocks());

  it("显示 Huma 字段校验原因", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              title: "Unprocessable Entity",
              errors: [{ location: "body.name", message: "不能为空" }],
            }),
            { status: 422, headers: { "content-type": "application/problem+json" } }
          )
      )
    );

    await expect(apiRequest("/api/test")).rejects.toThrow("body.name: 不能为空");
  });

  it("上传 FormData 时不覆盖浏览器生成的 Content-Type", async () => {
    const fetchMock = vi.fn(async (_url: string, options?: RequestInit) => {
      expect(new Headers(options?.headers).has("Content-Type")).toBe(false);
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" },
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    await apiRequest("/api/upload", { method: "POST", body: new FormData() });
    expect(fetchMock).toHaveBeenCalledOnce();
  });
});
