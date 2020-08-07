package logger

import (
    "log"
)

var logger Logger

//1 = silent
//2 = prod
//3 = message
//4 = game
//5 = debug
type Logger struct {
    Level int
}

func InitLogger(level int) {
	logger = Logger{Level:level}
}

func Log(level int, v ...interface{}) {
	if level <= logger.Level {
		log.Println(v)
	}
}
