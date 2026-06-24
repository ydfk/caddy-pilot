import { Braces, SlidersHorizontal } from "lucide-react";
import { Controller, type Control, type FieldErrors } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import type { SiteFormValues } from "./site-form-data";

type Props = {
  control: Control<SiteFormValues>;
  errors: FieldErrors<SiteFormValues>;
  mode: SiteFormValues["configMode"];
  format: SiteFormValues["customFormat"];
};

export function ConfigModeEditor({ control, errors, mode, format }: Props) {
  return (
    <div className="space-y-3">
      <Controller
        control={control}
        name="configMode"
        render={({ field }) => (
          <div className="grid grid-cols-2 gap-2 rounded-lg bg-muted p-1">
            <ModeButton active={field.value === "visual"} onClick={() => field.onChange("visual")}>
              <SlidersHorizontal /> 可视化配置
            </ModeButton>
            <ModeButton active={field.value === "custom"} onClick={() => field.onChange("custom")}>
              <Braces /> 自定义配置
            </ModeButton>
          </div>
        )}
      />

      {mode === "custom" ? (
        <Controller
          control={control}
          name="customConfig"
          render={({ field }) => (
            <Field data-invalid={Boolean(errors.customConfig) || undefined}>
              <div className="flex items-center justify-between gap-3">
                <FieldLabel htmlFor="customConfig">站点配置</FieldLabel>
                <Controller
                  control={control}
                  name="customFormat"
                  render={({ field: formatField }) => (
                    <Tabs value={formatField.value} onValueChange={formatField.onChange}>
                      <TabsList className="h-8">
                        <TabsTrigger value="caddyfile" className="h-7 px-3 text-xs">
                          Caddyfile
                        </TabsTrigger>
                        <TabsTrigger value="json" className="h-7 px-3 text-xs">
                          JSON
                        </TabsTrigger>
                      </TabsList>
                    </Tabs>
                  )}
                />
              </div>
              <Textarea
                id="customConfig"
                rows={18}
                className="min-h-80 resize-y font-mono text-xs leading-5"
                placeholder={
                  format === "caddyfile"
                    ? "example.com {\n\treverse_proxy 127.0.0.1:3000\n}"
                    : '{\n  "match": [{ "host": ["example.com"] }],\n  "handle": []\n}'
                }
                {...field}
              />
              <FieldError errors={[errors.customConfig]} />
            </Field>
          )}
        />
      ) : null}
    </div>
  );
}

function ModeButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <Button
      type="button"
      variant="ghost"
      className={cn("h-9 gap-2", active && "bg-background shadow-sm hover:bg-background")}
      onClick={onClick}
    >
      {children}
    </Button>
  );
}
