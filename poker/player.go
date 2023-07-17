package poker

import "log"

type Player struct {
	InTable     bool
	Chip        int64
	CurRoundBet int64 // 当前轮的下注
}

func (p *Player) Bet(chip int64) {
	if chip > p.Chip {
		log.Fatal("bet chip size invalid")
	}
	p.Chip -= chip
}
