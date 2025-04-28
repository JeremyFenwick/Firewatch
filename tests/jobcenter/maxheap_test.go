package jobcenter_test

import (
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/jobcenter"
	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Name  string
	Value int
}

func CreateMaxHeap() *jobcenter.MaxHeap[TestData, string] {
	compare := func(a, b TestData) bool {
		return a.Value > b.Value
	}
	identify := func(name string, data TestData) bool {
		return name == data.Name
	}
	return jobcenter.NewMaxHeap(compare, identify)
}

func TestMaxHeapPeek(t *testing.T) {
	mh := CreateMaxHeap()
	mh.Push(TestData{"First Entry", 1})
	assert.Equal(t, 1, mh.Size())
	peekValue, returnPeek := mh.Peek()
	assert.True(t, returnPeek)
	assert.Equal(t, TestData{"First Entry", 1}, peekValue)
	mh.Push(TestData{"Second Entry", 2})
	assert.Equal(t, 2, mh.Size())
	peekValue, returnPeek = mh.Peek()
	assert.True(t, returnPeek)
	assert.Equal(t, TestData{"Second Entry", 2}, peekValue)
	mh.Push(TestData{"Third Entry", 3})
	assert.Equal(t, 3, mh.Size())
	peekValue, returnPeek = mh.Peek()
	assert.True(t, returnPeek)
	assert.Equal(t, TestData{"Third Entry", 3}, peekValue)
	mh.Clear()
	assert.Equal(t, 0, mh.Size())
	_, returnPeek = mh.Peek()
	assert.False(t, returnPeek)
}

func TestMaxHeapPop(t *testing.T) {
	mh := CreateMaxHeap()
	mh.Push(TestData{"First Entry", 1})
	mh.Push(TestData{"Second Entry", 2})
	mh.Push(TestData{"Third Entry", 3})
	popValue, returnPop := mh.Pop()
	assert.True(t, returnPop)
	assert.Equal(t, TestData{"Third Entry", 3}, popValue)
	assert.Equal(t, 2, mh.Size())
	peekValue, returnPeek := mh.Peek()
	assert.True(t, returnPeek)
	assert.Equal(t, TestData{"Second Entry", 2}, peekValue)
	mh.Clear()
	assert.Equal(t, 0, mh.Size())
	_, returnPop = mh.Pop()
	assert.False(t, returnPop)
}

func TestMaxHeapDelete(t *testing.T) {
	mh := CreateMaxHeap()
	mh.Push(TestData{"First Entry", 1})
	mh.Push(TestData{"Second Entry", 2})
	mh.Push(TestData{"Third Entry", 3})
	assert.Equal(t, 3, mh.Size())
	assert.True(t, mh.Delete("Second Entry"))
	assert.Equal(t, 2, mh.Size())
	containsDeleted := mh.Contains("Second Entry")
	assert.False(t, containsDeleted)
	peekValue, returnPeek := mh.Peek()
	assert.True(t, returnPeek)
	assert.Equal(t, TestData{"Third Entry", 3}, peekValue)
}
