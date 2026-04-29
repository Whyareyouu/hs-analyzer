import type { HTMLAttributes, PropsWithChildren } from "react";
import { cn } from "@/shared/lib/cn";

type Tone = "info" | "success" | "warning" | "danger";

const toneClass: Record<Tone, string> = {
  info: "border-app-accent/45 bg-app-accent/10 text-app-accent",
  success: "border-app-success/45 bg-app-success/10 text-app-success",
  warning: "border-app-warning/45 bg-app-warning/10 text-app-warning",
  danger: "border-app-danger/45 bg-app-danger/10 text-app-danger",
};

export function Pill({
  tone = "info",
  className,
  children,
  ...props
}: PropsWithChildren<HTMLAttributes<HTMLSpanElement>> & { tone?: Tone }) {
  return (
    <span className={cn("inline-flex items-center rounded-full border px-2 py-1 text-xs font-medium", toneClass[tone], className)} {...props}>
      {children}
    </span>
  );
}
