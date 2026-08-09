import { ExternalLink } from "lucide-react";

import type { ProxySite } from "@/api/proxy-sites";
import { upstreamURL } from "@/components/proxy-sites/site-url";

export function UpstreamTarget({ site }: { site: ProxySite }) {
  if (site.config_mode === "custom") {
    return <span className="block truncate font-mono text-xs">完全自定义</span>;
  }
  if (site.site_type === "static") {
    return <span className="block truncate font-mono text-xs">{site.root_path}</span>;
  }

  return (
    <div className="flex max-w-56 flex-wrap items-center gap-x-2 gap-y-1">
      {site.site_type === "spa" ? (
        <span className="font-mono text-xs text-muted-foreground">{site.api_path} →</span>
      ) : null}
      {site.upstreams.map((upstream, index) => {
        const href = upstreamURL(upstream, site.upstream_type);
        return href ? (
          <a
            key={`${upstream}-${index}`}
            href={href}
            target="_blank"
            rel="noreferrer"
            className="inline-flex min-w-0 items-center gap-1 font-mono text-xs text-primary hover:underline"
            title={href}
          >
            <span className="truncate">{upstream}</span>
            <ExternalLink className="size-3 shrink-0" aria-hidden="true" />
          </a>
        ) : (
          <span key={`${upstream}-${index}`} className="truncate font-mono text-xs">
            {upstream}
          </span>
        );
      })}
    </div>
  );
}
