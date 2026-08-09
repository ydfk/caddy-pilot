export type PublicSitePorts = { http: number; https: number };
export type UpstreamType = "http" | "https" | "h2c" | "unix";

export function publicSiteURL(domain: string, enableHTTPS: boolean, ports: PublicSitePorts) {
  const host = domain.trim();
  if (!host || host.includes("*")) return null;
  const scheme = enableHTTPS ? "https" : "http";
  const port = enableHTTPS ? ports.https : ports.http;
  try {
    const url = new URL(`${scheme}://${host}`);
    if (!url.hostname || url.username || url.password || url.pathname !== "/") return null;
    url.port = port === (enableHTTPS ? 443 : 80) ? "" : String(port);
    return url.origin;
  } catch {
    return null;
  }
}

export function upstreamURL(upstream: string, type: UpstreamType) {
  if (type === "unix") return null;
  const address = upstream.trim().replace(/^(?:https?|h2c):\/\//i, "");
  if (!address) return null;
  const scheme = type === "https" ? "https" : "http";
  try {
    const url = new URL(`${scheme}://${address}`);
    if (!url.hostname || url.username || url.password) return null;
    return url.pathname === "/" && !url.search && !url.hash ? url.origin : url.href;
  } catch {
    return null;
  }
}
