package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	APPNAME      = "unix2date"
	MIN_UNIXTIME = 1000000000000 // 2001-09-09T01:46:40.000Z
	MAX_UNIXTIME = 2999999999999 // 2065-01-24T05:19:59.999Z
)

var (
	textPattern string         = `(?:^|[^0-9a-zA-Z-_])([12](?:\d{9}|\d{12}))(?:[^0-9a-zA-Z-_]|$)`
	regexText   *regexp.Regexp = regexp.MustCompile(textPattern)
	format10    string         = "2006-01-02T15:04:05Z"
	format13    string         = "2006-01-02T15:04:05.000Z"
	jsonPattern string         = `(?:" *:) *([12](?:\d{9}|\d{12})) *(?:[,}]|$)`
	regexJSON   *regexp.Regexp = regexp.MustCompile(jsonPattern)
)

type Parameter struct {
	filterFlag   bool
	noConvFlag   bool
	invertFlag   bool
	summaryFlag  bool
	filterFrom   string
	filterTo     string
	filterFromMS int64
	filterToMS   int64
}

type Summary struct {
	TotalNumberOfLines           int64  `json:"TotalNumberOfLines"`
	TotalNumberOfUnixtime        int64  `json:"TotalNumberOfUnixtime"`
	NumberOfLinesContainUnixtime int64  `json:"NumberOfLinesContainUnixtime"`
	NumberOfLinesWithoutUnixtime int64  `json:"NumberOfLinesWithoutUnixtime"`
	OldestUnixtime               int64  `json:"-"`
	OldestDatetime               string `json:"OldestDatetime,omitempty"`
	NewestUnixtime               int64  `json:"-"`
	NewestDatetime               string `json:"NewestDatetime,omitempty"`
	FilterCommandExample         string `json:"FilterCommandExample,omitempty"`
}

type Result struct {
	Text         string
	ShouldOutput bool
}

type Replacement struct {
	UnixtimeStr string
	StartIndex  int
	EndIndex    int
	TimeFormat  string
	JSONFormat  bool
}

func main() {
	s := Summary{}
	p := initParameters()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		result := replaceUnixtimeToDatetime(line, &s, &p)
		if result.ShouldOutput {
			fmt.Println(result.Text)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if p.summaryFlag {
		outputSummary(&s)
	}
}

func initParameters() Parameter {
	var p Parameter
	testing.Init()
	flag.CommandLine.Init(APPNAME, flag.ExitOnError)
	flag.CommandLine.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Usage:\n")
		fmt.Fprintf(o, "  %s [-s]\n", flag.CommandLine.Name())
		fmt.Fprintf(o, "  %s [-ni] [-f YYYY-mm-ddTHH:MM:SS(.NNN)Z] [-t YYYY-mm-ddTHH:MM:SS(.NNN)Z]\n", flag.CommandLine.Name())
		fmt.Fprintf(o, "Options:\n")
		fmt.Fprintf(o, "  -s (--summary)         Output only summary. (this option cannot be used together other options\n")
		fmt.Fprintf(o, "  -n (--no-convert)      Output without converting unixtime\n")
		fmt.Fprintf(o, "  -i (--invert-filter)   Invert and output filtering results\n")
		fmt.Fprintf(o, "  -f (--filter-from) [filter start date (ex. 2024-07-01T00:30:00Z)]\n")
		fmt.Fprintf(o, "  -t (--filter-to)   [filter end date   (ex. 2024-07-01T01:00:00Z)]\n")
		fmt.Fprintf(o, "                         Output only lines containing unixtime within specified period\n")
	}

	flag.StringVar(&p.filterFrom, "filter-from", "", "")
	flag.StringVar(&p.filterFrom, "f", "", "")
	flag.StringVar(&p.filterTo, "filter-to", "", "")
	flag.StringVar(&p.filterTo, "t", "", "")
	flag.BoolVar(&p.noConvFlag, "no-convert", false, "")
	flag.BoolVar(&p.noConvFlag, "n", false, "")
	flag.BoolVar(&p.invertFlag, "invert-filter", false, "")
	flag.BoolVar(&p.invertFlag, "i", false, "")
	flag.BoolVar(&p.summaryFlag, "summary", false, "")
	flag.BoolVar(&p.summaryFlag, "s", false, "")

	flag.Parse()

	if p.filterFrom != "" {
		p.filterFlag = true
		if len(p.filterFrom) == 20 {
			p.filterFrom = strings.Replace(p.filterFrom, "Z", ".000Z", 1)
		}
		unixtime, err := parsedUnixtime(p.filterFrom)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			flag.CommandLine.Usage()
			os.Exit(2)
		}
		p.filterFromMS = unixtime
	} else {
		p.filterFromMS = MIN_UNIXTIME
	}

	if p.filterTo != "" {
		p.filterFlag = true
		if len(p.filterTo) == 20 {
			p.filterTo = strings.Replace(p.filterTo, "Z", ".999Z", 1)
		}
		unixtime, err := parsedUnixtime(p.filterTo)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			flag.CommandLine.Usage()
			os.Exit(2)
		}
		p.filterToMS = unixtime
	} else {
		p.filterToMS = MAX_UNIXTIME
	}

	if p.filterToMS < p.filterFromMS {
		fmt.Fprintln(os.Stderr, "--filter-from(-f) value cannot be set larger than --filter-to(-t) value")
		flag.CommandLine.Usage()
		os.Exit(2)
	}

	if p.summaryFlag &&
		(p.filterFlag || p.invertFlag || p.noConvFlag) {
		fmt.Fprintln(os.Stderr, "--summary(-s) option cannot be used together other options")
		flag.CommandLine.Usage()
		os.Exit(2)
	}
	return p
}

