//go:build !launchdarkly_easyjson
// +build !launchdarkly_easyjson

package jreader

// This file defines the default implementation of the low-level JSON tokenizer. If the launchdarkly_easyjson
// build tag is enabled, we use the easyjson adapter in token_reader_easyjson.go instead. These have the same
// methods so the Reader code does not need to know which implementation we're using; however, we don't
// actually define an interface for these, because calling the methods through an interface would limit
// performance.

import (
	"bytes"
	"io"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	tokenNull  = []byte("null")  //nolint:gochecknoglobals
	tokenTrue  = []byte("true")  //nolint:gochecknoglobals
	tokenFalse = []byte("false") //nolint:gochecknoglobals
)

type token struct {
	kind        tokenKind
	boolValue   bool
	numberValue float64
	stringValue []byte
	delimiter   byte
}

type tokenKind int

const (
	nullToken      tokenKind = iota
	boolToken      tokenKind = iota
	numberToken    tokenKind = iota
	stringToken    tokenKind = iota
	delimiterToken tokenKind = iota
)

func (t token) valueKind() ValueKind {
	if t.kind == delimiterToken {
		if t.delimiter == '[' {
			return ArrayValue
		}
		if t.delimiter == '{' {
			return ObjectValue
		}
	}
	return valueKindFromTokenKind(t.kind)
}

func (t token) description() string {
	if t.kind == delimiterToken && t.delimiter != '[' && t.delimiter != '{' {
		return "'" + string(t.delimiter) + "'"
	}
	return t.valueKind().String()
}

type tokenReader struct {
	data        []byte
	pos         int
	len         int
	hasUnread   bool
	unreadToken token
	lastPos     int
}

func newTokenReader(data []byte) tokenReader {
	tr := tokenReader{
		data: data,
		pos:  0,
		len:  len(data),
	}
	return tr
}

// EOF returns true if we are at the end of the input (not counting whitespace).
func (r *tokenReader) EOF() bool {
	if r.hasUnread {
		return false
	}
	_, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return true
	}
	r.unreadByte()
	return false
}

// LastPos returns the byte offset within the input where we most recently started parsing a token.
func (r *tokenReader) LastPos() int {
	return r.lastPos
}

func (r *tokenReader) getPos() int {
	if r.hasUnread {
		return r.lastPos
	}
	return r.pos
}

// Null returns (true, nil) if the next token is a null (consuming the token); (false, nil) if the next
// token is not a null (not consuming the token); or (false, error) if the next token is not a valid
// JSON value.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Null() (bool, error) {
	t, err := r.next()
	if err != nil {
		return false, err
	}
	if t.kind == nullToken {
		return true, nil
	}
	r.putBack(t)
	if t.kind == delimiterToken && t.delimiter != '[' && t.delimiter != '{' {
		return false, SyntaxError{Message: errMsgUnexpectedChar, Value: string(t.delimiter), Offset: r.getPos()}
	}
	return false, nil
}

// Bool requires that the next token is a JSON boolean, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON boolean.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Bool() (bool, error) {
	t, err := r.consumeScalar(boolToken)
	return t.boolValue, err
}

// Bool requires that the next token is a JSON number, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON number.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Number() (float64, error) {
	t, err := r.consumeScalar(numberToken)
	return t.numberValue, err
}

// String requires that the next token is a JSON string, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON string.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) String() (string, error) {
	stringValue, err := r.StringAsBytes()
	return string(stringValue), err
}

// StringAsBytes requires that the next token is a JSON string, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON string.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) StringAsBytes() ([]byte, error) {
	t, err := r.consumeScalar(stringToken)
	return t.stringValue, err
}

// PropertyName requires that the next token is a JSON string and the token after that is a colon,
// returning the string as a byte slice if successful, or an error otherwise.
//
// Returning the string as a byte slice avoids the overhead of allocating a string, since normally
// the names of properties will not be retained as strings but are only compared to constants while
// parsing an object.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) PropertyName() ([]byte, error) {
	t, err := r.consumeScalar(stringToken)
	if err != nil {
		return nil, err
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return nil, io.EOF
	}
	if b != ':' {
		r.unreadByte()
		return nil, r.syntaxErrorOnNextToken(errMsgExpectedColon)
	}
	return t.stringValue, nil
}

