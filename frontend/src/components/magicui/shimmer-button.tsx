import * as React from "react";

import { cn } from "@/lib/utils";

type ShimmerButtonProps = React.ComponentProps<"button">;

export function ShimmerButton({
  className,
  children,
  type = "button",
  ...props
}: ShimmerButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        "relative inline-flex items-center justify-center overflow-hidden rounded-md border border-border bg-foreground px-4 py-2 text-sm font-medium text-background shadow-sm transition-colors",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "disabled:cursor-not-allowed disabled:opacity-60",
        "after:pointer-events-none after:absolute after:inset-0 after:-translate-x-full after:bg-gradient-to-r after:from-transparent after:via-white/35 after:to-transparent after:[animation:shimmer_2.2s_infinite]",
        "motion-reduce:after:animate-none disabled:after:hidden",
        className
      )}
      {...props}
    >
      <span className="relative z-10">{children}</span>
    </button>
  );
}
