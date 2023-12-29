// Package vars maintains a list of metrics. These can be used to
// monitor behavior of an application.
package vars

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Metrics holds a set of metric values that can be updated
// atomically.
type Metrics struct {
	mu     sync.Mutex
	Detail map[string]interface{}
}

// New establishes a group of metrics.
func New() *Metrics {
	return &Metrics{Detail: make(map[string]interface{})}
}

// ErrInvalid and ErrNotNumber are standard errors returned by this
// package.
var (
	ErrInvalid   = errors.New("undefined metrics")
	ErrNotNumber = errors.New("not a number")
)

// Set sets the value of a specific metric.
func (m *Metrics) Set(k string, value interface{}) error {
	if m == nil {
		return ErrInvalid
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Detail[k] = value
	return nil
}

// Get returns the current value of a specific metric.
func (m *Metrics) Get(k string) interface{} {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Detail[k]
}

// getNumberLocked returns a numerical value for a specific metric, or
// an error.
func (m *Metrics) getNumberLocked(k string) (float64, error) {
	v := m.Detail[k]
	switch v.(type) {
	case int:
		return float64(v.(int)), nil
	case int32:
		return float64(v.(int32)), nil
	case int64:
		return float64(v.(int64)), nil
	case uint:
		return float64(v.(uint)), nil
	case uint32:
		return float64(v.(uint32)), nil
	case uint64:
		return float64(v.(uint64)), nil
	case float64:
		return v.(float64), nil
	default:
		return 0, ErrNotNumber
	}
}

// GetNumber returns the numerical value of a metric or, in the case
// the metric is not a number, it indicates this with an error value.
func (m *Metrics) GetNumber(k string) (float64, error) {
	if m == nil {
		return 0, ErrNotNumber
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getNumberLocked(k)
}

// Add ads one to a metric or, in the case the metric was not
// previously numerical, it replaces the metric with the provided
// number.
func (m *Metrics) Add(k string, n float64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	x, err := m.getNumberLocked(k)
	if err != nil {
		m.Detail[k] = n
	} else {
		m.Detail[k] = n + x
	}
}

// Snapshot holds a timestamped snapshot of metrics.
type Snapshot struct {
	When   time.Time
	Values *Metrics
}

// Snap snapshots all of the current metric values.
func (m *Metrics) Snap() *Snapshot {
	s := &Snapshot{
		Values: New(),
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.Detail {
		s.Values.Detail[k] = v
	}
	s.When = time.Now()
	return s
}

// DumpMDTable returns a byte array of markdown text that represents a
// table of the current values of all the metrics.
func (m *Metrics) DumpMDTable() []byte {
	if m == nil {
		return nil
	}
	s := m.Snap()
	var ks []string
	for x := range s.Values.Detail {
		ks = append(ks, x)
	}
	sort.Strings(ks)

	for i, x := range ks {
		ks[i] = fmt.Sprintf("%s | %v", x, s.Values.Detail[x])
	}

	return []byte(strings.Join(append([]string{fmt.Sprintf("key | value at %s\n----|------", s.When.Format(time.UnixDate))}, ks...), "\n") + "\n")
}
