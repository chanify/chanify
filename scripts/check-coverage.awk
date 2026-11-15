# Cross-package profiles can repeat a block; combine its execution counts before
# enforcing complete statement coverage.
$1 == "mode:" { next }
{
    statements[$1] = $2
    executions[$1] += $3
}
END {
    for (block in statements) {
        total += statements[block]
        if (executions[block] > 0) {
            covered += statements[block]
        } else if (statements[block] > 0) {
            printf "Uncovered: %s (%d statements)\n", block, statements[block] > "/dev/stderr"
        }
    }
    if (total == 0) {
        print "Coverage profile contains no statements" > "/dev/stderr"
        exit 1
    }
    printf "Coverage: %.1f%% (%d/%d statements)\n", 100 * covered / total, covered, total
    if (covered != total) {
        print "Required coverage: 100%" > "/dev/stderr"
        exit 1
    }
}
