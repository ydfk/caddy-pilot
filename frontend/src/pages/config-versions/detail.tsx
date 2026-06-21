import { useEffect, useState } from "react";
import { ArrowLeft, RotateCcw } from "lucide-react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";

import { getConfigVersion, rollbackConfigVersion, type ConfigVersion } from "@/api/config-versions";
import { formatVersionTime, VersionStatus } from "@/components/config-versions/version-status";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function ConfigVersionDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [version, setVersion] = useState<ConfigVersion | null>(null);

  useEffect(() => {
    if (!id) return;
    getConfigVersion(id)
      .then(setVersion)
      .catch((error) => toast.error(error instanceof Error ? error.message : "读取配置版本失败"));
  }, [id]);

  async function rollback() {
    if (!version) return;
    try {
      const created = await rollbackConfigVersion(version.id);
      toast.success(`已创建回滚版本 v${created.version}`);
      navigate(`/config-versions/${created.id}`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "回滚配置失败");
    }
  }

  if (!version) return <Skeleton className="h-96 w-full" />;

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow="CONFIG / SNAPSHOT"
        title={`配置版本 v${version.version}`}
        description={`${version.reason || "无发布原因"} · ${formatVersionTime(version.created_at)}`}
        actions={
          <>
            <Button variant="outline" asChild>
              <Link to="/config-versions">
                <ArrowLeft data-icon="inline-start" /> 返回
              </Link>
            </Button>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button>
                  <RotateCcw data-icon="inline-start" /> 回滚到此版本
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>回滚到 v{version.version}？</AlertDialogTitle>
                  <AlertDialogDescription>
                    该操作会立即调用本地 Caddy Admin API，并创建一条新的回滚记录。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>取消</AlertDialogCancel>
                  <AlertDialogAction onClick={() => void rollback()}>确认回滚</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </>
        }
      />

      {version.error_message ? (
        <Alert variant="destructive">
          <AlertTitle>操作失败</AlertTitle>
          <AlertDescription>{version.error_message}</AlertDescription>
        </Alert>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            v{version.version} <VersionStatus status={version.status} />
          </CardTitle>
          <CardDescription>发布时间：{formatVersionTime(version.published_at)}</CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="caddy">
            <TabsList>
              <TabsTrigger value="caddy">Caddy JSON</TabsTrigger>
              <TabsTrigger value="business">Business Config</TabsTrigger>
            </TabsList>
            <TabsContent value="caddy">
              <JSONBlock value={version.caddy_json} />
            </TabsContent>
            <TabsContent value="business">
              <JSONBlock value={version.business_config} />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}

function JSONBlock({ value }: { value: unknown }) {
  return (
    <pre className="mt-4 max-h-[62vh] overflow-auto rounded-lg bg-muted p-4 font-mono text-xs leading-relaxed">
      {JSON.stringify(value, null, 2)}
    </pre>
  );
}
