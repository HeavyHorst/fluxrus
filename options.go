package fluxrus

import (
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
)

// Option configures the Hook
type Option func(*InfluxHook)

// WithBatchInterval sets the batchInterval
func WithBatchInterval(interval int) Option {
	return func(o *InfluxHook) {
		o.batchInterval = time.Duration(interval) * time.Second
	}
}

// WithBatchSize sets the batchSize
func WithBatchSize(count int) Option {
	return func(o *InfluxHook) {
		o.batchSize = count
	}
}

// WithPrecision sets the precision
func WithPrecision(p string) Option {
	return func(o *InfluxHook) {
		o.precision = p
	}
}

// WithTags sets the tags that we will extract from the log fields
func WithTags(t []string) Option {
	return func(o *InfluxHook) {
		o.tags = t
	}
}

// WithClient overwrites the default influx.Client
func WithClient(c influx.Client) Option {
	return func(o *InfluxHook) {
		o.client = c
	}
}
