package fluxrus

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	influx "github.com/influxdata/influxdb/client/v2"
)

type InfluxHook struct {
	client        influx.Client
	database      string
	measurement   string
	tags          []string
	precision     string
	batchSize     int
	batchInterval time.Duration
	batchChan     chan *influx.Point
	flushChan     chan struct{}
	flushed       chan struct{}
	err           error
	errLock       sync.RWMutex
}

func (h *InfluxHook) setError(e error) {
	h.errLock.Lock()
	h.err = e
	h.errLock.Unlock()
}

func New(url, db, measurement string, opts ...Option) (*InfluxHook, error) {
	hook := &InfluxHook{
		database:      db,
		measurement:   measurement,
		precision:     "ns",
		batchSize:     200,
		batchInterval: 5 * time.Second,
		flushChan:     make(chan struct{}),
		flushed:       make(chan struct{}),
		err:           nil,
		errLock:       sync.RWMutex{},
	}

	for _, o := range opts {
		o(hook)
	}

	// we have no client - create one
	if hook.client == nil {
		influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
			Addr: url,
		})
		if err != nil {
			return nil, err
		}
		hook.client = influxClient
	}

	hook.batchChan = make(chan *influx.Point, hook.batchSize)

	go func() {
		var err error
		var batch influx.BatchPoints

		ticker := time.NewTicker(hook.batchInterval)
		batch, err = influx.NewBatchPoints(influx.BatchPointsConfig{
			Database:  hook.database,
			Precision: hook.precision,
		})
		if err != nil {
			logrus.Errorf("Could not create the InfluxDB batch of points: %v", err)
		}

		flushAndClear := func() {
			err := hook.client.Write(batch)
			hook.setError(err)

			// only clear the buffer if all data is written to the server
			if err == nil {
				batch, err = influx.NewBatchPoints(influx.BatchPointsConfig{
					Database:  hook.database,
					Precision: hook.precision,
				})
			}
		}

		for {
			select {
			case <-ticker.C:
				flushAndClear()
			case p := <-hook.batchChan:
				batch.AddPoint(p)
				if len(batch.Points()) >= hook.batchSize {
					flushAndClear()
				}
			case <-hook.flushChan:
				for p := range hook.batchChan {
					batch.AddPoint(p)
				}
				flushAndClear()
				hook.flushed <- struct{}{}
			}
		}

	}()

	return hook, nil
}

func (h *InfluxHook) Close() {
	close(h.batchChan)
	h.flushChan <- struct{}{}
	<-h.flushed
}

func (h *InfluxHook) Fire(entry *logrus.Entry) error {
	tags := map[string]string{
		"level": entry.Level.String(),
	}
	for _, tag := range h.tags {
		if tagValue, ok := getTag(entry.Data, tag); ok {
			tags[tag] = tagValue
		}
	}

	fields := map[string]interface{}{
		"message": entry.Message,
	}
	for k, v := range entry.Data {
		fields[k] = v
	}
	for _, tag := range h.tags {
		delete(fields, tag)
	}

	pt, err := influx.NewPoint(h.measurement, tags, fields, entry.Time)
	if err != nil {
		return fmt.Errorf("Could not create new InfluxDB point: %v", err)
	}
	h.batchChan <- pt

	h.errLock.RLock()
	err = h.err
	h.errLock.RUnlock()
	return err
}

// Levels implementation allows for level logging.
func (h *InfluxHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// Helper function.
func getTag(fields logrus.Fields, tag string) (string, bool) {
	value, ok := fields[tag]
	if ok {
		return fmt.Sprintf("%v", value), ok
	}
	return "", ok
}
