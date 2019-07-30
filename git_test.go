package main

import (
	"log"
	"os"
	"testing"
)

func Test_saveGitToken(t *testing.T) {
	defer cleanDatabase()
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Correct Token", args{"1421117574054999149514211175740549991495"}, false},
		{"Too Long", args{"14211175740549991495142111757405499914951"}, true},
		{"Too Short", args{"142111757405499914951421117574054999149"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := saveGitToken(tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("saveGitToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getToken(t *testing.T) {
	defer cleanDatabase()

	_, err := getToken()
	if err == nil {
		t.Error("Expected error got nil instead")
	}

	tokenInsert := "1421117574054999149514211175740549991495"
	_ = saveGitToken(tokenInsert)

	token, err := getToken()

	if err != nil {
		t.Errorf("Error: %s, expected nil", err)
	}

	if token != tokenInsert {
		t.Errorf("Got %s, Expected %s", token, tokenInsert)
	}

}

func Test_saveIssuesToDatabaseImplementation(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	defer cleanDatabase()
	saveIssuesToDatabaseImplementation("https://api.github.com/repos/golang/go/issues", "")

	var count int
	row := testDB.QueryRow("SELECT COUNT(*) FROM tasks")
	err := row.Scan(&count)

	if err != nil {
		log.Fatal(err)
	}

	// We dont know how many issues we will get, so we just want to check
	// that at least one exists
	if count == 0 {
		t.Errorf("Got %d, expected at least 1", count)
	}
}
