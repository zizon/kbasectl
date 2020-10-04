package generate

import (
	"testing"

	"github.com/zizon/kbasectl/pkg/panichain"
)

func TestGenreateTempalte(t *testing.T) {
	panichain.Propogate(NewTemplateCommand().Execute())
}
