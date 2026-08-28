package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func EncodeSearch(report SearchReport) string {
	data, err := json.Marshal(report)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func EncodeImport(report ImportReport) string {
	data, err := json.Marshal(report)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func Lines(report SearchReport) []string {
	lines := make([]string, 0, len(report.IDs)+1)
	lines = append(lines, fmt.Sprintf("%s: %d (%s)", report.Query, report.Count, report.Message))
	ids := append([]string(nil), report.IDs...)
	sort.Strings(ids)
	lines = append(lines, ids...)
	return lines
}

func Join(lines []string) string { return strings.Join(lines, "\n") }

func EncodeDashboard(dashboard Dashboard) string {
	data, err := json.Marshal(dashboard)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func DashboardLines(dashboard Dashboard) []string {
	lines := []string{fmt.Sprintf("total=%d", dashboard.Total), fmt.Sprintf("active=%d", dashboard.Active), fmt.Sprintf("archived=%d", dashboard.Archived)}
	statuses := make([]string, 0, len(dashboard.ByStatus))
	for status := range dashboard.ByStatus {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	for _, status := range statuses {
		lines = append(lines, StatusLine(status, dashboard.ByStatus[status]))
	}
	return lines
}

func ImportLines(report ImportReport) []string {
	lines := []string{fmt.Sprintf("imported=%d", report.Imported), fmt.Sprintf("rejected=%d", report.Rejected)}
	lines = append(lines, report.IDs...)
	lines = append(lines, report.Reasons...)
	return lines
}

func Compact(report SearchReport) string {
	if report.Count == 0 {
		return "empty"
	}
	return fmt.Sprintf("%d:%s", report.Count, strings.Join(report.IDs, ","))
}
