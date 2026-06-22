import { Plus } from "lucide-react";

import type { BasicAuthCredential, BasicAuthCredentialPayload } from "@/api/basic-auth";
import { BasicAuthDialog } from "@/components/basic-auth-dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { FieldDescription } from "@/components/ui/field";

type CredentialSelectorProps = {
  credentials: BasicAuthCredential[];
  selected: string[];
  onChange: (value: string[]) => void;
  onCreate: (payload: BasicAuthCredentialPayload) => Promise<BasicAuthCredential>;
};

export function CredentialSelector({
  credentials,
  selected,
  onChange,
  onCreate,
}: CredentialSelectorProps) {
  return (
    <div className="grid gap-2">
      {credentials.length ? (
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
      ) : (
        <FieldDescription>密码本为空，可在这里创建第一个账号。</FieldDescription>
      )}
      <BasicAuthDialog
        trigger={
          <Button type="button" variant="outline" size="sm" className="w-fit">
            <Plus data-icon="inline-start" />
            新增密码条目
          </Button>
        }
        onSave={async (payload) => {
          const created = await onCreate(payload);
          onChange([...selected, created.id]);
        }}
      />
    </div>
  );
}
