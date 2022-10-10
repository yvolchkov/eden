#!/bin/bash

set -e

function main() {
    local eve_src_path="$(readlink -f ~/src/eve)"
    local eve_version="$(make DEV=n --no-print-directory --directory "${eve_src_path}" version)"

    echo "using eve version ${eve_version}"

    make clean
    make eden
    ./eden config add
    ./eden setup --eve-tag "${eve_version}"
    ./eden start
    ./eden eve onboard
}

main "${@}"
