#!/usr/bin/env bash
# Start a throwaway Warpgate for the acceptance tests and print the environment
# they need.
#
#   eval "$(scripts/test-warpgate.sh start)"
#   TF_ACC=1 go test ./internal/provider/ -run TestAcc -v
#   scripts/test-warpgate.sh stop
#
# Uses the release binary rather than a container image: ghcr.io/warp-tech/warpgate
# has no 0.27 tag yet.
set -euo pipefail

VERSION="${WARPGATE_VERSION:-0.27.3}"
PORT="${WARPGATE_TEST_PORT:-18888}"
TOKEN="${WARPGATE_TEST_TOKEN:-acctest-admin-token}"
RUNDIR="${WARPGATE_TEST_DIR:-${TMPDIR:-/tmp}/warpgate-acctest}"

BIN="$RUNDIR/warpgate"
PIDFILE="$RUNDIR/warpgate.pid"

detect_asset() {
  local os arch
  case "$(uname -s)" in
    Linux) os=linux ;;
    Darwin) os=macos ;;
    *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) arch=x86_64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
  echo "warpgate-v${VERSION}-${arch}-${os}"
}

start() {
  mkdir -p "$RUNDIR/data"

  if [ ! -x "$BIN" ]; then
    local asset
    asset="$(detect_asset)"
    curl --fail --location --silent --show-error -o "$BIN" \
      "https://github.com/warp-tech/warpgate/releases/download/v${VERSION}/${asset}" >&2
    chmod +x "$BIN"
  fi

  if [ ! -f "$RUNDIR/warpgate.yaml" ]; then
    WARPGATE_ADMIN_PASSWORD='AccTestPassword123!' "$BIN" --config "$RUNDIR/warpgate.yaml" \
      unattended-setup --data-path "$RUNDIR/data" --http-port "$PORT" --external-host localhost >&2
  fi

  WARPGATE_ADMIN_TOKEN="$TOKEN" nohup "$BIN" --config "$RUNDIR/warpgate.yaml" \
    run --enable-admin-token > "$RUNDIR/warpgate.log" 2>&1 &
  echo $! > "$PIDFILE"

  local i
  for i in $(seq 1 60); do
    if curl -sk -o /dev/null "https://localhost:${PORT}/@warpgate/admin/api/parameters"; then
      break
    fi
    if [ "$i" = 60 ]; then
      echo "Warpgate did not come up; see $RUNDIR/warpgate.log" >&2
      exit 1
    fi
    sleep 1
  done

  echo "export WARPGATE_HOST=https://localhost:${PORT}"
  echo "export WARPGATE_TOKEN=${TOKEN}"
  echo "export WARPGATE_INSECURE_SKIP_VERIFY=true"
}

stop() {
  if [ -f "$PIDFILE" ]; then
    kill "$(cat "$PIDFILE")" 2>/dev/null || true
    rm -f "$PIDFILE"
  fi
}

case "${1:-start}" in
  start) start ;;
  stop) stop ;;
  clean) stop; rm -rf "$RUNDIR" ;;
  *) echo "usage: $0 {start|stop|clean}" >&2; exit 1 ;;
esac
