package main

import (
	"sync"
	"testing"
)

func TestValidateFlagVariables(t *testing.T) {
	tests := []struct {
		name    string
		fv      *FlagVariables
		isValid bool
	}{
		{"not specified option", &FlagVariables{}, true},
		{"-f invalid datetime", &FlagVariables{filterFrom: "a"}, false},
		{"-f empty string", &FlagVariables{filterFrom: ""}, true},
		{"-t invalid datetime", &FlagVariables{filterTo: "a"}, false},
		{"-t empty string", &FlagVariables{filterTo: ""}, true},
		{"-s with -n", &FlagVariables{summaryFlag: true, noConvFlag: true}, false},
		{"-s with -i", &FlagVariables{summaryFlag: true, invertFlag: true}, false},
		{"-s with -f", &FlagVariables{summaryFlag: true, filterFrom: "a"}, false},
		{"-s with -t", &FlagVariables{summaryFlag: true, filterTo: "a"}, false},
		{"-s with -qt", &FlagVariables{summaryFlag: true, quotations: "a"}, true},
		{"-s with -sp", &FlagVariables{summaryFlag: true, separators: "a"}, true},
		{"out-of-range for -f", &FlagVariables{filterFrom: "1950-12-24T00:00:00Z"}, false},
		{"within-range for -f", &FlagVariables{filterFrom: "2014-12-24T00:00:00Z"}, true},
		{"out-of-range for -t", &FlagVariables{filterTo: "2080-12-24T00:00:00Z"}, false},
		{"within-range for -t", &FlagVariables{filterTo: "2014-12-24T00:00:00Z"}, true},
		{"-f newer than -t", &FlagVariables{filterFrom: "2014-12-24T00:00:00Z", filterTo: "2014-12-23T23:59:59Z"}, false},
		{"-t newer than -f", &FlagVariables{filterTo: "2014-12-24T00:00:00Z", filterFrom: "2014-12-23T23:59:59Z"}, true},
		{"millisec for -f", &FlagVariables{filterFrom: "2014-12-24T00:00:00.000Z"}, true},
		{"millisec for -t", &FlagVariables{filterFrom: "2014-12-24T00:00:00.999Z"}, true},
	}
	for _, tt := range tests {
		initializeFlagVariables(tt.fv)
		if _, err := validateFlagVariables(tt.fv); (err == nil) != tt.isValid {
			t.Errorf("%20s [ NG ] => expect: %v actual: %v", tt.name, tt.isValid, (err == nil))
		}
	}
}

func initializeFlagVariables(fv *FlagVariables) {
	if fv.quotations == "" {
		fv.quotations = DEF_QUOTATIONS
	}
	if fv.separators == "" {
		fv.separators = DEF_SEPARATORS
	}
}

func TestReplaceUnixtimeToDatetimeFilterTest(t *testing.T) {
	s := &Summary{mu: &sync.Mutex{}}
	tests := []struct {
		name   string
		fv     *FlagVariables
		input  string
		expect bool
	}{
		{"not include unixtime with filter",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.123Z"},
			"", false},
		{"not include unixtime with filter",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.123Z"},
			"test", false},
		{"include unixtime within filter period #1",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z", filterTo: "2009-02-13T23:31:30.000Z"},
			"1234567890000", true},
		{"include unixtime within filter period #2",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z"},
			"1234567890123", true},
		{"include unixtime within filter period #3",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z"},
			"2345678890", true},
		{"include unixtime within filter period #4",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.999Z", filterTo: "2009-02-13T23:31:30.999Z"},
			"1234567890999", true},
		{"include unixtime within filter period #5",
			&FlagVariables{filterTo: "2009-02-13T23:31:30.999Z"},
			"1234567890", true},
		{"include unixtime within filter period #6",
			&FlagVariables{filterTo: "2009-02-13T23:31:30.999Z"},
			"1123456789", true},
		{"include unixtime not within filter period #1",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z"},
			"1234567889999", false},
		{"include unixtime not within filter period #2",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z", filterTo: "2009-02-13T23:31:30.000Z"},
			"1234567890001", false},
		{"include unixtime not within filter period #3",
			&FlagVariables{filterTo: "2009-02-13T23:31:30.000Z"},
			"1234567890001", false},
		{"summary",
			&FlagVariables{summaryFlag: true},
			"1234567890001", false},
		{"include unixtime with invert flag within filter period",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z", filterTo: "2009-02-13T23:31:30.000Z", invertFlag: true},
			"1234567890000", false},
		{"include unixtime with invert flag not within filter period",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z", filterTo: "2009-02-13T23:31:30.001Z", invertFlag: true},
			"1234567890002", true},
		{"include unixtime with noConvert flag",
			&FlagVariables{noConvFlag: true},
			"1234567890001", true},
		{"both unixtimes are within filter period",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.000Z", filterTo: "2009-02-13T23:31:30.003Z"},
			"1234567890001 1234567890002 ", true},
		{"one of two unixtimes is within filter period #1",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.001Z", filterTo: "2009-02-13T23:31:30.003Z"},
			"1234567890000 1234567890002 ", true},
		{"one of two unixtimes is within filter period #2",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.001Z", filterTo: "2009-02-13T23:31:30.003Z"},
			"1234567890004 1234567890002 ", true},
		{"both unixtimes are not within filter period",
			&FlagVariables{filterFrom: "2009-02-13T23:31:30.001Z", filterTo: "2009-02-13T23:31:30.003Z"},
			"1234567890000 1234567890004 ", false},
	}
	for _, tt := range tests {
		initializeFlagVariables(tt.fv)
		p, _ := validateFlagVariables(tt.fv)
		input := &Input{Index: 0, Text: tt.input}
		if actual := replaceUnixtimeToDatetime(input, s, p); actual.NeedToOutput != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.NeedToOutput)
		} else {
			t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.NeedToOutput)
		}
	}
}

