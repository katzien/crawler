package crawler

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestGetEdges(t *testing.T) {

	result := getEdges(getTestSitemap())

	expected := [][]string{
		{"https://test.com", "https://test.com"},
		{"https://test.com", "https://test.com/foo"},
		{"https://test.com", "https://test.com/bar"},
		{"https://test.com/foo", "https://test.com/bar"},
		{"https://test.com/foo", "https://test.com/baz"},
		{"https://test.com/bar", "https://test.com/foo"},
	}

	if len(result) != len(expected) {
		t.Errorf("getEdges(): expected to have %d edges, got %d edges", len(expected), len(result))
		t.Errorf("expected: %v", expected)
		t.Errorf("actual: %v", result)
		t.FailNow()
	}

	allEdgesFound := true
	for _, ii := range expected {
		found := false
		for _, jj := range result {
			if ii[0] == jj[0] && ii[1] == jj[1] || ii[0] == jj[1] && ii[1] == jj[0] {
				found = true
			}
		}
		if !found {
			t.Errorf("getEdges(): expected %v\nto contain %v", result, ii)
			allEdgesFound = false
		}
	}

	if !allEdgesFound {
		t.Fail()
	}
}

func TestText(t *testing.T) {
	actual, err := Text(getTestSitemap())

	if err != nil {
		t.Errorf("Text(): expected no errors returned, got %s\n", err.Error())
		t.FailNow()
	}

	expectedLines := []string{
		"pages:",
		"https://test.com",
		"https://test.com/foo",
		"https://test.com/bar",
		"https://test.com/baz",
		"links:",
		"https://test.com -> https://test.com",
		"https://test.com -> https://test.com/foo",
		"https://test.com -> https://test.com/bar",
		"https://test.com/foo -> https://test.com/bar",
		"https://test.com/foo -> https://test.com/baz",
		"https://test.com/bar -> https://test.com/foo",
	}

	// Since we're using the range operator in Text() to loop over the sitemap and edges slices,
	// the order in which the edges and nodes get written to the file may be different every time.
	// Hence we're just checking here that the output contains all the expected lines, in any order.
	for _, ll := range expectedLines {
		if !strings.Contains(actual, ll) {
			t.Errorf("Text(): expected %s, got %s\n", ll, actual)
		}
	}
}

func TestWriteDot(t *testing.T) {

	var b bytes.Buffer

	err := writeDot(&b, getTestSitemap())
	if err != nil {
		t.Errorf("WriteDot(): expected no error returned, got %s", err.Error())
	}

	actual := b.String()

	expectedLines := []string{
		"digraph G",
		`"https://test.com"->"https://test.com";`,
		`"https://test.com"->"https://test.com/foo";`,
		`"https://test.com"->"https://test.com/bar";`,
		`"https://test.com/foo"->"https://test.com/bar"`,
		`"https://test.com/foo"->"https://test.com/baz"`,
		`"https://test.com/bar"->"https://test.com/foo";`,
		`"https://test.com";`,
		`"https://test.com/foo";`,
		`"https://test.com/bar";`,
		`"https://test.com/baz";`,
		`}`,
	}

	// Since we're using the range operator in WriteDot() to loop over the sitemap and edges slices,
	// the order in which the edges and nodes get written to the file may be different every time.
	// Hence we're just checking here that the output contains all the expected lines, in any order.
	for _, ll := range expectedLines {
		if !strings.Contains(actual, ll) {
			t.Errorf("WriteDot(): expected output %s to contain line %s", actual, ll)
		}
	}
}

type errWriter struct {
	err error
}

func (w errWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestWriteDotReturnsErrorsFromWriter(t *testing.T) {
	expected := errors.New("io.Writer error")

	err := writeDot(errWriter{err: expected}, getTestSitemap())
	if err == nil {
		t.Errorf("WriteDot(errWriter): expected error %s to be returned, got no error", expected.Error())
		t.FailNow()
	}

	if err.Error() != expected.Error() {
		t.Errorf("WriteDot(errWriter): expected to get error %s, got %s", expected.Error(), err.Error())
	}
}

func getTestSitemap() Sitemap {
	s := Sitemap{}

	l1 := Links{"https://test.com", "https://test.com/foo", "https://test.com/bar"}
	l2 := Links{"https://test.com/bar", "https://test.com/baz"}
	l3 := Links{"https://test.com/foo"}

	s[CanonicalURL("https://test.com")] = l1
	s[CanonicalURL("https://test.com/foo")] = l2
	s[CanonicalURL("https://test.com/bar")] = l3
	s[CanonicalURL("https://test.com/baz")] = Links{}

	return s
}
