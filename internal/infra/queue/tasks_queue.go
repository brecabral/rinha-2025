package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	key = "tasks_queue"
)

type TasksQueue struct {
	redisClient *redis.Client
}

func NewTasksQueue(rc *redis.Client) *TasksQueue {
	return &TasksQueue{redisClient: rc}
}

type TaskInfo struct {
	ID          string
	Amount      float64
	RequestedAt time.Time
}

func (q *TasksQueue) Enqueue(task TaskInfo) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return q.redisClient.RPush(ctx, key, taskJSON).Err()
}

func (q *TasksQueue) Dequeue() (TaskInfo, error) {
	var task TaskInfo
	result, err := q.redisClient.BLPop(ctx, 0, key).Result()
	if err != nil {
		return task, err
	}
	taskJSON := result[1]
	err = json.Unmarshal([]byte(taskJSON), &task)
	return task, err
}

func (q *TasksQueue) Prepend(task TaskInfo) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return q.redisClient.LPush(ctx, key, taskJSON).Err()
}
