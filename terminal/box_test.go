package terminal

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestWidthHeight(t *testing.T) {
	s, err := tcell.NewScreen()
	if err != nil {
		t.Errorf("NewScreen %v", err)
	}
	b := NewBox(s, 0, 0, 100, 100)
	if b.Width() != 98 {
		t.Errorf("Width %d != 98", b.Width())
	}
}
