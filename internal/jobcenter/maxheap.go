package jobcenter

import (
	"log"
	"slices"
	"sync"
)

type MaxHeap[T any, I any] struct {
	mutex       sync.RWMutex
	data        []T
	greaterThan func(a, b T) bool   // Comparator function (a > b)
	identify    func(a I, b T) bool // Function to identify the element
}

// API

// NewMaxHeap creates a new MaxHeap of values of type T with the given comparator function.
// Note the comparator function should return true if the first argument is GREATER than the second.
//
// Also takes an Identify function which returns true if the T matches given the argument I.
// Example: if you have a struct with an ID field, you can use this function to identify the element in the heap:
// func(id int, data T) bool {data.ID == id},
//
// The initial capacity of the heap can be specified. If not provided, a default capacity of 10 is used.
//
// This data structure is thread safe and can be used concurrently.
func NewMaxHeap[T any, I any](greaterThan func(a, b T) bool, Identify func(a I, b T) bool, initialCapacity ...int) *MaxHeap[T, I] {
	var capSize int
	if len(initialCapacity) > 0 {
		capSize = initialCapacity[0]
	} else {
		capSize = 10
	}
	return &MaxHeap[T, I]{data: make([]T, 0, capSize), greaterThan: greaterThan, identify: Identify}
}

// Heapify creates a new MaxHeap from the provided data slice.
//
// Note the greaterThan function should return true if the first argument is GREATER than the second.
//
// The identify function should return true if the T matches given the argument I.
// Example: if you have a struct with an ID field, you can use this function to identify the element in the heap:
// func(id int, data T) bool {data.ID == id},
//
// This data structure is thread safe and can be used concurrently.
func Heapify[T any, I any](data []T, greaterThan func(a, b T) bool, identify func(a I, b T) bool) *MaxHeap[T, I] {
	mh := NewMaxHeap(greaterThan, identify)
	// Create a copy of the data to prevent modifying the original data
	mh.data = slices.Clone(data) // This creates a new slice with the same elements as `data`

	// Perform siftDown starting from the last non-leaf node down to the root node
	for i := (mh.Size() - 2) / 2; i >= 0; i-- {
		mh.siftDown(i)
	}
	return mh
}

// IsEmpty checks if the heap is empty.
func (mh *MaxHeap[T, I]) IsEmpty() bool {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()

	return len(mh.data) == 0
}

// Size returns the number of elements in the heap.
func (mh *MaxHeap[T, I]) Size() int {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()

	return len(mh.data)
}

// Peek returns the maximum element in the heap without removing it.
func (mh *MaxHeap[T, I]) Peek() (T, bool) {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()

	if len(mh.data) == 0 {
		var none T
		return none, false
	}
	return mh.data[0], true
}

// Push adds a new element to the heap and maintains the heap property.
// Note that this function does not check for duplicates.
func (mh *MaxHeap[T, I]) Push(item T) {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()

	mh.data = append(mh.data, item)
	size := len(mh.data)
	mh.siftUp(size - 1)
}

// Pop removes and returns the maximum element from the heap, maintaining the heap property.
func (mh *MaxHeap[T, I]) Pop() (T, bool) {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()

	if len(mh.data) == 0 {
		var none T
		return none, false
	}
	lastIndex := len(mh.data) - 1
	mh.swap(0, lastIndex)
	item := mh.data[lastIndex]
	mh.data = mh.data[:lastIndex]
	mh.siftDown(0)
	return item, true
}

// Clear removes all elements from the heap.
func (mh *MaxHeap[T, I]) Clear() {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()

	mh.data = mh.data[:0]
}

