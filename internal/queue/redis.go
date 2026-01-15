package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Task represents a queue task
type Task struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	ConversationID string                 `json:"conversation_id"`
	EvaluatorTypes []string               `json:"evaluator_types,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// RedisQueue implements queue operations using Redis
type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisQueue creates a new Redis queue
func NewRedisQueue(redisURL string) (*RedisQueue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisQueue{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (q *RedisQueue) Close() error {
	return q.client.Close()
}

// Enqueue adds a task to the queue
func (q *RedisQueue) Enqueue(queueName string, task *Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	return q.client.RPush(q.ctx, queueName, data).Err()
}

// Dequeue removes and returns a task from the queue
func (q *RedisQueue) Dequeue(queueName string, timeout time.Duration) (*Task, error) {
	result, err := q.client.BLPop(q.ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No task available
		}
		return nil, fmt.Errorf("failed to dequeue task: %w", err)
	}

	if len(result) < 2 {
		return nil, nil
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// QueueLength returns the number of tasks in the queue
func (q *RedisQueue) QueueLength(queueName string) (int64, error) {
	return q.client.LLen(q.ctx, queueName).Result()
}

// Set stores a value with expiration
func (q *RedisQueue) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return q.client.Set(q.ctx, key, data, expiration).Err()
}

// Get retrieves a value
func (q *RedisQueue) Get(key string, dest interface{}) error {
	data, err := q.client.Get(q.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil // Key not found
		}
		return fmt.Errorf("failed to get value: %w", err)
	}
	return json.Unmarshal(data, dest)
}

// Delete removes a key
func (q *RedisQueue) Delete(key string) error {
	return q.client.Del(q.ctx, key).Err()
}

// Publish publishes a message to a channel
func (q *RedisQueue) Publish(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return q.client.Publish(q.ctx, channel, data).Err()
}
