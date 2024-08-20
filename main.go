package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	Version     = "unset"
	VersionFlag bool
)

const (
	APPNAME           = "unix2date"
	MIN_UNIXTIME      = 1000000000000 // 2001-09-09T01:46:40.000Z
	MAX_UNIXTIME      = 2999999999999 // 2065-01-24T05:19:59.999Z
	DEF_QUOTATIONS    = `"`
	DEF_SEPARATORS    = ` ,\t`
	DATETIME_FORMAT10 = "2006-01-02T15:04:05Z"
	DATETIME_FORMAT13 = "2006-01-02T15:04:05.000Z"
	UNIXTIME_PATTERN  = `([12](?:\d{12}|\d{9}))`
	TYPE_JSON         = iota
	TYPE_QT
	TYPE_SP
)

type FlagVariables struct {
	noConvFlag  bool
	invertFlag  bool
	summaryFlag bool
	filterFrom  string
	filterTo    string
	quotations  string
	separators  string
}

type Parameter struct {
	filterFlag      bool
	noConvFlag      bool
	invertFlag      bool
	summaryFlag     bool
	filterFromMS    int64
	filterToMS      int64
	replacePatterns []ReplacePattern
}

type ReplacePattern struct {
	Regexp *regexp.Regexp
	Type   int
}

type Summary struct {
	mu                           *sync.Mutex
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

type Input struct {
	Index int64
	Text  string
}

type Result struct {
	Index        int64
	Text         string
	NeedToOutput bool
}

type Output struct {
	mu         *sync.Mutex
	Index      int64
	BufResults map[int64]*Result
}

type ReplaceInfo struct {
	UnixtimeStr string
	StartIndex  int
	EndIndex    int
	TimeFormat  string
	NeedQuote   bool
}

func main() {
	s := &Summary{mu: &sync.Mutex{}}
	fv, fs := parseFlagSet()
	if VersionFlag {
		fmt.Println(Version)
		os.Exit(0)
	}
	p, err := validateFlagVariables(fv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fs.Usage()
		os.Exit(2)
	}

	var wg sync.WaitGroup
	var lineCount int64
	output := &Output{mu: &sync.Mutex{}, BufResults: map[int64]*Result{}}
	limiter := make(chan struct{}, runtime.NumCPU())
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		input := &Input{Index: lineCount, Text: line}
		limiter <- struct{}{}
		wg.Add(1)
		go func(input *Input, output *Output) {
			defer func() {
				<-limiter
				wg.Done()
			}()
			result := replaceUnixtimeToDatetime(input, s, p)
			outputLines(output, result)
		}(input, output)
		lineCount++
	}
	wg.Wait()
	outputLines(output, nil)

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if fv.summaryFlag {
		outputSummary(s)
	}
}

func outputLines(output *Output, result *Result) {
	output.mu.Lock()
	defer output.mu.Unlock()
	if result != nil {
		output.BufResults[result.Index] = result
	}
	for len(output.BufResults) != 0 {
		if tmpRes, ok := output.BufResults[output.Index]; ok {
			if tmpRes.NeedToOutput {
				fmt.Println(tmpRes.Text)
			}
			delete(output.BufResults, output.Index)
			output.Index++
		} else {
			break
		}
	}
}

