import type { PropsWithChildren } from "react";
import { ServicesProvider } from "@/shared/api/services-context";

export function AppProviders({ children }: PropsWithChildren) {
  return <ServicesProvider>{children}</ServicesProvider>;
}
