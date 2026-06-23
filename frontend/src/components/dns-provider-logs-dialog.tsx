import { useCallback, useEffect, useState, type ReactNode } from "react";
import { RefreshCw } from "lucide-react";
import { toast } from "sonner";

import type { DNSProvider } from "@/api/dns-providers";
import { listLogs, type LogEntry } from "@/api/logs";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Spinner } from "@/components/ui/spinner";

type Props = {
  provider: DNSProvider;
  trigger: ReactNode;
};

export function DNSProviderLogsDialog({ provider, trigger }: Props) {
  const [open, setOpen] = useState(false);
  const [entries, setEntries] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const response = await listLogs("dns", 0, 200, provider.id);
      setEntries(response.entries);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取 DNS Provider 日志失败");
    } finally {
      setLoading(false);
    }
  }, [provider.id]);

  useEffect(() => {
    if (open) void load();
  }, [load, open]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="max-h-[85vh] max-w-5xl overflow-hidden">
        <DialogHeader>
          <DialogTitle>{provider.name} · DNS 日志</DialogTitle>
          <DialogDescription>
            仅显示该 Provider 的脱敏调用记录；旧版本产生的日志没有 Provider ID，不会出现在这里。
          </DialogDescription>
        </DialogHeader>
        <div className="flex items-center justify-between">
          <Badge variant="outline">最近 {entries.length} 条</Badge>
          <Button variant="outline" size="sm" disabled={loading} onClick={() => void load()}>
            {loading ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <RefreshCw data-icon="inline-start" />
            )}
            刷新
          </Button>
        </div>
        <div className="h-[min(62vh,640px)] overflow-auto rounded-lg bg-zinc-950 p-3 font-mono text-xs text-zinc-300">
          {entries.length > 0 ? (
            entries.map((entry) => (
              <div
                key={entry.id}
                className="grid gap-1 border-b border-white/5 py-2 sm:grid-cols-[10rem_4rem_1fr] sm:gap-3"
              >
                <span className="text-zinc-500">{formatTime(entry.timestamp)}</span>
                <span className={entry.level === "ERROR" ? "text-red-400" : "text-emerald-400"}>
                  {entry.level || "LOG"}
                </span>
                <span className="break-all whitespace-pre-wrap">
                  {entry.message}
                  {entry.fields && Object.keys(entry.fields).length > 0 ? (
                    <span className="mt-1 block text-zinc-500">{JSON.stringify(entry.fields)}</span>
                  ) : null}
                </span>
              </div>
            ))
          ) : (
            <div className="flex h-full items-center justify-center text-zinc-600">
              {loading ? "正在读取日志…" : "暂无该 Provider 的日志"}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function formatTime(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString("zh-CN", { hour12: false });
}
