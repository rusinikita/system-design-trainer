package metrics

import "time"

type span struct {
	key         string
	start       time.Time
	metricsChan metricChan
}

func newSpan(key string, metricChan metricChan) *span {
	return &span{
		key:         key,
		start:       time.Now(),
		metricsChan: metricChan,
	}
}

func (s *span) Done(err error) {
	r := result{
		key:      s.key,
		start:    s.start,
		duration: s.start.Sub(time.Now()),
		err:      err,
	}

	s.metricsChan <- r
}
