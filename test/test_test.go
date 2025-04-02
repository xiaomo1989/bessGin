package main

import (
	"testing"
)

func BenchmarkDemo(b *testing.B) {
	demoMap := map[string]string{
		"a": "a",
		"b": "b",
	}
	// 模拟并发写map
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			demoMap["a"] = "aa"
		}
	})
}

// BenchmarkDemo
// fatal error: concurrent map writes
// fatal error: concurrent map writes
