package runner

import (
    "regexp"
    "strings"
)

var nsjailLineRE = regexp.MustCompile(`(?m)^\[[A-Z]\].*$`)

// extractNsjailLogs finds lines that look like nsjail diagnostics
// (e.g. "[I][...]") and returns (cleaned, extracted).
func extractNsjailLogs(s string) (string, string) {
    if s == "" {
        return "", ""
    }

    matches := nsjailLineRE.FindAllString(s, -1)
    if len(matches) == 0 {
        return s, ""
    }

    // remove matched lines
    cleaned := nsjailLineRE.ReplaceAllString(s, "")
    // also trim excessive blank lines
    cleaned = strings.TrimSpace(cleaned)
    extracted := strings.Join(matches, "\n")
    return cleaned, extracted
}
