package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const gitApi = "https://api.github.com/issues"

// GithubToken represents an oauth token for Github
type GithubToken struct {
	token string
}

// Issue represents an issue of a repo from Github
type Issue struct {
	Title string `json:"title"`
	State string `json:"state"`
	Repo  Repo   `json:"repository"`
}

// Repo represents a repo from Github
type Repo struct {
	Name string `json:"name"`
}

// saveIssuesToDatabase saves all open issues assigned to the user of the oauth token
// as tasks under the category "github" to the database
func saveIssuesToDatabase() {
	token, err := getToken()
	if err != nil {
		log.Fatalln(err)
		return
	}
	saveIssuesToDatabaseImplementation(gitApi, token)

}

// saveIssuesToDatabaseImplementation holds the logic of the saveIssuesToDatabase function
func saveIssuesToDatabaseImplementation(url string, token string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "token " + token)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		x := fmt.Sprintf("Something went wrong, got a status code of %d, expected 200", res.StatusCode)
		panic(x)
	}

	var issues []Issue
	err = json.Unmarshal([]byte(body), &issues)
	if err != nil {
		panic(err)
	}
	categoryName := "Github"
	for _, i := range issues {
		description := fmt.Sprintf("%s: %s", i.Repo.Name, i.Title)
		SaveTask(categoryName, description, -1, -1)
	}

}

// getToken returns the oauth github token saved in the database
func getToken() (string, error) {
	githubTokenTable := `CREATE TABLE IF NOT EXISTS githubToken (
					id integer primary key check(id = 0), 
					token text not null
				);`
	_, err := db.Exec(githubTokenTable)
	checkErrorQueries(err, githubTokenTable)
	row := db.QueryRow("SELECT token FROM githubToken;")
	var githubToken GithubToken
	switch err := row.Scan(&githubToken.token); err {
	case sql.ErrNoRows:
		err = errors.New("no data in database, Use the -gittoken flag to add a token")
		return "", err
	case nil:
		return githubToken.token, nil
	default:
		return "", err
	}
}

// saveGitToken saves the given github token to the database
func saveGitToken(token string) error {
	lenToken := len(token)
	if lenToken != 40 {
		return errors.New(fmt.Sprintf("Github Token consists of 40 chars, your token was %d chars long", lenToken))

	}

	sqlStmt := fmt.Sprintf(`INSERT OR REPLACE INTO githubToken(id, token) VALUES ('%d', '%s');`, 0, token)
	_, err := db.Exec(sqlStmt)
	checkErrorQueries(err, sqlStmt)
	return nil
}
