import { publicSiteURL } from "./site-url";

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
