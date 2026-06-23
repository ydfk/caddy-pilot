import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { CirclePause, CirclePlay, RefreshCw, Search, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { listLogs, type LogEntry, type LogSource } from "@/api/logs";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function LogsPage() {
  const [source, setSource] = useState<LogSource>("system");
  const [entries, setEntries] = useState<LogEntry[]>([]);
  const [cursor, setCursor] = useState(0);
  const [paused, setPaused] = useState(false);
  const [level, setLevel] = useState("ALL");
  const [keyword, setKeyword] = useState("");
  const loadingRef = useRef(false);

  const load = useCallback(
    async (reset = false) => {
      if (loadingRef.current) return;
      loadingRef.current = true;
      try {
        const response = await listLogs(source, reset ? 0 : cursor);
        setCursor(response.next_cursor);
        setEntries((current) =>
          reset ? response.entries : [...current, ...response.entries].slice(-1000)
        );
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "读取日志失败");
      } finally {
        loadingRef.current = false;
      }
    },
    [cursor, source]
  );

  useEffect(() => {
    setEntries([]);
    setCursor(0);
    void load(true);
  }, [source]);

  useEffect(() => {
    if (paused) return;
    const timer = window.setInterval(() => void load(), 2000);
    return () => window.clearInterval(timer);
  }, [load, paused]);

  const filtered = useMemo(() => {
    const query = keyword.trim().toLowerCase();
    return entries.filter((entry) => {
      const levelMatches = level === "ALL" || entry.level === level;
      return levelMatches && (!query || entry.message.toLowerCase().includes(query));
    });
  }, [entries, keyword, level]);

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-4">
      <PageHeader
        eyebrow="RUNTIME / LOG STREAM"
        title="日志"
        description="在线查看 CaddyPilot 系统日志与托管 Caddy 进程日志。"
      />
      <Card className="overflow-hidden border-primary/20">
        <CardHeader className="border-b bg-primary/[0.035] p-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Tabs value={source} onValueChange={(value) => setSource(value as LogSource)}>
              <TabsList>
                <TabsTrigger value="system">系统日志</TabsTrigger>
                <TabsTrigger value="caddy">Caddy 日志</TabsTrigger>
                <TabsTrigger value="dns">DNS Provider</TabsTrigger>
              </TabsList>
            </Tabs>
            <div className="flex flex-wrap gap-2">
              <Select value={level} onValueChange={setLevel}>
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {["ALL", "DEBUG", "INFO", "WARN", "ERROR"].map((item) => (
                    <SelectItem key={item} value={item}>
                      {item}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
                <Input
                  className="w-52 pl-8"
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                  placeholder="过滤关键词"
                />
              </div>
              <Button
                variant="outline"
                size="icon"
                aria-label={paused ? "继续刷新" : "暂停刷新"}
                onClick={() => setPaused((value) => !value)}
              >
                {paused ? <CirclePlay /> : <CirclePause />}
              </Button>
              <Button
                variant="outline"
                size="icon"
                aria-label="立即刷新"
                onClick={() => void load()}
              >
                <RefreshCw />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                aria-label="清空当前视图"
                onClick={() => setEntries([])}
              >
                <Trash2 />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {source === "dns" ? (
            <div className="border-b bg-amber-500/5 px-3 py-2 text-xs text-muted-foreground">
              DNS Provider 审计日志会记录域名、记录名称、类型和 TXT 值，但不会记录 AccessKey、Token、签名或认证头。
            </div>
          ) : null}
          <div className="h-[min(68vh,720px)] overflow-auto bg-zinc-950 p-3 font-mono text-xs text-zinc-300">
            {filtered.length ? (
              filtered.map((entry) => (
                <div
                  key={entry.id}
                  className="grid grid-cols-[10rem_4rem_1fr] gap-3 border-b border-white/5 py-1.5"
                >
                  <span className="text-zinc-500">{formatTime(entry.timestamp)}</span>
                  <span className={levelClass(entry.level)}>{entry.level || "LOG"}</span>
                  <span className="break-all whitespace-pre-wrap">
                    {entry.message}
                    {entry.fields && Object.keys(entry.fields).length ? (
                      <span className="mt-1 block text-zinc-500">
                        {JSON.stringify(entry.fields)}
                      </span>
                    ) : null}
                  </span>
                </div>
              ))
            ) : (
              <div className="flex h-full items-center justify-center text-zinc-600">
                暂无匹配日志
              </div>
            )}
          </div>
          <div className="flex items-center justify-between border-t px-3 py-2 text-xs text-muted-foreground">
            <span>{paused ? "已暂停" : "每 2 秒自动刷新"}</span>
            <Badge variant="outline">当前视图 {filtered.length} 条</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function formatTime(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString("zh-CN", { hour12: false });
}

function levelClass(level?: string) {
  if (level === "ERROR") return "text-red-400";
  if (level === "WARN") return "text-amber-400";
  if (level === "INFO") return "text-emerald-400";
  return "text-sky-400";
}
