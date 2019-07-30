package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	. "github.com/logrusorgru/aurora"
)

var ctx = context.Background()
var defaultCategoryID int64 = 1
var db *sql.DB

// Task represents a task of the user
type Task struct {
	Id           int64
	Description  string
	Created      int64
	Until        int64
	Done         bool
	CategoryId   int64
	CategoryName string
}

// Category represents a category which tasks can be assigned to
type Category struct {
	Id   int64
	Name string
}

func (task *Task) getCheckBox() string {

	checkBox := ""
	if task.Done {
		checkBox = Green("\u2713").String()
	} else {
		checkBox = Red("\u2A09").String()
	}

	return checkBox
}

func (task *Task) StringArray() []string {

	desc := task.Description
	id := Bold(task.Id).String()
	untilString := timeUntil(task.Until)
	done := task.Done
	catName := task.CategoryName
	var checkBox string

	if done {
		checkBox = Green("\u2713").String()
	} else {
		checkBox = Red("\u2A09").String()
	}

	return []string{checkBox, id, desc, untilString, catName}

}

func SaveTask(categoryName string, description string, day int64, hour int64) *Task {
	now := time.Now().Unix()

	if categoryName != "" {
		var err error
		categoryId, err = GetOrCreateCategory(categoryName)
		if err != nil {
			panic(err)
		}
	}

	task := Task{Description: description, Created: now, CategoryId: categoryId}
	until := now
	if hour != -1 {
		until += hour * 60 * 60
	}
	if day != -1 {
		until += day * 60 * 60 * 24
	}
	if until != now {
		task.Until = until
	}

	insertTask(&task)

	return &task
}

func (c *Category) StringArray() []string {
	return []string{fmt.Sprintf("%d", c.Id), c.Name}
}

// setDb sets the database variable
// useful since we want to have two different database
// one for running the application and one for testing
func setDB(database *sql.DB) {
	db = database
}

// CreateTableTaskCategory creates a task table in the database if it does not exist
func CreateTableTaskCategory() {
	taskTable := `CREATE TABLE IF NOT EXISTS tasks (
					id integer not null primary key, 
					description text not null UNIQUE,
					done boolean DEFAULT false,
					created integer,
					until integer,
					category_id integer,
					FOREIGN KEY(category_id) REFERENCES categories(id)
				);`
	categoryTable := `CREATE TABLE IF NOT EXISTS categories(
			id integer not null primary key,
			name text not null unique
	);`

	defaultCategory := fmt.Sprintf(`INSERT OR IGNORE INTO categories(id, name) VALUES (%d, 'default');`, defaultCategoryID)

	_, err := db.Exec(categoryTable)
	checkErrorQueries(err, categoryTable)

	_, err = db.Exec(taskTable)
	checkErrorQueries(err, taskTable)

	_, err = db.Exec(defaultCategory)
	checkErrorQueries(err, defaultCategory)
}

// GetOrCreateCategory creates a new category if it is not present in the database
// and then returns it or the already existing one
func GetOrCreateCategory(name string) (id int64, err error) {
	name = strings.ToLower(name)
	sqlStmt := fmt.Sprintf(`SELECT id FROM categories WHERE name="%s";`, name)
	row := db.QueryRow(sqlStmt)

	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		sqlStmt = fmt.Sprintf(`INSERT INTO "categories" (name) VALUES('%s');`, name)
		res, err := db.Exec(sqlStmt)
		checkErrorQueries(err, sqlStmt)
		id, err = res.LastInsertId()
		return id, err
	case nil:
		return id, err
	default:
		panic(err)
	}
}

// NewTasks inserts a new task in the database
func insertTask(t *Task) *Task {

	if t.CategoryId <= 0 {
		t.CategoryId = defaultCategoryID
	}

	sqlStmt := "INSERT OR IGNORE INTO tasks (description, created, until, category_id) VALUES ($1, $2, $3, $4)"
	res, err := db.Exec(sqlStmt, t.Description, t.Created, t.Until, t.CategoryId)
	checkErrorQueries(err, sqlStmt)

	// Update the Id of Task
	id, err := res.LastInsertId()

	if err != nil {
		panic(err)
	}
	t.Id = id

	return t
}

