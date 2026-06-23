import { Download } from "lucide-react";

import type { ProxySitePreview } from "@/api/proxy-sites";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  preview: ProxySitePreview | null;
};

export function SiteConfigPreviewDialog({ open, onOpenChange, title, preview }: Props) {
  function downloadCaddyfile() {
    if (!preview) return;
    const url = URL.createObjectURL(
      new Blob([preview.caddyfile], { type: "text/plain;charset=utf-8" })
    );
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "Caddyfile";
    anchor.click();
    URL.revokeObjectURL(url);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[88vh] max-w-5xl overflow-hidden">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>同一站点配置的只读 JSON 与 Caddyfile 视图。</DialogDescription>
        </DialogHeader>
        <Tabs defaultValue="json" className="min-h-0">
          <TabsList>
            <TabsTrigger value="json">Caddy JSON</TabsTrigger>
            <TabsTrigger value="caddyfile">Caddyfile</TabsTrigger>
          </TabsList>
          <TabsContent value="json">
            <pre className="max-h-[62vh] overflow-auto rounded-lg bg-muted p-4 font-mono text-xs leading-relaxed">
              {JSON.stringify(preview?.caddy_json ?? {}, null, 2)}
            </pre>
          </TabsContent>
          <TabsContent value="caddyfile">
            <pre className="max-h-[62vh] overflow-auto rounded-lg bg-muted p-4 font-mono text-xs leading-relaxed">
              {preview?.caddyfile ?? ""}
            </pre>
          </TabsContent>
        </Tabs>
        <DialogFooter>
          <Button variant="outline" onClick={downloadCaddyfile} disabled={!preview}>
            <Download data-icon="inline-start" /> 导出 Caddyfile
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
