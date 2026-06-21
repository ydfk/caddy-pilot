import { useCallback, useEffect, useState, type FormEvent } from "react";
import { KeyRound, Pencil, Plus, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import {
  createBasicAuthCredential,
  deleteBasicAuthCredential,
  listBasicAuthCredentials,
  updateBasicAuthCredential,
  type BasicAuthCredential,
} from "@/api/basic-auth";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

export default function BasicAuthPage() {
  const [credentials, setCredentials] = useState<BasicAuthCredential[]>([]);
  const [editing, setEditing] = useState<BasicAuthCredential | null>(null);
  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);

  const load = useCallback(async () => {
    try {
      setCredentials(await listBasicAuthCredentials());
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取密码本失败");
    }
  }, []);

  useEffect(() => void load(), [load]);

  function edit(credential: BasicAuthCredential) {
    setEditing(credential);
    setName(credential.name);
    setUsername(credential.username);
    setPassword("");
  }

  function reset() {
    setEditing(null);
    setName("");
    setUsername("");
    setPassword("");
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!editing && password.length < 6) {
      toast.error("密码至少需要 6 个字符");
      return;
    }
    setPending(true);
    try {
      const payload = { name: name.trim(), username: username.trim(), password };
      if (editing) await updateBasicAuthCredential(editing.id, payload);
      else await createBasicAuthCredential(payload);
      toast.success(editing ? "密码条目已更新" : "密码条目已创建");
      reset();
      await load();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存密码条目失败");
    } finally {
      setPending(false);
    }
  }

  async function remove(credential: BasicAuthCredential) {
    if (!window.confirm(`删除密码条目“${credential.name}”？`)) return;
    try {
      await deleteBasicAuthCredential(credential.id);
      if (editing?.id === credential.id) reset();
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
        description="统一维护代理站点使用的账号；密码只保存为 bcrypt 哈希。"
      />
      <div className="grid gap-4 lg:grid-cols-[minmax(18rem,0.75fr)_minmax(0,1.25fr)]">
        <Card>
          <CardHeader>
            <CardTitle>{editing ? "编辑密码条目" : "新增密码条目"}</CardTitle>
            <CardDescription>{editing ? "密码留空时保持原密码。" : "创建后可被多个站点引用。"}</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={submit}>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="credential-name">名称</FieldLabel>
                  <Input id="credential-name" value={name} onChange={(event) => setName(event.target.value)} required />
                  <FieldDescription>例如：运维账号、只读访问。</FieldDescription>
                </Field>
                <Field>
                  <FieldLabel htmlFor="credential-username">用户名</FieldLabel>
                  <Input id="credential-username" value={username} onChange={(event) => setUsername(event.target.value)} required />
                </Field>
                <Field>
                  <FieldLabel htmlFor="credential-password">密码</FieldLabel>
                  <Input id="credential-password" type="password" autoComplete="new-password" minLength={editing ? undefined : 6} value={password} onChange={(event) => setPassword(event.target.value)} required={!editing} />
                </Field>
                <div className="flex gap-2">
                  {editing ? <Button type="button" variant="ghost" onClick={reset}><X data-icon="inline-start" />取消</Button> : null}
                  <Button type="submit" disabled={pending}>{pending ? <Spinner data-icon="inline-start" /> : <Plus data-icon="inline-start" />}{editing ? "保存修改" : "创建条目"}</Button>
                </div>
              </FieldGroup>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>密码条目</CardTitle>
            <CardDescription>不会显示密码明文或哈希。</CardDescription>
          </CardHeader>
          <CardContent>
            {credentials.length === 0 ? (
              <Empty>
                <EmptyHeader><EmptyMedia variant="icon"><KeyRound /></EmptyMedia><EmptyTitle>密码本为空</EmptyTitle><EmptyDescription>创建第一个 Basic Auth 账号。</EmptyDescription></EmptyHeader>
              </Empty>
            ) : (
              <div className="divide-y rounded-lg border">
                {credentials.map((credential) => (
                  <div key={credential.id} className="flex items-center gap-3 p-3">
                    <KeyRound className="size-4 text-muted-foreground" />
                    <div className="min-w-0 flex-1"><p className="truncate text-sm font-medium">{credential.name}</p><p className="truncate font-mono text-xs text-muted-foreground">{credential.username}</p></div>
                    <Button variant="ghost" size="icon" aria-label={`编辑 ${credential.name}`} onClick={() => edit(credential)}><Pencil /></Button>
                    <Button variant="ghost" size="icon" aria-label={`删除 ${credential.name}`} onClick={() => void remove(credential)}><Trash2 /></Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
