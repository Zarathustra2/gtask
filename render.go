package main

import (
	"fmt"
	"github.com/gookit/color"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// RenderTableTasks renders the table with the tasks
func RenderTableTasks(orderBy string, sorted string) {

	tasks := AllTasks(orderBy, sorted)
	data := make([][]string, len(tasks))

	todo := 0

	for i := range data {
		task := tasks[i]

		if !task.Done {
			todo++
		}

		data[i] = task.StringArray()

	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"  ", "ID", "Description", "Until", "Category"})
	table.SetFooter([]string{"", "", "", "ToDo", strconv.Itoa(todo)})

	table.SetHeaderColor(
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.FgHiGreenColor},
		tablewriter.Colors{tablewriter.FgHiGreenColor},
		tablewriter.Colors{tablewriter.FgHiGreenColor},
		tablewriter.Colors{tablewriter.FgHiGreenColor},
	)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
	)

	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Println()
	table.Render()
}

// RenderTableCategories renders the table with all existing categories
func RenderTableCategories() {

	categories := AllCategories()
	data := make([][]string, len(categories))
	for i := range data {
		data[i] = categories[i].StringArray()
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name"})

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}

// AlignedOutputCategory represents a category and all tasks with the given category
// It also saves the amount of tasks for the category as well as the amount of tasks which
// have been finished/marked as done
type AlignedOutputCategory struct {
	total    int
	Tasks    []Task
	Category string
	Done     int
}

// RenderAligned renders the categories with its tasks out in the following format
func RenderAligned(orderBy string, sorted string) {

	tasks := AllTasks(orderBy, sorted)
	m := make(map[string]*AlignedOutputCategory)

	for i := range tasks {
		task := tasks[i]
		category := task.CategoryName

		if _, ok := m[category]; !ok {
			a := &AlignedOutputCategory{1, make([]Task, 1), category, 0}
			a.Tasks[0] = task
			m[category] = a
			if task.Done {
				a.Done++
			}
		} else {
			a := m[category]
			a.Tasks = append(m[category].Tasks, task)
			a.total++
			if task.Done {
				a.Done++
			}
		}
	}

	total, done := 0, 0
	fmt.Println()
	for _, value := range m {
		t, d := value.Render()
		fmt.Println()
		total += t
		done += d
	}

	fmt.Printf("%d left, %d done\n\n", total-done, done)

}

// Render renders a single AlignedOutputCategory in the following format
// Default - [0/2]
//        1. Clean House
//        2. Clean Dishes
func (a *AlignedOutputCategory) Render() (int, int) {
	color.OpUnderscore.Printf("%s", strings.Title(a.Category))
	fmt.Printf(" - [%d/%d]\n", a.Done, a.total)
	for _, t := range a.Tasks {
		d := t.Description
		if t.Done {
			d = color.OpStrikethrough.Sprint(d)
		}
		fmt.Printf("%15s  %d %s\n", t.getCheckBox(), t.Id, d)
	}

	return a.total, a.Done

}
