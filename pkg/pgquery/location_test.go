package pgquery_test

import (
	"testing"

	"github.com/aprimetechnology/derisk-sql/pkg/pgquery"
)

// we use an actual dummy migration to ensure that the TextLocation
// abilities match what we would expect from a code editor's (LineNum, CharPos)
const SampleText = `-- migrate:up
CREATE TABLE dummy_table_with_index (dummy_column varchar(20));
INSERT INTO dummy_table_with_index (dummy_column) VALUES ('Hello world!');
CREATE INDEX idx_dummy_column ON dummy_table(dummy_column);

-- migrate:down
DROP INDEX CONCURRENTLY idx_dummy_column;
DELETE FROM dummy_table_with_index;
DROP TABLE dummy_table_with_index; -- dummy comment
`

func TestSkipSpace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      string
		byteOffset int
		expected   int
	}{
		{
			"noop",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			1,
			1,
		},
		{
			"new line is not skipped",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			3,
			3,
		},
		{
			"new lines are not skipped",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			10,
			10,
		},
		{
			"skip one space",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			5,
			6,
		},
		{
			"skip multiple spaces",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			14,
			17,
		},
		{
			"skip one tab",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			8,
			9,
		},
		{
			"skip multiple tabs",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			19,
			22,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			t.Log(test.name)
			if got := pgquery.SkipSpace(test.input, test.byteOffset); got != test.expected {
				t.Fatalf("SkipSpace(%d) returned %v; expected %v", test.byteOffset, got, test.expected)
			}
		})
	}
}

func TestSkipUntilNewLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      string
		byteOffset int
		expected   int
	}{
		{
			"noop",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			0,
			0,
		},
		{
			"skip some letters",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			1,
			3,
		},
		{
			"skip letters and whitespace",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			4,
			11,
		},
		{
			"noop in multiple newlines",
			"\nab\nc de\tfg\n\nh   ji\t\t\tk  \t  \t\t  lmno",
			12,
			12,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			t.Log(test.name)
			if got := pgquery.SkipUntilNewLine(test.input, test.byteOffset); got != test.expected {
				t.Fatalf("SkipUntilNewLine(%d) returned %v; expected %v", test.byteOffset, got, test.expected)
			}
		})
	}
}

func TestSkipComments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      string
		byteOffset int
		expected   int
	}{
		{
			"skip start of file comment",
			SampleText,
			0,
			13,
		},
		{
			"noop since not comment",
			SampleText,
			13,
			13,
		},
		{
			"noop since not comment - middle of line",
			SampleText,
			20,
			20,
		},
		{
			"do NOT skip comment if positioned at empty line",
			SampleText,
			212,
			212,
		},
		{
			"do skip comment if positioned at new line before comment",
			SampleText,
			213,
			229,
		},
		{
			"skip comment in middle of file, at start of line",
			SampleText,
			214,
			229,
		},
		{
			"do not skip comment if in middle of --",
			SampleText,
			215,
			215,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			t.Log(test.name)
			if got := pgquery.SkipComments(test.input, test.byteOffset); got != test.expected {
				t.Fatalf("SkipComments(%d) returned %v; expected %v", test.byteOffset, got, test.expected)
			}
		})
	}
}

