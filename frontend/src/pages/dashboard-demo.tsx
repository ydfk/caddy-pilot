import { useState } from "react";
import {
  ArrowRight,
  CalendarDays,
  ChevronRight,
  CircleAlert,
  CircleCheckBig,
  LayoutTemplate,
  MoreHorizontal,
  Plus,
  Rocket,
  Settings2,
  Sparkles,
  TrendingUp,
} from "lucide-react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  XAxis,
} from "recharts";
import { toast } from "sonner";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge, type BadgeProps } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Calendar } from "@/components/ui/calendar";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";

const metricCards = [
  {
    title: "本周收入",
    value: "¥ 248,900",
    change: "+18.4%",
    caption: "较上周新增 21 个活跃账户",
  },
  {
    title: "待跟进线索",
    value: "164",
    change: "+12",
    caption: "来自站内表单和顾问转介",
  },
  {
    title: "自动化任务",
    value: "27",
    change: "92%",
    caption: "流程执行成功率保持稳定",
  },
  {
    title: "上线准备度",
    value: "81 / 100",
    change: "+9",
    caption: "还差法务确认和知识库补充",
  },
] as const;

const revenueData = [
  { label: "Mon", income: 32, target: 20 },
  { label: "Tue", income: 46, target: 29 },
  { label: "Wed", income: 41, target: 34 },
  { label: "Thu", income: 68, target: 42 },
  { label: "Fri", income: 74, target: 47 },
  { label: "Sat", income: 59, target: 39 },
  { label: "Sun", income: 83, target: 58 },
];

const velocityData = [
  { label: "Growth", completed: 18 },
  { label: "Sales", completed: 14 },
  { label: "Support", completed: 11 },
  { label: "Ops", completed: 9 },
];

const channelData = [
  { channel: "organic", value: 42, fill: "var(--color-organic)" },
  { channel: "outbound", value: 27, fill: "var(--color-outbound)" },
  { channel: "partner", value: 19, fill: "var(--color-partner)" },
  { channel: "existing", value: 12, fill: "var(--color-existing)" },
];

const revenueChartConfig = {
  income: {
    label: "实际收入",
    color: "var(--chart-1)",
  },
  target: {
    label: "目标收入",
    color: "var(--chart-2)",
  },
} satisfies ChartConfig;

const velocityChartConfig = {
  completed: {
    label: "完成事项",
    color: "var(--chart-3)",
  },
} satisfies ChartConfig;

const channelChartConfig = {
  organic: {
    label: "官网转化",
    color: "var(--chart-1)",
  },
  outbound: {
    label: "顾问外呼",
    color: "var(--chart-2)",
  },
  partner: {
    label: "伙伴渠道",
    color: "var(--chart-3)",
  },
  existing: {
    label: "老客扩购",
    color: "var(--chart-4)",
  },
} satisfies ChartConfig;

const readinessItems = [
  { label: "品牌视觉与落地页", value: 92 },
  { label: "客户分层和名单清洗", value: 84 },
  { label: "销售话术与客服预案", value: 76 },
  { label: "法务和定价确认", value: 58 },
] as const;

const pipelineRows = [
  {
    name: "Northwind SaaS",
    owner: "AL",
    stage: "Discovery",
    amount: "¥ 72,000",
    priority: "高优先级",
  },
  {
    name: "Horizon Studio",
    owner: "MS",
    stage: "Proposal",
    amount: "¥ 48,000",
    priority: "待审批",
  },
  {
    name: "Beacon Retail",
    owner: "QY",
    stage: "Pilot",
    amount: "¥ 96,000",
    priority: "推进中",
  },
  {
    name: "Summit Cloud",
    owner: "ZW",
    stage: "Renewal",
    amount: "¥ 31,000",
    priority: "回访中",
  },
] as const;

const activityFeed = [
  {
    title: "客服知识库已同步到最新版本",
    time: "10 分钟前",
  },
  {
    title: "Horizon Studio 完成第二轮报价确认",
    time: "42 分钟前",
  },
  {
    title: "自动化规则“Leads > 48h”触发 7 次",
    time: "今天 08:30",
  },
] as const;

const scheduleItems = [
  { time: "09:30", title: "渠道复盘", badge: "内部" },
  { time: "13:00", title: "销售培训彩排", badge: "协作" },
  { time: "16:30", title: "发布前 QA Check", badge: "发布" },
] as const;

