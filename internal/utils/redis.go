package utils

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func WriteToCache(key string, data interface{}, cache *redis.Client) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return cache.Set(context.Background(), key, bytes, time.Hour).Err()
}

func GetFromCache[T any](key string, cache *redis.Client) (*T, error) {
	cached, err := cache.Get(context.Background(), key).Result()
	if cached != "" && err == nil {
		var data T
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			return &data, nil
		}
	}
	return nil, err
}

func DeleteFromCache(key string, cache *redis.Client) error {
	return cache.Del(context.Background(), key).Err()
}

// keys function
func GetHomeCacheKey(homeID int) string {
	return "home:" + strconv.Itoa(homeID)
}

func GetTaskKey(taskID int) string {
	return "task:" + strconv.Itoa(taskID)
}

func GetTasksForHomeKey(homeID int) string {
	return "tasks:home:" + strconv.Itoa(homeID)
}

func GetAssignmentKey(assignmentID int) string {
	return "assignment:" + strconv.Itoa(assignmentID)
}

func GetAssignmentsForUserKey(userID int) string {
	return "assignments:user:" + strconv.Itoa(userID)
}

func GetClosestAssignmentsForUserKey(userID int) string {
	return "assignment:user:" + strconv.Itoa(userID)
}

func GetBillKey(billID int) string {
	return "bill:" + strconv.Itoa(billID)
}

func GetRoomKey(roomID int) string {
	return "room:" + strconv.Itoa(roomID)
}

func GetRoomsForHomeKey(homeID int) string {
	return "rooms:home:" + strconv.Itoa(homeID)
}
