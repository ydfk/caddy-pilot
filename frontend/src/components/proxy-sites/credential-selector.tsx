import { Link } from "react-router-dom";

import type { BasicAuthCredential } from "@/api/basic-auth";
import { Checkbox } from "@/components/ui/checkbox";
import { FieldDescription } from "@/components/ui/field";

type CredentialSelectorProps = {
  credentials: BasicAuthCredential[];
  selected: string[];
  onChange: (value: string[]) => void;
};

export function CredentialSelector({ credentials, selected, onChange }: CredentialSelectorProps) {
  if (credentials.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
        密码本还是空的。请先前往
        <Link className="mx-1 font-medium text-primary hover:underline" to="/basic-auth">
          Basic Auth 密码本
        </Link>
        创建账号。
      </div>
    );
  }

  return (
    <div className="grid gap-2 sm:grid-cols-2">
      {credentials.map((credential) => {
        const checked = selected.includes(credential.id);
        return (
          <label
            key={credential.id}
            className="flex cursor-pointer items-start gap-3 rounded-lg border p-3 hover:bg-accent/40"
          >
            <Checkbox
              checked={checked}
              onCheckedChange={(next) =>
                onChange(
                  next
                    ? [...selected, credential.id]
                    : selected.filter((id) => id !== credential.id)
                )
              }
            />
            <span>
              <span className="block text-sm font-medium">{credential.name}</span>
              <FieldDescription className="font-mono">{credential.username}</FieldDescription>
            </span>
          </label>
        );
      })}
    </div>
  );
}
