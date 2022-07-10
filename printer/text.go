package printer

import (
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
