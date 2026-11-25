#!/usr/bin/env bash
# Copyright 2026 High Five Labs
# SPDX-License-Identifier: Apache-2.0
# Download the vendor's OpenAPI specs at the commits vendor.yaml pins.
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p specs
[ -s specs/confluent.yaml ] || curl -fsSL -o specs/confluent.yaml "https://docs.confluent.io/cloud/current/openapi.yaml"
echo "4d2a5f295d93781c8bc125c319c3b9f4aa22052390021b710291545e5c01c956  specs/confluent.yaml" | sha256sum -c --quiet - || { echo "specs/confluent.yaml changed upstream: review, regenerate, and update sha256 in vendor.yaml" >&2; exit 1; }
