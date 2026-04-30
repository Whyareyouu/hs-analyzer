#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: ./scripts/run-app.sh [options]

Options:
  --logs-path PATH        Path to Hearthstone logs or a local logs fixture.
  --cards-db-path PATH    Path to cards.json (default: data/cards.json).
  --backend-addr ADDR     Backend listen address (default: 127.0.0.1:8080).
  --live-api-addr ADDR    Live API listen address (default: 127.0.0.1:8765).
  --database-url URL      PostgreSQL URL for backend storage.
  --rebuild-go            Force rebuild of Go binaries before launch.
  --rebuild-client        Force rebuild of the Electron client before launch.
  -h, --help              Show this help message.
EOF
}

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
client_dir="$repo_root/apps/hs-analyzer-client"
bin_dir="$repo_root/bin"
run_dir="$repo_root/.tmp/run"
output_dir="$repo_root/output"
cards_db_path="data/cards.json"
logs_path=""
backend_addr="127.0.0.1:8080"
live_api_addr="127.0.0.1:8765"
database_url="${HS_ANALYZER_DATABASE_URL:-}"
rebuild_go=0
rebuild_client=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --logs-path)
      logs_path="${2:-}"
      shift 2
      ;;
    --cards-db-path)
      cards_db_path="${2:-}"
      shift 2
      ;;
    --backend-addr)
      backend_addr="${2:-}"
      shift 2
      ;;
    --live-api-addr)
      live_api_addr="${2:-}"
      shift 2
      ;;
    --database-url)
      database_url="${2:-}"
      shift 2
      ;;
    --rebuild-go)
      rebuild_go=1
      shift
      ;;
    --rebuild-client)
      rebuild_client=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

