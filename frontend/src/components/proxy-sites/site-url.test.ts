import { publicSiteURL, upstreamURL } from "./site-url";

describe("代理站点公开地址", () => {
  test("标准端口不写入地址", () => {
    expect(publicSiteURL("example.com", false, { http: 80, https: 443 })).toBe(
      "http://example.com"
    );
    expect(publicSiteURL("example.com", true, { http: 80, https: 443 })).toBe(
      "https://example.com"
    );
  });

  test("非标准端口写入地址", () => {
    expect(publicSiteURL("example.com", false, { http: 18080, https: 18443 })).toBe(
      "http://example.com:18080"
    );
    expect(publicSiteURL("example.com", true, { http: 18080, https: 18443 })).toBe(
      "https://example.com:18443"
    );
  });

  test("通配符和无效域名不生成链接", () => {
    expect(publicSiteURL("*.example.com", true, { http: 80, https: 443 })).toBeNull();
    expect(publicSiteURL("", false, { http: 80, https: 443 })).toBeNull();
  });
});

describe("代理站点上游地址", () => {
  test("根据上游类型生成浏览器可访问地址", () => {
    expect(upstreamURL("192.168.1.10:3000", "http")).toBe("http://192.168.1.10:3000");
    expect(upstreamURL("https://nas.local:9443", "https")).toBe("https://nas.local:9443");
    expect(upstreamURL("app.internal:8080", "h2c")).toBe("http://app.internal:8080");
  });

  test("支持带路径的地址并拒绝 Unix Socket", () => {
    expect(upstreamURL("192.168.1.10:3000/health", "http")).toBe("http://192.168.1.10:3000/health");
    expect(upstreamURL("/run/app.sock", "unix")).toBeNull();
  });

  test("拒绝空地址和包含认证信息的地址", () => {
    expect(upstreamURL("", "http")).toBeNull();
    expect(upstreamURL("user:password@192.168.1.10", "http")).toBeNull();
  });
});
