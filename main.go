package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	pattern := `(^|[^0-9]?)([0-9]{10,13})([^0-9]?|$)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(text, -1)

	var unixtimeList []string
	for _, match := range matches {
		unixtimeStr := match[2]
		if !contains(unixtimeList, unixtimeStr) {
			unixtimeList = append(unixtimeList, unixtimeStr)
		}
	}
	sort.Sort(byLength(unixtimeList))

	for _, unixtimeStr := range unixtimeList {
		var targetTime time.Time
		var timeFormat string
		unixtime, _ := strconv.Atoi(unixtimeStr)
		if len(unixtimeStr) == 10 {
			targetTime = time.Unix(int64(unixtime), 0)
			timeFormat = "2006-01-02T15:04:05Z"
		} else if len(unixtimeStr) == 13 {
			targetTime = time.Unix(0, int64(unixtime)*int64(time.Millisecond))
			timeFormat = "2006-01-02T15:04:05.000Z"
		} else {
			continue
		}
		if isInConvertiblePeriod(targetTime, now, daysStart, daysEnd) {
			text = strings.Replace(text, unixtimeStr, targetTime.UTC().Format(timeFormat), -1)
		}
	}
	return text
}

func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
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

// sort.Interface implementation for sorting slice by length
type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) > len(s[j])
}
