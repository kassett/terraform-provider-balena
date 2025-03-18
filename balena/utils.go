package balena

type IDWrapper struct {
	ID int `json:"__id"`
}

func is200Level(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
