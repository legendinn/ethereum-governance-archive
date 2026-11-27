#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

block=0xdd782654955f5b0ffb41ab185b43a92e7220d0c07901fb29ee676577956c5a72
tx=0xf9df4d7016a592d1135dee34cacae43030c27c3eb69a12281385faa6bd487f56
proxy=0xc3d688b66703497daa19211eedff47f25384cdc3
implementation=0x2c045e9a1afb125898d6e736d1ff09fa28788e9d
proof=examples/compound-usdc/upgrade-proof.json

verify() {
  ./govarchive verify-upgrade -block "$block" -tx "$tx" \
    -address "$proxy" -implementation "$implementation" "$1"
}

if [ ! -x ./govarchive ]; then
  printf '%s\n' 'Build the CLI with go build -o govarchive ./cmd/govarchive' >&2
  exit 1
fi

verify "$proof"

scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
trap 'exit 1' HUP INT TERM
python3 - "$proof" "$scratch/tampered.json" "$implementation" <<'PY'
import json
import sys
from pathlib import Path

source, destination, implementation = sys.argv[1:]
proof = json.loads(Path(source).read_text(encoding="utf-8"))
address = implementation[2:]
changed = ("3" if address[0] == "2" else "2") + address[1:]
for index, node in enumerate(proof["receiptProof"]):
    if address in node:
        proof["receiptProof"][index] = node.replace(address, changed, 1)
        break
else:
    sys.exit("Implementation address is missing from the receipt proof")
Path(destination).write_text(json.dumps(proof), encoding="utf-8")
PY

if error=$(verify "$scratch/tampered.json" 2>&1); then
  printf '%s\n' 'Error: verification accepted an altered receipt proof' >&2
  exit 1
fi
case "$error" in
  'receipt proof:'*) printf 'Tampered receipt rejected (%s)\n' "$error" ;;
  *) printf 'Unexpected verification error (%s)\n' "$error" >&2; exit 1 ;;
esac
