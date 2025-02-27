package metrics

import (
	"fmt"
	"slices"
	"time"
)

type result struct {
	key      string
	start    time.Time
	duration time.Duration
	err      error
}

type logData struct {
	key       string
	count     int64
	ts        time.Time
	durations []time.Duration
}

func (d logData) appendRecord(bytes []byte) []byte {
	slices.Sort(d.durations)

	p99 := d.durations[int(float32(len(d.durations))*0.99)-1]
	p98 := d.durations[int(float32(len(d.durations))*0.98)-1]
	p95 := d.durations[int(float32(len(d.durations))*0.95)-1]

	return fmt.Appendf(bytes, "%s,%d,%d,%d,%d,%d\n", d.key, d.ts.UnixMilli(), d.count, p99, p98, p95)
}

func newLogs(results []result) logRecords {
	m := map[string]*logData{}

	for _, r := range results {
		d, ok := m[r.key]
		if !ok {
			d = &logData{
				key: r.key,
				ts:  r.start,
			}
			m[r.key] = d
		}

		if r.start.After(d.ts) {
			d.ts = r.start
		}

		d.count++
		d.durations = append(d.durations, r.duration)
	}

	bytes := make([]byte, 0, 1000)
	for _, d := range m {
		bytes = d.appendRecord(bytes)
	}

	return bytes
}