// Delimiter checks whether the next token is the specified ASCII delimiter character. If so, it
// returns (true, nil) and consumes the token. If it is a delimiter, but not the same one, it
// returns (false, nil) and does not consume the token. For anything else, it returns an error.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Delimiter(delimiter byte) (bool, error) {
	if r.hasUnread {
		if r.unreadToken.kind == delimiterToken && r.unreadToken.delimiter == delimiter {
			r.hasUnread = false
			return true, nil
		}
		return false, nil
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return false, nil
	}
	if b == delimiter {
		return true, nil
	}
	r.unreadByte() // we'll back up and try to parse a token, to see if it's valid JSON or not
	token, err := r.next()
	if err != nil {
		return false, err // it was malformed JSON
	}
	r.putBack(token) // it was valid JSON, we just haven't hit that delimiter
	return false, nil
}

// EndDelimiterOrComma checks whether the next token is the specified ASCII delimiter character
// or a comma. If it is the specified delimiter, it returns (true, nil) and consumes the token.
// If it is a comma, it returns (false, nil) and consumes the token. For anything else, it
// returns an error. The delimiter parameter will always be either '}' or ']'.
func (r *tokenReader) EndDelimiterOrComma(delimiter byte) (bool, error) {
	if r.hasUnread {
		if r.unreadToken.kind == delimiterToken &&
			(r.unreadToken.delimiter == delimiter || r.unreadToken.delimiter == ',') {
			r.hasUnread = false
			return r.unreadToken.delimiter == delimiter, nil
		}
		return false, SyntaxError{Message: badArrayOrObjectItemMessage(delimiter == '}'),
			Value: r.unreadToken.description(), Offset: r.lastPos}
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return false, io.EOF
	}
	if b == delimiter || b == ',' {
		return b == delimiter, nil
	}
	r.unreadByte()
	t, err := r.next()
	if err != nil {
		return false, err
	}
	return false, SyntaxError{Message: badArrayOrObjectItemMessage(delimiter == '}'),
		Value: t.description(), Offset: r.lastPos}
}

func badArrayOrObjectItemMessage(isObject bool) string {
	if isObject {
		return errMsgBadObjectItem
	}
	return errMsgBadArrayItem
}

// Any checks whether the next token is either a valid JSON scalar value or the opening delimiter of
// an array or object value. If so, it returns (AnyValue, nil) and consumes the token; if not, it
// returns an error. Unlike Reader.Any(), for array and object values it does not create an
// ArrayState or ObjectState.
func (r *tokenReader) Any() (AnyValue, error) {
	return r.any(false)
}

func (r *tokenReader) any(ignoreString bool) (AnyValue, error) {
	t, err := r.next()
	if err != nil {
		return AnyValue{}, err
	}
	switch t.kind {
	case boolToken:
		return AnyValue{Kind: BoolValue, Bool: t.boolValue}, nil
	case numberToken:
		return AnyValue{Kind: NumberValue, Number: t.numberValue}, nil
	case stringToken:
		var s string
		if !ignoreString {
			s = string(t.stringValue)
		}
		return AnyValue{Kind: StringValue, String: s}, nil
	case delimiterToken:
		if t.delimiter == '[' {
			return AnyValue{Kind: ArrayValue}, nil
		}
		if t.delimiter == '{' {
			return AnyValue{Kind: ObjectValue}, nil
		}
		return AnyValue{},
			SyntaxError{Message: errMsgUnexpectedChar, Value: string(t.delimiter), Offset: r.lastPos}
	default:
		return AnyValue{Kind: NullValue}, nil
	}
}

