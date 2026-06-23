import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ConfigManagement } from "./config-management";

vi.mock("@/api/config-versions", () => ({
  listConfigVersions: vi.fn().mockResolvedValue([]),
  rollbackConfigVersion: vi.fn(),
}));

vi.mock("@/api/caddy", () => ({
  getCurrentCaddyConfig: vi.fn(),
  previewCaddyfile: vi.fn(),
  publishCaddyConfig: vi.fn(),
  validateCaddyConfig: vi.fn(),
}));

describe("ConfigManagement", () => {
  it("只保留一个配置管理卡片标题并集中预览操作", () => {
    render(
      <ConfigManagement
        status={{
          dirty: false,
          state: "in_sync",
          runtime_in_sync: true,
          persistent_config_in_sync: true,
          active_version: 3,
        }}
        refreshKey={0}
        onChanged={vi.fn()}
      />
    );

    expect(screen.getAllByText("配置管理")).toHaveLength(1);
    expect(screen.queryByText("配置发布")).not.toBeInTheDocument();
    expect(screen.queryByText("配置版本")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "当前 JSON" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "预览 Caddyfile" })).toBeInTheDocument();
  });
});
