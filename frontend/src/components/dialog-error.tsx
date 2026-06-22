import { AlertCircle } from "lucide-react";

export function DialogError({ message }: { message: string }) {
  return (
    <div className="flex gap-2 rounded-lg border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
      <AlertCircle className="mt-0.5 size-4 shrink-0" />
      <p className="break-words">{message}</p>
    </div>
  );
}
