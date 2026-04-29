import type { HTMLAttributes, PropsWithChildren } from "react";
import { cn } from "@/shared/lib/cn";

export function Panel({ className, children, ...props }: PropsWithChildren<HTMLAttributes<HTMLDivElement>>) {
  return (
    <div
      className={cn(
        "rounded-xl border border-app-border/70 bg-[linear-gradient(180deg,rgba(21,31,54,0.94),rgba(12,18,31,0.92))] p-5 shadow-panel backdrop-blur-md",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}
