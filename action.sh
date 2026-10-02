#!/usr/bin/env bash
set -euo pipefail

tool="fmd2json"
tag="v0.0.7"
version="${tag#v}"

# X86 needs conversion to ia32, but it's unsupported, so no issue.
cache_arch=$(printf '%s' "${RUNNER_ARCH:?}" | tr '[:upper:]' '[:lower:]')

# Follow the @actions/tool-cache convention.
tool_root="${RUNNER_TOOL_CACHE:?}/${tool}/${version}/${cache_arch}"
marker="${tool_root}.complete"
bin_dir="${tool_root}/bin"

if [[ ! -f "${marker}" ]]; then
  sh "${GITHUB_ACTION_PATH:?}/install.sh" -b "${bin_dir}" "${tag}"
  : > "${marker}"
fi

echo "${bin_dir}" >> "${GITHUB_PATH:?}"