// AllTasks returns all tasks in the database.
// You have to pass an orderBy and sorted argument for the query
func AllTasks(orderBy string, sorted string) []Task {

	sqlStmt := fmt.Sprintf("SELECT t.id, t.description, t.created, t.until, t.done, c.name FROM tasks as t INNER JOIN categories As c ON (t.category_id=c.id) ORDER BY t.%s %s;", orderBy, sorted)

	rows, err := db.QueryContext(ctx, sqlStmt)

	defer rows.Close()
	tasks := make([]Task, 0)

	switch err {

	case sql.ErrNoRows:
		fmt.Println("No Tasks in the Database yet, add some")

	case nil:

		for rows.Next() {
			var task Task
			err = rows.Scan(
				&task.Id,
				&task.Description,
				&task.Created,
				&task.Until,
				&task.Done,
				&task.CategoryName,
			)

			if err != nil {
				log.Fatal(err)
			}
			tasks = append(tasks, task)
		}

	default:
		checkErrorQueries(err, sqlStmt)

	}

	return tasks

}

// AllCategories returns all categories present in the database
func AllCategories() []Category {

	sqlStmt := `SELECT id, name FROM categories`

	rows, err := db.QueryContext(ctx, sqlStmt)

	defer rows.Close()
	categories := make([]Category, 0)

	switch err {

	case sql.ErrNoRows:
		fmt.Println("No Categories in the Database yet, add some")

	case nil:

		for rows.Next() {
			var category Category
			err = rows.Scan(&category.Id, &category.Name)
			if err != nil {
				log.Fatal(err)
			}

			categories = append(categories, category)
		}

	default:
		checkErrorQueries(err, sqlStmt)

	}

	return categories

}

// UpdateCategory updates the category for all tasks given by id
func UpdateCategory(catId int64, ids idFlags) {
	sqlStmt := fmt.Sprintf(`UPDATE tasks set category_id=%d where id in (%s)`, catId, ids.String())
	_, err := db.Exec(sqlStmt)
	checkErrorQueries(err, sqlStmt)
}

// DeleteDoneTasks deletes all tasks in the database
// where done is set true
func DeleteDoneTasks() {
	sqlStmt := `DELETE FROM tasks WHERE done=true`
	_, err := db.Exec(sqlStmt)
	checkErrorQueries(err, sqlStmt)
}

// DeleteTasksById deletes all tasks which were specified
func DeleteTasksById(ids idFlags) {
	sqlStmt := fmt.Sprintf("DELETE FROM tasks WHERE id in (%s)", ids.String())
	_, err := db.Exec(sqlStmt)
	checkErrorQueries(err, sqlStmt)
}

// TaskDone marks tasks as done in the database
func TaskDone(ids idFlags) {
	sqlStmt := fmt.Sprintf("UPDATE tasks set done=TRUE WHERE id in (%s)", ids.String())
	_, err := db.Exec(sqlStmt)
	checkErrorQueries(err, sqlStmt)
}

// getDbPath returns the path of the database given by name
func getDbPath(dbName string) (dir string, db string, fullPath string, err error) {
	//https://stackoverflow.com/questions/32163425/how-to-get-the-directory-of-the-package-the-file-is-in-not-the-current-working
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No Caller Information")
	}
	p := path.Dir(filename)
	db = dbName + ".db"
	dir, _ = filepath.Abs(fmt.Sprintf("%s", p))
	fullPath = fmt.Sprintf("%s/%s", dir, db)
	return dir, db, fullPath, err
}

// getORCreateDb creates a new sqlite database
// if it does not exist already and then returns it
// Database gets created in the directory of this package
func getOrCreateDb(dir string, dbName string, fullPath string) {
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {

		// ToDo: Add Support for Windows
		cmd := exec.Command("touch", dbName)
		cmd.Dir, _ = filepath.Abs(dir)
		_, err := cmd.Output()

		if err != nil {
			panic(err)
		}

	}
}

// checkErrorQueries logs the sql statement if an error occurred
func checkErrorQueries(err error, stmt string) {
	if err != nil {
		log.Printf("%q: %s\n", err, stmt)
	}
}
