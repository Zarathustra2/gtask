package main

import (
	"fmt"
	"time"

	. "github.com/logrusorgru/aurora"
)

// converts a unixDateTime to time.Time
func convertDate(unixTimestamp int64) time.Time {
	tm := time.Unix(unixTimestamp, 0)
	return tm
}

// timeUntil converts a Time into a readable until-string
// it returns either the days left, hours left or minutes left
func timeUntil(unixTimestamp int64) string {

	if unixTimestamp == 0 {
		return "-"
	}

	tm := convertDate(unixTimestamp)
	since := time.Until(tm).Seconds()

	daysUntil := int(since / 86400)
	intResult := daysUntil
	formatString := ""

	if daysUntil >= 7 {
		formatString = Green(fmt.Sprintf("%dw", daysUntil)).String()
	} else if daysUntil >= 0 && daysUntil <= 1 {
		hours := int(since / 3600)
		intResult = hours
		formatString = Red(fmt.Sprintf("%dh", intResult)).String()
	} else {
		formatString = Magenta(fmt.Sprintf("%dd", intResult)).String()
	}

	return formatString
}
