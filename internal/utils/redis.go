package utils

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Dragodui/diploma-server/internal/metrics"
	"github.com/redis/go-redis/v9"
)

func WriteToCache(ctx context.Context, key string, data interface{}, cache *redis.Client) error {
	start := time.Now()
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = cache.Set(ctx, key, bytes, time.Hour).Err()
	duration := time.Since(start).Seconds()
	metrics.RedisOperationDuration.WithLabelValues("set").Observe(duration)
	if err != nil {
		metrics.RedisOperationsTotal.WithLabelValues("set", "error").Inc()
	} else {
		metrics.RedisOperationsTotal.WithLabelValues("set", "success").Inc()
	}
	return err
}

func GetFromCache[T any](ctx context.Context, key string, cache *redis.Client) (*T, error) {
	start := time.Now()
	cached, err := cache.Get(ctx, key).Result()
	duration := time.Since(start).Seconds()
	metrics.RedisOperationDuration.WithLabelValues("get").Observe(duration)
	if cached != "" && err == nil {
		var data T
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			metrics.RedisOperationsTotal.WithLabelValues("get", "hit").Inc()
			metrics.RedisCacheHits.Inc()
			return &data, nil
		}
	}
	if err != nil && err != redis.Nil {
		metrics.RedisOperationsTotal.WithLabelValues("get", "error").Inc()
	} else {
		metrics.RedisOperationsTotal.WithLabelValues("get", "miss").Inc()
		metrics.RedisCacheMisses.Inc()
	}
	return nil, err
}

func DeleteFromCache(ctx context.Context, key string, cache *redis.Client) error {
	start := time.Now()
	err := cache.Del(ctx, key).Err()
	duration := time.Since(start).Seconds()
	metrics.RedisOperationDuration.WithLabelValues("del").Observe(duration)
	if err != nil {
		metrics.RedisOperationsTotal.WithLabelValues("del", "error").Inc()
	} else {
		metrics.RedisOperationsTotal.WithLabelValues("del", "success").Inc()
	}
	return err
}

// keys function
func GetHomeCacheKey(homeID int) string {
	return "home:" + strconv.Itoa(homeID)
}

func GetUserHomeKey(userID int) string {
	return "home:user:" + strconv.Itoa(userID)
}

func GetUserHomesKey(userID int) string {
	return "homes:user:" + strconv.Itoa(userID)
}

func GetHomeCurrencyKey(homeID int) string {
	return "homes:currency" + strconv.Itoa(homeID)
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

func GetAssignmentsForUserKey(userID int, homeID int) string {
	return "assignments:user:" + strconv.Itoa(userID) + ":home:" + strconv.Itoa(homeID)
}

func GetClosestAssignmentsForUserKey(userID int, homeID int) string {
	return "assignment:user:" + strconv.Itoa(userID) + ":home:" + strconv.Itoa(homeID)
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

// GetUserNotificationsKey builds the cache key for a user's notifications,
// optionally scoped to a single home (pass nil for the unscoped, all-homes view).
func GetUserNotificationsKey(userID int, homeID *int) string {
	if homeID == nil {
		return "notification:user:" + strconv.Itoa(userID) + ":home:all"
	}
	return "notification:user:" + strconv.Itoa(userID) + ":home:" + strconv.Itoa(*homeID)
}

func GetHomeNotificationsKey(homeID int) string {
	return "notification:home:" + strconv.Itoa(homeID)
}

func GetBillCategoryKey(categoryID int) string {
	return "bill_category:" + strconv.Itoa(categoryID)
}

func GetBillCategoriesKey(homeID, userID int, public bool) string {
	scope := "public"
	if !public {
		scope = "private:user:" + strconv.Itoa(userID)
	}
	return "bill_categories:home:" + strconv.Itoa(homeID) + ":" + scope
}

func GetNoteKey(noteID int) string {
	return "note:" + strconv.Itoa(noteID)
}

func GetAllNotesForHomeKey(homeID int) string {
	return "notes:home:" + strconv.Itoa(homeID)
}

func GetNoteCategoryKey(categoryID int) string {
	return "note_category:" + strconv.Itoa(categoryID)
}

func GetAllNoteCategoriesForHomeKey(homeID int) string {
	return "note_categories:home:" + strconv.Itoa(homeID)
}

