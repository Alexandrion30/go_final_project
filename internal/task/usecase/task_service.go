package usecase

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"test/internal/task"
	"test/internal/task/repository"
	"time"
)

const FormatDate = "20060102"

type TaskService struct {
	taskRepository *repository.TaskRepository
}

func NewTaskService(taskRepository *repository.TaskRepository) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
	}
}

func (ts *TaskService) NextDate(now string, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("правило пустое")
	}

	startDate, err := time.Parse(FormatDate, date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	nowDate, err := time.Parse(FormatDate, now)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты %v", err)
	}

	parts := strings.Fields(repeat)
	if len(parts) < 1 {
		return "", errors.New("неверный формат правила повторения")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("неверный формат правила d")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval <= 0 || interval > 365 {
			return "", errors.New("неверный интервал")
		}
		for {
			startDate = startDate.AddDate(0, 0, interval)
			if startDate.After(nowDate) {
				break
			}
		}
	case "w":
		if len(parts) != 2 {
			return "", errors.New("неверный формат правила w")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval < 1 || interval > 7 {
			return "", errors.New("неверный интервал недели")
		}
		for {
			startDate = startDate.AddDate(0, 0, 1)
			if int(startDate.Weekday()) == interval && startDate.After(nowDate) {
				break
			}
		}
	case "m":
		if len(parts) != 2 {
			return "", errors.New("неверный формат m")
		}
		dayParts := strings.Split(parts[1], ",")
		if len(dayParts) != 2 {
			return "", errors.New("неверные дни месяцев")
		}
		monthOffset, err := strconv.Atoi(dayParts[0])
		if err != nil {
			return "", errors.New("ошибка конвертации")
		}
		dayOfMonth, err := strconv.Atoi(dayParts[1])
		if err != nil || dayOfMonth < 1 || dayOfMonth > 31 {
			return "", errors.New("неверный день месяца")
		}
		for {
			startDate = startDate.AddDate(0, monthOffset, 0)
			if startDate.Day() != dayOfMonth {
				startDate = time.Date(startDate.Year(), startDate.Month(), dayOfMonth, 0, 0, 0, 0, startDate.Location())
			}
			if startDate.After(nowDate) {
				break
			}
		}
	case "y":
		if len(parts) != 1 {
			return "", errors.New("неверный формат правила y")
		}
		for {
			startDate = startDate.AddDate(1, 0, 0)
			if startDate.After(nowDate) {
				break
			}
		}
	default:
		return "", errors.New("неизвестный тип правила")
	}

	return startDate.Format(FormatDate), nil
}

func (ts *TaskService) Create(task *task.Task) (error, string) {
	err, id := ts.taskRepository.Insert(task)
	if err != nil {
		return err, ""
	}

	task.ID = id

	return nil, id
}

func (ts *TaskService) GetAll(search string) (error, *task.List) {
	if search != "" {
		_, err := time.Parse("02.01.2006", search)
		if err == nil {
			return ts.taskRepository.GetByDate(search)
		}

		return ts.taskRepository.GetByTitleOrComment(search)
	}

	err, taskList := ts.taskRepository.GetAll()
	if err != nil {
		return err, nil
	}

	return nil, taskList
}

func (ts *TaskService) GetById(id int) (error, *task.Task) {
	err, t := ts.taskRepository.GetById(id)
	if err != nil {
		return err, nil
	}

	return nil, t
}

func (ts *TaskService) Update(t *task.Task) (error, *task.Task) {
	err, t := ts.taskRepository.UpdateById(t)
	if err != nil {
		return err, nil
	}

	return nil, t
}

func (ts *TaskService) Delete(id int) error {
	err := ts.taskRepository.DeleteById(id)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TaskService) Done(paramId string) error {
	if paramId == "" {
		return errors.New("задача не найдена")
	}
	id, err := strconv.Atoi(paramId)
	if err != nil {
		return errors.New("некорректный параметр id")
	}
	err, t := ts.GetById(id)
	if err != nil {
		return errors.New("задача не найдена")
	}
	if t.Repeat == "" {
		err = ts.Delete(id)
		if err != nil {
			return errors.New("ошибка при удалении")
		}

		return nil
	}

	newDate, err := ts.NextDate(time.Now().Format(FormatDate), t.Date, t.Repeat)
	if err != nil {
		return errors.New("ошибка при вычислении даты")
	}
	t.Date = newDate
	err = ts.taskRepository.Done(t)
	if err != nil {
		return err
	}

	return nil
}
