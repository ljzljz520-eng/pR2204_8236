package report

import (
	"fmt"
	"sort"
	"strings"

	"examvault/internal/domain"
)

type SearchReport struct {
	Query   string
	Count   int
	IDs     []string
	Status  map[string]int
	Message string
}

type ImportReport struct {
	Imported int
	Rejected int
	IDs      []string
	Reasons  []string
	Digest   string
}

func SearchSummary(query string, records []domain.Record) SearchReport {
	ids := make([]string, 0, len(records))
	status := make(map[string]int)
	for _, record := range records {
		ids = append(ids, record.ID)
		status[record.Status]++
	}
	sort.Strings(ids)
	message := "records found"
	if len(ids) == 0 {
		message = "no records found"
	}
	return SearchReport{Query: strings.TrimSpace(query), Count: len(ids), IDs: ids, Status: status, Message: message}
}

func ImportSummary(result domain.ImportResult, digest string) ImportReport {
	ids := append([]string(nil), result.IDs...)
	reasons := append([]string(nil), result.Reasons...)
	sort.Strings(ids)
	sort.Strings(reasons)
	return ImportReport{Imported: result.Imported, Rejected: result.Rejected, IDs: ids, Reasons: reasons, Digest: digest}
}

func PermissionSummary(records []domain.Record) map[string]int {
	result := map[string]int{domain.PermissionPublic: 0, domain.PermissionStaff: 0, domain.PermissionRestricted: 0}
	for _, record := range records {
		if _, ok := result[record.Permission]; ok {
			result[record.Permission]++
		}
	}
	return result
}

func Active(records []domain.Record) []domain.Record {
	result := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if record.IsActive() {
			result = append(result, record)
		}
	}
	domain.SortRecords(result)
	return result
}

type Dashboard struct {
	Total        int
	Active       int
	Archived     int
	ByStatus     map[string]int
	ByPermission map[string]int
	ByOwner      map[string]int
}

func DashboardSummary(records []domain.Record) Dashboard {
	dashboard := Dashboard{ByStatus: make(map[string]int), ByPermission: make(map[string]int), ByOwner: make(map[string]int)}
	for _, record := range records {
		dashboard.Total++
		dashboard.ByStatus[record.Status]++
		dashboard.ByPermission[record.Permission]++
		dashboard.ByOwner[record.Owner]++
		if record.Status == domain.StatusArchived {
			dashboard.Archived++
		} else {
			dashboard.Active++
		}
	}
	return dashboard
}

func SortByRevision(records []domain.Record) []domain.Record {
	result := domain.CloneRecords(records)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Revision == result[j].Revision {
			return result[i].ID < result[j].ID
		}
		return result[i].Revision < result[j].Revision
	})
	return result
}

func LatestRevision(records []domain.Record) int {
	latest := 0
	for _, record := range records {
		if record.Revision > latest {
			latest = record.Revision
		}
	}
	return latest
}

func StatusLine(status string, count int) string {
	return fmt.Sprintf("%s=%d", domain.StatusLabel(status), count)
}

func PermissionLines(records []domain.Record) []string {
	counts := PermissionSummary(records)
	lines := make([]string, 0, len(counts))
	for _, permission := range domain.AllPermissions() {
		lines = append(lines, fmt.Sprintf("%s=%d", domain.PermissionLabel(permission), counts[permission]))
	}
	return lines
}

func OwnerLines(records []domain.Record) []string {
	counts := make(map[string]int)
	for _, record := range records {
		counts[record.Owner]++
	}
	owners := make([]string, 0, len(counts))
	for owner := range counts {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	lines := make([]string, 0, len(owners))
	for _, owner := range owners {
		lines = append(lines, fmt.Sprintf("%s=%d", owner, counts[owner]))
	}
	return lines
}

func EmptySearch(query string) SearchReport { return SearchSummary(query, []domain.Record{}) }

func MergeReports(left, right SearchReport) SearchReport {
	ids := append(append([]string(nil), left.IDs...), right.IDs...)
	sort.Strings(ids)
	status := make(map[string]int)
	for key, value := range left.Status {
		status[key] += value
	}
	for key, value := range right.Status {
		status[key] += value
	}
	return SearchReport{Query: left.Query, Count: len(ids), IDs: ids, Status: status, Message: fmt.Sprintf("%d records found", len(ids))}
}
