import type { LucideIcon } from "lucide-react";
import {
  Activity,
  BellRing,
  CalendarDays,
  CircleDotDashed,
  MessagesSquare,
  PanelsLeftBottom,
  FormInput,
  LayoutDashboard,
  MousePointerClick,
  PanelsTopLeft,
  SlidersHorizontal,
  Sparkles,
  SquareCheckBig,
  SquareStack,
  TableProperties,
  ToggleLeft,
  Users,
} from "lucide-react";

export type NavigationItem = {
  title: string;
  url: string;
  icon: LucideIcon;
};

export type NavigationGroup = {
  label: string;
  items: NavigationItem[];
};

export const navigationGroups: NavigationGroup[] = [
  {
    label: "控制台",
    items: [
      {
        title: "总览",
        url: "/",
        icon: LayoutDashboard,
      },
      {
        title: "用户管理",
        url: "/user-management",
        icon: Users,
      },
    ],
  },
  {
    label: "组件示例",
    items: [
      {
        title: "按钮",
        url: "/components/button",
        icon: MousePointerClick,
      },
      {
        title: "输入组件",
        url: "/components/input",
        icon: FormInput,
      },
      {
        title: "多选框",
        url: "/components/checkbox",
        icon: SquareCheckBig,
      },
      {
        title: "单选组",
        url: "/components/radio-group",
        icon: CircleDotDashed,
      },
      {
        title: "开关",
        url: "/components/switch",
        icon: ToggleLeft,
      },
      {
        title: "滑块",
        url: "/components/slider",
        icon: SlidersHorizontal,
      },
      {
        title: "选项卡",
        url: "/components/tabs",
        icon: PanelsTopLeft,
      },
      {
        title: "进度条",
        url: "/components/progress",
        icon: Activity,
      },
      {
        title: "卡片",
        url: "/components/card",
        icon: SquareStack,
      },
      {
        title: "对话框",
        url: "/components/dialog",
        icon: PanelsLeftBottom,
      },
      {
        title: "表格",
        url: "/components/table",
        icon: TableProperties,
      },
      {
        title: "通知与菜单",
        url: "/components/dropdown",
        icon: BellRing,
      },
      {
        title: "表单",
        url: "/components/form",
        icon: Sparkles,
      },
      {
        title: "选择器",
        url: "/components/select",
        icon: FormInput,
      },
      {
        title: "提示反馈",
        url: "/components/sonner",
        icon: MessagesSquare,
      },
      {
        title: "日历",
        url: "/components/calendar",
        icon: CalendarDays,
      },
      {
        title: "Magic UI",
        url: "/components/magicui",
        icon: Sparkles,
      },
    ],
  },
];

const routeMeta = new Map(
  navigationGroups.flatMap((group) =>
    group.items.map((item) => [
      item.url,
      {
        section: group.label,
        title: item.title,
      },
    ])
  )
);

export function getRouteMeta(pathname: string) {
  return (
    routeMeta.get(pathname) ?? {
      section: "控制台",
      title: "组件展示",
    }
  );
}
