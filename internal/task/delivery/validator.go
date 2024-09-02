package delivery

import (
	"errors"
	"test/internal/task"
	"test/internal/task/usecase"
	"time"
)

func ValidateForCreate(task *task.Task, dateNow string) error {
	if task.Title == "" {
		return errors.New("title is empty")
	}
	if task.Date == "" {
		task.Date = dateNow
		return nil
	}
	_, err := time.Parse(usecase.FormatDate, task.Date)
	if err != nil {
		return errors.New("date is invalid")
	}
	if task.Date < dateNow && task.Repeat == "" {
		task.Date = dateNow
	}

	return nil
}
