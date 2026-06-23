import { useCallback, useEffect, useState } from "react";
import {
  Braces,
  ChevronLeft,
  ChevronRight,
  Copy,
  FileUp,
  Pencil,
  Plus,
  Route,
  Trash2,
} from "lucide-react";
import { Link } from "react-router-dom";
import { toast } from "sonner";

import {
  deleteProxySite,
  listProxySites,
  importNginxConfig,
  previewProxySite,
  setProxySiteEnabled,
  type ProxySite,
  type ProxySitePreview,
} from "@/api/proxy-sites";
import { getCaddyChangeStatus, publishCaddyConfig, validateCaddyConfig } from "@/api/caddy";
import { NginxImportDialog } from "@/components/proxy-sites/nginx-import-dialog";
import { ProxySitePublishActions } from "@/components/proxy-sites/proxy-site-publish-actions";
import { SiteConfigPreviewDialog } from "@/components/proxy-sites/site-config-preview-dialog";
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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busyID, setBusyID] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<ProxySite | null>(null);
  const [preview, setPreview] = useState<{ name: string; value: ProxySitePreview } | null>(null);
  const [validated, setValidated] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [validating, setValidating] = useState(false);
  const [publishing, setPublishing] = useState(false);

  const loadSites = useCallback(async () => {
    setLoading(true);
    try {
      const [sitePage, status] = await Promise.all([listProxySites(page), getCaddyChangeStatus()]);
      setSites(sitePage.items);
      setTotal(sitePage.total);
      setTotalPages(sitePage.total_pages);
      setHasChanges(status.dirty);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取代理站点失败");
    } finally {
      setLoading(false);
    }
  }, [page]);

  async function refreshChangeStatus() {
    const status = await getCaddyChangeStatus();
    setHasChanges(status.dirty);
    if (!status.dirty) setValidated(false);
  }

  useEffect(() => {
    void loadSites();
  }, [loadSites]);

  async function toggleSite(site: ProxySite, enabled: boolean) {
    setBusyID(site.id);
    try {
      const updated = await setProxySiteEnabled(site.id, enabled);
      setSites((current) => current.map((item) => (item.id === updated.id ? updated : item)));
      setValidated(false);
      await refreshChangeStatus();
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
      setPreview({ name: site.domains[0] ?? "站点", value: result });
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
      if (sites.length === 1 && page > 1) setPage((current) => current - 1);
      else await loadSites();
      setValidated(false);
      await refreshChangeStatus();
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

  async function publishConfig(reason = "从代理站点页面发布") {
    setPublishing(true);
    try {
      const version = await publishCaddyConfig(reason);
      setValidated(false);
      setHasChanges(false);
      toast.success(`配置 v${version.version} 已发布`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "发布配置失败");
    } finally {
      setPublishing(false);
    }
  }

  async function regenerateConfig() {
    setPublishing(true);
    try {
      await validateCaddyConfig();
      const version = await publishCaddyConfig("手动重新生成并发布完整配置");
      setValidated(false);
      setHasChanges(false);
      toast.success(`已重新生成并发布配置 v${version.version}`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "重新生成配置失败");
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
            <ProxySitePublishActions
              hasChanges={hasChanges}
              validated={validated}
              validating={validating}
              publishing={publishing}
              onValidate={() => void validateConfig()}
              onPublish={() => void publishConfig()}
              onRegenerate={() => void regenerateConfig()}
            />
            <NginxImportDialog
              trigger={
                <Button variant="outline">
                  <FileUp data-icon="inline-start" /> 导入 Nginx
                </Button>
              }
              onImport={importNginxConfig}
              onImported={(result) => {
                if (page === 1) void loadSites();
                else setPage(1);
                toast.success(`已导入 ${result.sites.length} 个停用站点`);
                if (result.warnings.length > 0) {
                  toast.warning(result.warnings.join("；"));
                }
              }}
            />
            <Button asChild>
              <Link to="/proxy-sites/new">
                <Plus data-icon="inline-start" /> 新增站点
              </Link>
            </Button>
          </>
        }
      />

      {hasChanges ? (
        <div className="flex items-center justify-between gap-3 rounded-lg border border-primary/25 bg-primary/5 px-4 py-3 text-sm">
          <div>
            <p className="font-medium">检测到未发布的站点变更</p>
            <p className="text-xs text-muted-foreground">先校验完整配置，再确认发布到 Caddy。</p>
          </div>
          <Badge variant={validated ? "default" : "outline"}>
            {validated ? "校验通过" : "需要校验"}
          </Badge>
        </div>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>站点清单</CardTitle>
          <CardDescription>共 {total} 个未删除站点，每页 20 个</CardDescription>
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
                    <TableHead>域名</TableHead>
                    <TableHead>类型</TableHead>
                    <TableHead>目标</TableHead>
                    <TableHead>HTTPS</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>更新时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sites.map((site) => (
                    <TableRow key={site.id}>
                      <TableCell className="max-w-56">
                        <span className="block truncate font-mono text-xs">
                          {site.domains.join(", ")}
                        </span>
                        {site.description ? (
                          <span className="mt-1 block truncate text-xs text-muted-foreground">
                            {site.description}
                          </span>
                        ) : null}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{siteTypeLabel(site.site_type)}</Badge>
                      </TableCell>
                      <TableCell className="max-w-56">
                        <span className="block truncate font-mono text-xs">
                          {site.site_type === "static"
                            ? site.root_path
                            : site.site_type === "spa"
                              ? `${site.api_path} → ${site.upstreams.join(", ")}`
                              : site.upstreams.join(", ")}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge variant={site.enable_https ? "secondary" : "outline"}>
                          {site.enable_https
                            ? site.certificate_type === "wildcard"
                              ? "通配符"
                              : "单域名"
                            : "关闭"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Switch
                            aria-label={`${site.enabled ? "停用" : "启用"}${site.domains[0]}`}
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
                      <TableCell className="min-w-48 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="sm" asChild>
                            <Link to={`/proxy-sites/${site.id}/edit`}>
                              <Pencil /> <span className="hidden xl:inline">编辑</span>
                            </Link>
                          </Button>
                          <Button variant="ghost" size="sm" asChild>
                            <Link to={`/proxy-sites/${site.id}/clone`}>
                              <Copy /> <span className="hidden xl:inline">克隆</span>
                            </Link>
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={busyID === site.id}
                            onClick={() => void showPreview(site)}
                          >
                            {busyID === site.id ? <Spinner /> : <Braces />}
                            <span className="hidden xl:inline">预览</span>
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="text-destructive hover:text-destructive"
                            aria-label={`删除 ${site.domains[0]}`}
                            onClick={() => setDeleteTarget(site)}
                          >
                            <Trash2 />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
          {!loading && totalPages > 1 ? (
            <div className="mt-4 flex items-center justify-between border-t pt-4 text-sm">
              <span className="text-muted-foreground">
                第 {page} / {totalPages} 页
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage((current) => current - 1)}
                >
                  <ChevronLeft /> 上一页
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => setPage((current) => current + 1)}
                >
                  下一页 <ChevronRight />
                </Button>
              </div>
            </div>
          ) : null}
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
              “{deleteTarget?.domains[0]}”将被软删除。已发布配置不会自动改变，需要另行发布。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void removeSite()}>确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <SiteConfigPreviewDialog
        open={Boolean(preview)}
        onOpenChange={(open) => !open && setPreview(null)}
        title={`${preview?.name ?? "站点"} · 配置预览`}
        preview={preview?.value ?? null}
      />
    </div>
  );
}

function siteTypeLabel(siteType: ProxySite["site_type"]) {
  if (siteType === "static") return "静态目录";
  if (siteType === "spa") return "SPA + API";
  return "反向代理";
}
