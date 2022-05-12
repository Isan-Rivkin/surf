package printer

import (
	"time"

	"github.com/briandowns/spinner"
)

type Loader interface {
	Start(prefix, suffix, color string)
	Stop()
}

type SpinnerApi struct {
	s *spinner.Spinner
}

// spinnerType - https://github.com/briandowns/spinner#available-character-sets
func newSpinner(spinnerType int, prefix, suffix, color string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[spinnerType], 100*time.Millisecond)
	s.Prefix = prefix
	s.Suffix = suffix
	s.Color(color, "bold")
	return s
}

func (sp *SpinnerApi) Start(prefix, suffix, color string) {
	s := newSpinner(9, prefix+" ", suffix, color)
	sp.s = s
	sp.s.Start()
}

func (sp *SpinnerApi) Stop() {
	if sp.s != nil {
		sp.s.Stop()
	}
}
