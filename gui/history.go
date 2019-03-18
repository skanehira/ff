package gui

// HistoryManager have the move history
type HistoryManager struct {
	idx       int
	histories []string
}

// NewHistoryManager new history manager
func NewHistoryManager() *HistoryManager {
	return &HistoryManager{}
}

// Save save the move history
func (p *HistoryManager) Save(path string) {
	count := len(p.histories)

	// if not have history
	if p.idx == count-1 || count == 0 {
		p.histories = append(p.histories, path)
		p.idx++
	} else {
		p.histories = append(p.histories[:p.idx+1], path)
		p.idx = len(p.histories) - 1
	}
}

// Previous return the previous history
func (p *HistoryManager) Previous() string {
	count := len(p.histories)
	if count == 0 {
		return ""
	}

	p.idx--
	if p.idx < 0 {
		p.idx = 0
	}
	return p.histories[p.idx]
}

// Next return the next history
func (p *HistoryManager) Next() string {
	count := len(p.histories)
	if count == 0 {
		return ""
	}

	p.idx++
	if p.idx >= count {
		p.idx = count - 1
	}
	return p.histories[p.idx]
}
