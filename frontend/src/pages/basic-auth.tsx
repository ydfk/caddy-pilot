import { useCallback, useEffect, useState } from "react";
import { KeyRound, Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";

import {
  createBasicAuthCredential,
  deleteBasicAuthCredential,
  listBasicAuthCredentials,
  updateBasicAuthCredential,
  type BasicAuthCredential,
  type BasicAuthCredentialPayload,
} from "@/api/basic-auth";
import { BasicAuthDialog } from "@/components/basic-auth-dialog";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

export default function BasicAuthPage() {
  const [credentials, setCredentials] = useState<BasicAuthCredential[]>([]);

  const load = useCallback(async () => {
    try {
      setCredentials(await listBasicAuthCredentials());
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取密码本失败");
    }
  }, []);

  useEffect(() => void load(), [load]);

  async function save(payload: BasicAuthCredentialPayload, credential?: BasicAuthCredential) {
    try {
      if (credential) await updateBasicAuthCredential(credential.id, payload);
      else await createBasicAuthCredential(payload);
      await load();
      toast.success(credential ? "密码条目已更新" : "密码条目已创建");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存密码条目失败");
      throw error;
    }
  }

  async function remove(credential: BasicAuthCredential) {
    if (!window.confirm(`删除密码条目“${credential.name}”？`)) return;
    try {
      await deleteBasicAuthCredential(credential.id);
      await load();
      toast.success("密码条目已删除");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除密码条目失败");
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow="ACCESS / BASIC AUTH"
        title="Basic Auth 密码本"
        description="统一维护站点账号；也可以在编辑站点时直接新增。"
        actions={
          <BasicAuthDialog
            trigger={
              <Button>
                <Plus data-icon="inline-start" />
                新增密码条目
              </Button>
            }
            onSave={(payload) => save(payload)}
          />
        }
      />
      <Card>
        <CardHeader>
          <CardTitle>密码条目</CardTitle>
          <CardDescription>不会显示密码明文或哈希。</CardDescription>
        </CardHeader>
        <CardContent>
          {credentials.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <KeyRound />
                </EmptyMedia>
                <EmptyTitle>密码本为空</EmptyTitle>
                <EmptyDescription>创建第一个 Basic Auth 账号。</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="divide-y rounded-lg border">
              {credentials.map((credential) => (
                <div key={credential.id} className="flex items-center gap-3 p-3">
                  <KeyRound className="size-4 text-muted-foreground" />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{credential.name}</p>
                    <p className="truncate font-mono text-xs text-muted-foreground">
                      {credential.username}
                    </p>
                  </div>
                  <BasicAuthDialog
                    credential={credential}
                    trigger={
                      <Button variant="ghost" size="icon" aria-label={`编辑 ${credential.name}`}>
                        <Pencil />
                      </Button>
                    }
                    onSave={(payload) => save(payload, credential)}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`删除 ${credential.name}`}
                    onClick={() => void remove(credential)}
                  >
                    <Trash2 />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
