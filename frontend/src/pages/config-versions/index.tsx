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
import { PageHeader } from "@/components/page-header";
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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

export default function ConfigVersionsPage() {
  const [versions, setVersions] = useState<ConfigVersionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [rollbackTarget, setRollbackTarget] = useState<ConfigVersionSummary | null>(null);

  const loadVersions = useCallback(async () => {
    setLoading(true);
    try {
      setVersions(await listConfigVersions());
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取配置版本失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadVersions();
  }, [loadVersions]);

  async function rollback() {
    if (!rollbackTarget) return;
    const target = rollbackTarget;
    setRollbackTarget(null);
    try {
      const version = await rollbackConfigVersion(target.id);
      toast.success(`已创建回滚版本 v${version.version}`);
      await loadVersions();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "回滚配置失败");
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        eyebrow="CONFIG / HISTORY"
        title="配置版本"
        description="每次发布和回滚都会产生不可变的历史记录，失败操作同样留痕。"
      />
      <Card>
        <CardHeader>
          <CardTitle>版本历史</CardTitle>
          <CardDescription>按版本号倒序排列</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex flex-col gap-3">
              {Array.from({ length: 5 }, (_, index) => (
                <Skeleton key={index} className="h-12 w-full" />
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
                    <TableHead>版本号</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>原因</TableHead>
                    <TableHead>发布时间</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {versions.map((version) => (
                    <TableRow key={version.id}>
                      <TableCell className="font-mono font-medium">v{version.version}</TableCell>
                      <TableCell>
                        <VersionStatus status={version.status} />
                      </TableCell>
                      <TableCell className="max-w-64 truncate">{version.reason || "—"}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {formatVersionTime(version.published_at)}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {formatVersionTime(version.created_at)}
                      </TableCell>
                      <TableCell>
                        <div className="flex justify-end gap-2">
                          <Button variant="outline" size="sm" asChild>
                            <Link to={`/config-versions/${version.id}`}>查看 JSON</Link>
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setRollbackTarget(version)}
                          >
                            <RotateCcw data-icon="inline-start" /> 回滚
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
      </Card>

      <AlertDialog
        open={Boolean(rollbackTarget)}
        onOpenChange={(open) => !open && setRollbackTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>回滚到 v{rollbackTarget?.version}？</AlertDialogTitle>
            <AlertDialogDescription>
              历史 Caddy JSON 会先补齐受保护管理入口，再发布到本地节点，并创建一个新的回滚版本。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void rollback()}>确认回滚</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
