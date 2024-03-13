#!/bin/bash
set -euo pipefail

scriptdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"


mosquitto -c $scriptdir/server.conf
