#!/usr/bin/env bash
# The bring-in lab (issue #1155): adopt a sample repository into a new
# workspace, import the other samples into it, and check that each one installs,
# builds, tests, lints and passes smith doctor.
#
# A step passes when its command exits 0, or when its failure is named in the
# "Left to do" list that adopt or import printed, or when known-failing.tsv
# lists it with an open issue. A crash ("panic:", "bug in smith") always fails.
# The lab exits 1 when any step fails.
#
# Usage: tools/bring-in-lab/run.sh [lab-folder]
#   SMITH       the smith binary to test (default: smith on PATH)
#   LAB_ONLY    comma-separated scenario ids to run after the adopt one (default: all)
#   LAB_KEEP=1  keep the lab folder after the run
#
# See docs/contributing/bring-in-lab.md.
set -uo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
smith="${SMITH:-smith}"
lab="${1:-$(mktemp -d "${TMPDIR:-/tmp}/bring-in-lab.XXXXXX")}"
ws="$lab/ws"
logs="$lab/logs"
mkdir -p "$logs"

export GIT_AUTHOR_NAME="bring-in lab" GIT_AUTHOR_EMAIL="lab@example.invalid"
export GIT_COMMITTER_NAME="$GIT_AUTHOR_NAME" GIT_COMMITTER_EMAIL="$GIT_AUTHOR_EMAIL"
export CI="${CI:-true}" NO_COLOR=1

steps=(bring-in install build test lint doctor ignored)
results=()   # "id|step|status|detail"
failed=0

# rows prints the non-comment lines of a tab-separated file.
rows() { grep -v '^[[:space:]]*#' "$1" | grep -v '^[[:space:]]*$'; }

# known_issue prints the issue number of a known-failing step, or nothing.
known_issue() {
  rows "$here/known-failing.tsv" | awk -F'\t' -v id="$1" -v step="$2" '$1==id && $2==step {print $3; exit}'
}

# left_to_do prints the "Left to do" items of a bring-in log, one per line.
left_to_do() {
  [ -f "$1" ] || return 0
  awk '/Left to do \(/ {on=1; next} on && /^  next:/ {on=0} on' "$1"
}

# step_named reports whether the Left to do list names a failed step.
step_named() {
  local list="$1" step="$2"
  [ -n "$list" ] || return 1
  case "$step" in
    install) grep -qiE 'install|lockfile' <<<"$list" ;;
    build)   grep -qiE 'build' <<<"$list" ;;
    test)    grep -qiE 'test' <<<"$list" ;;
    lint)    grep -qiE 'lint|eslint|ruff' <<<"$list" ;;
    # doctor checks the same workspace again: any named item may be its finding.
    doctor)  return 0 ;;
    *)       return 1 ;;
  esac
}

# record classifies one step and adds it to the results.
record() {
  local id="$1" step="$2" code="$3" log="$4" left="$5" issue status detail=""
  issue="$(known_issue "$id" "$step")"
  if grep -qiE '^panic:|goroutine [0-9]+ \[|bug in smith' "$log" 2>/dev/null; then
    status="FAIL"; detail="smith crashed"
  elif [ "$code" = 0 ]; then
    status="pass"
    [ -n "$issue" ] && detail="known #$issue passes now: remove it from known-failing.tsv"
  elif [ -n "$issue" ]; then
    status="known"; detail="#$issue"
  elif step_named "$left" "$step"; then
    status="left to do"
  else
    status="FAIL"; detail="exit $code, not in Left to do"
  fi
  # A known crash stays known: the issue names it.
  if [ "$status" = FAIL ] && [ -n "$issue" ]; then status="known"; detail="#$issue"; fi
  [ "$status" = FAIL ] && failed=1
  results+=("$id|$step|$status|$detail")
  printf '  %-10s %-11s %s\n' "$step" "$status" "$detail"
}

# run_step runs one command in the workspace, logs it and records the result.
run_step() {
  local id="$1" step="$2" left="$3"; shift 3
  local log="$logs/$id-$step.log" code
  (cd "$ws" && "$@") >"$log" 2>&1 </dev/null
  code=$?
  record "$id" "$step" "$code" "$log" "$left"
  return "$code"
}

