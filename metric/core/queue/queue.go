package queue

import (
	"container/heap"
	"core/helpers"
	"core/models"
	"core/repositories"
	"core/structures"
	"core/tasks"
	"errors"
	"time"
)

type Queue struct {
	Limit        int // Maximum cumulative computational weight of all tasks ran at the same time
	load         int // Current cumulative computational weight
	snapshotRepo *repositories.SnapshotRepository

	/*
		    used for storing active tasks' data. Used in case the server crashes.
			In that case, the running tasks are restored from the database.
	*/
	taskRepo *repositories.TaskRepository

	/*
		queue is the task queue. It consists of pointers to models.Task defined by Queue.tasks.
		It shouldn't modify tasks, only rearrange them.
	*/
	queue structures.PriorityQueue

	/*
		    tasks is the task map. Required because we sometimes need to change some task data
			when it's not in queue (is running). It's supposed to be the single source of truth for task data.
	*/
	tasks structures.TaskMap // Map of tasks

	/*
		    a channel that will send finished tasks through gRPC. Used in ForceExecute to
			wait until the task is finished
	*/
	taskFinished chan *models.Task
}

func NewQueue(limit int, snapshotRepo *repositories.SnapshotRepository, taskRepo *repositories.TaskRepository) *Queue {
	return &Queue{Limit: limit, load: 0, snapshotRepo: snapshotRepo, taskRepo: taskRepo, taskFinished: make(chan *models.Task)}
}

func (q *Queue) Start() {
	// Initialize structures
	q.queue = structures.PriorityQueue{}
	q.tasks = *structures.NewTaskMap()
	heap.Init(&q.queue)

	// Fetch tasks from DB.
	tasks, err := q.taskRepo.GetAll()

	// No point in continuing if we have troubles with DB
	if err != nil {
		panic(err)
	}

	// If there were tasks in DB, add them to structures
	for _, task := range tasks {
		heap.Push(&q.queue, task)
		q.tasks.AddTask(task)
	}

	// Start the infinite loop for advancing the queue
	go q.AdvanceTasks()
}

func (q *Queue) AddTask(data *models.TaskCreate) *models.Task {

	// Check if the task is already running
	activeTasks := q.tasks.GetByGroups(data.Groups)
	for _, task := range activeTasks {
		if task.Metric == data.Metric {
			// Task is already running
			return nil
		}
	}

	// Initialize task
	task := models.NewTask(data)
	task.AttemptedAt = task.UpdatedAt

	// Add task to structures
	q.tasks.AddTask(&task)
	heap.Push(&q.queue, &task)

	// Add task to DB
	q.taskRepo.Create(&task)

	return &task
}

func (q *Queue) UpdateGroupName(old string, new string) error {
	// Update task in the map
	updatedTasks := q.tasks.UpdateGroupName(old, new)

	// Update task in the DB
	for _, task := range updatedTasks {
		q.taskRepo.Update(task)
	}

	// Update all snapshots that use this group name
	err := q.snapshotRepo.UpdateGroupName(old, new)

	return err
}

func (q *Queue) UpdateByGroupName(group string, createdAt time.Time, deletedAt time.Time) []*models.Task {

	// Update tasks in the map
	updatedTasks := q.tasks.UpdateByGroupName(group, createdAt, deletedAt)

	// Update tasks in the DB
	for _, task := range updatedTasks {
		q.taskRepo.Update(task)
	}

	return updatedTasks
}

func (q *Queue) ForceExecute(task *models.TaskCreate, groups []string) (*models.Task, error) {

	// Update AttemptedAt and UpdatedAt in the map
	_, err := q.tasks.ForceUpdate(task.Metric, groups)

	// If we couldn't find the task in the map, add it
	if err != nil {
		q.AddTask(task)
	}

	// Rebuild the queue to account for the new AttemptedAt
	q.queue.Rebuild()

	// Wait for the task to be finished
	resultChan := make(chan *models.Task, 1)

	for task := range q.taskFinished {

		// Check if this is the task we're looking for
		if !helpers.ContainsAllElements(task.Groups, groups) {
			continue
		}
		resultChan <- task
		break
	}

	select {
	case task := <-resultChan:
		// The task has finished
		return task, nil
	case <-time.After(10 * time.Second):
		// The task hasn't finished in 10 seconds
		return nil, errors.New("timeout")
	}
}