func TestReplaceUnixtimeToDatetimeQuotationsAndSeparatorsTest(t *testing.T) {
	s := &Summary{mu: &sync.Mutex{}}
	tests := []struct {
		name   string
		fv     *FlagVariables
		input  string
		expect string
	}{
		{"use comma as separatos #1", &FlagVariables{separators: ","},
			",1234567890", ",2009-02-13T23:31:30Z"},
		{"use comma as separatos #2", &FlagVariables{separators: ","},
			",1234567890,", ",2009-02-13T23:31:30Z,"},
		{"use comma as separatos #3", &FlagVariables{separators: ","},
			"1234567890,", "2009-02-13T23:31:30Z,"},
		{"use space as separatos #1", &FlagVariables{separators: " "},
			" 1234567890", " 2009-02-13T23:31:30Z"},
		{"use space as separatos #2", &FlagVariables{separators: " "},
			" 1234567890 ", " 2009-02-13T23:31:30Z "},
		{"use space as separatos #3", &FlagVariables{separators: " "},
			"1234567890 ", "2009-02-13T23:31:30Z "},
		{"use tab as separatos #1", &FlagVariables{separators: "\t"},
			"	1234567890	", "	2009-02-13T23:31:30Z	"},
		{"use double-quote as quotations #1", &FlagVariables{quotations: "\""},
			"\"1234567890\"", "\"2009-02-13T23:31:30Z\""},
		{"use double-quote as quotations #2", &FlagVariables{quotations: "\""},
			" 1234567890\"", " 1234567890\""},
		{"use multi-kind quote as quotations #1", &FlagVariables{quotations: "\"'"},
			"'1234567890\"", "'2009-02-13T23:31:30Z\""},
	}
	for _, tt := range tests {
		initializeFlagVariables(tt.fv)
		p, _ := validateFlagVariables(tt.fv)
		input := &Input{Index: 0, Text: tt.input}
		if actual := replaceUnixtimeToDatetime(input, s, p); actual.Text != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		} else {
			t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		}
	}
}

func TestReplaceUnixtimeToDatetimeNoConvTest(t *testing.T) {
	fv := &FlagVariables{filterFrom: "2009-02-13T23:31:30.123Z"}
	initializeFlagVariables(fv)
	p, _ := validateFlagVariables(fv)
	s := &Summary{mu: &sync.Mutex{}}
	tests := []struct {
		name       string
		noConvFlag bool
		input      string
		expect     string
	}{
		{"include unixtime with noConvert flag", true,
			"1234567890123 2345678901234", "1234567890123 2345678901234"},
		{"include unixtime without noConvert flag", false,
			"1234567890123 2345678901234", "2009-02-13T23:31:30.123Z 2044-05-01T01:28:21.234Z"},
	}
	for _, tt := range tests {
		p.noConvFlag = tt.noConvFlag
		input := &Input{Index: 0, Text: tt.input}
		if actual := replaceUnixtimeToDatetime(input, s, p); actual.Text != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		} else {
			t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		}
	}
}

