package tui

type SizedModel struct {
	width  int
	height int
}

func (s *SizedModel) SetSize(width, height int) {
	s.width = width
	s.height = height
}
