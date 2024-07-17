package main

import (
	"testing"
)

func TestReplaceUnixtimeToDatetimeFilterTest(t *testing.T) {
	tests := []struct {
		name   string
		param  *Parameter
		input  string
		expect bool
	}{
		{"not include unixtime with filter",
			&Parameter{filterFlag: true, filterFromMS: 1234567890123, filterToMS: MAX_UNIXTIME},
			"", false},
		{"not include unixtime with filter",
			&Parameter{filterFlag: true, filterFromMS: 1234567890123, filterToMS: MAX_UNIXTIME},
			"test", false},
		{"include unixtime within filter period #1",
			&Parameter{filterFlag: true, filterFromMS: 1234567890000, filterToMS: 1234567890000},
			"1234567890000", true},
		{"include unixtime within filter period #2",
			&Parameter{filterFlag: true, filterFromMS: 1234567890000, filterToMS: MAX_UNIXTIME},
			"1234567890123", true},
		{"include unixtime within filter period #3",
			&Parameter{filterFlag: true, filterFromMS: 1234567890000, filterToMS: MAX_UNIXTIME},
			"2345678890", true},
		{"include unixtime within filter period #4",
			&Parameter{filterFlag: true, filterFromMS: 1234567890999, filterToMS: 1234567890999},
			"1234567890999", true},
		{"include unixtime within filter period #5",
			&Parameter{filterFlag: true, filterFromMS: MIN_UNIXTIME, filterToMS: 1234567890999},
			"1234567890", true},
		{"include unixtime within filter period #6",
			&Parameter{filterFlag: true, filterFromMS: MIN_UNIXTIME, filterToMS: 1234567890999},
			"1123456789", true},
		{"include unixtime not within filter period #1",
			&Parameter{filterFlag: true, filterFromMS: 1234567890000, filterToMS: MAX_UNIXTIME},
			"1234567889999", false},
		{"include unixtime not within filter period #2",
			&Parameter{filterFlag: true, filterFromMS: 1234567890000, filterToMS: 1234567890000},
			"1234567890001", false},
		{"include unixtime not within filter period #3",
			&Parameter{filterFlag: true, filterFromMS: MIN_UNIXTIME, filterToMS: 1234567890000},
			"1234567890001", false},
		{"summary",
			&Parameter{summaryFlag: true},
			"1234567890001", false},
		{"include unixtime with invert flag within filter period",
			&Parameter{filterFlag: true, filterFromMS: 1234567890000, filterToMS: 1234567890000, invertFlag: true},
			"1234567890000", false},
		{"include unixtime with invert flag not within filter period",
			&Parameter{filterFlag: true, filterFromMS: 1234567890000, filterToMS: 1234567890001, invertFlag: true},
			"1234567890002", true},
		{"include unixtime with noConvert flag",
			&Parameter{noConvFlag: true},
			"1234567890001", true},
	}
	for _, tt := range tests {
		if actual := replaceUnixtimeToDatetime(tt.input, &Summary{}, tt.param); actual.ShouldOutput != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		} else {
			//t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		}
	}
}

func TestReplaceUnixtimeToDatetimeNoConvTest(t *testing.T) {
	tests := []struct {
		name   string
		param  *Parameter
		input  string
		expect string
	}{
		{"include unixtime with noConvert flag",
			&Parameter{filterFlag: true, filterFromMS: 1234567890123, filterToMS: MAX_UNIXTIME, noConvFlag: true},
			"1234567890123 2345678901234", "1234567890123 2345678901234"},
		{"include unixtime without noConvert flag",
			&Parameter{filterFlag: true, filterFromMS: 1234567890123, filterToMS: MAX_UNIXTIME, noConvFlag: false},
			" 1234567890123 2345678901234 ", " 2009-02-13T23:31:30.123Z 2044-05-01T01:28:21.234Z "},
	}
	for _, tt := range tests {
		if actual := replaceUnixtimeToDatetime(tt.input, &Summary{}, tt.param); actual.Text != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		} else {
			//t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		}
	}
}

func TestReplaceUnixtimeToDatetimeWithJSON(t *testing.T) {
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
		if actual := replaceUnixtimeToDatetime(tt.input, &Summary{}, &Parameter{}); actual.Text != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		} else {
			//t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		}
	}
}

func TestReplaceUnixtimeToDatetime(t *testing.T) {
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
		{"space before unixtime", " 1720999990", " 2024-07-14T23:33:10Z"},
		{"space after unixtime", "1720999990 ", "2024-07-14T23:33:10Z "},
		{"same 10 digits", "1720999999 1720999999", "2024-07-14T23:33:19Z 2024-07-14T23:33:19Z"},
		{"same 13 digits", "1722543769134 1722543769134", "2024-08-01T20:22:49.134Z 2024-08-01T20:22:49.134Z"},
		{"dot + 10 digits", ".1234567890 ", ".1234567890 "},
		{"underscore + 10 digits", "_1234567890 ", "_1234567890 "},
		{"hyphen + 10 digits", "-1234567890 ", "-1234567890 "},
		{"a-z + 10 digits", "abc1234567890 ", "abc1234567890 "},
		{"10 digits + A-Z", "1234567890ABC", "1234567890ABC"},
		{"A-Z + 13 digits + a-z", "ABC1234567890123abc", "ABC1234567890123abc"},
		{"comma separated 10 digits", "a,1720999999,1722543769", "a,2024-07-14T23:33:19Z,2024-08-01T20:22:49Z"},
		{"space separated 13 digits", "1720999999000 1722543769134 ", "2024-07-14T23:33:19.000Z 2024-08-01T20:22:49.134Z "},
		{"10 digits and 13 digits", "1720999999 1722543769876", "2024-07-14T23:33:19Z 2024-08-01T20:22:49.876Z"},
		{"13 digits and 10 digits", "1720999999111 1722543769", "2024-07-14T23:33:19.111Z 2024-08-01T20:22:49Z"},
		{"multi bytes #1", "あ1722543769･1722543769876／", "あ2024-08-01T20:22:49Z･2024-08-01T20:22:49.876Z／"},
		{"multi bytes #2", "１７２２５４３７６９", "１７２２５４３７６９"},
	}
	for _, tt := range tests {
		if actual := replaceUnixtimeToDatetime(tt.input, &Summary{}, &Parameter{}); actual.Text != tt.expect {
			t.Errorf("[ NG ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		} else {
			//t.Logf("[ OK ] => %s\n   input: %v\n  expect: %v\n  actual: %v", tt.name, tt.input, tt.expect, actual)
		}
	}
}
