package words

import (
	"fmt"
	"strconv"
)

// sprint is a placeholder value as text: strings as they are, integers
// without formatting, anything else as %v prints it.
func sprint(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case fmt.Stringer:
		return x.String()
	}
	return fmt.Sprint(v)
}