// Attempts to parse and consume the next token, ignoring whitespace. A token is either a valid JSON scalar
// value or an ASCII delimiter character. If a token was previously unread using putBack, it consumes that
// instead.
func (r *tokenReader) next() (token, error) {
	if r.hasUnread {
		r.hasUnread = false
		return r.unreadToken, nil
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return token{}, io.EOF
	}

	switch {
	// We can get away with reading bytes instead of runes because the JSON spec doesn't allow multi-byte
	// characters except within a string literal.
	case b >= 'a' && b <= 'z':
		n := r.consumeASCIILowercaseAlphabeticChars() + 1
		id := r.data[r.lastPos : r.lastPos+n]
		if b == 'f' && bytes.Equal(id, tokenFalse) {
			return token{kind: boolToken, boolValue: false}, nil
		}
		if b == 't' && bytes.Equal(id, tokenTrue) {
			return token{kind: boolToken, boolValue: true}, nil
		}
		if b == 'n' && bytes.Equal(id, tokenNull) {
			return token{kind: nullToken}, nil
		}
		return token{}, SyntaxError{Message: errMsgUnexpectedSymbol, Value: string(id), Offset: r.lastPos}
	case (b >= '0' && b <= '9') || b == '-':
		if n, ok := r.readNumber(b); ok {
			return token{kind: numberToken, numberValue: n}, nil
		}
		return token{}, SyntaxError{Message: errMsgInvalidNumber, Offset: r.lastPos}
	case b == '"':
		s, err := r.readString()
		if err != nil {
			return token{}, err
		}
		return token{kind: stringToken, stringValue: s}, nil
	case b == '[', b == ']', b == '{', b == '}', b == ':', b == ',':
		return token{kind: delimiterToken, delimiter: b}, nil
	}

	return token{}, SyntaxError{Message: errMsgUnexpectedChar, Value: string(b), Offset: r.lastPos}
}

func (r *tokenReader) putBack(token token) {
	r.unreadToken = token
	r.hasUnread = true
}

func (r *tokenReader) consumeScalar(kind tokenKind) (token, error) {
	t, err := r.next()
	if err != nil {
		return token{}, err
	}
	if t.kind == kind {
		return t, nil
	}
	if t.kind == delimiterToken && t.delimiter != '[' && t.delimiter != '{' {
		return token{}, SyntaxError{Message: errMsgUnexpectedChar, Value: string(t.delimiter), Offset: r.LastPos()}
	}
	return token{}, TypeError{Expected: valueKindFromTokenKind(kind),
		Actual: t.valueKind(), Offset: r.LastPos()}
}

func (r *tokenReader) readByte() (byte, bool) {
	if r.pos >= r.len {
		return 0, false
	}
	b := r.data[r.pos]
	r.pos++
	return b, true
}

func (r *tokenReader) unreadByte() {
	r.pos--
}

func (r *tokenReader) skipWhitespaceAndReadByte() (byte, bool) {
	for {
		ch, ok := r.readByte()
		if !ok {
			return 0, false
		}
		// JSON permits only these four whitespace characters between tokens. Any other
		// character (including other Unicode spaces) is the start of a token.
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			r.lastPos = r.pos - 1
			return ch, true
		}
	}
}