func (q *Queue) DeleteTask(metric string, groups []string, deleteSnapshots bool) (*models.Task, error) {

	// Delete snapshots from DB if needed
	if deleteSnapshots {
		err := q.snapshotRepo.DeleteByMetric(metric, groups)

		if err != nil {
			return nil, err
		}
	}

	// Set DeletedAt to current time
	// Not deleting immediately in case this task is currently running
	// The task deletion check is in AdvanceTasks
	task, err := q.tasks.MarkDelete(metric, groups)

	// Task is already deleted
	if err != nil {
		return nil, err
	}

	// Delete task from DB
	q.taskRepo.Delete(task.Id)

	return task, nil
}

func (q *Queue) ListTasks(groups []string) []*models.Task {
	tasks := q.tasks.GetByGroups(groups)

	return tasks
}

func (q *Queue) UpdateTask(task *models.TaskCreate) (*models.Task, error) {
	// Update task in the map
	// Note that it may cause issues if the task is currently running
	result, err := q.tasks.UpdateTask(task)

	// task not found
	if err != nil {
		return nil, err
	}

	// Update task in the DB
	q.taskRepo.Update(result)

	// Rebuild the queue just in case
	q.queue.Rebuild()

	return result, nil
}

func (q *Queue) onFinish(task models.Task) {
	// Free the computational load
	q.load -= task.Weight

	// If the task is still in the map and is not yet deleted
	if q.tasks.GetTask(task.Id) != nil && (q.tasks.GetTask(task.Id).DeletedAt.IsZero() || q.tasks.GetTask(task.Id).DeletedAt.After(time.Now())) {

		// The task has just been updated
		q.tasks.GetTask(task.Id).UpdatedAt = time.Now()
		q.tasks.GetTask(task.Id).AttemptedAt = time.Now()

		// Add the task back to the queue
		heap.Push(&q.queue, q.tasks.GetTask(task.Id))

		// Update the task data in the DB
		q.taskRepo.Update(q.tasks.GetTask(task.Id))
	} else {
		// The task has been deleted while it was running
		// AdvanceTasks will never be called again for this task,
		// so we delete it here
		q.taskRepo.Delete(task.Id)
	}

	// Send the task to the taskFinished channel
	q.taskFinished <- &task
}

func (q *Queue) AdvanceTasks() {
	for {
		// If we have no tasks, wait for a second and try again
		if q.queue.Len() == 0 {
			time.Sleep(time.Second)
			continue
		}

		// only try to start a new task if we have computational capacity
		for q.load < q.Limit {

			peek := q.queue.Peek()
			// If the queue is empty, restart the loop
			if peek == nil || q.queue.Len() == 0 || (q.load+peek.Weight > q.Limit) {
				break
			}

			// Make a copy to avoid issues with concurrent access
			// I honestly don't remember why this is needed, but I'm sure it is
			oldestUpdated := helpers.CopyTask(heap.Pop(&q.queue).(*models.Task))

			// If the task is not yet "created", pretend that we just finished it and move on
			if !oldestUpdated.CreatedAt.IsZero() && oldestUpdated.CreatedAt.After(time.Now()) {
				q.tasks.GetTask(oldestUpdated.Id).AttemptedAt = time.Now()
				heap.Push(&q.queue, q.tasks.GetTask(oldestUpdated.Id))
				continue
			}

			// If the task is marked as deleted, delete it for good
			if !oldestUpdated.DeletedAt.IsZero() && oldestUpdated.DeletedAt.Before(time.Now()) {
				q.tasks.DeleteTask(oldestUpdated.Id)
				q.taskRepo.Delete(oldestUpdated.Id)

				continue
			}

			// If we found the task name in the list
			if tasks.List[oldestUpdated.Metric] != nil {
				// If enough time for UpdateRate has passed
				if time.Since(oldestUpdated.UpdatedAt) > oldestUpdated.UpdateRate {
					// Start the task
					tasks.Run(*oldestUpdated, tasks.List[oldestUpdated.Metric], q.onFinish, q.snapshotRepo)

					// Add the task weight to the load
					q.load += oldestUpdated.Weight

				} else if q.tasks.GetTask(oldestUpdated.Id) != nil {
					// If not enough time has passed, pretend that we just finished it and move on
					q.tasks.GetTask(oldestUpdated.Id).AttemptedAt = time.Now()
					heap.Push(&q.queue, q.tasks.GetTask(oldestUpdated.Id))
				} else {
					// Triggers if the task is deleted midway through the loop
					q.tasks.DeleteTask(oldestUpdated.Id)
				}
			}
		}

		// Wait a second and try again
		time.Sleep(time.Second)
	}
}

