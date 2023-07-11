package terminal

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type SelectItemInfo interface {
	String() string
	Name() string
	Id() string
}

type SelectItem struct {
	IsSelect bool
	Info     SelectItemInfo
}

func NewEmptySelect(t *Terminal, StartX, StartY, EndX, EndY int) *Select {
	box := t.NewBox(
		StartX,
		StartY,
		EndX,
		EndY,
		StyleDefault,
	)
	s := &Select{
		t:           t,
		Box:         box,
		AnchorIndex: 0,
		Items:       make([]*SelectItem, 0),
	}
	return s.init()
}

func NewSelect(t *Terminal, StartX, StartY int, items []*SelectItem) *Select {
	var maxInfoWidth int
	for _, item := range items {
		length := runewidth.StringWidth(item.Info.String())
		if length > maxInfoWidth {
			maxInfoWidth = length
		}
	}
	s := &Select{
		t:     t,
		Box:   t.NewBox(StartX, StartY, StartX+maxInfoWidth+1, StartY+len(items)+1, StyleDefault),
		Items: items,
	}
	return s.init()
}

type Select struct {
	Box    *Box
	StartX int
	StartY int
	// MaxWidth  int
	// MaxHeight int
	Items []*SelectItem
	// 选中行的样式
	StyleSelect tcell.Style
	// 选中行索引
	SelectIndex int
	// 选中后的执行方法
	SelectFn func(*SelectItem)
	// 上下移动时光标固定的锚点
	AnchorIndex int
	// Items 为空时需要填充的数据
	EmptyFillText string
	LoadingText   string

	t *Terminal
}

func (s *Select) init() *Select {
	s.StyleSelect = StyleSelect
	return s
}

func (s *Select) SetItems(items []*SelectItem) *Select {
	s.Items = items
	return s
}

func (s *Select) Filter(filter string) *Select {
	filterItems := make([]*SelectItem, 0)
	items := s.Items
	for _, item := range items {
		if strings.Contains(item.Info.String(), filter) {
			filterItems = append(filterItems, item)
		}
	}
	s.SetItems(filterItems)
	s.SelectIndex = 0
	return s
}

func (s *Select) SetSelectIndex(i int) *Select {
	s.SelectIndex = i
	return s
}

func (s *Select) SetEmptyFillText(t string) *Select {
	s.EmptyFillText = t
	return s
}

func (s *Select) SetSelectFn(fn func(*SelectItem)) *Select {
	s.SelectFn = fn
	return s
}

func (s *Select) SetLoadingText(t string) *Select {
	s.LoadingText = t
	return s
}

func (s *Select) SetAnchorIndex(i int) *Select {
	s.AnchorIndex = i
	return s
}

func (s Select) GetSeleteItem() *SelectItem {
	if s.Items == nil || len(s.Items) <= s.SelectIndex {
		return nil
	}
	return s.Items[s.SelectIndex]
}

func (s Select) Length() int {
	return len(s.Items)
}

func (s *Select) GetDrawItems(offset int) []*SelectItem {
	if len(s.Items) > s.Box.Height() {
		maxEnd := len(s.Items) - 1
		if offset > maxEnd {
			offset = maxEnd
		}
		end := s.Box.Height() + offset
		if end > maxEnd {
			end = maxEnd + 1
		}
		return s.Items[offset:end]
	}
	return s.Items
}

func (s *Select) MoveDownSelect(step int) (isChange bool) {
	var minH = len(s.Items)
	if s.SelectIndex+step < minH {
		s.SelectIndex += step
		isChange = true
	} else if s.SelectIndex < minH && s.SelectIndex+step >= minH {
		s.SelectIndex = minH - 1
		isChange = true
	}
	Log.Infof("MoveDownSelect step: %d index %d isChange %v", step, s.SelectIndex, isChange)
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
	Log.Infof("MoveUpSelect step: %d index %d isChange %v", step, s.SelectIndex, isChange)
	return
}

func (s *Select) IsMoveEnd() bool {
	if s.SelectIndex == len(s.Items)-1 {
		return true
	}
	return false
}

func (s *Select) DrawLoading() {
	s.Box.Draw()
	s.Box.DrawLineText(0, StyleDefault, s.LoadingText)
	s.t.S.Show()
}

func (s *Select) Draw() {
	s.Box.Draw()

	// 计算获取需要绘制的列表开始索引
	// TODO: 需要重构
	var offset int = 0
	selectItemLen := len(s.Items)
	selectMaxH := s.Box.Height()
	if selectItemLen > selectMaxH {
		selectI := s.SelectIndex
		selectOffset := selectMaxH - s.AnchorIndex
		if selectI+selectOffset > selectMaxH {
			offset = selectI + selectOffset - selectMaxH
			if offset > selectItemLen-selectMaxH {
				offset = selectItemLen - selectMaxH
			}
		}
	}
	Log.Infof("Select Draw Items offset %d StyleSelect %v", offset, s.StyleSelect)
	drawItems := s.GetDrawItems(offset)
	if drawItems == nil || len(drawItems) == 0 {
		// 没有内容时进行填充
		s.Box.DrawLineText(0, StyleDefault, s.EmptyFillText)
		s.t.S.Show()
		return
	}
	// 填充 select 信息
	for i, item := range drawItems {
		text := item.Info.String()
		style := StyleDefault
		if item.IsSelect {
			text = " " + text
		}
		if i+offset == s.SelectIndex {
			style = s.StyleSelect
		}
		s.Box.DrawLineText(i, style, text)
	}
	s.t.S.Show()
	// 绘制选中状态的额外操作
	selectItem := s.GetSeleteItem()
	if selectItem != nil && s.SelectFn != nil {
		s.SelectFn(selectItem)
	}
}
