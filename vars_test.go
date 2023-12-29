package vars

import (
	"bytes"
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
