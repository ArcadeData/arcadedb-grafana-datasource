package plugin

import (
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestExpandMacros_TimeFrom(t *testing.T) {
	tr := backend.TimeRange{
		From: time.UnixMilli(1711900800000),
		To:   time.UnixMilli(1711987200000),
	}
	result := ExpandMacros("SELECT * FROM t WHERE ts >= $__timeFrom", tr, time.Minute)
	expected := "SELECT * FROM t WHERE ts >= 1711900800000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandMacros_TimeTo(t *testing.T) {
	tr := backend.TimeRange{
		From: time.UnixMilli(1711900800000),
		To:   time.UnixMilli(1711987200000),
	}
	result := ExpandMacros("SELECT * FROM t WHERE ts <= $__timeTo", tr, time.Minute)
	expected := "SELECT * FROM t WHERE ts <= 1711987200000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandMacros_TimeFilter(t *testing.T) {
	tr := backend.TimeRange{
		From: time.UnixMilli(1000),
		To:   time.UnixMilli(5000),
	}
	result := ExpandMacros("SELECT * FROM t WHERE $__timeFilter(ts)", tr, time.Minute)
	expected := "SELECT * FROM t WHERE ts >= 1000 AND ts <= 5000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandMacros_Interval(t *testing.T) {
	tr := backend.TimeRange{
		From: time.UnixMilli(0),
		To:   time.UnixMilli(1000),
	}

	tests := []struct {
		interval time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{2 * time.Hour, "2h"},
	}

	for _, tc := range tests {
		result := ExpandMacros("BUCKET $__interval", tr, tc.interval)
		exp := "BUCKET " + tc.expected
		if result != exp {
			t.Errorf("interval %v: expected %q, got %q", tc.interval, exp, result)
		}
	}
}

func TestExpandMacros_NoMacros(t *testing.T) {
	tr := backend.TimeRange{
		From: time.UnixMilli(0),
		To:   time.UnixMilli(1000),
	}
	query := "SELECT name FROM Person LIMIT 10"
	result := ExpandMacros(query, tr, time.Minute)
	if result != query {
		t.Errorf("expected no changes, got %q", result)
	}
}
