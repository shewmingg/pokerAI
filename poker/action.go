package poker

import (
	"fmt"
	"log"
	"pokerAI/util"
	"strconv"
)

type Action struct {
	Player      string
	Action      string // three option: Call Fold Raise
	Chip        int64
	Immune2Chip bool // 由于有的input在识别玩家筹码时已经将第一次下注给扣除了，所以需要在此情况下不对玩家筹码做扣减
}

const Call = "called"
const Fold = "folded"
const Raise = "raised"
const Check = "checked"
const SmallBlind = "SB"
const BigBlind = "BB"

// 1. call/check: c 2. raise: r100 3. fold:f
func NewAction(player string, s string, callBet int64, playerBet int64) Action {
	var chip int64
	var action string
	// 尝试action是否直接是数字
	num, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		chip = num
		if chip == callBet {
			action = Call
		} else if chip == 0 {
			action = Check
		} else {
			action = Raise
		}
		return Action{
			Player: player,
			Action: action,
			Chip:   chip,
		}
	}
	switch s {
	case "c", "call":
		chip = callBet
		if chip == 0 {
			action = Check
		} else {
			action = Call
			chip += playerBet
		}
	case "f", "fold":
		chip = 0
		action = Fold
	default:
		if s[0] != 'r' {
			log.Fatal("invalid action")
		}
		var err error
		chip, err = util.ParseNumber(s[1:])
		action = Raise
		if err != nil {
			log.Fatal(err)
		}
	}
	return Action{
		Player: player,
		Action: action,
		Chip:   chip,
	}
}

func (a Action) String() string {
	if a.Action == Raise {
		return fmt.Sprintf("%s %s to %d", a.Player, a.Action, a.Chip)
	} else if a.Chip != 0 {
		return fmt.Sprintf("%s %s %d", a.Player, a.Action, a.Chip)
	} else {
		return fmt.Sprintf("%s %s", a.Player, a.Action)
	}
}

func (action Action) Execute(table *Table) {
	// 将操作记录下来
	table.Actions[table.Round.WhichRound] = append(table.Actions[table.Round.WhichRound], action)
	// 根据不同动作执行
	if action.Action == Call {
		Bet(action, table)
	} else if action.Action == Fold {
		table.Players[table.Round.CurActionPos].InTable = false
	} else if action.Action == Raise {
		Bet(action, table)
		table.Round.FinalPosition = table.PreviousValidPosition(table.Round.CurActionPos)

	} else if action.Action == SmallBlind {
		Bet(action, table)
	} else if action.Action == BigBlind {
		Bet(action, table)
		table.Round.FinalPosition = table.Round.CurActionPos // 大盲是最后一位
	}
}

func Bet(action Action, table *Table) {
	// 如果是raise，需要去掉之前bet过的chip
	actualBettingSize := action.Chip - table.Round.PlayerBet[table.Round.CurActionPos]
	if action.Immune2Chip {
		table.Players[table.Round.CurActionPos].Bet(actualBettingSize)
	}
	// 当前轮用户所下的筹码需要记录下来
	table.Round.PlayerBet[table.Round.CurActionPos] += actualBettingSize
	table.Pot += actualBettingSize
	table.Round.CallBetSize = table.Round.PlayerBet[table.Round.CurActionPos]
}
