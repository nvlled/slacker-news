package main

import (
	"fmt"
	"testing"
	"time"
)

func TestNextUpdate(t *testing.T) {
	cm := NewCacheManager()
	cm.LastUpdate = time.Date(2024, time.January, 5 /*time:*/, 8, 23, 0, 0, time.Local).Unix()
	fmt.Printf("%v -> %v\n", time.Unix(cm.LastUpdate, 0), time.Unix(cm.GetNextUpdate(), 0))

	cm.LastUpdate = time.Date(2024, time.January, 5 /*time:*/, 8, 30, 0, 0, time.Local).Unix()
	fmt.Printf("%v -> %v\n", time.Unix(cm.LastUpdate, 0), time.Unix(cm.GetNextUpdate(), 0))

	cm.LastUpdate = time.Date(2024, time.January, 5 /*time:*/, 8, 31, 0, 0, time.Local).Unix()
	fmt.Printf("%v -> %v\n", time.Unix(cm.LastUpdate, 0), time.Unix(cm.GetNextUpdate(), 0))

	cm.LastUpdate = time.Date(2024, time.January, 5 /*time:*/, 8, 41, 0, 0, time.Local).Unix()
	fmt.Printf("%v -> %v\n", time.Unix(cm.LastUpdate, 0), time.Unix(cm.GetNextUpdate(), 0))

	cm.LastUpdate = time.Date(2024, time.January, 5 /*time:*/, 8, 51, 0, 0, time.Local).Unix()
	fmt.Printf("%v -> %v\n", time.Unix(cm.LastUpdate, 0), time.Unix(cm.GetNextUpdate(), 0))

	cm.LastUpdate = time.Date(2024, time.January, 5 /*time:*/, 9, 00, 0, 0, time.Local).Unix()
	fmt.Printf("%v -> %v\n", time.Unix(cm.LastUpdate, 0), time.Unix(cm.GetNextUpdate(), 0))

	cm.LastUpdate = time.Date(2024, time.January, 5 /*time:*/, 9, 01, 0, 0, time.Local).Unix()
	fmt.Printf("%v -> %v\n", time.Unix(cm.LastUpdate, 0), time.Unix(cm.GetNextUpdate(), 0))
}
