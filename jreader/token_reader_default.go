package jreader

// This file defines the low-level JSON tokenizer. We don't define an interface for these methods,
// because calling them through an interface would limit performance.

import (
	"bytes"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"
)

var (
	tokenNull  = []byte("null")  //nolint:gochecknoglobals
	tokenTrue  = []byte("true")  //nolint:gochecknoglobals
	tokenFalse = []byte("false") //nolint:gochecknoglobals
)

// whitespaceChars marks the bytes that are skipped between tokens. Each byte is classified the
// way unicode.IsSpace classifies its Latin-1 code point, matching a per-byte rune conversion of
// the input (so the single bytes 0x85 and 0xA0 also count as whitespace).
var whitespaceChars = makeWhitespaceChars() //nolint:gochecknoglobals

func makeWhitespaceChars() (t [256]bool) {
	for c := 0; c < 256; c++ {
		if unicode.IsSpace(rune(c)) {
			t[c] = true
		}
	}
	return
}

// plainStringChars marks the bytes that pass through a decoded string unchanged: the ASCII
// characters other than the quote mark and the backslash. Everything else ends a scan of plain
// characters and is handled individually.
var plainStringChars = makePlainStringChars() //nolint:gochecknoglobals

func makePlainStringChars() (t [256]bool) {
	for c := 0; c < utf8.RuneSelf; c++ {
		if c != '"' && c != '\\' {
			t[c] = true
		}
	}
	return
}

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
	data      []byte
	pos       int
	len       int
	hasUnread bool
	tok       token // the most recently parsed token; the unread token when hasUnread is true
	lastPos   int
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
	if err := r.next(); err != nil {
		return false, err
	}
	if r.tok.kind == nullToken {
		return true, nil
	}
	r.putBack()
	if r.tok.kind == delimiterToken && r.tok.delimiter != '[' && r.tok.delimiter != '{' {
		return false, SyntaxError{Message: errMsgUnexpectedChar, Value: string(r.tok.delimiter), Offset: r.getPos()}
	}
	return false, nil
}

// Bool requires that the next token is a JSON boolean, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON boolean.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Bool() (bool, error) {
	if !r.hasUnread {
		b, ok := r.skipWhitespaceAndReadByte()
		if !ok {
			return false, io.EOF
		}
		if b == 't' || b == 'f' {
			// A keyword starting with these letters can only be a boolean or malformed.
			_, boolValue, err := r.readKeyword(b)
			return boolValue, err
		}
		r.unreadByte()
	}
	if err := r.consumeScalar(boolToken); err != nil {
		return false, err
	}
	return r.tok.boolValue, nil
}

// Number requires that the next token is a JSON number, returning its value if successful (consuming
// the token), or an error if the next token is anything other than a JSON number.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Number() (float64, error) {
	if !r.hasUnread {
		b, ok := r.skipWhitespaceAndReadByte()
		if !ok {
			return 0, io.EOF
		}
		if (b >= '0' && b <= '9') || b == '-' {
			if n, ok2 := r.readNumber(b); ok2 {
				return n, nil
			}
			return 0, SyntaxError{Message: errMsgInvalidNumber, Offset: r.lastPos}
		}
		r.unreadByte()
	}
	if err := r.consumeScalar(numberToken); err != nil {
		return 0, err
	}
	return r.tok.numberValue, nil
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
	if !r.hasUnread {
		b, ok := r.skipWhitespaceAndReadByte()
		if !ok {
			return nil, io.EOF
		}
		if b == '"' {
			return r.readString()
		}
		r.unreadByte()
	}
	if err := r.consumeScalar(stringToken); err != nil {
		return nil, err
	}
	return r.tok.stringValue, nil
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
	name, err := r.StringAsBytes()
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
	return name, nil
}

