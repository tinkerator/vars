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
	ErrNotFound  = errors.New("not found")
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

// AsNumber returns a numerical value for an interface{} value, or an
// error.
func AsNumber(v interface{}) (float64, error) {
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
	v := m.Detail[k]
	m.mu.Unlock()
	return AsNumber(v)
}

// Add adds a number to a metric or, in the case the metric was not
// previously numerical, it replaces the metric with the provided
// number, n.
func (m *Metrics) Add(k string, n float64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	x, ok := m.Detail[k]
	v, err := AsNumber(x)
	if !ok || err != nil {
		m.Detail[k] = n
	} else {
		m.Detail[k] = n + v
	}
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
	s.When = time.Now()
	for k, v := range m.Detail {
		s.Values.Detail[k] = v
	}
	return s
}

// Trim removes redundant entries from an array of Snapshots.  The
// returned value includes the most recently valid timestamp for all
// entries. That is, the most recent snapshot of the trimmed slice is
// a full snapshot. The slice is edited in place and the length of the
// slice may also reduce.
func Trim(snaps []*Snapshot) (results []*Snapshot) {
	latest := make(map[string]string)
	for i := 0; i < len(snaps)-1; i++ {
		m := snaps[i].Values
		var ks []string
		m.mu.Lock()
		for k, v := range m.Detail {
			s := fmt.Sprint(v)
			if latest[k] == s {
				ks = append(ks, k)
			} else {
				latest[k] = s
			}
		}
		for _, k := range ks {
			delete(m.Detail, k)
		}
		m.mu.Unlock()
		if len(m.Detail) == 0 {
			snaps = append(snaps[:i], snaps[i+1:]...)
			i--
		}
	}
	results = snaps
	return
}

// Infer returns the most current value for a specified key at the
// requested time, indicating the time when the returned value was
// recorded.
func Infer(snaps []*Snapshot, t time.Time, k string) (index int, v interface{}, err error) {
	if len(snaps) == 0 || snaps[0].When.After(t) {
		err = ErrNotFound
		return
	}
	before := sort.Search(len(snaps), func(a int) bool {
		return snaps[a].When.After(t)
	})
	var ok bool
	for i := before - 1; before > 0; before-- {
		if v, ok = snaps[i].Values.Detail[k]; ok {
			index = i
			return
		}
	}
	err = ErrNotFound
	return
}

// ExtractNumbers returns an array of number values. The first column
// holds the number of timeunits since the epoch associated with the
// measured value.
func ExtractNumbers(snaps []*Snapshot, timeunits time.Duration, from, to time.Time, vars []string) ([][]float64, error) {
	starts := make(map[string]float64)
	var minI int
	for j, k := range vars {
		i, v, err := Infer(snaps, from, k)
		if err != nil {
			return nil, fmt.Errorf("error for %q at %v: %v", k, from, err)
		}
		n, err := AsNumber(v)
		if err != nil {
			return nil, fmt.Errorf("error for %q at %v: %v", k, from, err)
		}
		if j == 0 || i > minI {
			minI = i
		}
		starts[k] = n
	}
	lastTS := float64(0)
	ts := float64(from.UnixNano() / int64(timeunits))
	var lines [][]float64
	done := false
	for i := minI + 1; i <= len(snaps); i++ {
		vs := []float64{ts}
		for _, k := range vars {
			vs = append(vs, starts[k])
		}
		if ts == lastTS {
			lines[len(lines)-1] = vs
		} else {
			lines = append(lines, vs)
		}
		if done || i == len(snaps) {
			break
		}
		lastTS = ts
		s := snaps[i]
		if s.When.Before(to) {
			ts = float64(s.When.UnixNano() / int64(timeunits))
		} else if tts := float64(to.UnixNano() / int64(timeunits)); ts == tts {
			break
		} else {
			ts = tts
			done = true
			continue
		}
		for k, x := range s.Values.Detail {
			v, err := AsNumber(x)
			if err != nil {
				return nil, fmt.Errorf("snapshot[%d][%q] = %v: %v", i, k, x, err)
			}
			starts[k] = v
		}
	}
	return lines, nil
}

// Sample holds an (Timestamp,X) value.
type Sample struct {
	When  time.Time
	Value float64
}

// Rate is a convenience function for determining the rate of change
// of some variable, at v[1]. It tries to compute the gradient of the
// two points around the center point, but fails over to a best guess
// based on two or fewer samples.
func Rate(v ...Sample) float64 {
	if len(v) <= 1 {
		return 0
	}
	t0 := float64(v[0].When.UnixNano())
	var t1, v1 float64
	if len(v) == 2 {
		t1 = float64(v[1].When.UnixNano())
		v1 = v[1].Value
	} else {
		t1 = float64(v[2].When.UnixNano())
		v1 = v[2].Value
	}
	return (v1 - v[0].Value) / (t1 - t0) * float64(time.Second/time.Nanosecond)
}
