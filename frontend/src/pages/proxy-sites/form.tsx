import { useEffect, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";

import {
  cloneProxySite,
  createProxySite,
  getProxySite,
  updateProxySite,
  type ProxySite,
} from "@/api/proxy-sites";
import { publishCaddyConfig } from "@/api/caddy";
import { listBasicAuthCredentials, type BasicAuthCredential } from "@/api/basic-auth";
import { JSONDialog } from "@/components/json-dialog";
import { SiteForm } from "@/components/proxy-sites/site-form";
import {
  defaultSiteValues,
  draftPreview,
  formValuesFromSite,
  payloadFromForm,
  type SiteFormValues,
} from "@/components/proxy-sites/site-form-data";
import { Skeleton } from "@/components/ui/skeleton";

export default function ProxySiteFormPage() {
  const { id } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const mode = !id ? "new" : location.pathname.endsWith("/clone") ? "clone" : "edit";
  const [values, setValues] = useState<SiteFormValues>(defaultSiteValues);
  const [loading, setLoading] = useState(Boolean(id));
  const [pending, setPending] = useState(false);
  const [preview, setPreview] = useState<unknown>(null);
  const [credentials, setCredentials] = useState<BasicAuthCredential[]>([]);

  useEffect(() => {
    listBasicAuthCredentials()
      .then(setCredentials)
      .catch((error) => toast.error(error instanceof Error ? error.message : "读取密码本失败"));
  }, []);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    getProxySite(id)
      .then((site) => setValues(formValuesFromSite(site, mode === "clone")))
      .catch((error) => toast.error(error instanceof Error ? error.message : "读取代理站点失败"))
      .finally(() => setLoading(false));
  }, [id, mode]);

  async function save(formValues: SiteFormValues, publish: boolean) {
    setPending(true);
    try {
      const payload = payloadFromForm(formValues, mode === "clone");
      let saved: ProxySite;
      if (mode === "new") {
        saved = await createProxySite(payload);
      } else if (mode === "edit" && id) {
        saved = await updateProxySite(id, payload);
      } else if (id) {
        const cloned = await cloneProxySite(id, {
          name: payload.name,
          domains: payload.domains,
          upstreams: payload.upstreams,
        });
        saved = await updateProxySite(cloned.id, payload);
      } else {
        throw new Error("缺少原站点 ID");
      }
      if (publish) await publishCaddyConfig(`保存站点：${saved.name}`);
      toast.success(publish ? "站点已保存并发布" : "站点已保存");
      navigate("/proxy-sites");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存代理站点失败");
    } finally {
      setPending(false);
    }
  }

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  return (
    <>
      <SiteForm
        mode={mode}
        values={values}
        pending={pending}
        credentials={credentials}
        onSave={save}
        onPreview={(data) => setPreview(draftPreview(data))}
      />
      <JSONDialog
        open={preview !== null}
        onOpenChange={(open) => !open && setPreview(null)}
        title="站点草稿 · Caddy JSON"
        description="此处为表单草稿片段；完整配置请在 Caddy 状态页预览。"
        value={preview}
      />
    </>
  );
}
