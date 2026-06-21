import { useCallback, useEffect, useState } from "react";
import {
  Braces,
  Copy,
  MoreHorizontal,
  Pencil,
  Plus,
  Rocket,
  Route,
  ShieldCheck,
  Trash2,
} from "lucide-react";
import { Link } from "react-router-dom";
import { toast } from "sonner";

import {
  deleteProxySite,
  listProxySites,
  previewProxySite,
  setProxySiteEnabled,
  type ProxySite,
} from "@/api/proxy-sites";
import { publishCaddyConfig, validateCaddyConfig } from "@/api/caddy";
import { JSONDialog } from "@/components/json-dialog";
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
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export default function ProxySitesPage() {
  const [sites, setSites] = useState<ProxySite[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyID, setBusyID] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<ProxySite | null>(null);
  const [preview, setPreview] = useState<{ name: string; value: unknown } | null>(null);
  const [validated, setValidated] = useState(false);
  const [validating, setValidating] = useState(false);
  const [publishing, setPublishing] = useState(false);

  const loadSites = useCallback(async () => {
    setLoading(true);
    try {
      setSites(await listProxySites());
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取代理站点失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadSites();
  }, [loadSites]);

  async function toggleSite(site: ProxySite, enabled: boolean) {
    setBusyID(site.id);
    try {
      const updated = await setProxySiteEnabled(site.id, enabled);
      setSites((current) => current.map((item) => (item.id === updated.id ? updated : item)));
      setValidated(false);
      toast.success(enabled ? "站点已启用" : "站点已停用");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新站点状态失败");
    } finally {
      setBusyID("");
    }
  }

  async function showPreview(site: ProxySite) {
    setBusyID(site.id);
    try {
      const result = await previewProxySite(site.id);
      setPreview({ name: site.name, value: result.caddy_json });
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "生成预览失败");
    } finally {
      setBusyID("");
    }
  }

  async function removeSite() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    setDeleteTarget(null);
    setBusyID(target.id);
    try {
      await deleteProxySite(target.id);
      setSites((current) => current.filter((item) => item.id !== target.id));
      setValidated(false);
      toast.success("代理站点已删除");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除代理站点失败");
    } finally {
      setBusyID("");
    }
  }

  async function validateConfig() {
    setValidating(true);
    try {
      await validateCaddyConfig();
      setValidated(true);
      toast.success("配置校验通过，可以发布");
    } catch (error) {
      setValidated(false);
      toast.error(error instanceof Error ? error.message : "配置校验失败");
    } finally {
      setValidating(false);
    }
  }

  async function publishConfig() {
    setPublishing(true);
    try {
      const version = await publishCaddyConfig("从代理站点页面发布");
      setValidated(false);
      toast.success(`配置 v${version.version} 已发布`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "发布配置失败");
    } finally {
      setPublishing(false);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow="ROUTES / SITES"
        title="代理站点"
        description="维护域名、上游与发布状态。修改不会自动推送到 Caddy。"
        actions={
          <>
            <Button variant="outline" onClick={() => void validateConfig()} disabled={validating}>
              {validating ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <ShieldCheck data-icon="inline-start" />
              )}
              1. 校验配置
            </Button>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button disabled={!validated || publishing}>
                  {publishing ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <Rocket data-icon="inline-start" />
                  )}
                  2. 发布
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>发布已校验的站点配置？</AlertDialogTitle>
                  <AlertDialogDescription>
                    后端会重新生成完整配置、保留 :8080 管理入口并发布到本地 Caddy。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>取消</AlertDialogCancel>
                  <AlertDialogAction onClick={() => void publishConfig()}>
                    确认发布
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
            <Button variant="secondary" asChild>
              <Link to="/proxy-sites/new">
                <Plus data-icon="inline-start" /> 新增站点
              </Link>
            </Button>
          </>
        }
      />

      <div className="flex items-center justify-between gap-3 rounded-lg border bg-card px-4 py-3 text-sm">
        <div>
          <p className="font-medium">推荐流程：编辑站点 → 校验配置 → 发布</p>
          <p className="text-xs text-muted-foreground">任何站点状态变更后都需要重新校验。</p>
        </div>
        <Badge variant={validated ? "default" : "outline"}>
          {validated ? "校验通过" : "等待校验"}
        </Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>站点清单</CardTitle>
          <CardDescription>共 {sites.length} 个未删除站点</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex flex-col gap-3">
              {Array.from({ length: 4 }, (_, index) => (
                <Skeleton key={index} className="h-12 w-full" />
              ))}
            </div>
          ) : sites.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Route />
                </EmptyMedia>
                <EmptyTitle>还没有代理站点</EmptyTitle>
                <EmptyDescription>从一个域名和上游地址开始。</EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button asChild size="sm">
                  <Link to="/proxy-sites/new">新增第一个站点</Link>
                </Button>
              </EmptyContent>
            </Empty>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>名称</TableHead>
                    <TableHead>域名</TableHead>
                    <TableHead>上游</TableHead>
                    <TableHead>HTTPS</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>更新时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sites.map((site) => (
                    <TableRow key={site.id}>
                      <TableCell className="font-medium">{site.name}</TableCell>
                      <TableCell className="max-w-56">
                        <span className="block truncate font-mono text-xs">
                          {site.domains.join(", ")}
                        </span>
                      </TableCell>
                      <TableCell className="max-w-56">
                        <span className="block truncate font-mono text-xs">
                          {site.upstreams.join(", ")}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge variant={site.enable_https ? "secondary" : "outline"}>
                          {site.enable_https ? "启用" : "关闭"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Switch
                            aria-label={`${site.enabled ? "停用" : "启用"}${site.name}`}
                            checked={site.enabled}
                            disabled={busyID === site.id}
                            onCheckedChange={(enabled) => void toggleSite(site, enabled)}
                          />
                          <span className="text-xs text-muted-foreground">
                            {site.enabled ? "运行" : "停用"}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {new Intl.DateTimeFormat("zh-CN", {
                          month: "2-digit",
                          day: "2-digit",
                          hour: "2-digit",
                          minute: "2-digit",
                        }).format(new Date(site.updated_at))}
                      </TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" aria-label={`${site.name} 操作`}>
                              <MoreHorizontal />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuGroup>
                              <DropdownMenuItem asChild>
                                <Link to={`/proxy-sites/${site.id}/edit`}>
                                  <Pencil /> 编辑
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuItem asChild>
                                <Link to={`/proxy-sites/${site.id}/clone`}>
                                  <Copy /> 克隆
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuItem onSelect={() => void showPreview(site)}>
                                <Braces /> 预览配置
                              </DropdownMenuItem>
                            </DropdownMenuGroup>
                            <DropdownMenuSeparator />
                            <DropdownMenuGroup>
                              <DropdownMenuItem
                                variant="destructive"
                                onSelect={() => setDeleteTarget(site)}
                              >
                                <Trash2 /> 删除
                              </DropdownMenuItem>
                            </DropdownMenuGroup>
                          </DropdownMenuContent>
                        </DropdownMenu>
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
        open={Boolean(deleteTarget)}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除代理站点？</AlertDialogTitle>
            <AlertDialogDescription>
              “{deleteTarget?.name}”将被软删除。已发布配置不会自动改变，需要另行发布。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void removeSite()}>确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <JSONDialog
        open={Boolean(preview)}
        onOpenChange={(open) => !open && setPreview(null)}
        title={`${preview?.name ?? "站点"} · Caddy JSON 片段`}
        value={preview?.value}
      />
    </div>
  );
}