func (r *tokenReader) consumeASCIILowercaseAlphabeticChars() int {
	n := 0
	for {
		ch, ok := r.readByte()
		if !ok {
			break
		}
		if ch < 'a' || ch > 'z' {
			r.unreadByte()
			break
		}
		n++
	}
	return n
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// consumeDigits advances past any decimal digits in data starting at index i, returning the
// index of the first non-digit (or the end of the input).
func consumeDigits(data []byte, i, n int) int {
	for i < n && isDigit(data[i]) {
		i++
	}
	return i
}

func (r *tokenReader) readNumber(first byte) (float64, bool) {
	// The grammar is: [ - ] int [ frac ] [ exp ], where int is a single 0 or a 1-9 digit
	// followed by more digits, frac is '.' followed by at least one digit, and exp is
	// [eE][-+]? followed by at least one digit. We scan the input directly by index rather
	// than through readByte/unreadByte, then leave r.pos pointing just past the number.
	start := r.lastPos
	data, n := r.data, r.len
	i := start + 1 // the first byte has already been read
	isFloat := false

	// Optional minus sign, then the first digit of the integer part.
	if first == '-' {
		if i >= n || !isDigit(data[i]) {
			return 0, false
		}
		first = data[i]
		i++
	}

	// Integer part.
	if first == '0' {
		if i < n && isDigit(data[i]) {
			return 0, false // a leading zero cannot be followed by another digit
		}
	} else {
		i = consumeDigits(data, i, n)
	}

	// Fractional part: a decimal point must be followed by at least one digit.
	if i < n && data[i] == '.' {
		isFloat = true
		i++
		if i >= n || !isDigit(data[i]) {
			return 0, false
		}
		i = consumeDigits(data, i, n)
	}

	// Exponent part: [eE][-+]? followed by at least one digit.
	if i < n && (data[i] == 'e' || data[i] == 'E') {
		isFloat = true
		i++
		if i < n && (data[i] == '+' || data[i] == '-') {
			i++
		}
		if i >= n || !isDigit(data[i]) {
			return 0, false
		}
		i = consumeDigits(data, i, n)
	}

	r.pos = i
	chars := data[start:i]
	if !isFloat {
		if num, ok := parseIntFromBytes(chars); ok {
			return float64(num), true
		}
		// The integer literal overflows int64. Fall through to float parsing, which yields the
		// same value encoding/json produces for magnitudes within float64 range (and rejects
		// those beyond it, consistent with out-of-range float literals).
	}
	// Unfortunately, strconv.ParseFloat requires a string - there is no []byte equivalent. This means we can't
	// avoid a heap allocation here. Easyjson works around this by creating an unsafe string that points directly
	// at the existing bytes, but in our default implementation we can't use unsafe.
	num, err := strconv.ParseFloat(string(chars), 64)
	return num, err == nil
}

// beginEscapedCopy switches readString off its zero-copy fast path by copying the literal
// prefix [startPos:endPos) that has been validated so far into a fresh buffer with headroom
// for the decoded remainder.
func beginEscapedCopy(data []byte, startPos, endPos int) []byte {
	buf := make([]byte, endPos-startPos, endPos-startPos+20)
	if endPos > startPos {
		copy(buf, data[startPos:endPos])
	}
	return buf
}

func (r *tokenReader) readString() ([]byte, error) {
	startPos := r.pos // the opening quote mark has already been read
	var chars []byte
	haveEscaped := false
	var reader bytes.Reader // bytes.Reader understands multi-byte characters
	reader.Reset(r.data)
	_, _ = reader.Seek(int64(r.pos), io.SeekStart)

	for {
		ch, size, err := reader.ReadRune()
		if err != nil {
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
		if ch == '"' {
			break
		}
		if ch < 0x20 {
			// Control characters must be escaped inside a JSON string.
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
		if ch != '\\' {
			if ch == utf8.RuneError && size == 1 {
				// An invalid UTF-8 byte. encoding/json substitutes the Unicode replacement
				// character for these; do the same, which forces us off the zero-copy path.
				if !haveEscaped {
					chars = beginEscapedCopy(r.data, startPos, (r.len-reader.Len())-1)
					haveEscaped = true
				}
				chars = appendRune(chars, utf8.RuneError)
			} else if haveEscaped {
				chars = appendRune(chars, ch)
			}
			continue
		}
		if !haveEscaped {
			chars = beginEscapedCopy(r.data, startPos, (r.len-reader.Len())-1) // exclude the backslash just read
			haveEscaped = true
		}
		ch, _, err = reader.ReadRune()
		if err != nil {
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
		switch ch {
		case '"', '\\', '/':
			chars = appendRune(chars, ch)
		case 'b':
			chars = appendRune(chars, '\b')
		case 'f':
			chars = appendRune(chars, '\f')
		case 'n':
			chars = appendRune(chars, '\n')
		case 'r':
			chars = appendRune(chars, '\r')
		case 't':
			chars = appendRune(chars, '\t')
		case 'u':
			chars, err = r.readUnicodeEscape(&reader, chars)
			if err != nil {
				return nil, err
			}
		default:
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
	}
	r.pos = r.len - reader.Len()
	if haveEscaped {
		if len(chars) == 0 {
			return nil, nil
		}
		return chars, nil
	} else { //nolint:revive
		pos := r.pos - 1
		if pos <= startPos {
			return nil, nil
		}
		return r.data[startPos:pos], nil
	}
}

// readUnicodeEscape decodes a \u escape (the leading "\u" has already been consumed), combining
// a UTF-16 surrogate pair into a single code point when the escape is a surrogate followed by a
// valid pairing escape. A lone or invalid surrogate becomes the Unicode replacement character
// with the following bytes left to be parsed normally, matching encoding/json.
func (r *tokenReader) readUnicodeEscape(reader *bytes.Reader, chars []byte) ([]byte, error) {
	decoded, ok := readHexChar(reader)
	if !ok {
		return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
	}
	if !utf16.IsSurrogate(decoded) {
		return appendRune(chars, decoded), nil
	}
	mark := r.len - reader.Len()
	combined := unicode.ReplacementChar
	if b1, e1 := reader.ReadByte(); e1 == nil && b1 == '\\' {
		if b2, e2 := reader.ReadByte(); e2 == nil && b2 == 'u' {
			if low, lowOK := readHexChar(reader); lowOK {
				if pair := utf16.DecodeRune(decoded, low); pair != unicode.ReplacementChar {
					combined = pair
				}
			}
		}
	}
	if combined == unicode.ReplacementChar {
		_, _ = reader.Seek(int64(mark), io.SeekStart)
	}
	return appendRune(chars, combined), nil
}

func readHexChar(reader *bytes.Reader) (rune, bool) {
	var digits [4]byte
	for i := 0; i < 4; i++ {
		ch, err := reader.ReadByte()
		if err != nil || ((ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') && (ch < 'A' || ch > 'F')) {
			return 0, false
		}
		digits[i] = ch //nolint:gosec // G602 false positive: i is bounded by [0,4) matching the array size
	}
	n, _ := strconv.ParseUint(string(digits[:]), 16, 32)
	return rune(n), true
}

func (r *tokenReader) syntaxErrorOnLastToken(msg string) error { //nolint:unparam
	return SyntaxError{Message: msg, Offset: r.LastPos()}
}

func (r *tokenReader) syntaxErrorOnNextToken(msg string) error {
	t, err := r.next()
	if err != nil {
		return err
	}
	return SyntaxError{Message: msg, Value: t.description(), Offset: r.LastPos()}
}

// This is faster than creating a string to pass to strconv.Atoi.
func parseIntFromBytes(chars []byte) (int64, bool) {
	negate := false
	p := 0
	var ret int64
	if len(chars) == 0 {
		return 0, false
	}
	if chars[0] == '-' {
		negate = true
		p++
		if p == len(chars) {
			return 0, false
		}
	}
	const maxInt64 = 1<<63 - 1
	for p < len(chars) {
		d := int64(chars[p] - '0')
		// Signal overflow rather than silently wrapping; the caller falls back to float parsing.
		if ret > (maxInt64-d)/10 {
			return 0, false
		}
		ret = ret*10 + d
		p++
	}
	if negate {
		ret = -ret
	}
	return ret, true
}

func appendRune(out []byte, ch rune) []byte {
	var encodedRune [10]byte
	n := utf8.EncodeRune(encodedRune[0:10], ch)
	return append(out, encodedRune[0:n]...)
}

func valueKindFromTokenKind(k tokenKind) ValueKind {
	switch k {
	case nullToken:
		return NullValue
	case boolToken:
		return BoolValue
	case numberToken:
		return NumberValue
	case stringToken:
		return StringValue
	}
	return -1
}