func parseFlagSet() (*FlagVariables, *flag.FlagSet) {
	fv := FlagVariables{}
	flagSet := flag.NewFlagSet(APPNAME, flag.ExitOnError)
	flagSet.Usage = func() {
		o := flagSet.Output()
		fmt.Fprintf(o, "---\n")
		fmt.Fprintf(o, "Usage:\n")
		fmt.Fprintf(o, "  %s [-s]\n", flagSet.Name())
		fmt.Fprintf(o, "  %s [-ni] [-f YYYY-mm-ddTHH:MM:SS(.NNN)Z] [-t YYYY-mm-ddTHH:MM:SS(.NNN)Z]\n", flagSet.Name())
		fmt.Fprintf(o, "Options:\n")
		fmt.Fprintf(o, "  -s (--summary)         Output only summary. (this option cannot be used with {-n,-i,-f,-t} options\n")
		fmt.Fprintf(o, "  -n (--no-convert)      Output unixtime without converting\n")
		fmt.Fprintf(o, "  -i (--invert-filter)   Invert and output filtered results\n")
		fmt.Fprintf(o, "  -f (--filter-from) [filter start date (ex. 2024-07-01T00:30:00Z)]\n")
		fmt.Fprintf(o, "  -t (--filter-to)   [filter end date   (ex. 2024-07-01T01:00:00Z)]\n")
		fmt.Fprintf(o, "                         Output only lines containing unixtime within specified period\n")
		fmt.Fprintf(o, "  -qt (--quotations) [characters for quotations (default: `\"`)\n")
		fmt.Fprintf(o, "  -sp (--separators) [characters for separators (default: ` ,\\t`)\n")
		fmt.Fprintf(o, "                         Set characters to detect unixtime\n")
	}

	flagSet.BoolVar(&VersionFlag, "v", false, "")
	flagSet.StringVar(&fv.filterFrom, "filter-from", "", "")
	flagSet.StringVar(&fv.filterFrom, "f", "", "")
	flagSet.StringVar(&fv.filterTo, "filter-to", "", "")
	flagSet.StringVar(&fv.filterTo, "t", "", "")
	flagSet.BoolVar(&fv.noConvFlag, "no-convert", false, "")
	flagSet.BoolVar(&fv.noConvFlag, "n", false, "")
	flagSet.BoolVar(&fv.invertFlag, "invert-filter", false, "")
	flagSet.BoolVar(&fv.invertFlag, "i", false, "")
	flagSet.BoolVar(&fv.summaryFlag, "summary", false, "")
	flagSet.BoolVar(&fv.summaryFlag, "s", false, "")
	flagSet.StringVar(&fv.quotations, "quotations", DEF_QUOTATIONS, "")
	flagSet.StringVar(&fv.quotations, "qt", DEF_QUOTATIONS, "")
	flagSet.StringVar(&fv.separators, "separators", DEF_SEPARATORS, "")
	flagSet.StringVar(&fv.separators, "sp", DEF_SEPARATORS, "")

	flagSet.Parse(os.Args[1:])

	return &fv, flagSet
}

func validateFlagVariables(fv *FlagVariables) (*Parameter, error) {
	p := Parameter{noConvFlag: fv.noConvFlag, invertFlag: fv.invertFlag, summaryFlag: fv.summaryFlag}

	if fv.filterFrom != "" {
		p.filterFlag = true
		if len(fv.filterFrom) == 20 {
			fv.filterFrom = strings.Replace(fv.filterFrom, "Z", ".000Z", 1)
		}
		unixtime, err := parsedUnixtime(fv.filterFrom)
		if err != nil {
			return nil, err
		}
		p.filterFromMS = unixtime
	} else {
		p.filterFromMS = MIN_UNIXTIME
	}

	if fv.filterTo != "" {
		p.filterFlag = true
		if len(fv.filterTo) == 20 {
			fv.filterTo = strings.Replace(fv.filterTo, "Z", ".999Z", 1)
		}
		unixtime, err := parsedUnixtime(fv.filterTo)
		if err != nil {
			return nil, err
		}
		p.filterToMS = unixtime
	} else {
		p.filterToMS = MAX_UNIXTIME
	}

	if p.filterToMS < p.filterFromMS {
		return nil, fmt.Errorf("--filter-from(-f) value cannot be newer than --filter-to(-t) value")
	}

	if fv.summaryFlag &&
		(p.filterFlag || fv.invertFlag || fv.noConvFlag) {
		return nil, fmt.Errorf("--summary(-s) option cannot be used with other options")
	}

	if fv.invertFlag && !p.filterFlag {
		return nil, fmt.Errorf("--invert(-i) option must be used with --filter-from(-f) or --filter-to(-t) option")
	}

	p.replacePatterns = generateReplacePatternList(fv.quotations, fv.separators)

	return &p, nil
}

func generateReplacePatternList(quotations, separators string) []ReplacePattern {
	var replacePatterns []ReplacePattern
	if len(separators) > 0 {
		regexStr := `(?:^|[` + separators + `])` + UNIXTIME_PATTERN + `(?:[` + separators + `]|$)`
		replacePattern := ReplacePattern{
			Regexp: regexp.MustCompile(regexStr),
			Type:   TYPE_SP,
		}
		replacePatterns = append(replacePatterns, replacePattern)
	}
	if len(quotations) > 0 {
		regexStr := `(?:[` + quotations + `])` + UNIXTIME_PATTERN + `(?:[` + quotations + `])`
		replacePattern := ReplacePattern{
			Regexp: regexp.MustCompile(regexStr),
			Type:   TYPE_QT,
		}
		replacePatterns = append(replacePatterns, replacePattern)
	}
	replacePattern := ReplacePattern{
		Regexp: regexp.MustCompile(`(?:" *:) *` + UNIXTIME_PATTERN + ` *(?:[,}]|$)`),
		Type:   TYPE_JSON,
	}
	replacePatterns = append(replacePatterns, replacePattern)
	return replacePatterns
}

