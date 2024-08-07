package pgquery

const CommentPrefix = "--"

type Line struct {
	Number         int
	TextByteOffset int
}

type TextLocation struct {
	LineNumber       int
	LineCharPosition int
}

func IsValidOffset(input string, offset int) bool {
	return offset >= 0 && offset < len(input)
}

func SkipSpace(input string, offset int) int {
	for ; offset < len(input); offset += 1 {
		if input[offset] != ' ' && input[offset] != '\t' {
			return offset
		}
	}
	return offset
}

func IsNewline(input string, position int) bool {
	return IsValidOffset(input, position) && input[position] == '\n'
}

func SkipUntilNewLine(input string, offset int) int {
	for ; offset < len(input); offset += 1 {
		if IsNewline(input, offset) {
			return offset
		}
	}
	return offset
}

func IsCommentPrefix(input string, position int) bool {
	return IsValidOffset(input, position+len(CommentPrefix)) && input[position:position+len(CommentPrefix)] == CommentPrefix
}

func SkipComments(input string, offset int) int {
	// if invalid offset, or there's not enough bytes for a comment to exist, just return
	if !IsValidOffset(input, offset) || !IsValidOffset(input, offset+len(CommentPrefix)) {
		return offset
	}

	// two cases for detecting a comment:
	// 1) at a newline character, a comment immediately follows
	// (the pg_query library will put statement.StmtLocation at such a position)
	if IsNewline(input, offset) && IsCommentPrefix(input, offset+1) {
		offset += 1
	}

	// 2) presently at a comment
	if IsCommentPrefix(input, offset) {
		offset = SkipUntilNewLine(input, offset)
	}

	return offset
}

func SkipWhitespaceAndComments(input string, byteOffset int) int {
	position := byteOffset
	for position < len(input) {
		startingPosition := position

		// first eliminate any non-newline whitespace
		position = SkipSpace(input, position)
		if position >= len(input) {
			// if skipped through the end of the file, return last byte of the file
			return len(input) - 1
		}

		// skip any comments at this new position
		position = SkipComments(input, position)
		if position >= len(input) {
			// if skipped through the end of the file, return last byte of the file
			return len(input) - 1
		}

		// skip exactly one new line, so can loop against the next line
		if input[position] == '\n' {
			position += 1
		}

		// no net change, it's time to stop looping
		if position == startingPosition {
			break
		}
	}
	return position
}

func GetTextLocation(input string, byteOffset int) TextLocation {
	if !IsValidOffset(input, byteOffset) {
		return TextLocation{LineNumber: -1, LineCharPosition: -1}
	}

	goal := Line{Number: -1, TextByteOffset: -1}
	current := Line{Number: 1, TextByteOffset: -1}
	previous := Line{Number: 0, TextByteOffset: -1}

	for offset := 0; offset < len(input); offset += 1 {
		// skip until a newline character
		offset = SkipUntilNewLine(input, offset)

		// at every new line character, capture the byte position this occurred
		previous = current
		current.Number += 1
		current.TextByteOffset = offset

		// if the current line is PAST the byteOffset,
		// then the previous line is where this byteOffset can be found
		if current.TextByteOffset > byteOffset {
			// if byteOffset falls exactly on a '\n', we want to consider that part of the next line
			// eg, for 'hello\nworld' with a byteOffset of 5
			//      we want the text location to be (Line 2, Character 0)
			//      rather than (Line 1, Character 6) or (Line 2, Character -1) or (Line 2, Character 1)
			goal = previous
			break
		}
	}
	// if no line went PAST the byteOffset, this means the byteOffset is on the very last line
	if goal.Number == -1 {
		goal = current
	}

	return TextLocation{
		LineNumber:       goal.Number,
		LineCharPosition: byteOffset - goal.TextByteOffset,
	}
}
