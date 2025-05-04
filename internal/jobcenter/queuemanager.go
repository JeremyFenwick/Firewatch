package jobcenter

import (
	"log"
	"sync"
)

type QueueManager struct {
	Mutex        sync.Mutex
	Queues       map[string]*MaxHeap[*Job, int] // Maps queue name to MaxHeap
	JobLocations map[int]string                 // Maps job id to queue name
	CurrentId    int
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		Queues:       make(map[string]*MaxHeap[*Job, int]),
		JobLocations: make(map[int]string),
	}
}

func (qm *QueueManager) GetNextId() int {
	qm.Mutex.Lock()
	defer qm.Mutex.Unlock()

	qm.CurrentId++
	return qm.CurrentId
}

func (qm *QueueManager) QueueExists(queueName string) bool {
	_, exists := qm.Queues[queueName]
	return exists
}

// PutJob adds a job to the specified queue.
// If the queue does not exist, it creates a new one.
func (qm *QueueManager) PutJob(queueName string, job *Job) {
	if !qm.QueueExists(queueName) {
		qm.Queues[queueName] = NewMaxHeap(
			func(a, b *Job) bool { return a.Priority > b.Priority },
			func(id int, job *Job) bool { return id == job.Id },
		)
	}
	qm.Queues[queueName].Push(job)
	qm.JobLocations[job.Id] = queueName
}

// GetPriorityJob retrieves the job with the highest priority from the specified queues.
// If multiple queues are specified, it returns the job with the highest priority across all specified queues.
func (qm *QueueManager) GetPriorityJob(queues ...string) (*Job, bool) {
	if len(queues) == 0 {
		return nil, false
	}
	var candidate *Job
	for _, queue := range queues {
		if !qm.QueueExists(queue) {
			continue
		}
		job, found := qm.Queues[queue].Peek()
		if !found {
			continue
		}
		if candidate == nil || job.Priority > candidate.Priority {
			candidate = job
			continue
		}
	}
	if candidate == nil {
		return nil, false
	}
	delete(qm.JobLocations, candidate.Id)
	qm.Queues[candidate.Queue].Pop()
	return candidate, true
}

// GetJob retrieves a job from the specified queue.
// Returns whether the job was found and deleted successfully.
func (qm *QueueManager) DeleteJob(jobId int) bool {
	queueName, exists := qm.JobLocations[jobId]
	if !exists {
		return false
	}
	removed := qm.Queues[queueName].Delete(jobId)
	if !removed {
		log.Printf("Failed to delete job %d from queue %s. It was in the job locations map but not the queue itself.", jobId, queueName)
		return false
	}
	delete(qm.JobLocations, jobId)
	return true
}