function getPriorityVariant(priority: string): BadgeProps["variant"] {
  if (priority === "高优先级") {
    return "default";
  }

  if (priority === "推进中") {
    return "secondary";
  }

  return "outline";
}

function MetricCard({ title, value, change, caption }: (typeof metricCards)[number]) {
  return (
    <Card className="rounded-[28px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
      <CardHeader className="gap-3 pb-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <CardDescription className="text-sm text-muted-foreground">{title}</CardDescription>
            <CardTitle className="mt-2 text-3xl font-semibold tracking-tight">{value}</CardTitle>
          </div>
          <Badge variant="secondary" className="rounded-full px-3 py-1 text-xs font-medium">
            {change}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">{caption}</p>
      </CardContent>
    </Card>
  );
}

export default function DashboardDemo() {
  const [launchDialogOpen, setLaunchDialogOpen] = useState(false);
  const [campaignChannel, setCampaignChannel] = useState("product-launch");
  const [assistantMode, setAssistantMode] = useState(true);
  const [notifyOwners, setNotifyOwners] = useState(true);
  const [overviewDate, setOverviewDate] = useState<Date | undefined>(new Date("2026-04-02"));
  const [quickAudience, setQuickAudience] = useState("vip");
  const [quickAutomation, setQuickAutomation] = useState(true);

  const handleLaunchSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const projectName = String(formData.get("projectName") ?? "").trim();

    toast.success("发布节奏已创建", {
      description: `${projectName || "未命名项目"} 已进入 ${campaignChannel} 流程。`,
    });
    setLaunchDialogOpen(false);
  };

  const handleQuickCompose = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const campaignName = String(formData.get("campaignName") ?? "").trim();

    toast.success("草稿已保存", {
      description: `${campaignName || "未命名活动"} 已按 ${quickAudience} 受众建档。`,
    });
  };

  return (
    <div className="flex flex-col gap-6">
      <section className="rounded-[32px] border border-border/60 bg-card/85 px-5 py-5 shadow-sm backdrop-blur md:px-7 md:py-6">
        <div className="flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
          <div className="max-w-3xl">
            <Badge
              variant="outline"
              className="rounded-full border-primary/20 bg-primary/5 px-3 py-1 text-xs text-primary"
            >
              Admin Demo
            </Badge>
            <h1 className="mt-4 text-3xl font-semibold tracking-tight md:text-4xl">
              一个更像真实后台的 shadcn/ui 组件总览页
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground md:text-base">
              用统计卡片、图表、表格、表单、日历和弹窗把主要组件串进同一条工作流里，视觉上参考
              shadcn-admin 的轻量控制台风格。
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="rounded-2xl border-border/60 bg-background/80">
                  <Settings2 />
                  页面动作
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-52 rounded-2xl">
                <DropdownMenuItem onClick={() => toast("已创建共享链接")}>
                  复制分享链接
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => toast("已进入演示模式")}>
                  打开演示模式
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => toast("设计 token 已导出")}>
                  导出设计 token
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <Button className="rounded-2xl px-5" onClick={() => setLaunchDialogOpen(true)}>
              <Plus />
              创建发布节奏
            </Button>
          </div>
        </div>
        <div className="mt-6 grid gap-4 md:grid-cols-2 2xl:grid-cols-4">
          {metricCards.map((item) => (
            <MetricCard key={item.title} {...item} />
          ))}
        </div>
      </section>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.7fr)_minmax(320px,1fr)]">
        <Card className="rounded-[32px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
          <Tabs defaultValue="revenue">
            <CardHeader className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <CardTitle className="text-xl">核心趋势面板</CardTitle>
                <CardDescription>按营收、执行效率和渠道构成切换查看。</CardDescription>
              </div>
              <div className="w-full lg:w-auto">
                <TabsList className="grid w-full grid-cols-3 rounded-2xl lg:w-[320px]">
                  <TabsTrigger value="revenue">营收</TabsTrigger>
                  <TabsTrigger value="velocity">效率</TabsTrigger>
                  <TabsTrigger value="channels">渠道</TabsTrigger>
                </TabsList>
              </div>
            </CardHeader>
            <CardContent>
              <TabsContent value="revenue" className="mt-0">
                <ChartContainer config={revenueChartConfig} className="h-[320px] w-full">
                  <AreaChart data={revenueData} margin={{ left: 8, right: 8 }}>
                    <defs>
                      <linearGradient id="income-fill" x1="0" x2="0" y1="0" y2="1">
                        <stop offset="5%" stopColor="var(--color-income)" stopOpacity={0.35} />
                        <stop offset="95%" stopColor="var(--color-income)" stopOpacity={0.02} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid vertical={false} />
                    <XAxis axisLine={false} dataKey="label" tickLine={false} />
                    <ChartTooltip content={<ChartTooltipContent />} cursor={false} />
                    <Area
                      dataKey="target"
                      fill="none"
                      stroke="var(--color-target)"
                      strokeDasharray="6 6"
                      strokeWidth={2}
                      type="monotone"
                    />
                    <Area
                      dataKey="income"
                      fill="url(#income-fill)"
                      stroke="var(--color-income)"
                      strokeWidth={2.5}
                      type="monotone"
                    />
                  </AreaChart>
                </ChartContainer>
              </TabsContent>
              <TabsContent value="velocity" className="mt-0">
                <ChartContainer config={velocityChartConfig} className="h-[320px] w-full">
                  <BarChart data={velocityData} margin={{ left: 8, right: 8 }}>
                    <CartesianGrid vertical={false} />
                    <XAxis axisLine={false} dataKey="label" tickLine={false} />
                    <ChartTooltip content={<ChartTooltipContent />} cursor={false} />
                    <Bar
                      dataKey="completed"
                      fill="var(--color-completed)"
                      radius={[12, 12, 4, 4]}
                    />
                  </BarChart>
                </ChartContainer>
              </TabsContent>
              <TabsContent value="channels" className="mt-0">
                <ChartContainer config={channelChartConfig} className="h-[320px] w-full">
                  <PieChart>
                    <ChartTooltip
                      content={<ChartTooltipContent hideLabel nameKey="channel" />}
                      cursor={false}
                    />
                    <Pie
                      data={channelData}
                      dataKey="value"
                      innerRadius={76}
                      outerRadius={110}
                      paddingAngle={3}
                    >
                      {channelData.map((item) => (
                        <Cell key={item.channel} fill={item.fill} />
                      ))}
                    </Pie>
                    <ChartLegend
                      content={<ChartLegendContent nameKey="channel" />}
                      verticalAlign="bottom"
                    />
                  </PieChart>
                </ChartContainer>
              </TabsContent>
            </CardContent>
          </Tabs>
        </Card>

        <div className="grid gap-6">
          <Card className="rounded-[32px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
            <CardHeader>
              <div className="flex items-start justify-between gap-3">
                <div>
                  <CardTitle className="text-xl">发布准备度</CardTitle>
                  <CardDescription>适合展示 Progress、Badge 和状态说明。</CardDescription>
                </div>
                <Badge variant="secondary" className="rounded-full px-3 py-1">
                  81%
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="flex flex-col gap-5">
              {readinessItems.map((item) => (
                <div key={item.label} className="flex flex-col gap-2">
                  <div className="flex items-center justify-between gap-4 text-sm">
                    <span className="font-medium">{item.label}</span>
                    <span className="text-muted-foreground">{item.value}%</span>
                  </div>
                  <Progress value={item.value} className="h-2.5" />
                </div>
              ))}
            </CardContent>
          </Card>

          <Card className="rounded-[32px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
            <CardHeader>
              <CardTitle className="text-xl">提醒与动作</CardTitle>
              <CardDescription>典型后台里常见的消息提示卡片。</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div className="rounded-3xl border border-amber-500/20 bg-amber-500/8 p-4">
                <div className="flex items-start gap-3">
                  <CircleAlert className="mt-0.5 size-5 text-amber-600" />
                  <div className="flex flex-col gap-1">
                    <p className="font-medium">有 2 项发布前检查尚未完成</p>
                    <p className="text-sm text-muted-foreground">
                      建议今天 18:00 前完成法务确认，避免排期继续顺延。
                    </p>
                  </div>
                </div>
              </div>
              <div className="rounded-3xl border border-emerald-500/20 bg-emerald-500/8 p-4">
                <div className="flex items-start gap-3">
                  <CircleCheckBig className="mt-0.5 size-5 text-emerald-600" />
                  <div className="flex flex-col gap-1">
                    <p className="font-medium">数据回传链路已恢复稳定</p>
                    <p className="text-sm text-muted-foreground">
                      昨晚修复后已连续 12 小时无丢单告警。
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
            <CardFooter className="flex items-center justify-between">
              <Button variant="ghost" className="rounded-2xl">
                查看运行日志
              </Button>
              <Button variant="outline" className="rounded-2xl border-border/60">
                打开规则面板
              </Button>
            </CardFooter>
          </Card>
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(320px,0.95fr)]">
        <Card className="rounded-[32px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
          <CardHeader className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <CardTitle className="text-xl">重点机会表</CardTitle>
              <CardDescription>表格、头像、Badge 和下拉菜单的组合示例。</CardDescription>
            </div>
            <Button variant="outline" className="rounded-2xl border-border/60">
              <LayoutTemplate />
              切换视图
            </Button>
          </CardHeader>
          <CardContent>
            <div className="overflow-hidden rounded-3xl border border-border/60">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/40">
                    <TableHead>客户</TableHead>
                    <TableHead>阶段</TableHead>
                    <TableHead>负责人</TableHead>
                    <TableHead>金额</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pipelineRows.map((row) => (
                    <TableRow key={row.name}>
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <Avatar className="size-10 rounded-2xl">
                            <AvatarFallback className="rounded-2xl bg-primary/10 text-primary">
                              {row.owner}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">{row.name}</p>
                            <p className="text-sm text-muted-foreground">本周新增推进动作</p>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>{row.stage}</TableCell>
                      <TableCell>{row.owner}</TableCell>
                      <TableCell className="font-medium">{row.amount}</TableCell>
                      <TableCell>
                        <Badge
                          variant={getPriorityVariant(row.priority)}
                          className="rounded-full px-3 py-1"
                        >
                          {row.priority}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="rounded-xl">
                              <MoreHorizontal />
                              <span className="sr-only">更多操作</span>
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="rounded-2xl">
                            <DropdownMenuItem onClick={() => toast("已打开详情")}>
                              查看详情
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => toast("已指派跟进人")}>
                              指派负责人
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => toast("已加入本周复盘")}>
                              加入复盘
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        <div className="grid gap-6">
          <Card className="rounded-[32px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
            <CardHeader>
              <CardTitle className="text-xl">快速创建卡片</CardTitle>
              <CardDescription>Input、Select、Switch、Checkbox、Textarea 的组合。</CardDescription>
            </CardHeader>
            <CardContent>
              <form className="flex flex-col gap-4" onSubmit={handleQuickCompose}>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="campaign-name">活动名称</Label>
                  <Input
                    id="campaign-name"
                    name="campaignName"
                    placeholder="Q2 产品沟通节奏"
                    className="rounded-2xl"
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label>受众分层</Label>
                  <Select value={quickAudience} onValueChange={setQuickAudience}>
                    <SelectTrigger className="rounded-2xl">
                      <SelectValue placeholder="选择受众" />
                    </SelectTrigger>
                    <SelectContent className="rounded-2xl">
                      <SelectItem value="vip">核心客户</SelectItem>
                      <SelectItem value="trial">试用客户</SelectItem>
                      <SelectItem value="partner">伙伴渠道</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="brief">投放备注</Label>
                  <Textarea
                    id="brief"
                    name="brief"
                    placeholder="补充说明目标、节奏和注意事项..."
                    className="min-h-28 rounded-2xl"
                  />
                </div>
                <div className="flex items-center justify-between rounded-2xl border border-border/60 bg-muted/30 px-4 py-3">
                  <div>
                    <p className="font-medium">自动生成执行清单</p>
                    <p className="text-sm text-muted-foreground">创建后同步给销售和客服。</p>
                  </div>
                  <Switch checked={quickAutomation} onCheckedChange={setQuickAutomation} />
                </div>
                <label className="flex items-center gap-3 text-sm text-muted-foreground">
                  <Checkbox defaultChecked />
                  同时创建复盘提醒
                </label>
                <Button type="submit" className="w-full rounded-2xl">
                  保存草稿
                  <ArrowRight />
                </Button>
              </form>
            </CardContent>
          </Card>

          <Card className="rounded-[32px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
            <CardHeader>
              <CardTitle className="text-xl">日程总览</CardTitle>
              <CardDescription>Calendar 与轻量日程列表组合示例。</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-5">
              <div className="rounded-3xl border border-border/60 bg-background/70 p-3">
                <Calendar
                  mode="single"
                  selected={overviewDate}
                  onSelect={setOverviewDate}
                  className="mx-auto"
                />
              </div>
              <div className="flex flex-col gap-3">
                {scheduleItems.map((item) => (
                  <div
                    key={item.time}
                    className="flex items-center justify-between gap-3 rounded-2xl border border-border/60 px-4 py-3"
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex size-10 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                        <CalendarDays className="size-4" />
                      </div>
                      <div>
                        <p className="font-medium">{item.title}</p>
                        <p className="text-sm text-muted-foreground">{item.time}</p>
                      </div>
                    </div>
                    <Badge variant="outline" className="rounded-full px-3 py-1">
                      {item.badge}
                    </Badge>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <Card className="rounded-[32px] border-border/60 bg-card/90 shadow-sm backdrop-blur">
        <CardHeader className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <CardTitle className="text-xl">团队活动</CardTitle>
            <CardDescription>适合放公告、事件流和团队状态的小组件区域。</CardDescription>
          </div>
          <Button variant="ghost" className="rounded-2xl">
            查看全部
            <ChevronRight />
          </Button>
        </CardHeader>
        <CardContent className="grid gap-4 lg:grid-cols-3">
          {activityFeed.map((item, index) => (
            <div
              key={item.title}
              className="rounded-3xl border border-border/60 bg-background/70 p-5"
            >
              <div className="flex items-center gap-3">
                <div className="flex size-11 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                  {index === 0 ? (
                    <Sparkles className="size-5" />
                  ) : (
                    <TrendingUp className="size-5" />
                  )}
                </div>
                <div>
                  <p className="font-medium">{item.title}</p>
                  <p className="text-sm text-muted-foreground">{item.time}</p>
                </div>
              </div>
              <Separator className="my-4" />
              <div className="flex items-center justify-between">
                <Badge variant="secondary" className="rounded-full px-3 py-1">
                  Live Feed
                </Badge>
                <Button variant="outline" className="rounded-2xl border-border/60">
                  查看
                </Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Dialog open={launchDialogOpen} onOpenChange={setLaunchDialogOpen}>
        <DialogContent className="rounded-[28px] border-border/60 p-0 sm:max-w-[560px]">
          <form onSubmit={handleLaunchSubmit}>
            <DialogHeader className="border-b border-border/60 px-6 py-5">
              <DialogTitle className="flex items-center gap-2 text-xl">
                <Rocket className="size-5 text-primary" />
                创建新的发布节奏
              </DialogTitle>
              <DialogDescription>
                这个弹窗把按钮、输入框、选择器、开关和多选框放进一个完整动作里。
              </DialogDescription>
            </DialogHeader>
            <div className="flex flex-col gap-4 px-6 py-5">
              <div className="flex flex-col gap-2">
                <Label htmlFor="project-name">项目名称</Label>
                <Input
                  id="project-name"
                  name="projectName"
                  placeholder="2026 春季增长计划"
                  className="rounded-2xl"
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label>流程模板</Label>
                <Select value={campaignChannel} onValueChange={setCampaignChannel}>
                  <SelectTrigger className="rounded-2xl">
                    <SelectValue placeholder="选择模板" />
                  </SelectTrigger>
                  <SelectContent className="rounded-2xl">
                    <SelectItem value="product-launch">产品发布</SelectItem>
                    <SelectItem value="sales-nurture">销售培育</SelectItem>
                    <SelectItem value="partner-revival">伙伴激活</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="launch-summary">项目说明</Label>
                <Textarea
                  id="launch-summary"
                  name="launchSummary"
                  placeholder="说明目标受众、节奏节点和联动角色..."
                  className="min-h-28 rounded-2xl"
                />
              </div>
              <div className="rounded-3xl border border-border/60 bg-muted/30 p-4">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="font-medium">启用 AI 助手草拟执行卡</p>
                    <p className="text-sm text-muted-foreground">
                      自动生成待办、话术建议和复盘提醒。
                    </p>
                  </div>
                  <Switch checked={assistantMode} onCheckedChange={setAssistantMode} />
                </div>
                <label className="mt-4 flex items-center gap-3 text-sm text-muted-foreground">
                  <Checkbox
                    checked={notifyOwners}
                    onCheckedChange={(checked) => setNotifyOwners(checked === true)}
                  />
                  创建完成后通知相关负责人
                </label>
              </div>
            </div>
            <DialogFooter className="border-t border-border/60 px-6 py-5">
              <Button
                type="button"
                variant="ghost"
                className="rounded-2xl"
                onClick={() => setLaunchDialogOpen(false)}
              >
                取消
              </Button>
              <Button type="submit" className="rounded-2xl">
                <Rocket />
                开始创建
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
