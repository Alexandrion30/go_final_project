package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"test/configs"
	"test/internal/task"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(dbConfig *configs.Database) (error, *TaskRepository) {
	install := false

	appPath, err := os.Executable()
	if err != nil {
		return err, nil
	}
	dbFile := filepath.Join(filepath.Dir(appPath), dbConfig.DatabaseName)
	_, err = os.Stat(dbFile)

	if err != nil {
		install = true
	}

	if !install {
		db, err := sql.Open(dbConfig.DriverName, dbConfig.DatabaseName)
		if err != nil {
			return err, nil
		}

		log.Print("Database initialized")

		return nil, &TaskRepository{
			db: db,
		}
	}

	db, err := sql.Open(dbConfig.DriverName, dbConfig.DatabaseName)
	if err != nil {
		return err, nil
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, date VARCHAR NOT NULL, title VARCHAR NOT NULL, comment VARCHAR NOT NULL,  repeat VARCHAR(128) NOT NULL)")
	if err != nil {
		return err, nil
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date)")
	if err != nil {
		return err, nil
	}

	log.Print("Database initialized")

	return nil, &TaskRepository{
		db: db,
	}
}

func (ts *TaskRepository) Insert(task *task.Task) (error, string) {
	query := "insert into scheduler (date, title, comment, repeat) values ($1, $2, $3, $4)"
	res, err := ts.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return err, ""
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err, ""
	}

	return nil, fmt.Sprint(id)
}

func (ts *TaskRepository) GetAll() (error, *task.List) {
	query := "select * from scheduler order by date"
	rows, err := ts.db.Query(query)
	if err != nil {
		return err, nil
	}

	return ts.prepareTaskList(rows)
}

func (ts *TaskRepository) GetByDate(date string) (error, *task.List) {
	query := "select * from scheduler where `date` = :date"
	rows, err := ts.db.Query(query, date)
	if err != nil {
		return err, nil
	}

	return ts.prepareTaskList(rows)
}

func (ts *TaskRepository) GetByTitleOrComment(search string) (error, *task.List) {
	query := "SELECT * FROM scheduler WHERE (title LIKE ? OR comment LIKE ?) ORDER BY date"
	rows, err := ts.db.Query(query, "%"+search+"%", "%"+search+"%")
	if err != nil {
		return err, nil
	}

	return ts.prepareTaskList(rows)
}

func (ts *TaskRepository) GetById(id int) (error, *task.Task) {
	query := "select * from scheduler where id = ?"
	row := ts.db.QueryRow(query, id)
	var t task.Task
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return err, nil
	}

	return nil, &t
}

func (ts *TaskRepository) prepareTaskList(rows *sql.Rows) (error, *task.List) {
	taskList := make([]*task.Task, 0)
	for rows.Next() {
		var taskStruct task.Task
		err := rows.Scan(&taskStruct.ID, &taskStruct.Date, &taskStruct.Title, &taskStruct.Comment, &taskStruct.Repeat)
		if err != nil {
			return err, nil
		}
		taskList = append(taskList, &taskStruct)
	}

	return nil, &task.List{Task: taskList}
}

func (ts *TaskRepository) DeleteById(id int) error {
	query := "DELETE FROM scheduler WHERE id = $1"
	_, err := ts.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TaskRepository) UpdateById(t *task.Task) (error, *task.Task) {
	query := "update scheduler set date = $1, title = $2, comment = $3, repeat = $4 where id = $5"
	_, err := ts.db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		return err, nil
	}

	return nil, t
}

func (ts *TaskRepository) Done(task *task.Task) error {
	query := "update scheduler set date = $1 where id = $2"
	_, err := ts.db.Exec(query, task.Date, task.ID)
	if err != nil {
		return err
	}

	return nil
}
