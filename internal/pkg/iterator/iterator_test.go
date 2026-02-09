package iterator

import (
	"testing"
)

func TestIterator_NextWithEmptyItems(t *testing.T) {
	it := &Iterator[int]{}
	
	// Should not panic and should return zero value
	result := it.Next()
	if result != 0 {
		t.Errorf("Expected zero value (0) for empty iterator, got %d", result)
	}
}

func TestIterator_PeekWithEmptyItems(t *testing.T) {
	it := &Iterator[int]{}
	
	// Should not panic and should return zero value
	result := it.Peek()
	if result != 0 {
		t.Errorf("Expected zero value (0) for empty iterator, got %d", result)
	}
}

func TestIterator_NextWithItems(t *testing.T) {
	it := &Iterator[int]{
		Items: []int{1, 2, 3},
	}
	
	// Test round-robin behavior
	expected := []int{2, 3, 1, 2} // First call increments to 1, so starts at index 1
	for i, exp := range expected {
		result := it.Next()
		if result != exp {
			t.Errorf("Call %d: expected %d, got %d", i+1, exp, result)
		}
	}
}

func TestIterator_PeekWithItems(t *testing.T) {
	it := &Iterator[int]{
		Items: []int{1, 2, 3},
	}
	
	// Peek should return the same value without advancing
	for i := 0; i < 3; i++ {
		result := it.Peek()
		if result != 1 {
			t.Errorf("Peek call %d: expected 1, got %d", i+1, result)
		}
	}
}

func TestIterator_NextWithPointers(t *testing.T) {
	type testStruct struct {
		value int
	}
	
	it := &Iterator[*testStruct]{}
	
	// Should not panic and should return nil for empty iterator
	result := it.Next()
	if result != nil {
		t.Errorf("Expected nil for empty iterator, got %v", result)
	}
}

func TestIterator_NextWithPowerOfTwo(t *testing.T) {
	// Test with power of 2 length (uses bitwise operation)
	it := &Iterator[int]{
		Items: []int{1, 2, 3, 4}, // length is 4 (power of 2)
	}
	
	expected := []int{2, 3, 4, 1} // Should cycle through items
	for i, exp := range expected {
		result := it.Next()
		if result != exp {
			t.Errorf("Call %d: expected %d, got %d", i+1, exp, result)
		}
	}
}
