package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"test/configs"
	"test/internal/task/delivery"
	"test/internal/task/repository"
	"test/internal/task/usecase"
)

func main() {
	config := configs.New()

	err, taskRepository := repository.NewTaskRepository(config.Database)
	if err != nil {
		log.Fatal(err)
	}

	taskService := usecase.NewTaskService(taskRepository)
	taskHttp := delivery.NewTaskHttp(taskService)

	r := mux.NewRouter()
	r.HandleFunc("/api/nextdate", taskHttp.HandleTime).Methods("GET")
	r.HandleFunc("/api/task", taskHttp.Create).Methods("POST")
	r.HandleFunc("/api/tasks", taskHttp.GetList).Methods("GET")
	r.HandleFunc("/api/task", taskHttp.Show).Methods("GET")
	r.HandleFunc("/api/task", taskHttp.Delete).Methods("DELETE")
	r.HandleFunc("/api/task", taskHttp.Edit).Methods("PUT")
	r.HandleFunc("/api/task/done", taskHttp.Done).Methods("POST")

	startServer(config.Server, r)
}

func startServer(server *configs.Server, r *mux.Router) {
	fs := http.FileServer(http.Dir("./web"))
	r.PathPrefix("").Handler(http.StripPrefix("", fs))

	fmt.Printf("Server is running on port %s\n", server.Port)
	err := http.ListenAndServe(":"+server.Port, r)
	if err != nil {
		log.Fatal(err)
	}
}
