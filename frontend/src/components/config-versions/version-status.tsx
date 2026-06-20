import type { ConfigVersionSummary } from "@/api/config-versions";
import { Badge } from "@/components/ui/badge";

const statusMeta = {
  draft: { label: "草稿", variant: "outline" },
  published: { label: "已发布", variant: "default" },
  failed: { label: "失败", variant: "destructive" },
  rollback: { label: "回滚", variant: "secondary" },
} as const;

export function VersionStatus({ status }: Pick<ConfigVersionSummary, "status">) {
  const meta = statusMeta[status];
  return <Badge variant={meta.variant}>{meta.label}</Badge>;
}

export function formatVersionTime(value: string | null) {
  if (!value) return "—";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
