package utils

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func WriteToCache(ctx context.Context, key string, data interface{}, cache *redis.Client) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return cache.Set(ctx, key, bytes, time.Hour).Err()
}

func GetFromCache[T any](ctx context.Context, key string, cache *redis.Client) (*T, error) {
	cached, err := cache.Get(ctx, key).Result()
	if cached != "" && err == nil {
		var data T
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			return &data, nil
		}
	}
	return nil, err
}

func DeleteFromCache(ctx context.Context, key string, cache *redis.Client) error {
	return cache.Del(ctx, key).Err()
}

// keys function
func GetHomeCacheKey(homeID int) string {
	return "home:" + strconv.Itoa(homeID)
}

func GetUserHomeKey(userID int) string {
	return "home:user:" + strconv.Itoa(userID)
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

func GetCategoryKey(categoryID int) string {
	return "shopping_category:" + strconv.Itoa(categoryID)
}

func GetAllCategoriesForHomeKey(homeID int) string {
	return "shopping_categories:home:" + strconv.Itoa(homeID)
}

func GetPollKey(pollID int) string {
	return "poll:" + strconv.Itoa(pollID)
}

func GetAllPollsForHomeKey(homeID int) string {
	return "poll:home:" + strconv.Itoa(homeID)
}

func GetUserNotificationsKey(userID int) string {
	return "notification:user:" + strconv.Itoa(userID)
}

func GetHomeNotificationsKey(homeID int) string {
	return "notification:home:" + strconv.Itoa(homeID)
}

func GetBillCategoryKey(categoryID int) string {
	return "bill_category:" + strconv.Itoa(categoryID)
}

func GetBillCategoriesKey(homeID int) string {
	return "bill_categories:home:" + strconv.Itoa(homeID)
}
