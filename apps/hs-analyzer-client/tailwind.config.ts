import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        app: {
          bg: "hsl(var(--color-bg-app) / <alpha-value>)",
          panel: "hsl(var(--color-bg-panel) / <alpha-value>)",
          panelStrong: "hsl(var(--color-bg-panel-strong) / <alpha-value>)",
          border: "hsl(var(--color-border-soft) / <alpha-value>)",
          text: "hsl(var(--color-text-primary) / <alpha-value>)",
          muted: "hsl(var(--color-text-muted) / <alpha-value>)",
          accent: "hsl(var(--color-accent-primary) / <alpha-value>)",
          success: "hsl(var(--color-accent-success) / <alpha-value>)",
          danger: "hsl(var(--color-accent-danger) / <alpha-value>)",
          warning: "hsl(var(--color-accent-warning) / <alpha-value>)"
        }
      },
      borderRadius: {
        sm: "var(--radius-sm)",
        md: "var(--radius-md)",
        lg: "var(--radius-lg)",
        xl: "var(--radius-xl)"
      },
      spacing: {
        1: "var(--space-1)",
        2: "var(--space-2)",
        3: "var(--space-3)",
        4: "var(--space-4)",
        5: "var(--space-5)",
        6: "var(--space-6)"
      },
      boxShadow: {
        panel: "var(--shadow-panel)",
        glow: "var(--shadow-glow)"
      },
      transitionDuration: {
        fast: "var(--motion-fast)",
        base: "var(--motion-base)",
        slow: "var(--motion-slow)"
      }
    }
  },
  plugins: []
};

export default config;
