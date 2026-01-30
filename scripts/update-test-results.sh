#!/bin/bash
# Update test-results.json with coverage and/or mutation data
# Usage: ./update-test-results.sh [--coverage coverage.out] [--mutation mutation.json]

set -euo pipefail

RESULTS_FILE="${RESULTS_FILE:-docs/test-results.json}"
COVERAGE_FILE=""
MUTATION_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage) COVERAGE_FILE="$2"; shift 2 ;;
        --mutation) MUTATION_FILE="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Initialize results file if it doesn't exist
if [[ ! -f "$RESULTS_FILE" ]]; then
    echo '{"coverage":{"total":null,"packages":{},"lastUpdated":null},"mutation":{"summary":{},"results":[],"lastUpdated":null}}' > "$RESULTS_FILE"
fi

TIMESTAMP=$(date -u +"%Y-%m-%d %H:%M UTC")

# Update coverage if provided
if [[ -n "$COVERAGE_FILE" && -f "$COVERAGE_FILE" ]]; then
    echo "Updating coverage from $COVERAGE_FILE..."

    # Parse coverage.out to get per-package coverage
    COVERAGE_DATA=$(go tool cover -func="$COVERAGE_FILE" | awk '
        /^total:/ { total = $3; gsub(/%/, "", total) }
        /\.go:/ {
            split($1, parts, "/")
            # Get package path (everything before the filename)
            pkg = ""
            for (i = 1; i < length(parts); i++) {
                if (pkg != "") pkg = pkg "/"
                pkg = pkg parts[i]
            }
            # Track coverage per package
            if (!(pkg in pkg_funcs)) {
                pkg_funcs[pkg] = 0
                pkg_covered[pkg] = 0
            }
            pct = $3
            gsub(/%/, "", pct)
            pkg_funcs[pkg]++
            pkg_covered[pkg] += pct
        }
        END {
            printf "{\"total\":%s,\"packages\":{", total
            first = 1
            for (pkg in pkg_funcs) {
                if (!first) printf ","
                avg = pkg_covered[pkg] / pkg_funcs[pkg]
                printf "\"%s\":%.1f", pkg, avg
                first = 0
            }
            printf "}}"
        }
    ')

    # Merge coverage into results
    jq --argjson cov "$COVERAGE_DATA" --arg ts "$TIMESTAMP" '
        .coverage = ($cov + {lastUpdated: $ts})
    ' "$RESULTS_FILE" > "$RESULTS_FILE.tmp" && mv "$RESULTS_FILE.tmp" "$RESULTS_FILE"

    echo "Coverage updated: $(jq -r '.coverage.total' "$RESULTS_FILE")%"
fi

# Update mutation if provided
if [[ -n "$MUTATION_FILE" && -f "$MUTATION_FILE" ]]; then
    echo "Updating mutation from $MUTATION_FILE..."

    # Get files tested in new run
    NEW_FILES=$(jq -r '.results[]?.file // .results[]?.File // empty' "$MUTATION_FILE" 2>/dev/null | sort -u || true)

    if [[ -n "$NEW_FILES" ]]; then
        FILE_FILTER=$(echo "$NEW_FILES" | jq -R -s 'split("\n") | map(select(length > 0))')

        # Merge: keep existing results for untested files, replace tested files
        jq --argjson newFiles "$FILE_FILTER" --slurpfile new "$MUTATION_FILE" --arg ts "$TIMESTAMP" '
            # Get existing results for files NOT in new run
            (.mutation.results // []) | map(
                select(
                    (.file // .File) as $f |
                    ($newFiles | index($f)) == null
                )
            ) as $kept |

            # Combine with new results
            ($kept + ($new[0].results // [])) as $merged |

            # Update mutation section
            .mutation = {
                summary: {
                    TotalMutants: ($merged | length),
                    Killed: ($merged | map(select((.status // .Status | ascii_downcase) == "killed")) | length),
                    Survived: ($merged | map(select((.status // .Status | ascii_downcase) == "survived")) | length),
                    Timeout: ($merged | map(select((.status // .Status | ascii_downcase) == "timeout")) | length),
                    Errors: ($merged | map(select((.status // .Status | ascii_downcase) == "error")) | length),
                    MutationScore: (
                        ($merged | map(select((.status // .Status | ascii_downcase) == "killed")) | length) as $killed |
                        ($merged | length) as $total |
                        if $total > 0 then (($killed / $total * 100) | . * 10 | floor / 10) else null end
                    )
                },
                results: $merged,
                lastUpdated: $ts
            }
        ' "$RESULTS_FILE" > "$RESULTS_FILE.tmp" && mv "$RESULTS_FILE.tmp" "$RESULTS_FILE"

        TOTAL=$(jq -r '.mutation.summary.TotalMutants' "$RESULTS_FILE")
        SCORE=$(jq -r '.mutation.summary.MutationScore // "N/A"' "$RESULTS_FILE")
        echo "Mutation updated: $TOTAL mutants, score: $SCORE%"
    else
        echo "No mutation results in $MUTATION_FILE"
    fi
fi

echo "Results saved to $RESULTS_FILE"
