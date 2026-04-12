#!/bin/bash
# Run const_gen from repo root so paths like db/const_db.sql resolve regardless of cwd.

set -euo pipefail
cd "$(dirname "$0")/.."
go run ./cmd/const_gen
