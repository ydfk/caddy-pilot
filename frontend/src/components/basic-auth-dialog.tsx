import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { Save } from "lucide-react";

import type { BasicAuthCredential, BasicAuthCredentialPayload } from "@/api/basic-auth";
import { DialogError } from "@/components/dialog-error";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

type Props = {
  credential?: BasicAuthCredential | null;
  trigger: ReactNode;
  onSave: (payload: BasicAuthCredentialPayload) => Promise<unknown>;
};

export function BasicAuthDialog({ credential, trigger, onSave }: Props) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    if (!open) return;
    setName(credential?.name ?? "");
    setUsername(credential?.username ?? "");
    setPassword("");
    setErrorMessage("");
  }, [credential, open]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    setErrorMessage("");
    try {
      await onSave({ name: name.trim(), username: username.trim(), password });
      setOpen(false);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "保存密码条目失败");
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={submit}>
          <DialogHeader>
            <DialogTitle>{credential ? "编辑密码条目" : "新增密码条目"}</DialogTitle>
            <DialogDescription>密码只用于生成 bcrypt 哈希，不会保存或回传明文。</DialogDescription>
          </DialogHeader>
          <FieldGroup className="py-5">
            {errorMessage ? <DialogError message={errorMessage} /> : null}
            <Field>
              <FieldLabel htmlFor="credential-dialog-name">名称</FieldLabel>
              <Input
                id="credential-dialog-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="运维账号"
                required
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="credential-dialog-username">用户名</FieldLabel>
              <Input
                id="credential-dialog-username"
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                required
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="credential-dialog-password">密码</FieldLabel>
              <Input
                id="credential-dialog-password"
                type="password"
                autoComplete="new-password"
                minLength={credential ? undefined : 6}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required={!credential}
              />
              {credential ? <FieldDescription>留空时保持原密码。</FieldDescription> : null}
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button type="submit" disabled={pending}>
              {pending ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
