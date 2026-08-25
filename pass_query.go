package main

import (
	"sort"
	"strings"
)

type PassQuery struct {
	Satellite string
	Station   string
	State     string
	Page      int
	PageSize  int
}

type PassPage struct {
	Items    []PassWindow `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int          `json:"total"`
	HasNext  bool         `json:"has_next"`
}

func passFilterBySatellite(items []PassWindow, satellite string) []PassWindow {
	if satellite == "" {
		return items
	}
	out := make([]PassWindow, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Satellite), strings.ToLower(satellite)) {
			out = append(out, item)
		}
	}
	return out
}

func passFilterByState(items []PassWindow, state string) []PassWindow {
	if state == "" {
		return items
	}
	out := make([]PassWindow, 0, len(items))
	for _, item := range items {
		if item.State == state {
			out = append(out, item)
		}
	}
	return out
}

func passMatch(item PassWindow, query PassQuery) bool {
	if query.Satellite != "" && !strings.Contains(strings.ToLower(item.Satellite), strings.ToLower(query.Satellite)) {
		return false
	}
	if query.Station != "" && item.Station != query.Station {
		return false
	}
	if query.State != "" && item.State != query.State {
		return false
	}
	return true
}

func sortPasses(items []PassWindow) {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].Start.Equal(items[j].Start) {
			return items[i].Start.Before(items[j].Start)
		}
		return items[i].ID < items[j].ID
	})
}

func passQueryDefaults(q PassQuery) PassQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 25
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q
}

func passBounds(total, page, size int) (int, int) {
	q := passQueryDefaults(PassQuery{Page: page, PageSize: size})
	start := (q.Page - 1) * q.PageSize
	if start > total {
		start = total
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return start, end
}

func passPageCount(total, size int) int {
	if size < 1 || total == 0 {
		return 0
	}
	return (total + size - 1) / size
}

func passQueryKey(q PassQuery) string {
	return strings.Join([]string{q.Satellite, q.Station, q.State}, "|")
}

func clonePassPage(p PassPage) PassPage {
	p.Items = append([]PassWindow(nil), p.Items...)
	return p
}
