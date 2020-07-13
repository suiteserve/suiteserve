package httpheader

// ConnectionTokens parses each given value as a list defined by RFC 2616's
// '#rule' BNF, which would usually appear as the value to the HTTP Connection
// header, and then flattens those lists and returns the elements. This list BNF
// for each value conforms to:
//   *1( *LWS connection-token *( *LWS "," *LWS connection-token ))
// where:
//   connection-token = token
//   token            = 1*<any CHAR except CTLs or separators>
//   separators       = "(" | ")" | "<" | ">" | "@"
//                    | "," | ";" | ":" | "\" | <">
//                    | "/" | "[" | "]" | "?" | "="
//                    | "{" | "}" | SP | HT
//   CHAR             = <any US-ASCII character (octets 0 - 127)>
//   CTL              = <any US-ASCII control character
//                      (octets 0 - 31) and DEL (127)>
//   SP               = <US-ASCII SP, space (32)>
//   HT               = <US-ASCII HT, horizontal-tab (9)>
//   LWS              = [CRLF] 1*( SP | HT )
// Unlike the RFC, ConnectionTokens allows the list to be empty. In addition,
// passing nil or an empty slice will return an empty slice. Whitespace around
// elements is trimmed.
func ConnectionTokens(values []string) []string {
	elements := make([]string, 0, len(values))
	for _, v := range values {
		elements = append(elements, consumeList([]rune(v), consumeToken)...)
	}
	return elements
}

// UpgradeTokens parses each given value as a list defined by RFC 2616's '#rule'
// BNF, which would usually appear as the value to the HTTP Upgrade header, and
// then flattens those lists and returns the elements. This list BNF for each
// value conforms to:
//   *1( *LWS product *( *LWS "," *LWS product ))
// where:
//   product         = token ["/" product-version]
//   product-version = token
//   token           = 1*<any CHAR except CTLs or separators>
//   separators      = "(" | ")" | "<" | ">" | "@"
//                   | "," | ";" | ":" | "\" | <">
//                   | "/" | "[" | "]" | "?" | "="
//                   | "{" | "}" | SP | HT
//   CHAR            = <any US-ASCII character (octets 0 - 127)>
//   CTL             = <any US-ASCII control character
//                     (octets 0 - 31) and DEL (127)>
//   SP              = <US-ASCII SP, space (32)>
//   HT              = <US-ASCII HT, horizontal-tab (9)>
//   LWS             = [CRLF] 1*( SP | HT )
// Unlike the RFC, UpgradeTokens allows the list to be empty. In addition,
// passing nil or an empty slice will return an empty slice. Whitespace around
// elements is trimmed.
func UpgradeTokens(values []string) []string {
	elements := make([]string, 0, len(values))
	for _, v := range values {
		elements = append(elements, consumeList([]rune(v), consumeProduct)...)
	}
	return elements
}

// consumeList consumes:
//   *1( *LWS element *( *LWS "," *LWS element ))
func consumeList(s []rune, elementFunc func(int, []rune) int) []string {
	var i, begin int
	var elements []string
	for i < len(s) {
		i = consumeLws(i, s)
		begin = i
		i = elementFunc(i, s)
		if i == begin {
			// we didn't consume an element
			i++
			continue
		}
		elements = append(elements, string(s[begin:i]))
		i = consumeLws(i, s)
		if i == len(s) || s[i] != ',' {
			break
		}
		i++
	}
	return elements
}

// consumeLws consumes:
//   *LWS
//   LWS = [CRLF] 1*( SP | HT )
func consumeLws(i int, s []rune) int {
	// base case since this function is recursive: exit when s[i] doesn't start
	// with ( CR | SP | HT )
	if len(s)-i == 0 || (s[i] != '\r' && s[i] != ' ' && s[i] != '\t') {
		return i
	}
	begin := i
	// consume [CRLF]
	if s[i] == '\r' {
		if len(s)-i == 1 || s[i+1] != '\n' {
			return begin
		}
		i += 2
	}
	if i == len(s) {
		return begin
	}
	// consume 1*( SP | HT )
	for ; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			break
		}
	}
	return consumeLws(i, s)
}

// consumeProduct consumes:
//   product         = token ["/" product-version]
//   product-version = token
func consumeProduct(i int, s []rune) int {
	i = consumeToken(i, s)
	if i < len(s) && s[i] == '/' {
		return consumeToken(i+1, s)
	}
	return i
}

var separators = map[rune]bool{
	'(':  true,
	')':  true,
	'<':  true,
	'>':  true,
	'@':  true,
	',':  true,
	';':  true,
	':':  true,
	'\\': true,
	'"':  true,
	'/':  true,
	'[':  true,
	']':  true,
	'?':  true,
	'=':  true,
	'{':  true,
	'}':  true,
	' ':  true,
	'\t': true,
}

// consumeToken consumes:
//   token = 1*<any CHAR except CTLs or separators>
func consumeToken(i int, s []rune) int {
	var ch rune
	for ; i < len(s); i++ {
		ch = s[i]
		if ch <= 31 || 127 <= ch || separators[ch] {
			break
		}
	}
	return i
}
