import { useEffect, useState, type FormEvent } from "react";
import { AlertCircle, ArrowRight } from "lucide-react";
import { useLocation, useNavigate } from "react-router-dom";

import { getSetupStatus, login, register } from "@/api/auth";
import { BrandLogo } from "@/components/brand-logo";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { useAuthStore } from "@/store/auth-store";

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const setToken = useAuthStore((state) => state.setToken);
  const [screen, setScreen] = useState<"loading" | "setup" | "login">("loading");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    getSetupStatus()
      .then(({ initialized }) => {
        if (active) setScreen(initialized ? "login" : "setup");
      })
      .catch((reason) => {
        if (!active) return;
        setScreen("login");
        setError(reason instanceof Error ? reason.message : "无法检查初始化状态");
      });
    return () => {
      active = false;
    };
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    try {
      if (screen === "setup") await register({ username, password });
      const result = await login({ username, password });
      setToken(result.token);
      const from = (location.state as { from?: string } | null)?.from;
      navigate(from || "/dashboard", { replace: true });
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "操作失败，请稍后重试");
    } finally {
      setPending(false);
    }
  }

  const isSetup = screen === "setup";

  return (
    <main className="flex min-h-svh items-center justify-center p-4 sm:p-6">
      <Card className="w-full max-w-3xl overflow-hidden border-border/70 shadow-xl shadow-foreground/5">
        <div className="grid md:grid-cols-[minmax(18rem,1fr)_minmax(22rem,1.1fr)]">
          <section className="relative flex min-h-64 flex-col overflow-hidden border-b bg-sidebar p-7 md:min-h-[30rem] md:border-r md:border-b-0">
            <div className="absolute inset-0 opacity-50 [background-image:linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [background-size:34px_34px]" />
            <div className="relative flex items-center gap-3">
              <BrandLogo className="size-10 shadow-md shadow-primary/15" eager />
              <div>
                <p className="font-mono text-sm font-semibold tracking-[0.16em]">CADDYPILOT</p>
              </div>
            </div>
            <div className="relative my-auto max-w-sm py-8">
              <h1 className="text-3xl font-semibold leading-tight tracking-tight">
                可视化管理 Caddy，
                <br />
                从站点到证书一次完成。
              </h1>
              <p className="mt-4 text-sm leading-6 text-muted-foreground">
                CaddyPilot 集中管理反向代理、静态站点、证书、DNS Provider、配置发布与回滚。
              </p>
              <div className="mt-6 flex flex-wrap gap-2 text-xs text-muted-foreground">
                <span className="rounded-full border bg-background/50 px-3 py-1">配置预览</span>
                <span className="rounded-full border bg-background/50 px-3 py-1">安全发布</span>
                <span className="rounded-full border bg-background/50 px-3 py-1">版本回滚</span>
              </div>
            </div>
          </section>

          <section>
            <CardHeader className="gap-1 p-6 pb-4">
              <CardTitle className="text-2xl">
                {screen === "loading"
                  ? "正在检查实例"
                  : isSetup
                    ? "初始化 CaddyPilot"
                    : "进入控制台"}
              </CardTitle>
              <CardDescription>
                {screen === "loading"
                  ? "正在确认此实例是否已创建管理员。"
                  : isSetup
                    ? "当前没有管理员。创建此实例唯一的管理员账户。"
                    : "使用已有的管理员账户继续。"}
              </CardDescription>
            </CardHeader>
            <CardContent className="p-6 pt-0">
              {screen === "loading" ? (
                <div className="flex items-center gap-3 py-10 text-sm text-muted-foreground">
                  <Spinner />
                  正在读取初始化状态…
                </div>
              ) : (
                <form className="flex flex-col gap-5" onSubmit={submit}>
                  {error ? (
                    <Alert variant="destructive">
                      <AlertCircle />
                      <AlertTitle>无法继续</AlertTitle>
                      <AlertDescription>{error}</AlertDescription>
                    </Alert>
                  ) : null}
                  <FieldGroup>
                    <Field>
                      <FieldLabel htmlFor="username">用户名</FieldLabel>
                      <Input
                        id="username"
                        autoComplete="username"
                        value={username}
                        onChange={(event) => setUsername(event.target.value)}
                        required
                        autoFocus
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="password">密码</FieldLabel>
                      <Input
                        id="password"
                        type="password"
                        autoComplete={isSetup ? "new-password" : "current-password"}
                        minLength={6}
                        value={password}
                        onChange={(event) => setPassword(event.target.value)}
                        required
                      />
                      <FieldDescription>至少 6 个字符。</FieldDescription>
                    </Field>
                  </FieldGroup>
                  <Button type="submit" disabled={pending}>
                    {pending ? <Spinner data-icon="inline-start" /> : null}
                    {pending
                      ? isSetup
                        ? "正在初始化"
                        : "正在验证"
                      : isSetup
                        ? "创建管理员并登录"
                        : "登录"}
                    {!pending ? <ArrowRight data-icon="inline-end" /> : null}
                  </Button>
                </form>
              )}
            </CardContent>
          </section>
        </div>
      </Card>
    </main>
  );
}
