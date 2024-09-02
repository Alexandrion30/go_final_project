package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	task2 "test/internal/task"
	"test/internal/task/usecase"
	"time"
)

type TaskHttp struct {
	taskService *usecase.TaskService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewTaskHttp(taskService *usecase.TaskService) *TaskHttp {
	return &TaskHttp{
		taskService: taskService,
	}
}

func (th *TaskHttp) HandleTime(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	newDate, err := th.taskService.NextDate(now, date, repeat)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	response, _ := strconv.Atoi(newDate)
	writeResponse(response, w, false)
}

func (th *TaskHttp) Create(w http.ResponseWriter, r *http.Request) {
	var task task2.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		errorResponse := ErrorResponse{Error: "Ошибка десериализации JSON"}
		writeResponse(errorResponse, w, true)
		return
	}

	dateNow := time.Now().Format(usecase.FormatDate)
	err = ValidateForCreate(&task, dateNow)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	if task.Repeat == "" || task.Date == dateNow {
		task.Date = dateNow
	} else {
		task.Date, err = th.taskService.NextDate(dateNow, task.Date, task.Repeat)
		if err != nil {
			errorResponse := ErrorResponse{Error: err.Error()}
			writeResponse(errorResponse, w, true)
			return
		}
	}

	err, id := th.taskService.Create(&task)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	response := task2.Task{ID: id}
	writeResponse(response, w, false)
}

func (th *TaskHttp) GetList(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")
	err, taskList := th.taskService.GetAll(search)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	writeResponse(taskList, w, false)
}

func (th *TaskHttp) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	err, task := th.taskService.GetById(id)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	writeResponse(task, w, false)
}

func (th *TaskHttp) Delete(w http.ResponseWriter, r *http.Request) {
	paramId := r.FormValue("id")
	if paramId == "" {
		errorResponse := ErrorResponse{Error: "задача не найдена"}
		writeResponse(errorResponse, w, true)
		return
	}

	id, err := strconv.Atoi(paramId)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	err = th.taskService.Delete(id)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	writeResponse(task2.Task{}, w, false)
}

func (th *TaskHttp) Edit(w http.ResponseWriter, r *http.Request) {
	var task task2.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		errorResponse := ErrorResponse{Error: "Ошибка десериализации JSON"}
		writeResponse(errorResponse, w, true)
		return
	}

	dateNow := time.Now().Format(usecase.FormatDate)
	err = ValidateForCreate(&task, dateNow)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	task.Date, err = th.taskService.NextDate(dateNow, task.Date, task.Repeat)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	err, updatedTask := th.taskService.Update(&task)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	writeResponse(updatedTask, w, false)
}

func (th *TaskHttp) Done(w http.ResponseWriter, r *http.Request) {
	paramId := r.FormValue("id")
	err := th.taskService.Done(paramId)
	if err != nil {
		errorResponse := ErrorResponse{Error: err.Error()}
		writeResponse(errorResponse, w, true)
		return
	}

	writeResponse(task2.Task{}, w, false)
}

func writeResponse(data interface{}, w http.ResponseWriter, isError bool) {
	toJson, err := json.Marshal(&data)
	if err != nil {
		http.Error(w, `Marshal error`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(toJson)
	if err != nil {
		http.Error(w, `Failed response`, http.StatusInternalServerError)
		return
	}

	if !isError {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
