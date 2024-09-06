package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"test/configs"
	"test/internal/task/delivery"
	"test/internal/task/repository"
	"test/internal/task/usecase"
)

func main() {
	config := configs.New()

	db, err := initDB(config.Database)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}(db)

	taskRepository := repository.NewTaskRepository(db)
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

	// Создаем канал для получения сигналов прерывания
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в фоновом режиме
	srv := &http.Server{Addr: ":" + server.Port, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen and serve: %v", err)
		}
	}()

	// Ожидаем сигнал прерывания
	<-quit

	// Производим безопасное завершение сервера
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}

func initDB(dbConfig *configs.Database) (*sql.DB, error) {
	install := false

	appPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dbFile := filepath.Join(filepath.Dir(appPath), dbConfig.DatabaseName)
	_, err = os.Stat(dbFile)

	if err != nil {
		install = true
	}

	if !install {
		db, err := sql.Open(dbConfig.DriverName, dbConfig.DatabaseName)
		if err != nil {
			return nil, err
		}

		log.Print("Database initialized")

		return db, nil
	}

	db, err := sql.Open(dbConfig.DriverName, dbConfig.DatabaseName)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, date VARCHAR NOT NULL, title VARCHAR NOT NULL, comment VARCHAR NOT NULL,  repeat VARCHAR(128) NOT NULL)")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date)")
	if err != nil {
		return nil, err
	}

	log.Print("Database initialized")

	return db, nil
}
