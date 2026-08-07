export type PublicSitePorts = { http: number; https: number };

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
