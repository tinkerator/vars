package vars

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	var m *Metrics
	if err := m.Set("x", "y"); err == nil {
		t.Fatal("setting nil metrics worked!?")
	}
	m = New()
	m.Set("a", 1)
	m.Set("b", "this")
	if got := m.Get("c"); got != nil {
		t.Errorf("expected an error, but got=%v", got)
	}
	if got := m.Get("b"); got == nil {
		t.Errorf("failed to get \"b\" got=%v", got)
	} else if got != "this" {
		t.Errorf("failed to get the right value for \"b\", got=%q, want=\"this\"", got)
	}
	if got := m.Get("a"); got == nil {
		t.Errorf("failed to get \"a\": %v", got)
	} else if got != 1 {
		t.Errorf("failed to get the right value for \"b\", got=%v, want=1", got)
	}
}

func TestGetNumber(t *testing.T) {
	m := New()
	m.Set("a", 3)
	m.Set("b", "nothing")
	if v, err := m.GetNumber("c"); err == nil {
		t.Errorf("got a numerical value for c=%v", v)
	}
	if v, err := m.GetNumber("a"); err != nil {
		t.Errorf("got no numerical value for a=%v", v)
	}
	if v, err := m.GetNumber("b"); err == nil {
		t.Errorf("got a numerical value for b=%v and no error", v)
	}
}

func TestAdd(t *testing.T) {
	m := New()
	m.Set("a", 4)
	m.Set("b", "two")
	m.Add("a", 1)
	m.Add("b", 7)
	if v, err := m.GetNumber("a"); err != nil {
		t.Errorf("unexpected error reading \"a\": %v", err)
	} else if v != 5 {
		t.Errorf("expected \"a\"=5, got=%g", v)
	}
	if v, err := m.GetNumber("b"); err != nil {
		t.Errorf("unexpected error reading \"b\": %v", err)
	} else if v != 7 {
		t.Errorf("expected \"b\"=7, got=%g", v)
	}
}

func TestDumpMDTable(t *testing.T) {
	m := New()
	m.Set("a", 4)
	m.Set("b", "two")
	then := time.Now()
	d := m.DumpMDTable()
	after := time.Now()
	lines := bytes.Split(d, []byte("\n"))
	if len(lines) != 5 || len(lines[4]) != 0 {
		t.Error(lines)
		t.Fatalf("bad number of lines: got=%d, want=4", len(lines))
	}
	when := strings.Split(string(lines[0]), "|")
	if len(when) != 2 {
		t.Fatalf("cannot split 0th line: %q", string(lines[0]))
	}
	clock := strings.Trim(when[1], " ")
	stamp, err := time.Parse(time.UnixDate, strings.TrimPrefix(clock, "value at "))
	if err != nil {
		t.Fatalf("fatal parsing: %s: %v", clock, err)
	}
	if stamp.Before(then.Truncate(time.Second)) || stamp.After(after.Truncate(time.Second)) {
		t.Errorf("wanted %q <= %q <= %q", then.Format(time.UnixDate), stamp, after.Format(time.UnixDate))
	}
	expect := []string{
		"----|------",
		"a | 4",
		"b | two",
	}
	for i, x := range lines[1:4] {
		if s := string(x); s != expect[i] {
			t.Errorf("got=%q want=%q", s, expect[i])
		}
	}
}

func TestSnaps(t *testing.T) {
	vs := New()
	for i := 10; i < 16; i++ {
		vs.Set(fmt.Sprintf("%X", i), i)
	}
	snaps := []*Snapshot{vs.Snap()}
	for i := 0; i < 48; i++ {
		time.Sleep(1 * time.Millisecond)
		vs.Add(fmt.Sprintf("%X", 10+(i%6)), float64(i))
		snaps = append(snaps, vs.Snap())
		snaps = Trim(snaps)
		delta := 0
		if i >= 1 {
			delta = -1
		}
		if i > 1 && len(snaps[i-1].Values.Detail) != 1 {
			t.Errorf("for [%d] expecting 1 entry, got %v", i, snaps[i-1].Values.Detail)
		}
		if got, want := len(snaps), 2+delta+i; got != want {
			t.Fatalf("wrong number of snapshots: got=%d, want=%d", got, want)
		}
	}

	from := snaps[0].When
	to := snaps[len(snaps)-1].When
	ks := []string{"A", "C", "F"}
	nums, err := ExtractNumbers(snaps, time.Millisecond, from, to, ks)
	if err != nil {
		t.Errorf("ExtractNumbers failed: %v", err)
	}
	if got, want := len(nums), len(snaps); got != want {
		t.Errorf("bad number of numbers: got=%d, want=%d", got, want)
	}

	lastTS := 0.0
	for i := 0; i < len(nums); i++ {
		a := nums[i]
		if a[0] <= lastTS {
			t.Errorf("[%d] non increasing timestamp: got=%f, want>%f", i, a[0], lastTS)
		}
		lastTS = a[0]
		if i == 0 {
			continue
		}
		changed := 0
		for j := 1; j < len(a); j++ {
			if was, is := nums[i-1][j], a[j]; was < is {
				changed++
			}
		}
		if changed > 1 {
			t.Errorf("[%d] too many changed, %d, want <= 1", i, changed)
		}
	}
}

func TestRate(t *testing.T) {
	samples := []struct {
		dt time.Duration
		v  float64
		r  float64
	}{
		{dt: time.Second, v: 1, r: 0},
		{dt: 2 * time.Second, v: 2, r: 1},
		{dt: 5 * time.Second, v: 4, r: 0.75},
	}
	now := time.Now()
	var pts []Sample
	for i, s := range samples {
		x := Sample{When: now.Add(s.dt), Value: s.v}
		pts = append(pts, x)
		if got, want := Rate(pts...), s.r; got != want {
			t.Errorf("[%d] mismatch: got=%f, want=%f", i, got, want)
		}
	}
}
