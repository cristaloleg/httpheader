package httpheader

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func ExampleAllow() {
	header := http.Header{"Allow": []string{"GET, HEAD, OPTIONS"}}
	fmt.Printf("%q", Allow(header))
	// Output: ["GET" "HEAD" "OPTIONS"]
}

func ExampleSetAllow() {
	header := http.Header{}
	SetAllow(header, []string{"GET", "HEAD", "OPTIONS"})
	fmt.Printf("%q", header)
	// Output: map["Allow":["GET, HEAD, OPTIONS"]]
}

func TestAllow(t *testing.T) {
	tests := []struct {
		header http.Header
		result []string
	}{
		// Valid headers.
		{
			http.Header{},
			nil,
		},
		{
			http.Header{"Foo": []string{"bar"}},
			nil,
		},
		{
			http.Header{"Allow": []string{""}},
			[]string{},
		},
		{
			http.Header{"Allow": []string{
				"",
				"",
			}},
			[]string{},
		},
		{
			http.Header{"Allow": []string{"GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"GET,"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{",GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"  ,\t ,, GET, , "}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"GET,HEAD"}},
			[]string{"GET", "HEAD"},
		},
		{
			http.Header{"Allow": []string{"GET, HEAD"}},
			[]string{"GET", "HEAD"},
		},
		{
			http.Header{"Allow": []string{
				"GET",
				"HEAD",
				"OPTIONS",
			}},
			[]string{"GET", "HEAD", "OPTIONS"},
		},
		{
			http.Header{"Allow": []string{
				"GET\t,\t  HEAD\t",
				"\tOPTIONS",
			}},
			[]string{"GET", "HEAD", "OPTIONS"},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Allow": []string{";;;"}},
			[]string{},
		},
		{
			http.Header{"Allow": []string{";;;,GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"GET;;;whatever, HEAD"}},
			[]string{"GET", "HEAD"},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Allow(test.header))
		})
	}
}

func ExampleVary() {
	header := http.Header{"Vary": []string{"cookie, accept-encoding"}}
	fmt.Printf("%q", Vary(header))
	// Output: ["Cookie" "Accept-Encoding"]
}

func ExampleVia() {
	header := http.Header{"Via": []string{
		"1.1 foo.example.com:8080 (corporate)",
		"2 bar.example.net",
	}}
	fmt.Printf("%q", Via(header))
	// Output: [{"HTTP/1.1" "foo.example.com:8080" "corporate"} {"HTTP/2" "bar.example.net" ""}]
}

func ExampleAddVia() {
	header := http.Header{"Via": []string{"1.0 foo"}}
	AddVia(header, ViaEntry{
		ReceivedProto: "HTTP/1.1",
		ReceivedBy:    "bar",
	})
	fmt.Printf("%q", header)
	// Output: map["Via":["1.0 foo" "1.1 bar"]]
}