// Delete removes the first matching element from the heap based on the provided value.
// The value is identified using the Identify function.
//
// Note, this has O(n) time complexity due to the need to find the element first.
func (mh *MaxHeap[T, I]) Delete(value I) bool {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()

	if len(mh.data) == 0 {
		return false
	}

	// Find the index first. Use a variable outside the loop scope.
	foundIndex := -1
	for i, item := range mh.data {
		if mh.identify(value, item) {
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
		mh.data = mh.data[:lastIndex]
	} else {
		// Case 2: Deleting an element before the end.
		// Swap with the last element.
		mh.swap(foundIndex, lastIndex)
		// Shrink the slice (removes the element originally at foundIndex).
		mh.data = mh.data[:lastIndex]

		// Now, heapify the element that moved into foundIndex.
		// It's safe to access mh.Data[foundIndex] because foundIndex < lastIndex here.
		parentIndex := (foundIndex - 1) / 2
		if foundIndex > 0 && mh.greaterThan(mh.data[foundIndex], mh.data[parentIndex]) {
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
//
// Note, this has O(n) time complexity due to the linear scan.
func (mh *MaxHeap[T, I]) Contains(value I) bool {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()

	if len(mh.data) == 0 {
		return false
	}
	for _, item := range mh.data {
		if mh.identify(value, item) {
			return true
		}
	}
	return false
}

// Update finds an element in the heap using the Identify function, updates it with the new value,
// then restores the heap property. Returns true if the element was found and updated.
//
// Note that this has O(n) time complexity due to the need to find the element first.
func (mh *MaxHeap[T, I]) Update(id I, newValue T) bool {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()

	if len(mh.data) == 0 {
		return false
	}

	// Find the element
	index := -1
	for i, item := range mh.data {
		if mh.identify(id, item) {
			index = i
			break
		}
	}

	if index == -1 {
		return false // Element not found
	}

	// Save the old value for comparison
	oldValue := mh.data[index]

	// Update the value
	mh.data[index] = newValue

	// Restore heap property
	if index > 0 {
		parentIndex := (index - 1) / 2
		if mh.greaterThan(mh.data[index], mh.data[parentIndex]) {
			// If the new value is greater than its parent, sift up
			mh.siftUp(index)
			return true
		}
	}

	// If we didn't sift up, we might need to sift down
	// Only sift down if the new value is less than the old value
	if mh.greaterThan(oldValue, newValue) {
		mh.siftDown(index)
	}

	return true
}

// Compact reduces the capacity of the underlying slice to match its length,
// potentially freeing memory when many elements have been removed.
// This is useful after many removals to reduce memory usage.
func (mh *MaxHeap[T, I]) Compact() {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()

	if len(mh.data) == 0 {
		mh.data = make([]T, 0) // Reset to empty slice with minimal capacity
		return
	}

	// Create a new slice with exact capacity needed and copy elements
	newData := make([]T, len(mh.data))
	copy(newData, mh.data)
	mh.data = newData
}

// UpdateOrPush updates an element if it exists in the heap (identified by id),
// or pushes the new element if it doesn't exist.
// Returns true if an update occurred, false if a push occurred.
//
// Note that this has O(n) time complexity due to the need to find the element first.
func (mh *MaxHeap[T, I]) UpdateOrPush(id I, value T) bool {
	if mh.Update(id, value) {
		return true // Updated existing element
	}

	// Element not found, push the new value
	mh.Push(value)
	return false
}

// INTERNAL OPERATIONS

func (mh *MaxHeap[T, I]) lastIndex() int {
	if len(mh.data) == 0 {
		return -1
	}
	return len(mh.data) - 1
}

func (mh *MaxHeap[T, I]) swap(i, j int) {
	size := len(mh.data)
	if i < 0 || i >= size || j < 0 || j >= size {
		log.Printf("Swap indices %d and %d out of bounds for heap of size %d", i, j, size)
		return
	}
	mh.data[i], mh.data[j] = mh.data[j], mh.data[i]
}

func (mh *MaxHeap[T, I]) siftUp(index int) {
	size := len(mh.data)
	if index < 0 || index >= size {
		log.Printf("Index %d out of bounds for heap of size %d", index, size)
		return
	}
	for index > 0 {
		parentIndex := (index - 1) / 2
		if mh.greaterThan(mh.data[parentIndex], mh.data[index]) {
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

		if leftChildIndex <= lastIndex && mh.greaterThan(mh.data[leftChildIndex], mh.data[largestIndex]) {
			largestIndex = leftChildIndex
		}
		if rightChildIndex <= lastIndex && mh.greaterThan(mh.data[rightChildIndex], mh.data[largestIndex]) {
			largestIndex = rightChildIndex
		}
		if largestIndex == index {
			break
		}
		mh.swap(index, largestIndex)
		index = largestIndex
	}
}
