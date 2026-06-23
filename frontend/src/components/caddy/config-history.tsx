import { useCallback, useEffect, useState } from "react";
import { FileClock, RotateCcw } from "lucide-react";
import { Link } from "react-router-dom";
import { toast } from "sonner";

import {
  listConfigVersions,
  rollbackConfigVersion,
  type ConfigVersionSummary,
} from "@/api/config-versions";
import { formatVersionTime, VersionStatus } from "@/components/config-versions/version-status";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { CardContent } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function ConfigHistory({
  refreshKey,
  onRollback,
  activeVersion,
}: {
  refreshKey: number;
  onRollback: () => Promise<void>;
  activeVersion?: number;
}) {
  const [versions, setVersions] = useState<ConfigVersionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [rollbackTarget, setRollbackTarget] = useState<ConfigVersionSummary | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setVersions(await listConfigVersions());
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取配置版本失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => void load(), [load, refreshKey]);

  async function rollback() {
    if (!rollbackTarget) return;
    const target = rollbackTarget;
    setRollbackTarget(null);
    try {
      const version = await rollbackConfigVersion(target.id);
      toast.success(`已创建回滚版本 v${version.version}`);
      await load();
      await onRollback();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "回滚配置失败");
    }
  }

  return (
    <>
      <CardContent className="border-t pt-5">
        {loading ? (
          <div className="grid gap-3">
            {Array.from({ length: 4 }, (_, index) => (
              <Skeleton key={index} className="h-11 w-full" />
            ))}
          </div>
        ) : versions.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <FileClock />
              </EmptyMedia>
              <EmptyTitle>还没有配置版本</EmptyTitle>
              <EmptyDescription>首次发布后，版本会出现在这里。</EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>版本</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>说明</TableHead>
                  <TableHead>时间</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {versions.map((version) => (
                  <TableRow key={version.id}>
                    <TableCell>
                      <div className="flex items-center gap-2 font-mono font-medium">
                        v{version.version}
                        {activeVersion === version.version ? (
                          <span className="rounded border px-1.5 py-0.5 font-sans text-[10px] text-emerald-600">
                            当前运行
                          </span>
                        ) : null}
                      </div>
                    </TableCell>
                    <TableCell>
                      <VersionStatus status={version.status} />
                    </TableCell>
                    <TableCell className="max-w-72 truncate">{version.reason || "—"}</TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {formatVersionTime(version.published_at ?? version.created_at)}
                    </TableCell>
                    <TableCell>
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="sm" asChild>
                          <Link to={`/config-versions/${version.id}`}>查看</Link>
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setRollbackTarget(version)}
                        >
                          <RotateCcw data-icon="inline-start" />
                          回滚
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
      <AlertDialog
        open={Boolean(rollbackTarget)}
        onOpenChange={(open) => !open && setRollbackTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>回滚到 v{rollbackTarget?.version}？</AlertDialogTitle>
            <AlertDialogDescription>
              系统会保护管理入口、恢复历史 JSON，并创建新的回滚版本。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void rollback()}>确认回滚</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
