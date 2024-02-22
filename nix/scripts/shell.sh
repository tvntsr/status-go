#!/usr/bin/env bash

# This script is used by the Makefile to have an implicit nix-shell.
# The following environment variables modify the script behavior:
# - TARGET: This attribute is passed via --attr to Nix, defining the scope.
# - _NIX_PURE: This variable allows for making the shell pure with the use of --pure.
#     Take note that this makes Nix tools like `nix-build` unavailable in the shell.
# - _NIX_KEEP: This variable allows specifying which env vars to keep for Nix pure shell.

GIT_ROOT=$(cd "${BASH_SOURCE%/*}" && git rev-parse --show-toplevel)
source "${GIT_ROOT}/scripts/colors.sh"
source "${GIT_ROOT}/nix/scripts/source.sh"

export TERM=xterm # fix for colors
shift # we remove the first -c from arguments

entryPoint="default.nix"
nixArgs=(
    "--show-trace"
    "--attr shell"
)

# This variable allows specifying which env vars to keep for Nix pure shell
# The separator is a colon
if [[ -n "${_NIX_KEEP}" ]]; then
    nixArgs+=("--keep ${_NIX_KEEP//,/ --keep }")
fi

# Not all builds are ready to be run in a pure environment
if [[ -n "${_NIX_PURE}" ]]; then
    nixArgs+=("--pure")
    pureDesc='pure '
fi

echo -e "${GRN}Configuring ${pureDesc}Nix shell for target ...${RST}" 1>&2

# Save derivation from being garbage collected
"${GIT_ROOT}/nix/scripts/gcroots.sh" "shell"

# ENTER_NIX_SHELL is the fake command used when `make shell` is run.
# It is just a special string, not a variable, and a marker to not use `--run`.
if [[ "${@}" == "ENTER_NIX_SHELL" ]]; then
    export NIX_SHELL_TARGET="${TARGET}"
    exec nix-shell ${nixArgs[@]} --keep NIX_SHELL_TARGET ${entryPoint}
else
    exec nix-shell ${nixArgs[@]} --run "$@" ${entryPoint}
fi
