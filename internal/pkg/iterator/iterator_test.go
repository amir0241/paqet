package iterator

import (
	"testing"
)

func TestIterator_EmptyItems(t *testing.T) {
	// Test Next() with empty items
	iter := &Iterator[int]{}
	result := iter.Next()
	if result != 0 {
		t.Errorf("Next() with empty items should return zero value, got %v", result)
	}
	
	// Test Peek() with empty items
	result = iter.Peek()
	if result != 0 {
		t.Errorf("Peek() with empty items should return zero value, got %v", result)
	}
}

func TestIterator_EmptyItemsPointer(t *testing.T) {
	// Test with pointer type to ensure nil is returned
	type Item struct {
		Value int
	}
	
	iter := &Iterator[*Item]{}
	result := iter.Next()
	if result != nil {
		t.Errorf("Next() with empty items should return nil for pointer type, got %v", result)
	}
	
	result = iter.Peek()
	if result != nil {
		t.Errorf("Peek() with empty items should return nil for pointer type, got %v", result)
	}
}

func TestIterator_WithItems(t *testing.T) {
	// Test normal operation with items
	iter := &Iterator[int]{Items: []int{1, 2, 3}}
	
	// Test Next() cycles through items
	if got := iter.Next(); got != 2 {
		t.Errorf("Expected 2, got %d", got)
	}
	if got := iter.Next(); got != 3 {
		t.Errorf("Expected 3, got %d", got)
	}
	if got := iter.Next(); got != 1 {
		t.Errorf("Expected 1 (wrap around), got %d", got)
	}
}

func TestIterator_Peek(t *testing.T) {
	// Test Peek() doesn't advance the iterator
	iter := &Iterator[int]{Items: []int{10, 20, 30}}
	
	if got := iter.Peek(); got != 10 {
		t.Errorf("Expected 10, got %d", got)
	}
	if got := iter.Peek(); got != 10 {
		t.Errorf("Peek should not advance, expected 10, got %d", got)
	}
	
	// Now advance with Next()
	if got := iter.Next(); got != 20 {
		t.Errorf("Expected 20, got %d", got)
	}
	if got := iter.Peek(); got != 20 {
		t.Errorf("Expected 20 after Next(), got %d", got)
	}
}
