import { useRef, type ChangeEvent } from "react";
import { Upload } from "lucide-react";

import { Button } from "@/components/ui/button";

export function CaddyUploadButton({
  disabled,
  onUpload,
}: {
  disabled: boolean;
  onUpload: (file: File) => Promise<void>;
}) {
  const inputRef = useRef<HTMLInputElement>(null);

  async function selectFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (file) await onUpload(file);
  }

  return (
    <>
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        accept=".zip,.tar.gz,.tgz,.exe,application/octet-stream"
        onChange={(event) => void selectFile(event)}
      />
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={disabled}
        onClick={() => inputRef.current?.click()}
      >
        <Upload data-icon="inline-start" />
        上传 Caddy 安装包
      </Button>
    </>
  );
}
