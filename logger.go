package main

func init() {
	logger.Printf = func(format string, v ...interface{}) {}
}

var logger struct {
	Printf func(format string, v ...interface{})
}
