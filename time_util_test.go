package main

import (
	"reflect"
	"testing"
	"time"

	. "github.com/logrusorgru/aurora"
)

func Test_convertDate(t *testing.T) {
	type args struct {
		unixTimestamp int64
	}
	date := time.Date(1997, 1, 1, 1, 1, 1, 0, time.Local)
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"base-case", args{date.Unix()}, date},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertDate(tt.args.unixTimestamp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_timeUntil(t *testing.T) {
	type args struct {
		unixTimestamp int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{time.Now().Unix() + 10000}, Red("2h").String()},
		{"", args{time.Now().Unix() + 200000}, Magenta("2d").String()},
		{"", args{time.Now().Unix() + 700000}, Green("8w").String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := timeUntil(tt.args.unixTimestamp); got != tt.want {
				t.Errorf("timeUntil() = %v, want %v", got, tt.want)
			}
		})
	}
}
