package jobcenter

type Queue struct {
	MaxHeap *MaxHeap[Job, int]
}

type Job struct {
}
