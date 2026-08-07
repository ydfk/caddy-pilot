import { useCallback, useEffect, useState, type FormEvent } from "react";
import { Fingerprint, Pencil, Plus, ShieldCheck, Trash2 } from "lucide-react";
import { toast } from "sonner";

import {
  beginPasskeyRegistration,
  deletePasskey,
  finishPasskeyRegistration,
  listPasskeys,
  renamePasskey,
  type Passkey,
} from "@/api/auth";
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { createPasskey, passkeyErrorMessage } from "@/lib/passkey";

type EditorState = { mode: "add" | "rename"; credential?: Passkey } | null;

export default function AccountSecurityPage() {
  const [credentials, setCredentials] = useState<Passkey[]>([]);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState(false);
  const [editor, setEditor] = useState<EditorState>(null);
  const [deleteTarget, setDeleteTarget] = useState<Passkey | null>(null);
  const [name, setName] = useState("");

  const loadCredentials = useCallback(async () => {
    setLoading(true);
    try {
      const result = await listPasskeys();
      setCredentials(result.items);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取 Passkey 失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadCredentials();
  }, [loadCredentials]);

  function openEditor(state: NonNullable<EditorState>) {
    setName(state.credential?.name ?? defaultPasskeyName());
    setEditor(state);
  }

  async function submitEditor(event: FormEvent) {
    event.preventDefault();
    if (!editor) return;
    setPending(true);
    try {
      if (editor.mode === "rename" && editor.credential) {
        const updated = await renamePasskey(editor.credential.id, name.trim());
        setCredentials((current) =>
          current.map((credential) => (credential.id === updated.id ? updated : credential))
        );
        toast.success("Passkey 已重命名");
      } else {
        const challenge = await beginPasskeyRegistration(name.trim());
        const credential = await createPasskey(challenge.options.publicKey);
        const created = await finishPasskeyRegistration(challenge.session_id, credential);
        setCredentials((current) => [...current, created]);
        toast.success("Passkey 已添加");
      }
      setEditor(null);
    } catch (error) {
      toast.error(passkeyErrorMessage(error));
    } finally {
      setPending(false);
    }
  }

  async function removeCredential() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    setDeleteTarget(null);
    setPending(true);
    try {
      await deletePasskey(target.id);
      setCredentials((current) => current.filter((credential) => credential.id !== target.id));
      toast.success("Passkey 已删除");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除 Passkey 失败");
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="账户安全"
        actions={
          <Button onClick={() => openEditor({ mode: "add" })}>
            <Plus data-icon="inline-start" /> 添加 Passkey
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldCheck className="size-5 text-primary" /> 登录方式
          </CardTitle>
          <CardDescription>
            密码登录始终保留，可额外登记多个 Passkey。远程访问时浏览器要求管理后台使用 HTTPS。
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="grid gap-3">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : credentials.length === 0 ? (
            <Empty className="border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Fingerprint />
                </EmptyMedia>
                <EmptyTitle>还没有 Passkey</EmptyTitle>
                <EmptyDescription>添加后可使用设备解锁、指纹或安全密钥登录。</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="grid gap-3">
              {credentials.map((credential) => (
                <div
                  key={credential.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-lg border p-4"
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                      <Fingerprint className="size-5" />
                    </div>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <p className="truncate font-medium">{credential.name}</p>
                        {credential.last_used_at ? <Badge variant="outline">已使用</Badge> : null}
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">
                        添加于 {formatDate(credential.created_at)}
                        {credential.last_used_at
                          ? ` · 最近使用 ${formatDate(credential.last_used_at)}`
                          : ""}
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => openEditor({ mode: "rename", credential })}
                    >
                      <Pencil /> 重命名
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => setDeleteTarget(credential)}>
                      <Trash2 /> 删除
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={editor !== null} onOpenChange={(open) => !open && setEditor(null)}>
        <DialogContent>
          <form onSubmit={(event) => void submitEditor(event)}>
            <DialogHeader>
              <DialogTitle>
                {editor?.mode === "rename" ? "重命名 Passkey" : "添加 Passkey"}
              </DialogTitle>
              <DialogDescription>
                {editor?.mode === "rename"
                  ? "名称只用于区分设备，不会修改设备中的凭据。"
                  : "提交后浏览器会要求使用设备解锁、指纹或安全密钥完成登记。"}
              </DialogDescription>
            </DialogHeader>
            <Field className="py-5">
              <FieldLabel htmlFor="passkey-name">名称</FieldLabel>
              <Input
                id="passkey-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                maxLength={128}
                autoFocus
                required
              />
              <FieldDescription>例如“MacBook Touch ID”或“手机”。</FieldDescription>
            </Field>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditor(null)}>
                取消
              </Button>
              <Button type="submit" disabled={pending || !name.trim()}>
                {pending ? <Spinner data-icon="inline-start" /> : null}
                {editor?.mode === "rename" ? "保存" : "继续"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除 Passkey？</AlertDialogTitle>
            <AlertDialogDescription>
              删除“{deleteTarget?.name}”后，此凭据将不能再登录；密码及其他 Passkey 不受影响。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void removeCredential()}>确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function defaultPasskeyName() {
  const platform = navigator.platform?.trim();
  return platform ? `${platform} Passkey` : "我的 Passkey";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
