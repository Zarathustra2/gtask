package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

// declare our flags
var (
	database          *sql.DB
	amount            int
	hour              int64
	day               int64
	description       string
	idsLists          idFlags
	orderBy           string
	desc              bool
	updateIds         bool
	deleteIds         bool
	categoryName      string
	categoryId        int64
	delDoneTasks      bool
	renderCategories  bool
	gitIssuesDownload bool
	table             bool
	addGitToken       string
)

// idFlags represents a list of ids
type idFlags []string

func (i *idFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *idFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// interpreter chooses based on the given flags the right execution
func interpreter() {
	switch {

	case categoryId > 0:
		if len(idsLists) == 0 {
			panic("You need to provide IDs to update the category with the ids flag")
			return
		}
		UpdateCategory(categoryId, idsLists)

	case delDoneTasks:
		DeleteDoneTasks()

	case updateIds:
		if len(idsLists) == 0 {
			panic("You need to provide IDs to update done tasks with the ids flag")
			return
		}
		TaskDone(idsLists)

	case description != "":
		SaveTask(categoryName, description, day, hour)

	case deleteIds:
		if len(idsLists) == 0 {
			panic("You need to provide IDs to delete with the ids flag")
			return
		}
		DeleteTasksById(idsLists)

	case gitIssuesDownload:
		saveIssuesToDatabase()

	case addGitToken != "":
		err := saveGitToken(addGitToken)
		if err != nil {
			log.Fatalln(err)
		}

	default:
		renderTasks(orderBy, table, desc, renderCategories)

	}
}

// renderTasks renders the tasks. Either as aligned style or as table
func renderTasks(orderBy string, table bool, desc bool, renderCategories bool) {
	if renderCategories {
		RenderTableCategories()
	}
	sorted := "ASC"
	if desc {
		sorted = "DESC"
	}

	if table {
		RenderTableTasks(orderBy, sorted)
	} else {
		RenderAligned(orderBy, sorted)
	}
}

func main() {

	dir, dbName, fullPath, err := getDbPath("todo")
	getOrCreateDb(dir, dbName, fullPath)
	database, err = sql.Open("sqlite3", fullPath)
	if err != nil {
		panic(err)
	}

	defer database.Close()

	setDB(database)
	CreateTableTaskCategory()
	interpreter()

}

// init parses the flags
func init() {

	flag.BoolVar(&desc, "desc", false, "Sort Asc/Desc, default Asc")
	flag.IntVar(&amount, "amount", 10, "Amount of Tasks which are shown")

	flag.StringVar(&orderBy, "o", "ID", "Order by, default ID")

	flag.StringVar(&description, "i", "", "Description of Task")
	flag.StringVar(&categoryName, "cname", "", "Name of the Category")
	flag.Int64Var(&hour, "h", -1, "Until Hour")
	flag.Int64Var(&day, "d", -1, "Until Day")

	flag.BoolVar(&delDoneTasks, "delDone", false, "Delete finished tasks")

	flag.Int64Var(&categoryId, "cid", -1, "Id of the Category")
	flag.BoolVar(&renderCategories, "cshow", false, "Print Categories")

	flag.BoolVar(&updateIds, "done", false, "")
	flag.BoolVar(&deleteIds, "del", false, "")

	flag.BoolVar(&table, "table", false, "")

	flag.Var(&idsLists, "ids", "IDs of Tasks")

	flag.BoolVar(&gitIssuesDownload, "gitissues", false, "")
	flag.StringVar(&addGitToken, "gittoken", "", "Your Github Access Token")

	flag.Parse()
}
