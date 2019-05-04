package main

import (
	"testing"
)

// Not exactly sure how to write tests and test the rendering
// Maybe return a string, which then gets printed by the main method
// this way we could assert the string
// I am open for any ideas
// For now we just test that the functions "work"

func TestRenderAligned(t *testing.T) {
	defer cleanDatabase()
	createThreeTasks()
	RenderAligned("id", "")

}

func TestRenderTableTasks(t *testing.T) {
	defer cleanDatabase()
	createThreeTasks()

	RenderTableTasks("id", "")
}

func TestRenderTableCategories(t *testing.T) {
	defer cleanDatabase()
	createThreeTasks()

	RenderTableCategories()
}