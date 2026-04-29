export function formatWinrate(value: number): string {
  return `${Math.round(value * 100)}%`;
}

export function formatRecord(wins: number, losses: number): string {
  return `${wins}-${losses}`;
}

export function getMetricTone(value: number): "success" | "warning" | "danger" {
  if (value >= 0.6) {
    return "success";
  }
  if (value >= 0.45) {
    return "warning";
  }
  return "danger";
}
