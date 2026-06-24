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
  initialTab?: "json" | "caddyfile";
};

export function SiteConfigPreviewDialog({
  open,
  onOpenChange,
  title,
  preview,
  initialTab = "json",
}: Props) {
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
          <DialogDescription>
            可视化模式同时生成两种格式；自定义 Caddyfile 通过 Caddy 官方适配器生成 JSON。
          </DialogDescription>
        </DialogHeader>
        <Tabs key={initialTab} defaultValue={initialTab} className="min-h-0">
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