func outputSummary(s *Summary) {
	filterCommandExample := APPNAME
	if s.OldestUnixtime > 0 {
		s.OldestDatetime = time.Unix(0, s.OldestUnixtime*int64(time.Millisecond)).UTC().Format(DATETIME_FORMAT10)
		filterCommandExample += " -f " + s.OldestDatetime
	}
	if s.NewestUnixtime > 0 {
		s.NewestDatetime = time.Unix(0, s.NewestUnixtime*int64(time.Millisecond)).UTC().Format(DATETIME_FORMAT10)
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
	t, err := time.Parse(DATETIME_FORMAT13, datetimeStr)
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

func replaceUnixtimeToDatetime(input *Input, s *Summary, p *Parameter) *Result {
	text := input.Text
	orgText := input.Text
	lineContainUnixtime := false
	inFilterPeriod := false
	for {
		ri := getReplaceInfo(text, p.replacePatterns)
		if ri == nil {
			break
		}
		atomic.AddInt64(&s.TotalNumberOfUnixtime, 1)
		lineContainUnixtime = true

		var targetTime time.Time
		unixtime, _ := strconv.Atoi(ri.UnixtimeStr)
		if len(ri.UnixtimeStr) == 10 {
			targetTime = time.Unix(int64(unixtime), 0)
		} else if len(ri.UnixtimeStr) == 13 {
			targetTime = time.Unix(0, int64(unixtime)*int64(time.Millisecond))
		}
		datetimeStr := targetTime.UTC().Format(ri.TimeFormat)
		if ri.NeedQuote {
			datetimeStr = `"` + datetimeStr + `"`
		}
		text = text[:ri.StartIndex] + datetimeStr + text[ri.EndIndex:]

		unixMilli := targetTime.UnixMilli()
		if IsInFilterPeriod(unixMilli, p) {
			inFilterPeriod = true
		}
		updateUnixtimePeriod(unixMilli, s)
	}

	atomic.AddInt64(&s.TotalNumberOfLines, 1)
	if lineContainUnixtime {
		atomic.AddInt64(&s.NumberOfLinesContainUnixtime, 1)
	} else {
		atomic.AddInt64(&s.NumberOfLinesWithoutUnixtime, 1)
	}

	if p.summaryFlag {
		return &Result{input.Index, text, false}
	} else if p.filterFlag {
		if (p.invertFlag && !inFilterPeriod) || (!p.invertFlag && inFilterPeriod) {
			if p.noConvFlag {
				return &Result{input.Index, orgText, true}
			} else {
				return &Result{input.Index, text, true}
			}
		}
	} else {
		if p.noConvFlag {
			return &Result{input.Index, orgText, true}
		} else {
			return &Result{input.Index, text, true}
		}
	}
	return &Result{input.Index, text, false}
}

func updateUnixtimePeriod(unixtime int64, s *Summary) {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func getReplaceInfo(text string, replacePatterns []ReplacePattern) *ReplaceInfo {
	for _, rp := range replacePatterns {
		if textMatch := rp.Regexp.FindStringSubmatchIndex(text); textMatch != nil {
			startIndex := textMatch[2]
			endIndex := textMatch[3]
			unixtimeStr := text[startIndex:endIndex]
			var timeFormat string
			if len(unixtimeStr) == 10 {
				timeFormat = DATETIME_FORMAT10
			} else if len(unixtimeStr) == 13 {
				timeFormat = DATETIME_FORMAT13
			}
			replaceInfo := &ReplaceInfo{
				UnixtimeStr: unixtimeStr,
				StartIndex:  startIndex,
				EndIndex:    endIndex,
				TimeFormat:  timeFormat,
			}
			if rp.Type == TYPE_JSON {
				replaceInfo.NeedQuote = true
			}
			return replaceInfo
		}
	}
	return nil
}
