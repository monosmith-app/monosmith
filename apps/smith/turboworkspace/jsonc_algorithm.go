package turboworkspace

// StripJSONC removes the comments and trailing commas that turbo.json (and
// turbo.jsonc) allow, leaving plain JSON. String contents are never touched.
func StripJSONC(raw []byte) []byte {
	out := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c == '"':
			out = append(out, c)
			for i++; i < len(raw); i++ {
				out = append(out, raw[i])
				if raw[i] == '\\' && i+1 < len(raw) {
					i++
					out = append(out, raw[i])
					continue
				}
				if raw[i] == '"' {
					break
				}
			}
		case c == '/' && i+1 < len(raw) && raw[i+1] == '/':
			for i < len(raw) && raw[i] != '\n' {
				i++
			}
			if i < len(raw) {
				out = append(out, '\n')
			}
		case c == '/' && i+1 < len(raw) && raw[i+1] == '*':
			i += 2
			for i+1 < len(raw) && (raw[i] != '*' || raw[i+1] != '/') {
				i++
			}
			i++
		case c == ',':
			j := i + 1
			for j < len(raw) && (raw[j] == ' ' || raw[j] == '\t' || raw[j] == '\n' || raw[j] == '\r') {
				j++
			}
			if j < len(raw) && (raw[j] == '}' || raw[j] == ']') {
				continue
			}
			out = append(out, c)
		default:
			out = append(out, c)
		}
	}
	return out
}