func outputSummary(s *Summary) {
	filterCommandExample := APPNAME
	if s.OldestUnixtime > 0 {
		s.OldestDatetime = time.Unix(0, s.OldestUnixtime*int64(time.Millisecond)).UTC().Format(format10)
		filterCommandExample += " -f " + s.OldestDatetime
	}
	if s.NewestUnixtime > 0 {
		s.NewestDatetime = time.Unix(0, s.NewestUnixtime*int64(time.Millisecond)).UTC().Format(format10)
		filterCommandExample += " -t " + s.NewestDatetime
	}
	if s.OldestUnixtime > 0 || s.NewestUnixtime > 0 {
		s.FilterCommandExample = filterCommandExample
	}
	jsonOutput, err := jsonMarshalIndent(s)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("%s", string(jsonOutput))
}

func parsedUnixtime(datetimeStr string) (int64, error) {
	if len(datetimeStr) != 24 {
		return 0, fmt.Errorf("invalid datetime")
	}
	t, err := time.Parse(format13, datetimeStr)
	if err != nil {
		return 0, err
	}
	unixtime := t.UnixMilli()
	if unixtime < MIN_UNIXTIME || MAX_UNIXTIME < unixtime {
		return 0, fmt.Errorf("unacceptable date period")
	}
	return unixtime, nil
}

func jsonMarshalIndent(t interface{}) ([]byte, error) {
	marshalBuffer := &bytes.Buffer{}
	encoder := json.NewEncoder(marshalBuffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(t); err != nil {
		return nil, err
	}
	var indentBuffer bytes.Buffer
	err := json.Indent(&indentBuffer, marshalBuffer.Bytes(), "", "  ")
	return indentBuffer.Bytes(), err
}

func replaceUnixtimeToDatetime(text string, s *Summary, p *Parameter) *Result {
	orgText := text
	lineContainUnixtime := false
	inFilterPeriod := false
	for {
		ri := getReplaceInfo(text)
		if ri == nil {
			break
		}
		s.TotalNumberOfUnixtime++
		lineContainUnixtime = true

		var targetTime time.Time
		unixtime, _ := strconv.Atoi(ri.UnixtimeStr)
		if len(ri.UnixtimeStr) == 10 {
			targetTime = time.Unix(int64(unixtime), 0)
		} else if len(ri.UnixtimeStr) == 13 {
			targetTime = time.Unix(0, int64(unixtime)*int64(time.Millisecond))
		}
		datetimeStr := targetTime.UTC().Format(ri.TimeFormat)
		if ri.JSONFormat {
			datetimeStr = `"` + datetimeStr + `"`
		}
		text = text[:ri.StartIndex] + datetimeStr + text[ri.EndIndex:]

		unixMilli := targetTime.UnixMilli()
		if IsInFilterPeriod(unixMilli, p) {
			inFilterPeriod = true
		}
		updateUnixtimePeriod(unixMilli, s)
	}

	s.TotalNumberOfLines++
	if lineContainUnixtime {
		s.NumberOfLinesContainUnixtime++
	} else {
		s.NumberOfLinesWithoutUnixtime++
	}

	if p.summaryFlag {
		return &Result{text, false}
	} else if p.filterFlag {
		if (p.invertFlag && !inFilterPeriod) || (!p.invertFlag && inFilterPeriod) {
			if p.noConvFlag {
				return &Result{orgText, true}
			} else {
				return &Result{text, true}
			}
		}
	} else {
		if p.noConvFlag {
			return &Result{orgText, true}
		} else {
			return &Result{text, true}
		}
	}
	return &Result{text, false}
}

func updateUnixtimePeriod(unixtime int64, s *Summary) {
	if s.NewestUnixtime < unixtime {
		s.NewestUnixtime = unixtime
	}
	if s.OldestUnixtime > unixtime || s.OldestUnixtime == 0 {
		s.OldestUnixtime = unixtime
	}
}

func IsInFilterPeriod(unixtime int64, p *Parameter) bool {
	if p.filterFlag {
		if p.filterFromMS <= unixtime && unixtime <= p.filterToMS {
			return true
		}
	}
	return false
}

func getReplaceInfo(text string) *Replacement {
	if textMatch := regexText.FindStringSubmatchIndex(text); textMatch != nil {
		startIndex := textMatch[2]
		endIndex := textMatch[3]
		unixtimeStr := text[startIndex:endIndex]
		var timeFormat string
		if len(unixtimeStr) == 10 {
			timeFormat = format10
		} else if len(unixtimeStr) == 13 {
			timeFormat = format13
		}
		replacement := &Replacement{
			UnixtimeStr: unixtimeStr,
			StartIndex:  startIndex,
			EndIndex:    endIndex,
			TimeFormat:  timeFormat,
		}
		if jsonMatch := regexJSON.FindStringSubmatchIndex(text); jsonMatch != nil && startIndex == jsonMatch[2] {
			replacement.JSONFormat = true
		}
		return replacement
	}
	return nil
}
