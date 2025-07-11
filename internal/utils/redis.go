package utils

import "strconv"

func GetHomeCacheKey(id int) string {
	return "home:" + strconv.Itoa(id)
}
