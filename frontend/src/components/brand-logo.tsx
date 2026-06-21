import { cn } from "@/lib/utils";

type BrandLogoProps = {
  className?: string;
  eager?: boolean;
};

export function BrandLogo({ className, eager = false }: BrandLogoProps) {
  return (
    <img
      src="/caddypilot-logo.png"
      alt="CaddyPilot"
      className={cn("rounded-lg object-cover", className)}
      loading={eager ? "eager" : "lazy"}
      fetchPriority={eager ? "high" : "auto"}
    />
  );
}