func TestSkipWhitespaceAndComments(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		byteOffset int
		expected   int
	}{
		{
			"skip start of file comment",
			SampleText,
			0,
			14,
		},
		{
			"noop since not comment",
			SampleText,
			14,
			14,
		},
		{
			"noop since not comment - middle of line",
			SampleText,
			21,
			21,
		},
		{
			"do skip comment if positioned at empty line",
			SampleText,
			212,
			230,
		},
		{
			"empty line x 2",
			SampleText,
			212,
			230,
		},

		{
			"empty line x 1",
			SampleText,
			213,
			230,
		},
		{
			"skip comment in middle of file, at start of line",
			SampleText,
			214,
			230,
		},
		{
			"do not skip comment if in middle of --",
			SampleText,
			215,
			215,
		},
		{
			"skip from whitespace near end of file",
			SampleText,
			342,
			360,
		},
		{
			"skip from comment near end of file",
			SampleText,
			343,
			360,
		},
		{
			"do NOT skip partway through comment",
			SampleText,
			344,
			344,
		},
		{
			"handle skipping to end of file - newline",
			"hello\n \t --   \t ",
			5,
			15,
		},
		{
			"handle skipping to end of file - whitespace",
			"hello\n \t --   \t ",
			12,
			15,
		},
		{
			"handle skipping to end of file - comment",
			"hello\n \t --   \t ",
			9,
			15,
		},
		{
			"handle skipping to end of file - comment and whitespace",
			"hello\n \t --   \t ",
			6,
			15,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name)
			if got := pgquery.SkipWhitespaceAndComments(test.input, test.byteOffset); got != test.expected {
				t.Fatalf("SkipWhitespaceAndComments(%d) returned %v; expected %v", test.byteOffset, got, test.expected)
			}
		})
	}
}

func TestGetTextLocation(t *testing.T) {
	// marks TLog as capable of running in parallel with other tests
	t.Parallel()
	tests := []struct {
		byteOffset               int
		expectedLineNumber       int
		expectedLineCharPosition int
		expectedChar             string
		name                     string
	}{
		{0, 1, 1, "-", "line 1: start"},
		{7, 1, 8, "a", "line 1: in the middle"},
		{12, 1, 13, "p", "line 1: end"},
		{13, 2, 0, "\n", "line 1: EOL"},

		{14, 2, 1, "C", "line 2: start"},
		{45, 2, 32, "n", "line 2: in the middle"},
		{76, 2, 63, ";", "line 2: end"},
		{77, 3, 0, "\n", "line 2: EOL"},

		{78, 3, 1, "I", "line 3: start"},
		{113, 3, 36, "(", "line 3: in the middle"},
		{151, 3, 74, ";", "line 3: end"},
		{152, 4, 0, "\n", "line 3: EOL"},

		{153, 4, 1, "C", "line 4: start"},
		{191, 4, 39, "_", "line 4: in the middle"},
		{211, 4, 59, ";", "line 4: end"},
		{212, 5, 0, "\n", "line 4: EOL"},

		{213, 6, 0, "\n", "line 5: start + end + EOL"},

		{214, 6, 1, "-", "line 6: start"},
		{224, 6, 11, ":", "line 6: in the middle"},
		{228, 6, 15, "n", "line 6: end"},
		{229, 7, 0, "\n", "line 6: EOL"},

		{230, 7, 1, "D", "line 7: start"},
		{252, 7, 23, "Y", "line 7: in the middle"},
		{270, 7, 41, ";", "line 7: end"},
		{271, 8, 0, "\n", "line 7: EOL"},

		{272, 8, 1, "D", "line 8: start"},
		{281, 8, 10, "O", "line 8: in the middle"},
		{306, 8, 35, ";", "line 8: end"},
		{307, 9, 0, "\n", "line 8: EOL"},

		{308, 9, 1, "D", "line 9: start"},
		{325, 9, 18, "t", "line 9: in the middle"},
		{341, 9, 34, ";", "line 9: end"},
		{359, 10, 0, "\n", "line 9: EOL and EOF final byte"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// mark each test case as capable of running in parallel with each other
			t.Parallel()
			t.Log(test.name)

			input := SampleText
			if string(input[test.byteOffset]) != test.expectedChar {
				t.Fatalf("byteOffset %d is character `%s` expected `%s`", test.byteOffset, string(input[test.byteOffset]), test.expectedChar)
			}

			expected := pgquery.TextLocation{
				LineNumber:       test.expectedLineNumber,
				LineCharPosition: test.expectedLineCharPosition,
			}
			if got := pgquery.GetTextLocation(input, test.byteOffset); got != expected {
				t.Fatalf("GetTextLocation(%d) returned %v; expected %v", test.byteOffset, got, expected)
			}
		})
	}
}
