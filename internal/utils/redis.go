package utils

import "strconv"

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
