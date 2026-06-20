import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

type JSONDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  value: unknown;
};

export function JSONDialog({ open, onOpenChange, title, description, value }: JSONDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] max-w-4xl overflow-hidden">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description || "只读 JSON 预览"}</DialogDescription>
        </DialogHeader>
        <pre className="max-h-[65vh] overflow-auto rounded-lg bg-muted p-4 font-mono text-xs leading-relaxed">
          {JSON.stringify(value, null, 2)}
        </pre>
      </DialogContent>
    </Dialog>
  );
}
