package printer

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/jedib0t/go-pretty/v6/text"
)

// TODO: Consider using for all URL outputs:
// * Clickable URLs influenced of the termLink library https://github.com/savioxavier/termlink/blob/master/termlink.go#L165
// * return fmt.Sprintf("\x1b]8;;%s/ui/%s/kv/%s/edit\x07%s\x1b]8;;\x07", address, datacenter, key, key)
func FmtURL(url string) string {
	colorsURL := text.Colors{text.Underline, text.FgCyan}
	return colorsURL.Sprint(url)
	// transformer := text.NewURLTransformer()
	// return transformer(url)
}

func ColorFaint(txt string) string {
	col := text.Colors{text.Faint}
	return col.Sprint(txt)
}

func ColorHiBlue(txt string) string {
	col := text.Colors{text.FgHiBlue}
	return col.Sprint(txt)
}
func ColorHiMagenta(txt string) string {
	col := text.Colors{text.FgHiMagenta}
	return col.Sprint(txt)
}

func ColorHiYellow(txt string) string {
	col := text.Colors{text.FgHiYellow}
	return col.Sprint(txt)
}

func PrettyJson(js string) string {
	t := text.NewJSONTransformer("", "  ")
	return t(js)
}

func TruncateText(s string, max int, delimeters string) string {
	if max > len(s) {
		return s
	}
	if delimeters == "" {
		delimeters = " ,"
	}
	return s[:strings.LastIndexAny(s[:max], delimeters)] + "..."
}

func SanitizeASCII(s string) string {
	cleanVal := strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII || r < 32 {
			// return -1
			return ' '
		}
		return r
	}, s)

	rg := regexp.MustCompile(`\s+`)
	return rg.ReplaceAllString(cleanVal, " ")
}
