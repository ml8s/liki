package tianwen

import (
	"sync"
	"testing"
)

func TestBoundedMemoEvictsOldestInsertion(t *testing.T) {
	memo := newBoundedMemo[int, int](2)
	memo.put(1, 10)
	memo.put(2, 20)
	memo.put(3, 30)

	if _, exists := memo.get(1); exists {
		t.Fatal("oldest insertion was retained after eviction")
	}
	for key, want := range map[int]int{2: 20, 3: 30} {
		got, exists := memo.get(key)
		if !exists || got != want {
			t.Fatalf("get(%d) = (%d, %v), want (%d, true)", key, got, exists, want)
		}
	}
}

func TestLunarPureFunctionMemoIsTransparent(t *testing.T) {
	k := 1300.0
	first := getNewMoonJDTD(k)
	second := getNewMoonJDTD(k)
	if first != second {
		t.Fatalf("memoized new moon changed value: %v != %v", first, second)
	}
}

func TestBoundedMemoConcurrentAccessIsSafe(t *testing.T) {
	const workers = 8
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(k float64) {
			defer wg.Done()
			first := getNewMoonJDTD(k)
			second := getNewMoonJDTD(k)
			if first != second {
				t.Errorf("memoized new moon changed value for k=%v: %v != %v", k, first, second)
			}
		}(float64(1200 + i))
	}
	wg.Wait()
}
