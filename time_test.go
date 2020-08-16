package main

import (
	"testing"
)

func TestParseTime(t *testing.T) {
	cs := []struct {
		v   string
		exp string
	}{
		{v: "2020-08-16", exp: "2020-08-16T00:00:00Z"},
		{v: "2020-08-16+09:00", exp: "2020-08-15T15:00:00Z"},
		{v: "2020-08-16T00:00:00+09:00", exp: "2020-08-15T15:00:00Z"},
		{v: "2020-08-16T00:00:00Z", exp: "2020-08-16T00:00:00Z"},
		{v: "20200816", exp: "2020-08-16T00:00:00Z"},
		{v: "20200816+09:00", exp: "2020-08-15T15:00:00Z"},
	}
	for _, e := range cs {
		v := iso8601utc(parseTime(e.v))
		if e.exp != v {
			t.Fatalf("%v %v", e, v)
		}
	}
}