resolve_repo_path() {
  local value="$1"
  if [[ -z "$value" ]]; then
    printf '%s\n' ""
    return
  fi

  if [[ "$value" = /* || "$value" =~ ^[A-Za-z]:[\\/].* ]]; then
    printf '%s\n' "$value"
    return
  fi

  printf '%s\n' "$repo_root/$value"
}

resolve_auto_logs_path() {
  local candidates=()
  local local_app_data="${LOCALAPPDATA:-}"
  if [[ -n "$local_app_data" ]]; then
    candidates+=("$local_app_data/Blizzard/Hearthstone/Logs")
  fi

  local user_profile="${USERPROFILE:-}"
  if [[ -n "$user_profile" ]]; then
    candidates+=("$user_profile/AppData/Local/Blizzard/Hearthstone/Logs")
  fi

  local root=""
  for root in /mnt/[A-Za-z] /[A-Za-z]; do
    [[ -d "$root" ]] || continue
    candidates+=(
      "$root/Hearthstone/Logs"
      "$root/Games/Hearthstone/Logs"
      "$root/Blizzard/Hearthstone/Logs"
    )
  done

  if command -v powershell.exe >/dev/null 2>&1; then
    while IFS= read -r root; do
      [[ -z "$root" ]] && continue
      candidates+=(
        "$root/Hearthstone/Logs"
        "$root/Games/Hearthstone/Logs"
        "$root/Blizzard/Hearthstone/Logs"
      )
    done < <(powershell.exe -NoProfile -Command "Get-PSDrive -PSProvider FileSystem | ForEach-Object { \$_.Root }" 2>/dev/null | tr -d '\r')
  fi

  local candidate=""
  local seen="|"
  for candidate in "${candidates[@]}"; do
    [[ -z "$candidate" ]] && continue
    if [[ "$seen" == *"|$candidate|"* ]]; then
      continue
    fi
    seen="$seen$candidate|"
    local resolved
    resolved="$(resolve_logs_candidate "$candidate")"
    if [[ -n "$resolved" ]]; then
      printf '%s\n' "$resolved"
      return
    fi
  done

  printf '%s\n' ""
}

resolve_logs_candidate() {
  local logs_root="$1"
  if [[ ! -d "$logs_root" ]]; then
    printf '%s\n' ""
    return
  fi

  if [[ -f "$logs_root/Power.log" ]]; then
    printf '%s\n' "$logs_root"
    return
  fi

  local latest_dir=""
  local latest_mtime=0
  local candidate=""
  while IFS= read -r candidate; do
    [[ -z "$candidate" ]] && continue
    if [[ ! -f "$candidate/Power.log" ]]; then
      continue
    fi
    local mtime
    mtime=$(stat -c %Y "$candidate" 2>/dev/null || stat -f %m "$candidate" 2>/dev/null || printf '0')
    if (( mtime >= latest_mtime )); then
      latest_mtime=$mtime
      latest_dir="$candidate"
    fi
  done < <(find "$logs_root" -mindepth 1 -maxdepth 1 -type d 2>/dev/null)

  printf '%s\n' "$latest_dir"
}

resolve_deck_catalog_logs_path() {
  local value="$1"
  if [[ -z "$value" ]]; then
    printf '%s\n' ""
    return
  fi
  if [[ -d "$value" ]]; then
    printf '%s\n' "$value"
    return
  fi
  dirname "$value"
}

assert_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "Required command '$name' was not found in PATH." >&2
    exit 1
  fi
}

format_database_url_for_display() {
  local value="$1"
  if [[ -z "$value" ]]; then
    printf '%s\n' "(not set)"
    return
  fi
  printf '%s\n' "$value" | sed -E 's#(postgres(ql)?://[^:/@]+):([^@]+)@#\1:xxxxx@#'
}

test_database_tcp_endpoint() {
  local value="$1"
  [[ -n "$value" ]] || return 0

  local stripped="${value#*://}"
  stripped="${stripped%%/*}"
  stripped="${stripped##*@}"
  local host="${stripped%%:*}"
  local port="${stripped##*:}"
  if [[ "$port" == "$host" ]]; then
    port="5432"
  fi
  [[ -n "$host" ]] || return 0

  if command -v nc >/dev/null 2>&1; then
    if ! nc -z "$host" "$port" >/dev/null 2>&1; then
      echo "PostgreSQL is not reachable at $host:$port. Start PostgreSQL or update HS_ANALYZER_DATABASE_URL in .env." >&2
      exit 1
    fi
  fi
}

wait_http_endpoint() {
  local url="$1"
  local name="$2"
  local timeout_seconds="${3:-15}"
  local deadline=$((SECONDS + timeout_seconds))

  while (( SECONDS < deadline )); do
    if command -v curl >/dev/null 2>&1; then
      local status
      status="$(curl -s -o /dev/null -w '%{http_code}' "$url" || printf '000')"
      if [[ "$status" =~ ^[234][0-9][0-9]$ ]]; then
        return 0
      fi
    elif command -v powershell.exe >/dev/null 2>&1; then
      if powershell.exe -NoProfile -Command "try { \$r = Invoke-WebRequest -UseBasicParsing '$url' -TimeoutSec 2; if (\$r.StatusCode -ge 200 -and \$r.StatusCode -lt 500) { exit 0 } else { exit 1 } } catch { exit 1 }" >/dev/null 2>&1; then
        return 0
      fi
    fi
    sleep 0.25
  done

  echo "$name did not become ready at $url within ${timeout_seconds}s." >&2
  exit 1
}

load_dotenv() {
  local file="$1"
  [[ -f "$file" ]] || return 0

  local line name value
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -z "$line" || "${line:0:1}" == "#" ]] && continue
    [[ "$line" == *"="* ]] || continue

    name="${line%%=*}"
    value="${line#*=}"
    name="${name#"${name%%[![:space:]]*}"}"
    name="${name%"${name##*[![:space:]]}"}"
    value="${value#"${value%%[![:space:]]*}"}"
    value="${value%"${value##*[![:space:]]}"}"
    [[ -z "$name" ]] && continue

    if [[ "${value:0:1}" == "\"" && "${value: -1}" == "\"" ]] || [[ "${value:0:1}" == "'" && "${value: -1}" == "'" ]]; then
      value="${value:1:${#value}-2}"
    fi
    if [[ -z "${!name:-}" ]]; then
      export "$name=$value"
    fi
  done < "$file"
}

cleanup() {
  local code=$?
  if [[ -n "${client_pid:-}" ]] && kill -0 "$client_pid" 2>/dev/null; then
    kill "$client_pid" 2>/dev/null || true
  fi
  if [[ -n "${monitor_pid:-}" ]] && kill -0 "$monitor_pid" 2>/dev/null; then
    kill "$monitor_pid" 2>/dev/null || true
  fi
  if [[ -n "${backend_pid:-}" ]] && kill -0 "$backend_pid" 2>/dev/null; then
    kill "$backend_pid" 2>/dev/null || true
  fi
  exit "$code"
}

trap cleanup EXIT INT TERM

assert_command go
assert_command npm

if [[ ! -d "$client_dir" ]]; then
  echo "Electron client directory was not found: $client_dir" >&2
  exit 1
fi

load_dotenv "$repo_root/.env"
if [[ -z "$database_url" ]]; then
  database_url="${HS_ANALYZER_DATABASE_URL:-}"
fi

if [[ -z "$database_url" ]]; then
  echo "Database URL is not configured. Set HS_ANALYZER_DATABASE_URL in .env or pass --database-url." >&2
  exit 1
fi
test_database_tcp_endpoint "$database_url"

if [[ ! -d "$client_dir/node_modules" ]]; then
  echo "Client dependencies are missing. Run 'npm install' in $client_dir first." >&2
  exit 1
fi

cards_db_full="$(resolve_repo_path "$cards_db_path")"
if [[ ! -f "$cards_db_full" ]]; then
  echo "Cards database was not found: $cards_db_full" >&2
  exit 1
fi

logs_full="$(resolve_repo_path "$logs_path")"
if [[ -z "$logs_full" ]]; then
  logs_full="$(resolve_auto_logs_path)"
fi
if [[ -n "$logs_full" && ! -e "$logs_full" ]]; then
  echo "Logs path was not found: $logs_full" >&2
  exit 1
fi

mkdir -p "$bin_dir" "$run_dir" "$output_dir" "$repo_root/.tmp/gocache"
export GOCACHE="$repo_root/.tmp/gocache"

binary_suffix=""
case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*)
    binary_suffix=".exe"
    ;;
esac

backend_binary="$bin_dir/hs-analyzer-be${binary_suffix}"
cli_binary="$bin_dir/hs-analyzer-cli${binary_suffix}"
dist_entry="$client_dir/dist/index.html"
electron_entry="$client_dir/dist-electron/main/index.js"

if [[ $rebuild_go -eq 1 || ! -f "$backend_binary" ]]; then
  echo "Building backend binary..."
  (cd "$repo_root" && go build -o "$backend_binary" ./apps/hs-analyzer-be/cmd/hs-analyzer-be)
fi

if [[ $rebuild_go -eq 1 || ! -f "$cli_binary" ]]; then
  echo "Building CLI binary..."
  (cd "$repo_root" && go build -o "$cli_binary" ./apps/hs-analyzer-cli/cmd/hsanalyzer)
fi

if [[ $rebuild_client -eq 1 || ! -f "$dist_entry" || ! -f "$electron_entry" ]]; then
  echo "Building Electron client..."
  (cd "$client_dir" && npm run build:app)
fi

echo "Starting backend..."
backend_logs_path="$(resolve_deck_catalog_logs_path "$logs_full")"
(
  cd "$repo_root"
  if [[ -n "$backend_logs_path" ]]; then
    export HS_ANALYZER_LOGS_PATH="$backend_logs_path"
    export HS_ANALYZER_CARDS_DB_PATH="$cards_db_full"
  fi
  if [[ -n "$database_url" ]]; then
    export HS_ANALYZER_DATABASE_URL="$database_url"
  fi
  "$backend_binary" >"$run_dir/backend.stdout.log" 2>"$run_dir/backend.stderr.log"
) &
backend_pid=$!
sleep 2
if ! kill -0 "$backend_pid" 2>/dev/null; then
  echo "Backend exited early. Check logs in $run_dir." >&2
  exit 1
fi
wait_http_endpoint "http://$backend_addr/api/v1/matches" "backend API"

monitor_args=(
  monitor
  --cards-db-path "$cards_db_full"
  --live-api-addr "$live_api_addr"
  --match-upload-url "http://$backend_addr/api/v1/matches"
)
if [[ -n "$logs_full" ]]; then
  monitor_args+=(--logs-path "$logs_full")
fi

echo "Starting monitor..."
(cd "$repo_root" && "$cli_binary" "${monitor_args[@]}" >"$run_dir/monitor.stdout.log" 2>"$run_dir/monitor.stderr.log") &
monitor_pid=$!
sleep 2
if ! kill -0 "$monitor_pid" 2>/dev/null; then
  echo "Monitor exited early. Check logs in $run_dir." >&2
  exit 1
fi
wait_http_endpoint "http://$live_api_addr/api/v1/live/overlay-snapshot" "live overlay API"

echo
echo "Local services are ready:"
echo "  backend: http://$backend_addr"
echo "  live api: http://$live_api_addr"
echo "  hearthstone logs: ${logs_full:-"(auto-discovery failed)"}"
echo "  deck catalog: ${backend_logs_path:-"(disabled)"}"
echo "  database: $(format_database_url_for_display "$database_url")"
echo "  logs: $run_dir"
echo

(cd "$client_dir" && npm run start) &
client_pid=$!
wait "$client_pid"