func TestReplaceUnixtimeToDatetimeWithJSON(t *testing.T) {
	fv := &FlagVariables{}
	initializeFlagVariables(fv)
	p, _ := validateFlagVariables(fv)
	s := &Summary{mu: &sync.Mutex{}}
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"add double quote when json format with eol",
			`{"test":1234567890123`, `{"test":"2009-02-13T23:31:30.123Z"`},
		{"add double quote when json format with trailing comma",
			`{"test":1234567890123,`, `{"test":"2009-02-13T23:31:30.123Z",`},
		{"add double quote when close json with curly bracket",
			`{"test": 1234567890123}`, `{"test": "2009-02-13T23:31:30.123Z"}`},
		{"dont add double quote",
			`{"test" :"1234567890123"}`, `{"test" :"2009-02-13T23:31:30.123Z"}`},
	}
	for _, tt := range tests {
		input := &Input{Index: 0, Text: tt.input}
		if actual := replaceUnixtimeToDatetime(input, s, p); actual.Text != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		} else {
			t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		}
	}
}

func TestReplaceUnixtimeToDatetime(t *testing.T) {
	fv := &FlagVariables{}
	initializeFlagVariables(fv)
	p, _ := validateFlagVariables(fv)
	s := &Summary{mu: &sync.Mutex{}}
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"empty", "", ""},
		{"one space", " ", " "},
		{"two spaces", " ", " "},
		{"9 digits", "172099999", "172099999"},
		{"10 digits", "1720999999", "2024-07-14T23:33:19Z"},
		{"12 digits", "172099999932", "172099999932"},
		{"13 digits", "1720999999321", "2024-07-14T23:33:19.321Z"},
		{"14 digits", "17209999993216", "17209999993216"},
		{"19 digits", "1720999999172099999", "1720999999172099999"},
		{"20 digits", "17209999991720999999", "17209999991720999999"},
		{"22 digits", "1720999999172099999945", "1720999999172099999945"},
		{"23 digits", "17209999991720999999321", "17209999991720999999321"},
		{"24 digits", "172099999917209999993217", "172099999917209999993217"},
		{"25 digits", "1720999999172099999932178", "1720999999172099999932178"},
		{"26 digits", "17209999991720999999321784", "17209999991720999999321784"},
		{"27 digits", "172099999917209999993217842", "172099999917209999993217842"},
		{"tab before unixtime", "	1720999990", "	2024-07-14T23:33:10Z"},
		{"space before unixtime", " 1720999990", " 2024-07-14T23:33:10Z"},
		{"space after unixtime", "1720999990 ", "2024-07-14T23:33:10Z "},
		{"dquote after unixtime", "1720999990\"", "1720999990\""},
		{"same 10 digits", "1720999999 1720999999", "2024-07-14T23:33:19Z 2024-07-14T23:33:19Z"},
		{"same 13 digits", "1722543769134 1722543769134", "2024-08-01T20:22:49.134Z 2024-08-01T20:22:49.134Z"},
		{"equal + 10 digits", "=1234567890 ", "=1234567890 "},
		{"dot + 10 digits", ".1234567890 ", ".1234567890 "},
		{"underscore + 10 digits", "_1234567890 ", "_1234567890 "},
		{"hyphen + 10 digits", "-1234567890 ", "-1234567890 "},
		{"a-z + 10 digits", "abc1234567890 ", "abc1234567890 "},
		{"10 digits + A-Z", "1234567890ABC", "1234567890ABC"},
		{"A-Z + 13 digits + a-z", "ABC1234567890123abc", "ABC1234567890123abc"},
		{"comma separated 10 digits", "a,1720999999,1722543769", "a,2024-07-14T23:33:19Z,2024-08-01T20:22:49Z"},
		{"space separated 13 digits", "1720999999000 1722543769134 ", "2024-07-14T23:33:19.000Z 2024-08-01T20:22:49.134Z "},
		{"10 digits and 13 digits", "1720999999 1722543769876", "2024-07-14T23:33:19Z 2024-08-01T20:22:49.876Z"},
		{"13 digits and 10 digits", "1720999999111  1722543769", "2024-07-14T23:33:19.111Z  2024-08-01T20:22:49Z"},
		{"multi bytes #1", "あ1722543769･1722543769876／", "あ1722543769･1722543769876／"},
		{"multi bytes #2", "１７２２５４３７６９", "１７２２５４３７６９"},
	}
	for _, tt := range tests {
		input := &Input{Index: 0, Text: tt.input}
		if actual := replaceUnixtimeToDatetime(input, s, p); actual.Text != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		} else {
			t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual.Text)
		}
	}
}
