package poker

import (
	"fmt"
	"log"
	"strconv"
)

// 桌子上的一切
type Table struct {
	MyCard    [2]*Card
	TableCard [5]*Card
	Actions   map[string][]Action
	Situation string
	Dealer    int
	Players   []Player
	TableSize int // 桌上位置数
	Pot       int64
	BetSize   int64
	Round     Round
}

func (t *Table) GetSelfName() string {
	return "P0"
}
func (t *Table) GetCurPlayerName() string {
	return "P" + strconv.Itoa(t.Round.CurActionPos)
}

func (t *Table) GetPlayerName(pos int) string {
	return "P" + strconv.Itoa(pos)
}

// 当前轮次信息
type Round struct {
	WhichRound    string  // 哪一轮次 ： preflop/ flop/ turn/ river
	CurActionPos  int     // 当前做动作的player下标
	CallBetSize   int64   // 想call的话，需要多少chip
	FinalPosition int     // 到谁可以结束这轮
	PlayerBet     []int64 // 每个player当前轮次的下注
}

func NewRound(whichRound string, finalPos int, startPos int, tableSize int) Round {
	pb := make([]int64, tableSize)
	return Round{WhichRound: whichRound, FinalPosition: finalPos, PlayerBet: pb, CurActionPos: startPos}
}

const Preflop = "preflop"
const Flop = "flop"
const Turn = "turn"
const River = "river"

func NewTable() Table {
	return Table{Actions: map[string][]Action{
		Preflop: []Action{},
		Flop:    []Action{},
		Turn:    []Action{},
		River:   []Action{},
	}}
}
func (t *Table) ShowSituation() {
	fmt.Println()
	fmt.Println("======================Table Situation======================")
	fmt.Printf("Table: %s %s %s %s %s\n", t.TableCard[0], t.TableCard[1], t.TableCard[2], t.TableCard[3], t.TableCard[4])
	fmt.Printf("Your Cards: %s %s\n", t.MyCard[0], t.MyCard[1])
	fmt.Printf("PotSize: %d\n", t.Pot)
	for _, v := range t.Actions[t.Round.WhichRound] {
		fmt.Printf("%s %s", v.Player, v.Action)
		if v.Chip > 0 {
			fmt.Printf(" %d, ", v.Chip)
		} else {
			fmt.Print(", ")
		}
	}
	fmt.Println()
	fmt.Println("===========================================================")
	fmt.Println()
}

func (t *Table) NextValidPosition(pos int) int {
	curPos := pos
	for {
		pos = (pos + 1) % t.TableSize
		if t.Players[pos].InTable {
			return pos
		}
		// 转了一圈说明有问题
		if pos == curPos {
			log.Println("next pos is me?")
			return -1
		}
	}
}

func (t *Table) PreviousValidPosition(pos int) int {
	curPos := pos
	for {
		pos = (pos - 1 + t.TableSize) % t.TableSize
		if t.Players[pos].InTable {
			return pos
		}
		// 转了一圈说明有问题
		if pos == curPos {
			log.Fatal("previous pos is me?")
		}
	}
	return 0
}

func (t *Table) ShowTableCards() {
	fmt.Printf("Table cards are: ")
	for i := 0; i < len(t.TableCard); i++ {
		fmt.Printf("%s, ", t.TableCard[i])
	}
	fmt.Println()
}
