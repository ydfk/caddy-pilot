import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { RuntimeCard } from "./runtime-card";

describe("RuntimeCard", () => {
  it("集中显示版本检查和两种安装方式", () => {
    render(
      <RuntimeCard
        status={{ online: true }}
        version={{
          current_version: "2.10.0",
          latest_version: "2.11.4",
          update_available: true,
          version_check_url: "https://example.com/manifest.json",
          download_url: "https://example.com/caddy.zip",
          release_url: "https://example.com/release",
          update_strategy: "managed",
        }}
        settings={{
          version_check_url: "https://example.com/manifest.json",
          download_url: "https://example.com/caddy.zip",
          checksum_url: "https://example.com/sha512sums.txt",
        }}
        checkingVersion={false}
        updatingVersion={false}
        updateTask={null}
        onCheckVersion={vi.fn()}
        onUpdateVersion={vi.fn()}
        onUploadVersion={vi.fn()}
        onSaveSettings={vi.fn()}
      />
    );

    expect(screen.getByRole("button", { name: "检查更新" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "在线更新到 2.11.4" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "上传 Caddy 安装包" })).toBeInTheDocument();
    expect(screen.queryByText("刷新状态")).not.toBeInTheDocument();
    expect(screen.queryByText("当前 JSON")).not.toBeInTheDocument();
  });
});
