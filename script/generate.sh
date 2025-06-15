#!/usr/bin/env bash

eval "$(pkgx dev --shellcode)"

set -eu

function main() {
    pkgx sqlc@1.29.0 generate
}

main