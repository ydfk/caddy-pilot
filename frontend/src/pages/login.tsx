import { useEffect, useState, type FormEvent } from "react";
import { AlertCircle, ArrowRight, Braces, LockKeyhole, Route } from "lucide-react";
import { useLocation, useNavigate } from "react-router-dom";

import { getSetupStatus, login, register } from "@/api/auth";
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
    <main className="grid min-h-svh lg:grid-cols-[minmax(0,1.15fr)_minmax(26rem,0.85fr)]">
      <section className="relative hidden overflow-hidden border-r border-border/60 bg-sidebar p-10 lg:flex lg:flex-col">
        <div className="absolute inset-0 opacity-70 [background-image:linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [background-size:42px_42px]" />
        <div className="relative flex items-center gap-3">
          <div className="flex size-11 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
            <Route className="size-5" />
          </div>
          <div>
            <p className="font-mono text-sm font-semibold tracking-[0.18em]">CADDYPILOT</p>
            <p className="text-xs text-muted-foreground">LOCAL CONTROL PLANE</p>
          </div>
        </div>
        <div className="relative my-auto max-w-xl">
          <p className="font-mono text-xs tracking-[0.2em] text-muted-foreground">
            {isSetup ? "FIRST RUN / LOCAL SETUP" : "ROUTE / INSPECT / LOAD"}
          </p>
          <h1 className="mt-5 text-balance text-5xl font-semibold leading-[1.05] tracking-[-0.04em]">
            {isSetup ? "先创建唯一管理员，再开始接管路由。" : "让每一次代理变更，都可见、可退。"}
          </h1>
          <div className="mt-10 grid grid-cols-2 gap-3 text-sm">
            <div className="rounded-xl border bg-background/70 p-4 backdrop-blur">
              <Braces className="mb-8 size-5 text-primary" />
              <p className="font-medium">{isSetup ? "单用户模式" : "JSON 预览"}</p>
              <p className="mt-1 text-muted-foreground">
                {isSetup ? "仅创建一个本地管理员" : "发布前检查完整配置"}
              </p>
            </div>
            <div className="rounded-xl border bg-background/70 p-4 backdrop-blur">
              <LockKeyhole className="mb-8 size-5 text-primary" />
              <p className="font-medium">{isSetup ? "本地初始化" : "入口保护"}</p>
              <p className="mt-1 text-muted-foreground">
                {isSetup ? "完成后自动关闭注册入口" : "发布与回滚始终保留管理面"}
              </p>
            </div>
          </div>
        </div>
        <p className="relative font-mono text-xs text-muted-foreground">ADMIN · 127.0.0.1:2019</p>
      </section>

      <section className="flex items-center justify-center p-5 sm:p-10">
        <Card className="w-full max-w-md border-border/70 shadow-xl shadow-foreground/5">
          <CardHeader className="gap-2">
            <div className="mb-3 flex items-center gap-3 lg:hidden">
              <div className="flex size-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <Route className="size-4" />
              </div>
              <span className="font-mono text-sm font-semibold tracking-[0.16em]">CADDYPILOT</span>
            </div>
            <CardTitle className="text-2xl">
              {screen === "loading" ? "正在检查实例" : isSetup ? "初始化 CaddyPilot" : "进入控制台"}
            </CardTitle>
            <CardDescription>
              {screen === "loading"
                ? "正在确认此实例是否已创建管理员。"
                : isSetup
                  ? "当前没有管理员。创建此实例唯一的本地管理员账户。"
                  : "使用已有的本地管理员账户继续。"}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {screen === "loading" ? (
              <div className="flex items-center gap-3 py-8 text-sm text-muted-foreground">
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
        </Card>
      </section>
    </main>
  );
}
