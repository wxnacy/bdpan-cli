package terminal

import "github.com/gdamore/tcell/v2"

type SelectItem struct {
	IsSelect bool
	Info     interface{}
}

type Select struct {
	StartX      int
	StartY      int
	MaxWidth    int
	MaxHeight   int
	Items       []*SelectItem
	SelectIndex int
	StyleSelect tcell.Style
}

func (s Select) GetSeleteItem() *SelectItem {
	return s.Items[s.SelectIndex]
}

func (s *Select) GetDrawItems() []*SelectItem {
	if len(s.Items) > s.MaxHeight {

	}
	return s.Items
}

func (s *Select) MoveDownSelect(step int) (isChange bool) {
	var minH = s.MaxHeight
	if len(s.Items) < minH {
		minH = len(s.Items)
	}
	if s.SelectIndex+1 < minH {
		s.SelectIndex += 1
		isChange = true
	}
	Log.Debugf("MoveDownSelect step: %d index %d", step, s.SelectIndex)
	return
}

func (s *Select) MoveUpSelect(step int) (isChange bool) {
	if s.SelectIndex != 0 {
		s.SelectIndex -= step
		if s.SelectIndex < 0 {
			s.SelectIndex = 0
		}
		isChange = true
	}
	Log.Debugf("MoveUpSelect step: %d index %d", step, s.SelectIndex)
	return
}
