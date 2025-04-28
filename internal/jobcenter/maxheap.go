package jobcenter

import (
	"log"
	"slices"
)

type MaxHeap[T any, I any] struct {
	Data        []T
	GreaterThan func(a, b T) bool   // Comparator function (a > b)
	Identify    func(a I, b T) bool // Function to identify the element
}

// API

// NewMaxHeap creates a new MaxHeap of values of type T with the given comparator function.
// Note the comparator function should return true if the first argument is GREATER than the second.
// Also takes an Identify function to identify the element in the heap.
// The identify function takes a type I and a type T and returns true if the element matches.
// Example: if you have a struct with an ID field, you can use this function to identify the element in the heap.
// The initial capacity of the heap can be specified. If not provided, a default capacity of 10 is used.
func NewMaxHeap[T any, I any](greaterThan func(a, b T) bool, Identify func(a I, b T) bool, initialCapacity ...int) *MaxHeap[T, I] {
	var capSize int
	if len(initialCapacity) > 0 {
		capSize = initialCapacity[0]
	} else {
		capSize = 10
	}
	return &MaxHeap[T, I]{Data: make([]T, 0, capSize), GreaterThan: greaterThan, Identify: Identify}
}

// Heapify creates a new MaxHeap from the provided data slice.
// Note the comparator function should return true if the first argument is greater than the second.
func Heapify[T any, I any](data []T, greaterThan func(a, b T) bool, identify func(a I, b T) bool) *MaxHeap[T, I] {
	mh := NewMaxHeap(greaterThan, identify)
	// Create a copy of the data to prevent modifying the original data
	mh.Data = slices.Clone(data) // This creates a new slice with the same elements as `data`

	// Perform siftDown starting from the last non-leaf node down to the root node
	for i := (mh.Size() - 2) / 2; i >= 0; i-- {
		mh.siftDown(i)
	}
	return mh
}

// IsEmpty checks if the heap is empty.
func (mh *MaxHeap[T, I]) IsEmpty() bool {
	return len(mh.Data) == 0
}

// Size returns the number of elements in the heap.
func (mh *MaxHeap[T, I]) Size() int {
	return len(mh.Data)
}

// Peek returns the maximum element in the heap without removing it.
func (mh *MaxHeap[T, I]) Peek() (T, bool) {
	if len(mh.Data) == 0 {
		var none T
		return none, false
	}
	return mh.Data[0], true
}

// Push adds a new element to the heap and maintains the heap property.
func (mh *MaxHeap[T, I]) Push(item T) {
	mh.Data = append(mh.Data, item)
	mh.siftUp(mh.Size() - 1)
}

// Pop removes and returns the maximum element from the heap, maintaining the heap property.
func (mh *MaxHeap[T, I]) Pop() (T, bool) {
	if mh.IsEmpty() {
		var none T
		return none, false
	}
	lastIndex := mh.Size() - 1
	mh.swap(0, lastIndex)
	item := mh.Data[lastIndex]
	mh.Data = mh.Data[:lastIndex]
	mh.siftDown(0)
	return item, true
}

// Clear removes all elements from the heap.
func (mh *MaxHeap[T, I]) Clear() {
	mh.Data = mh.Data[:0]
}

// Delete removes an element from the heap based on the provided value.
// The value is identified using the Identify function.
func (mh *MaxHeap[T, I]) Delete(value I) bool {
	if mh.IsEmpty() {
		return false
	}

	// Find the index first. Use a variable outside the loop scope.
	foundIndex := -1
	for i, item := range mh.Data {
		if mh.Identify(value, item) {
			foundIndex = i
			break // Exit loop once found
		}
	}

	// If not found, return false
	if foundIndex == -1 {
		return false
	}

	// Now perform the deletion logic using foundIndex
	lastIndex := mh.lastIndex()

	if foundIndex == lastIndex {
		// Case 1: Deleting the very last element.
		// Just shrink the slice. No swap/heapify needed.
		mh.Data = mh.Data[:lastIndex]
	} else {
		// Case 2: Deleting an element before the end.
		// Swap with the last element.
		mh.swap(foundIndex, lastIndex)
		// Shrink the slice (removes the element originally at foundIndex).
		mh.Data = mh.Data[:lastIndex]

		// Now, heapify the element that moved into foundIndex.
		// It's safe to access mh.Data[foundIndex] because foundIndex < lastIndex here.
		parentIndex := (foundIndex - 1) / 2
		if foundIndex > 0 && mh.GreaterThan(mh.Data[foundIndex], mh.Data[parentIndex]) {
			// If it's greater than its parent, sift it up.
			mh.siftUp(foundIndex)
		} else {
			// Otherwise, sift it down.
			mh.siftDown(foundIndex)
		}
	}

	return true // Item found and removed
}

// Contains checks if the heap contains an element based on the provided value.
func (mh *MaxHeap[T, I]) Contains(value I) bool {
	if mh.IsEmpty() {
		return false
	}
	for _, item := range mh.Data {
		if mh.Identify(value, item) {
			return true
		}
	}
	return false
}

// INTERNAL OPERATIONS

func (mh *MaxHeap[T, I]) lastIndex() int {
	if mh.IsEmpty() {
		return -1
	}
	return mh.Size() - 1
}

func (mh *MaxHeap[T, I]) swap(i, j int) {
	if i < 0 || i >= mh.Size() || j < 0 || j >= mh.Size() {
		log.Printf("Swap indices %d and %d out of bounds for heap of size %d", i, j, mh.Size())
		return
	}
	mh.Data[i], mh.Data[j] = mh.Data[j], mh.Data[i]
}

func (mh *MaxHeap[T, I]) siftUp(index int) {
	if index < 0 || index >= mh.Size() {
		log.Printf("Index %d out of bounds for heap of size %d", index, mh.Size())
		return
	}
	for index > 0 {
		parentIndex := (index - 1) / 2
		if mh.GreaterThan(mh.Data[parentIndex], mh.Data[index]) {
			break
		}
		mh.swap(index, parentIndex)
		index = parentIndex
	}
}

func (mh *MaxHeap[T, I]) siftDown(index int) {
	lastIndex := mh.lastIndex()
	if lastIndex < 0 {
		log.Printf("Attempting to sift down an empty heap")
		return
	}
	if index < 0 || index > lastIndex {
		log.Printf("Index %d out of bounds for heap of size %d", index, lastIndex+1)
		return
	}
	for {
		leftChildIndex := 2*index + 1
		rightChildIndex := 2*index + 2
		largestIndex := index

		if leftChildIndex <= lastIndex && mh.GreaterThan(mh.Data[leftChildIndex], mh.Data[largestIndex]) {
			largestIndex = leftChildIndex
		}
		if rightChildIndex <= lastIndex && mh.GreaterThan(mh.Data[rightChildIndex], mh.Data[largestIndex]) {
			largestIndex = rightChildIndex
		}
		if largestIndex == index {
			break
		}
		mh.swap(index, largestIndex)
		index = largestIndex
	}
}
