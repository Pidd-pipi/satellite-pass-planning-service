package main

import "fmt"

func seedOpsRecords() []OpsRecord {
	priorities := []OpsPriority{OpsPriorityLow, OpsPriorityNormal, OpsPriorityHigh, OpsPriorityCritical}
	statuses := []OpsStatus{OpsStatusQueued, OpsStatusActive, OpsStatusPaused, OpsStatusClosed}
	owners := []string{"li", "wang", "zhang", "chen"}
	subjects := []string{
		"pass scheduling review", "station antenna calibration", "downlink bandwidth plan",
		"uplink command verify", "orbital element update", "ground station maintenance",
		"pass conflict resolution", "telemetry checkout", "receive window audit",
		"station handover rehearsal",
	}
	records := make([]OpsRecord, 0, 40)
	for i := 1; i <= 40; i++ {
		records = append(records, OpsRecord{
			ID:        fmt.Sprintf("ops-%03d", i),
			Subject:   fmt.Sprintf("%s %d", subjects[(i-1)%len(subjects)], i),
			Owner:     owners[(i-1)%len(owners)],
			Status:    statuses[(i-1)%len(statuses)],
			Priority:  priorities[(i-1)%len(priorities)],
			Revision:  1,
			Labels:    map[string]string{"site": fmt.Sprintf("gs-%d", (i-1)%5+1)},
			CreatedAt: "2026-08-10T08:00:00Z",
			UpdatedAt: "2026-08-10T08:00:00Z",
		})
	}
	return records
}
