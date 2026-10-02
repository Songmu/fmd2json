#!/usr/bin/env bash
set -euo pipefail

tool="fmd2json"
tag="v0.0.6"
version="${tag#v}"

case "${RUNNER_ARCH:?}" in
  X64)
    cache_arch="x64"
    ;;
  ARM64)
    cache_arch="arm64"
    ;;
  X86)
    cache_arch="ia32"
    ;;
  ARM)
    cache_arch="arm"
    ;;
  *)
    echo "Unsupported runner architecture: ${RUNNER_ARCH}" >&2
    exit 1
    ;;
esac

# Follow the @actions/tool-cache convention.
tool_root="${RUNNER_TOOL_CACHE:?}/${tool}/${version}/${cache_arch}"
marker="${tool_root}.complete"
bin_dir="${tool_root}/bin"

if [[ ! -f "${marker}" ]]; then
  sh "${GITHUB_ACTION_PATH:?}/install.sh" -b "${bin_dir}" "${tag}"
  : > "${marker}"
fi

echo "${bin_dir}" >> "${GITHUB_PATH:?}"