# ignored checks that the bring-in committed no build output and that install,
# build and test left none untracked: the workspace's .gitignore must hide
# node_modules, dist, .venv, and the bin and obj folders of .NET.
# Only the paths of the scenario's project folder (<kind>/<project>/) count.
check_ignored() {
  local id="$1" project="$2" log="$logs/$id-ignored.log" code=0
  { (cd "$ws" && git ls-files && git status --porcelain --untracked-files=normal); } 2>&1 |
    grep -E "(^| )[a-z]+/$project/" >"$log"
  if grep -qE '/(node_modules|dist|\.venv|obj|bin/(Debug|Release))/' "$log"; then code=1; fi
  record "$id" ignored "$code" "$log" ""
  # Hide the output in .git/info/exclude so the next import, which needs a
  # clean tree, can run. The workspace files stay as smith wrote them.
  printf '%s\n' node_modules/ dist/ .venv/ obj/ bin/Debug/ bin/Release/ __pycache__/ >>"$ws/.git/info/exclude"
}

# settle commits what the steps changed (lockfiles), so the next import with
# --history keep finds a clean tree.
settle() {
  (cd "$ws" && git add -A && git commit -q --allow-empty -m "lab: after $1") >>"$logs/$1-settle.log" 2>&1
}

wanted() {
  [ -z "${LAB_ONLY:-}" ] && return 0
  [[ ",$LAB_ONLY," == *",$1,"* ]]
}

echo "bring-in lab in $lab with $("$smith" --version 2>&1)"
first=1
while IFS=$'\t' read -r id mode url sha project flags; do
  if [ "$first" = 0 ] && ! wanted "$id"; then continue; fi
  echo "== $id ($mode $url@${sha:0:12})"
  # shellcheck disable=SC2086 # flags is a list of words
  read -r -a extra <<<"$flags"
  bring="$logs/$id-bring-in.log"
  if [ "$mode" = adopt ]; then
    rm -rf "$ws"
    git clone -q "$url" "$ws" >"$logs/$id-clone.log" 2>&1 &&
      (cd "$ws" && git checkout -q -B main "$sha" && git remote remove origin) >>"$logs/$id-clone.log" 2>&1 ||
      { echo "cannot clone $url"; exit 2; }
    run_step "$id" bring-in "" "$smith" adopt --init --yes --commit "${extra[@]}"
  else
    run_step "$id" bring-in "" "$smith" import "$url" --ref "$sha" --history keep --yes --commit "${extra[@]}"
  fi
  code=$?
  left="$(left_to_do "$bring")"
  if [ "$code" != 0 ]; then
    for s in "${steps[@]:1}"; do results+=("$id|$s|skipped|bring-in failed"); done
    [ "$first" = 1 ] && break
    settle "$id"; first=0; continue
  fi
  run_step "$id" install "$left" "$smith" install
  run_step "$id" build "$left" "$smith" run build -p "$project"
  run_step "$id" test "$left" "$smith" run test -p "$project"
  run_step "$id" lint "$left" "$smith" run lint -p "$project"
  run_step "$id" doctor "$left" "$smith" doctor
  check_ignored "$id" "$project"
  settle "$id"
  first=0
done < <(rows "$here/repos.tsv")

# The result table, as Markdown, to stdout and to the GitHub job summary.
table() {
  printf '| scenario |'; for s in "${steps[@]}"; do printf ' %s |' "$s"; done; echo
  printf '|---|'; for _ in "${steps[@]}"; do printf -- '---|'; done; echo
  local ids=() r id
  for r in "${results[@]}"; do id="${r%%|*}"; [[ " ${ids[*]} " == *" $id "* ]] || ids+=("$id"); done
  for id in "${ids[@]}"; do
    printf '| %s |' "$id"
    for s in "${steps[@]}"; do
      local cell="-"
      for r in "${results[@]}"; do
        IFS='|' read -r rid rstep rstatus rdetail <<<"$r"
        if [ "$rid" = "$id" ] && [ "$rstep" = "$s" ]; then cell="$rstatus${rdetail:+ ($rdetail)}"; fi
      done
      printf ' %s |' "$cell"
    done
    echo
  done
}
echo
table | tee -a "${GITHUB_STEP_SUMMARY:-/dev/null}"
echo
echo "logs: $logs"
[ "${LAB_KEEP:-}" = 1 ] || [ -n "${1:-}" ] || rm -rf "$lab"
exit "$failed"