func TestVia(t *testing.T) {
	tests := []struct {
		header http.Header
		result []ViaEntry
	}{
		// Valid headers.
		{
			http.Header{"Via": []string{"1.0 foo"}},
			[]ViaEntry{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": []string{"1.0 \tfoo"}},
			[]ViaEntry{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": []string{"1.0 foo  "}},
			[]ViaEntry{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": []string{"1.0 foo  ,"}},
			[]ViaEntry{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": []string{"1.0 foo\t (comment)"}},
			[]ViaEntry{{"HTTP/1.0", "foo", "comment"}},
		},
		{
			http.Header{"Via": []string{
				"1.0 foo,1.0   bar\t, \t 1.0 baz,,",
				"1.1 qux",
			}},
			[]ViaEntry{
				{"HTTP/1.0", "foo", ""},
				{"HTTP/1.0", "bar", ""},
				{"HTTP/1.0", "baz", ""},
				{"HTTP/1.1", "qux", ""},
			},
		},
		{
			http.Header{"Via": []string{
				"HTTP/2 foo",
				"FSTR/3 bar (some new protocol)",
			}},
			[]ViaEntry{
				{"HTTP/2", "foo", ""},
				{"FSTR/3", "bar", "some new protocol"},
			},
		},
		{
			http.Header{"Via": []string{"1.1 foo (comment (with) nesting)"}},
			[]ViaEntry{{"HTTP/1.1", "foo", "comment (with) nesting"}},
		},
		{
			http.Header{"Via": []string{"1.1 foo (comment (with nesting))"}},
			[]ViaEntry{{"HTTP/1.1", "foo", "comment (with nesting)"}},
		},
		{
			http.Header{"Via": []string{`1.1 foo (comment with \) quoting)`}},
			[]ViaEntry{{"HTTP/1.1", "foo", "comment with ) quoting"}},
		},
		{
			http.Header{"Via": []string{
				`1.1 foo (comment (with \) quoting) and nesting)`,
			}},
			[]ViaEntry{
				{"HTTP/1.1", "foo", "comment (with ) quoting) and nesting"},
			},
		},
		{
			http.Header{"Via": []string{`1.1 foo (\strange quoting)`}},
			[]ViaEntry{{"HTTP/1.1", "foo", "strange quoting"}},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Via": []string{"1.0"}},
			[]ViaEntry{{"HTTP/1.0", "", ""}},
		},
		{
			http.Header{"Via": []string{"1.0, 1.1 foo, 1.2, 1.3 bar"}},
			[]ViaEntry{
				{"HTTP/1.0", "", ""},
				{"HTTP/1.1", "foo", ""},
				{"HTTP/1.2", "", ""},
				{"HTTP/1.3", "bar", ""},
			},
		},
		{
			http.Header{"Via": []string{
				"1.1 foo (unterminated",
				"1.1 bar",
			}},
			[]ViaEntry{
				{"HTTP/1.1", "foo", "unterminated"},
				{"HTTP/1.1", "bar", ""},
			},
		},
		{
			http.Header{"Via": []string{"1.1 foo (unterminated (with nesting)",}},
			[]ViaEntry{
				{"HTTP/1.1", "foo", "unterminated (with nesting)"},
			},
		},
		{
			http.Header{"Via": []string{
				`1.1 foo (unterminated with \quoting (and nesting`,
				"1.1 bar",
			}},
			[]ViaEntry{
				{"HTTP/1.1", "foo", "unterminated with quoting (and nesting"},
				{"HTTP/1.1", "bar", ""},
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Via(test.header))
		})
	}
}

func ExampleWarning() {
	header := http.Header{"Warning": []string{`299 gw1 "something is wrong"`}}
	fmt.Printf("%+v", Warning(header))
	// Output: [{Code:299 Agent:gw1 Text:something is wrong Date:0001-01-01 00:00:00 +0000 UTC}]
}

func ExampleAddWarning() {
	header := http.Header{}
	AddWarning(header, WarningEntry{
		Code:  299,
		Agent: "-",
		Text:  "something is fishy",
	})
	fmt.Printf("%q", header)
	// Output: map["Warning":["299 - \"something is fishy\""]]
}

func TestWarning(t *testing.T) {
	tests := []struct {
		header http.Header
		result []WarningEntry
	}{
		// Valid headers.
		{
			http.Header{"Warning": []string{`299 - "good"`}},
			[]WarningEntry{{299, "-", "good", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`299 example.net:80 "good"`}},
			[]WarningEntry{{299, "example.net:80", "good", time.Time{}}},
		},
		{
			// See RFC 6874.
			http.Header{"Warning": []string{`299 [fe80::a%25en1]:80 "good"`}},
			[]WarningEntry{{299, "[fe80::a%25en1]:80", "good", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`199 - "good", 299 - "better"`}},
			[]WarningEntry{
				{199, "-", "good", time.Time{}},
				{299, "-", "better", time.Time{}},
			},
		},
		{
			http.Header{"Warning": []string{`199 - "good" , 299 - "better"`}},
			[]WarningEntry{
				{199, "-", "good", time.Time{}},
				{299, "-", "better", time.Time{}},
			},
		},
		{
			http.Header{"Warning": []string{
				`299 - "good" "Sat, 06 Jul 2019 05:45:48 GMT"`,
			}},
			[]WarningEntry{{
				299, "-", "good",
				time.Date(2019, time.July, 6, 5, 45, 48, 0, time.UTC),
			}},
		},
		{
			http.Header{"Warning": []string{
				`199 - "good" "Sat, 06 Jul 2019 05:45:48 GMT",299 - "better"`,
			}},
			[]WarningEntry{
				{
					199, "-", "good",
					time.Date(2019, time.July, 6, 5, 45, 48, 0, time.UTC),
				},
				{
					299, "-", "better",
					time.Time{},
				},
			},
		},
		{
			http.Header{"Warning": []string{
				`199 - "good" "Sat, 06 Jul 2019 05:45:48 GMT"\t,299 - "better"`,
			}},
			[]WarningEntry{
				{
					199, "-", "good",
					time.Date(2019, time.July, 6, 5, 45, 48, 0, time.UTC),
				},
				{
					299, "-", "better",
					time.Time{},
				},
			},
		},
		{
			http.Header{"Warning": []string{`299 - "with \"escaped\" quotes"`}},
			[]WarningEntry{{299, "-", `with "escaped" quotes`, time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`299 - "\"escaped\" quotes"`}},
			[]WarningEntry{{299, "-", `"escaped" quotes`, time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`299 - "with \"escaped\""`}},
			[]WarningEntry{{299, "-", `with "escaped"`, time.Time{}}},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Warning": []string{"299"}},
			[]WarningEntry{{299, "", "", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{"299 -"}},
			[]WarningEntry{{299, "-", "", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{"299 - unquoted"}},
			[]WarningEntry{{299, "-", "", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`299  - "two spaces"`}},
			[]WarningEntry{{299, "-", "two spaces", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`?????,299 - "good"`}},
			[]WarningEntry{
				{0, "", "", time.Time{}},
				{299, "-", "good", time.Time{}},
			},
		},
		{
			http.Header{"Warning": []string{`299  bad, 299 - "good"`}},
			[]WarningEntry{
				{299, "bad", "", time.Time{}},
				{299, "-", "good", time.Time{}},
			},
		},
		{
			http.Header{"Warning": []string{`299 - "good" "bad date"`}},
			[]WarningEntry{{299, "-", "good", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`299 - "unterminated`}},
			[]WarningEntry{{299, "-", "unterminated", time.Time{}}},
		},
		{
			http.Header{"Warning": []string{`299 - "unterminated\"`}},
			[]WarningEntry{{299, "-", `unterminated"`, time.Time{}}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Warning(test.header))
		})
	}
}

func checkParse(t *testing.T, header http.Header, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("header: %#v\nexpected: %#v\nactual:   %#v",
			header, expected, actual)
	}
}
