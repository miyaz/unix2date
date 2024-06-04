package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	daysStart int
	daysEnd   int
	now       time.Time
)

func init() {
	flag.IntVar(&daysStart, "days-start", 365, "Start date of convertible period (days ago from now)")
	flag.IntVar(&daysEnd, "days-end", 365, "End date of convertible period (days later from now)")
	flag.Parse()
}

func main() {
	now = time.Now()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(replaceUnixtime2Datetime(line))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
}

func replaceUnixtime2Datetime(text string) string {
	pattern := `(^|[^0-9]?)([0-9]{10,13})([^0-9]?|$)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		timeStr := match[2]

		var timeObject time.Time
		var timeFormat string
		unixtime, _ := strconv.Atoi(timeStr)
		if len(timeStr) == 10 {
			timeObject = time.Unix(int64(unixtime), 0)
			timeFormat = "2006-01-02T15:04:05Z"
		} else if len(timeStr) == 13 {
			timeObject = time.Unix(0, int64(unixtime)*int64(time.Millisecond))
			timeFormat = "2006-01-02T15:04:05.000Z"
		} else {
			continue
		}
		if now.Sub(timeObject) > 0 {
			if int(now.Sub(timeObject).Hours()/24) > daysStart {
				continue
			}
		} else {
			if int(timeObject.Sub(now).Hours()/24) > daysEnd {
				continue
			}
		}
		text = strings.Replace(text, timeStr, timeObject.UTC().Format(timeFormat), -1)
	}
	return text
}
