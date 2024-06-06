package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	daysStart int
	daysEnd   int
)

func init() {
	flag.IntVar(&daysStart, "days-start", 365, "Start date of convertible period (days ago from now)")
	flag.IntVar(&daysEnd, "days-end", 365, "End date of convertible period (days later from now)")
}

func main() {
	flag.Parse()
	now := time.Now()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(replaceUnixtimeToDatetime(line, now))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
}

func replaceUnixtimeToDatetime(text string, now time.Time) string {
	for {
		ri := getReplaceInfo(text, now)
		if ri == nil {
			break
		}
		unixtimeStr := ri.UnixtimeStr
		startIndex := ri.StartIndex
		endIndex := ri.EndIndex
		timeFormat := ri.TimeFormat

		var targetTime time.Time
		unixtime, _ := strconv.Atoi(unixtimeStr)
		if len(unixtimeStr) == 10 {
			targetTime = time.Unix(int64(unixtime), 0)
		} else if len(unixtimeStr) == 13 {
			targetTime = time.Unix(0, int64(unixtime)*int64(time.Millisecond))
		}
		datetimeStr := targetTime.UTC().Format(timeFormat)
		text = text[:startIndex] + datetimeStr + text[endIndex:]
	}

	return text
}

type Replacement struct {
	UnixtimeStr string
	StartIndex  int
	EndIndex    int
	TimeFormat  string
}

func getReplaceInfo(text string, now time.Time) *Replacement {
	pattern := `(^|[^0-9])([0-9]{10,13})([^0-9]|$)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		startIndex := match[4]
		endIndex := match[5]
		unixtimeStr := text[startIndex:endIndex]
		var targetTime time.Time
		var timeFormat string
		unixtime, _ := strconv.Atoi(unixtimeStr)
		if len(unixtimeStr) == 10 {
			targetTime = time.Unix(int64(unixtime), 0)
			timeFormat = "2006-01-02T15:04:05Z"
		} else if len(unixtimeStr) == 13 {
			targetTime = time.Unix(0, int64(unixtime)*int64(time.Millisecond))
			timeFormat = "2006-01-02T15:04:05.000Z"
		}
		if isInConvertiblePeriod(targetTime, now, daysStart, daysEnd) {
			return &Replacement{
				UnixtimeStr: unixtimeStr,
				StartIndex:  startIndex,
				EndIndex:    endIndex,
				TimeFormat:  timeFormat,
			}
		}
	}
	return nil
}

func isInConvertiblePeriod(targetTime time.Time, now time.Time, daysStart int, daysEnd int) (isConvertible bool) {
	if now.Sub(targetTime) > 0 {
		if int(now.Sub(targetTime).Hours()/24) < daysStart {
			isConvertible = true
		}
	} else {
		if int(targetTime.Sub(now).Hours()/24) < daysEnd {
			isConvertible = true
		}
	}
	return
}
