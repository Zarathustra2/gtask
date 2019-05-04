package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	. "github.com/logrusorgru/aurora"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	createTestDB()
	CreateTableTaskCategory()
	rC := m.Run()
	_ = os.Remove("test.db")
	os.Exit(rC)
}

func TestCreateTableTaskCategory(t *testing.T) {
	// Gets already called in MainRunner
	CreateTableTaskCategory()

	taskTableCheck := "SELECT 1 FROM tasks LIMIT 1;"
	_, err := testDB.Exec(taskTableCheck)
	if err != nil {
		t.Errorf("Table tasks does not exist, err: %s", err)
	}
	categoryTableCheck := "SELECT 1 FROM categories LIMIT 1;"
	_, err = testDB.Exec(categoryTableCheck)
	if err != nil {
		t.Errorf("Table categories does not exist, err: %s", err)
	}
}

func TestGetOrCreateCategory(t *testing.T) {
	defer cleanDatabase()

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantId  int64
		wantErr bool
	}{
		// starting with id 2 because the default has already been created on creation of the tables
		{"Task", args{"Category"}, 2, false},
		{"Same Category should not be inserted", args{"Category"}, 2, false},
		{"", args{"NewCategory"}, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, err := GetOrCreateCategory(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrCreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("GetOrCreateCategory() = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	defer cleanDatabase()

	createThreeTasks()

	UpdateCategory(1, []string{"1", "2", "3"})
	var count int

	row := testDB.QueryRow("SELECT COUNT(*) FROM tasks where category_id=1")
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count != 3 {
		t.Errorf("Got %d in the database, expected %d", count, 3)
	}

}

func TestTaskDone(t *testing.T) {
	var count int

	createThreeTasks()
	TaskDone([]string{"2", "3"})

	row := testDB.QueryRow("SELECT COUNT(*) FROM tasks where done=TRUE ")
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count != 2 {
		t.Errorf("Got %d, expected %d", count, 2)
	}

}

func TestDeleteTasksById(t *testing.T) {
	defer cleanDatabase()

	createThreeTasks()

	DeleteTasksById([]string{"1", "2"})
	var count int
	row := testDB.QueryRow("SELECT COUNT(*) FROM tasks")
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Got %d, expected %d", count, 1)
	}
}

func Test_getDbPath(t *testing.T) {
	wantDir, _ := os.Getwd()
	wantFullPath := wantDir + "/" + "todo.db"
	type args struct {
		dbName string
	}
	tests := []struct {
		name         string
		args         args
		wantDir      string
		wantDb       string
		wantFullPath string
		wantErr      bool
	}{
		{"", args{"todo"}, wantDir, "todo.db", wantFullPath, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDir, gotDb, gotFullPath, err := getDbPath(tt.args.dbName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDbPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDir != tt.wantDir {
				t.Errorf("getDbPath() gotDir = %v, want %v", gotDir, tt.wantDir)
			}
			if gotDb != tt.wantDb {
				t.Errorf("getDbPath() gotDb = %v, want %v", gotDb, tt.wantDb)
			}
			if gotFullPath != tt.wantFullPath {
				t.Errorf("getDbPath() gotFullPath = %v, want %v", gotFullPath, tt.wantFullPath)
			}
		})
	}
}

func TestDeleteDoneTasks(t *testing.T) {
	defer cleanDatabase()

	createThreeTasks()

	TaskDone([]string{"2", "3"})
	DeleteDoneTasks()
	var count int
	row := testDB.QueryRow("SELECT COUNT(*) FROM tasks")
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Got %d, expected %d", count, 1)
	}
}

func TestAllCategories(t *testing.T) {
	defer cleanDatabase()

	createThreeTasks()
	c := AllCategories()

	expect := []Category{{1, "default"}, {2, "home"}, {3, "coding"}}
	fmt.Println(c)
	for i := range c {
		if c[i] != expect[i] {
			t.Errorf("Got %s, want %s", c[i].StringArray(), expect[i].StringArray())
		}
	}

}

func TestAllTasks(t *testing.T) {
	defer cleanDatabase()
	createThreeTasks()

	tasks := AllTasks("ID", "")

	if len(tasks) != 3 {
		t.Errorf("Got %d tasks, expected %d", len(tasks), 3)
	}

}

func TestTask_getCheckBox(t *testing.T) {
	type fields struct {
		Id           int64
		Description  string
		Created      int64
		Until        int64
		Done         bool
		CategoryId   int64
		CategoryName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"", fields{Done: true}, Green("\u2713").String()},
		{"", fields{Done: false}, Red("\u2A09").String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				Id:           tt.fields.Id,
				Description:  tt.fields.Description,
				Created:      tt.fields.Created,
				Until:        tt.fields.Until,
				Done:         tt.fields.Done,
				CategoryId:   tt.fields.CategoryId,
				CategoryName: tt.fields.CategoryName,
			}
			if got := task.getCheckBox(); got != tt.want {
				t.Errorf("Task.getCheckBox() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveTask(t *testing.T) {
	defer cleanDatabase()
	createThreeTasks()

	var count int
	row := testDB.QueryRow("SELECT COUNT(*) FROM tasks")
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count != 3 {
		t.Errorf("Got %d, expected %d", count, 3)
	}

}

func createTestDB() {
	dir, dbName, fullPath, err := getDbPath("test")
	if err != nil {
		panic(err)
	}

	getOrCreateDb(dir, dbName, fullPath)

	testDB, err = sql.Open("sqlite3", fullPath)
	if err != nil {
		panic(err)
	}

	setDB(testDB)
}

func cleanDatabase() {
	clearTasks := "DROP table tasks"
	clearCategories := "DROP table categories"
	clearGitHubToken := "DROP table githubToken"

	_, _ = testDB.Exec(clearTasks)
	_, _ = testDB.Exec(clearCategories)
	_, _ = testDB.Exec(clearGitHubToken)

	CreateTableTaskCategory()
	sqlStmt := `CREATE TABLE IF NOT EXISTS githubToken (
					id integer primary key check(id = 0), 
					token text not null
				);`

	_, _ = testDB.Exec(sqlStmt)

}

func createThreeTasks() {
	SaveTask("Home", "Clean Room", 0, 0)
	SaveTask("Coding", "Add Tests", 0, 0)
	SaveTask("Home", "Buy Present", 0, 0)
}

func TestCategory_StringArray(t *testing.T) {
	category := Category{1, "Coding"}
	got := category.StringArray()
	expect := []string{"1", "Coding"}

	for i := range got {
		if got[i] != expect[i] {
			t.Errorf("Got %s, expected %s", got[i], expect[i])
		}
	}
}

func TestTask_StringArray(t *testing.T) {
	type fields struct {
		Id           int64
		Description  string
		Created      int64
		Until        int64
		Done         bool
		CategoryId   int64
		CategoryName string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{

		{"",
			fields{1, "Fix Bugs", 0, 0, false, 1, "Coding"},
			[]string{Red("\u2A09").String(), Bold("1").String(), "Fix Bugs", "-", "Coding"},
		},
		{"",
			fields{1, "Fix Bugs", 0, 0, true, 1, "Coding"},
			[]string{Green("\u2713").String(), Bold("1").String(), "Fix Bugs", "-", "Coding"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				Id:           tt.fields.Id,
				Description:  tt.fields.Description,
				Created:      tt.fields.Created,
				Until:        tt.fields.Until,
				Done:         tt.fields.Done,
				CategoryId:   tt.fields.CategoryId,
				CategoryName: tt.fields.CategoryName,
			}
			if got := task.StringArray(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Task.StringArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
