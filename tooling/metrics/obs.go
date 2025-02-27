package metrics

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

type metricChan chan result
type logsChan chan logRecords
type logRecords []byte

type observability struct {
	file        io.Writer
	metricsChan metricChan
	logsChan    logsChan
}

func NewDefault() (Obs, error) {
	f, err := os.Create("log.csv")
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}

	return New(f), nil
}

func New(writer io.Writer) Obs {
	c := make(metricChan, 10000)
	l := make(logsChan, 10)

	return &observability{
		file:        writer,
		metricsChan: c,
		logsChan:    l,
	}
}

func (o *observability) StartSpan(name string) (span Span) {
	return newSpan(name, o.metricsChan)
}

func (o *observability) MakeLogs(cancel <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	var buffer []result

	for {
		select {
		case r := <-o.metricsChan:
			buffer = append(buffer, r)
		case <-ticker.C:
			logs := newLogs(buffer)
			clear(buffer)

			o.logsChan <- logs

		case <-cancel:
			return

		default:
		}
	}
}

func (o *observability) WriteLogs() {
	for ll := range o.logsChan {
		_, err := o.file.Write(ll)
		if err != nil {
			panic(fmt.Errorf("failed to write logs to file: %v", err))
		}
	}
}

func (o *observability) StartLogging(ctx context.Context) {
	go o.MakeLogs(ctx.Done())

	go o.WriteLogs()
}
