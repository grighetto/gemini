package gemini

import (
	"math/rand"
	"testing"
	"testing/quick"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func TestNonEmptyRandRange(t *testing.T) {
	f := func(x, y int) bool {
		r := nonEmptyRandIntRange(rnd, x, y, 10)
		return r > 0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestNonEmptyRandRange64(t *testing.T) {
	f := func(x, y int) bool {
		r := nonEmptyRandIntRange(rnd, x, y, 10)
		return r > 0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestNonEmptyRandFloat32Range(t *testing.T) {
	f := func(x, y float32) bool {
		r := nonEmptyRandFloat32Range(rnd, x, y, 10)
		return r > 0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestNonEmptyRandFloat64Range(t *testing.T) {
	f := func(x, y float64) bool {
		r := nonEmptyRandFloat64Range(rnd, x, y, 10)
		return r > 0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestNonEmptyRandString(t *testing.T) {
	// TODO: Figure out why this is so horribly slow...
	tt := time.Now()
	f := func(len int32) bool {
		r := nonEmptyRandStringWithTime(rnd, int(len), tt)
		return r != ""
	}
	cfg := &quick.Config{MaxCount: 10}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

var bench_r string

func BenchmarkNonEmptyRandStringWithTime(b *testing.B) {
	tt := time.Now()
	for i := 0; i < b.N; i++ {
		bench_r = nonEmptyRandStringWithTime(rnd, 30, tt)
	}
}

func BenchmarkNonEmptyRandStringWithTimeParallel(b *testing.B) {
	tt := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bench_r = nonEmptyRandStringWithTime(rnd, 30, tt)
		}
	})
}

var bench_rr int

func BenchmarkNonEmptyRandRange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bench_rr = nonEmptyRandIntRange(rnd, 0, 50, 30)
	}
}

var bench_rr64 int64

func BenchmarkNonEmptyRandRange64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bench_rr64 = nonEmptyRandInt64Range(rnd, 0, 50, 30)
	}
}