// Delimiter checks whether the next token is the specified ASCII delimiter character. If so, it
// returns (true, nil) and consumes the token. If it is a delimiter, but not the same one, it
// returns (false, nil) and does not consume the token. For anything else, it returns an error.
//
// This and all other tokenReader methods skip transparently past whitespace between tokens.
func (r *tokenReader) Delimiter(delimiter byte) (bool, error) {
	if r.hasUnread {
		if r.tok.kind == delimiterToken && r.tok.delimiter == delimiter {
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
	if err := r.next(); err != nil {
		return false, err // it was malformed JSON
	}
	r.putBack() // it was valid JSON, we just haven't hit that delimiter
	return false, nil
}

// EndDelimiterOrComma checks whether the next token is the specified ASCII delimiter character
// or a comma. If it is the specified delimiter, it returns (true, nil) and consumes the token.
// If it is a comma, it returns (false, nil) and consumes the token. For anything else, it
// returns an error. The delimiter parameter will always be either '}' or ']'.
func (r *tokenReader) EndDelimiterOrComma(delimiter byte) (bool, error) {
	if r.hasUnread {
		if r.tok.kind == delimiterToken &&
			(r.tok.delimiter == delimiter || r.tok.delimiter == ',') {
			r.hasUnread = false
			return r.tok.delimiter == delimiter, nil
		}
		return false, SyntaxError{Message: badArrayOrObjectItemMessage(delimiter == '}'),
			Value: r.tok.description(), Offset: r.lastPos}
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return false, io.EOF
	}
	if b == delimiter || b == ',' {
		return b == delimiter, nil
	}
	r.unreadByte()
	if err := r.next(); err != nil {
		return false, err
	}
	return false, SyntaxError{Message: badArrayOrObjectItemMessage(delimiter == '}'),
		Value: r.tok.description(), Offset: r.lastPos}
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
	if r.hasUnread {
		r.hasUnread = false
		return r.tokenToAnyValue(ignoreString)
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return AnyValue{}, io.EOF
	}
	switch {
	case b >= 'a' && b <= 'z':
		kind, boolValue, err := r.readKeyword(b)
		if err != nil {
			return AnyValue{}, err
		}
		if kind == boolToken {
			return AnyValue{Kind: BoolValue, Bool: boolValue}, nil
		}
		return AnyValue{Kind: NullValue}, nil
	case (b >= '0' && b <= '9') || b == '-':
		n, ok2 := r.readNumber(b)
		if !ok2 {
			return AnyValue{}, SyntaxError{Message: errMsgInvalidNumber, Offset: r.lastPos}
		}
		return AnyValue{Kind: NumberValue, Number: n}, nil
	case b == '"':
		stringValue, err := r.readString()
		if err != nil {
			return AnyValue{}, err
		}
		var s string
		if !ignoreString {
			s = string(stringValue)
		}
		return AnyValue{Kind: StringValue, String: s}, nil
	case b == '[':
		return AnyValue{Kind: ArrayValue}, nil
	case b == '{':
		return AnyValue{Kind: ObjectValue}, nil
	}
	return AnyValue{}, SyntaxError{Message: errMsgUnexpectedChar, Value: string(b), Offset: r.lastPos}
}

// tokenToAnyValue converts the token in r.tok, which has just been consumed, to an AnyValue.
func (r *tokenReader) tokenToAnyValue(ignoreString bool) (AnyValue, error) {
	switch r.tok.kind {
	case boolToken:
		return AnyValue{Kind: BoolValue, Bool: r.tok.boolValue}, nil
	case numberToken:
		return AnyValue{Kind: NumberValue, Number: r.tok.numberValue}, nil
	case stringToken:
		var s string
		if !ignoreString {
			s = string(r.tok.stringValue)
		}
		return AnyValue{Kind: StringValue, String: s}, nil
	case delimiterToken:
		if r.tok.delimiter == '[' {
			return AnyValue{Kind: ArrayValue}, nil
		}
		if r.tok.delimiter == '{' {
			return AnyValue{Kind: ObjectValue}, nil
		}
		return AnyValue{},
			SyntaxError{Message: errMsgUnexpectedChar, Value: string(r.tok.delimiter), Offset: r.lastPos}
	default:
		return AnyValue{Kind: NullValue}, nil
	}
}

// Attempts to parse and consume the next token, ignoring whitespace, leaving it in r.tok. A token
// is either a valid JSON scalar value or an ASCII delimiter character. If a token was previously
// unread using putBack, it consumes that instead. When an error is returned, r.tok is not
// meaningful.
func (r *tokenReader) next() error {
	if r.hasUnread {
		r.hasUnread = false
		return nil
	}
	b, ok := r.skipWhitespaceAndReadByte()
	if !ok {
		return io.EOF
	}

	switch {
	// We can get away with reading bytes instead of runes because the JSON spec doesn't allow multi-byte
	// characters except within a string literal.
	case b >= 'a' && b <= 'z':
		kind, boolValue, err := r.readKeyword(b)
		if err != nil {
			return err
		}
		r.tok = token{kind: kind, boolValue: boolValue}
		return nil
	case (b >= '0' && b <= '9') || b == '-':
		n, ok2 := r.readNumber(b)
		if !ok2 {
			return SyntaxError{Message: errMsgInvalidNumber, Offset: r.lastPos}
		}
		r.tok = token{kind: numberToken, numberValue: n}
		return nil
	case b == '"':
		s, err := r.readString()
		if err != nil {
			return err
		}
		r.tok = token{kind: stringToken, stringValue: s}
		return nil
	case b == '[', b == ']', b == '{', b == '}', b == ':', b == ',':
		r.tok = token{kind: delimiterToken, delimiter: b}
		return nil
	}

	return SyntaxError{Message: errMsgUnexpectedChar, Value: string(b), Offset: r.lastPos}
}

// putBack marks the token in r.tok, which must have just been parsed by next(), as unread, so
// that the next call to next() will consume it again instead of parsing new input.
func (r *tokenReader) putBack() {
	r.hasUnread = true
}

func (r *tokenReader) consumeScalar(kind tokenKind) error {
	if err := r.next(); err != nil {
		return err
	}
	if r.tok.kind == kind {
		return nil
	}
	if r.tok.kind == delimiterToken && r.tok.delimiter != '[' && r.tok.delimiter != '{' {
		return SyntaxError{Message: errMsgUnexpectedChar, Value: string(r.tok.delimiter), Offset: r.LastPos()}
	}
	return TypeError{Expected: valueKindFromTokenKind(kind),
		Actual: r.tok.valueKind(), Offset: r.LastPos()}
}

// readKeyword parses the remainder of a keyword token (true, false, or null) whose first letter
// has already been consumed, returning the token kind and, for a boolean, its value.
func (r *tokenReader) readKeyword(first byte) (tokenKind, bool, error) {
	n := r.consumeASCIILowercaseAlphabeticChars() + 1
	id := r.data[r.lastPos : r.lastPos+n]
	if first == 'f' && bytes.Equal(id, tokenFalse) {
		return boolToken, false, nil
	}
	if first == 't' && bytes.Equal(id, tokenTrue) {
		return boolToken, true, nil
	}
	if first == 'n' && bytes.Equal(id, tokenNull) {
		return nullToken, false, nil
	}
	return 0, false, SyntaxError{Message: errMsgUnexpectedSymbol, Value: string(id), Offset: r.lastPos}
}

func (r *tokenReader) unreadByte() {
	r.pos--
}

func (r *tokenReader) skipWhitespaceAndReadByte() (byte, bool) {
	data, n := r.data, r.len
	p := r.pos
	for p < n {
		ch := data[p]
		if whitespaceChars[ch] {
			p++
			continue
		}
		r.lastPos = p
		r.pos = p + 1
		return ch, true
	}
	r.pos = p
	return 0, false
}

func (r *tokenReader) consumeASCIILowercaseAlphabeticChars() int {
	data, n := r.data, r.len
	p := r.pos
	for p < n && data[p] >= 'a' && data[p] <= 'z' {
		p++
	}
	count := p - r.pos
	r.pos = p
	return count
}

func (r *tokenReader) readNumber(_ byte) (float64, bool) {
	// The digit run may contain at most one '.', and an exponent part must match
	// [eE][-+]?[0-9]+; beyond that the number's format is validated by the string-to-number
	// conversion at the end. We scan the input directly by index rather than through
	// readByte/unreadByte, updating r.pos at every return point so that the read position
	// always reflects exactly the bytes consumed.
	startPos := r.lastPos
	data, n := r.data, r.len
	p := startPos + 1 // the first byte has already been read
	isFloat := false

	// Integer and fractional digits. The byte that ends the run is consumed and then
	// reconsidered below.
	var ch byte
	consumedEnd := false
	for p < n {
		ch = data[p]
		p++
		if (ch < '0' || ch > '9') && (ch != '.' || isFloat) {
			consumedEnd = true
			break
		}
		if ch == '.' {
			isFloat = true
		}
	}

	if consumedEnd && (ch == 'e' || ch == 'E') {
		// The exponent marker must be followed by an optional sign and at least one digit.
		if p >= n {
			r.pos = p
			return 0, false
		}
		ch = data[p]
		p++
		if ch >= '0' && ch <= '9' {
			p-- // the digit is consumed by the loop below
		} else if ch != '+' && ch != '-' {
			r.pos = p
			return 0, false
		}
		hasExponent := false
		for p < n && data[p] >= '0' && data[p] <= '9' {
			p++
			hasExponent = true
		}
		if !hasExponent {
			r.pos = p
			return 0, false
		}
		isFloat = true
	} else if consumedEnd {
		p-- // the byte that ended the digit run is not part of the number
	}

	r.pos = p
	chars := data[startPos:p]
	if isFloat {
		// Unfortunately, strconv.ParseFloat requires a string - there is no []byte equivalent. This means we can't
		// avoid a heap allocation here. Easyjson works around this by creating an unsafe string that points directly
		// at the existing bytes, but in our default implementation we can't use unsafe.
		num, err := strconv.ParseFloat(string(chars), 64)
		return num, err == nil
	}
	num, ok := parseIntFromBytes(chars)
	return float64(num), ok
}

func (r *tokenReader) readString() ([]byte, error) {
	data, n := r.data, r.len
	start := r.pos // the opening quote mark has already been read
	p := start
	// Everything except the closing quote and the escape character passes through verbatim
	// (a quote or backslash byte can never occur inside a multi-byte UTF-8 sequence, so byte
	// comparisons are safe), which makes the whole string a zero-copy subslice of the input
	// unless it contains an escape.
	for p < n {
		b := data[p]
		if b == '"' {
			r.pos = p + 1
			if p == start {
				return nil, nil
			}
			return data[start:p], nil
		}
		if b == '\\' {
			return r.decodeString(start, p)
		}
		p++
	}
	return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
}

// decodeString handles a string that cannot be returned as a subslice of the input: the escape
// sequence at position p ends the plain prefix that began at start (just past the opening quote
// mark). It decodes the rest of the string into a new buffer in one forward scan, bulk-copying
// each run of plain characters.
func (r *tokenReader) decodeString(start, p int) ([]byte, error) {
	data, n := r.data, r.len
	buf := make([]byte, 0, decodedStringCapacity(data, start, n))
	buf = append(buf, data[start:p]...)
	for {
		runStart := p
		for p < n && plainStringChars[data[p]] {
			p++
		}
		buf = append(buf, data[runStart:p]...)
		if p >= n {
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
		b := data[p]
		if b == '"' {
			r.pos = p + 1
			return buf, nil
		}
		if b >= utf8.RuneSelf {
			// Decoding and re-encoding a character passes it through unchanged, except that
			// each invalid UTF-8 byte becomes the Unicode replacement character.
			_, size := utf8.DecodeRune(data[p:])
			if size > 1 {
				buf = append(buf, data[p:p+size]...)
				p += size
			} else {
				buf = utf8.AppendRune(buf, utf8.RuneError)
				p++
			}
			continue
		}
		// b == '\\'
		p++
		if p >= n {
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
		// All valid escape characters are ASCII, so reading a byte here is equivalent to reading
		// a character: any multi-byte or invalid sequence takes the default (error) branch.
		esc := data[p]
		p++
		switch esc {
		case '"', '\\', '/':
			buf = append(buf, esc)
		case 'b':
			buf = append(buf, '\b')
		case 'f':
			buf = append(buf, '\f')
		case 'n':
			buf = append(buf, '\n')
		case 'r':
			buf = append(buf, '\r')
		case 't':
			buf = append(buf, '\t')
		case 'u':
			ch, ok := readHex4(data, p, n)
			if !ok {
				return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
			}
			// AppendRune encodes a surrogate code point as the replacement character.
			buf = utf8.AppendRune(buf, ch)
			p += 4
		default:
			return nil, r.syntaxErrorOnLastToken(errMsgInvalidString)
		}
	}
}

// decodedStringCapacity returns a buffer capacity for decoding the string whose content starts at
// start (just past the opening quote mark): the string's raw length in the input. Decoding never
// needs more bytes than the raw form, except when invalid UTF-8 bytes are replaced (three bytes
// for one); the buffer grows in that rare case.
func decodedStringCapacity(data []byte, start, n int) int {
	p := start
	for p < n {
		switch data[p] {
		case '\\':
			p += 2
		case '"':
			return p - start
		default:
			p++
		}
	}
	return n - start
}

// readHex4 parses the four hex digits of a \u escape starting at data[p].
func readHex4(data []byte, p, n int) (rune, bool) {
	if p+4 > n {
		return 0, false
	}
	var v rune
	for i := 0; i < 4; i++ {
		c := data[p+i]
		switch {
		case c >= '0' && c <= '9':
			v = v<<4 + rune(c-'0')
		case c >= 'a' && c <= 'f':
			v = v<<4 + rune(c-'a'+10)
		case c >= 'A' && c <= 'F':
			v = v<<4 + rune(c-'A'+10)
		default:
			return 0, false
		}
	}
	return v, true
}

func (r *tokenReader) syntaxErrorOnLastToken(msg string) error { //nolint:unparam
	return SyntaxError{Message: msg, Offset: r.LastPos()}
}

func (r *tokenReader) syntaxErrorOnNextToken(msg string) error {
	if err := r.next(); err != nil {
		return err
	}
	return SyntaxError{Message: msg, Value: r.tok.description(), Offset: r.LastPos()}
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
	for p < len(chars) {
		ret = ret*10 + int64(chars[p]-'0')
		p++
	}
	if negate {
		ret = -ret
	}
	return ret, true
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
