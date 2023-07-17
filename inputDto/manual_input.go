package inputDto

import (
	"fmt"
	"log"
	"pokerAI/util"
)
import "pokerAI/poker"

type ManualInput struct {
}

func (mi *ManualInput) Init() {

}
func (mi *ManualInput) GetBetSize() (betSize int64) {
	fmt.Println("input bet size")
	var str string
	fmt.Scan(&str)
	betSize, err := util.ParseNumber(str)
	if err != nil {
		log.Fatal(err)
	}
	return
}
func (mi *ManualInput) GetTableSize() int {
	fmt.Println("input table size")
	var num int
	fmt.Scan(&num)
	return num
}

func (mi *ManualInput) GetDealer() int {
	fmt.Println("input dealer's place index")
	// 最后输入本轮庄家 （以下标决定)
	var dealer int
	fmt.Scan(&dealer)
	return dealer
}
func (mi *ManualInput) GetTurnCard() *poker.Card {
	fmt.Println("Input turn Card formatted as SuitValue")
	var suitValue string
	fmt.Scan(&suitValue)
	return poker.NewCard(suitValue)
}
func (mi *ManualInput) GetRiverCard() *poker.Card {
	fmt.Println("Input river Card formatted as SuitValue")
	var suitValue string
	fmt.Scan(&suitValue)
	return poker.NewCard(suitValue)
}

func (mi *ManualInput) GetFlopCards() (cards []*poker.Card) {
	fmt.Println("Input Flop Card formatted as SuitValue SuitValue SuitValue")
	var suitValue string
	fmt.Scan(&suitValue)
	cards = append(cards, poker.NewCard(suitValue))
	fmt.Scan(&suitValue)
	cards = append(cards, poker.NewCard(suitValue))
	fmt.Scan(&suitValue)
	cards = append(cards, poker.NewCard(suitValue))
	return
}

func (mi *ManualInput) GetSelfCard() (cards [2]*poker.Card) {
	var suitValue string
	fmt.Println("Input My Card formatted as SuitValue SuitValue")
	fmt.Scan(&suitValue)
	cards[0] = poker.NewCard(suitValue)
	fmt.Scan(&suitValue)
	cards[1] = poker.NewCard(suitValue)
	return
}

func (mi *ManualInput) InitPlayerWithChips(players []poker.Player, tableSize int) []poker.Player {
	// 先用0筹码代表没有人的位儿 从自己开始依次输入筹码数
	fmt.Printf("input %d players' chip\n", tableSize)
	var chip int64
	for i := 0; i < tableSize; i++ {
		fmt.Scan(&chip)
		inTable := true
		if chip == 0 {
			inTable = false
		}
		players = append(players, poker.Player{InTable: inTable, Chip: chip})
	}
	return players
}

// 每个轮次的下注
func (mi *ManualInput) Winner(table poker.Table) {
	fmt.Println("input who's winning")
	// TODO
	var num int
	fmt.Print("winner num: ")
	fmt.Scan(&num)
	//for i:=0;i<num;i++ {
	//	fmt.Scan()
	//}
}

func (mi *ManualInput) PlayerBetting(table *poker.Table, playerPos int) poker.Action {
	table.ShowSituation()
	fmt.Printf("%s betting...\n", table.GetCurPlayerName())
	if table.Round.CallBetSize == 0 {
		fmt.Println("nobody raised, you can check(c), raise(r***), fold(f)")
	} else {
		fmt.Printf("last one bets %d, you have already bets %d, you can call(c) with %d, raise(r***), fold(f)\n", table.Round.CallBetSize, table.Round.PlayerBet[table.Round.CurActionPos], table.Round.CallBetSize-table.Round.PlayerBet[table.Round.CurActionPos])
	}

	// 1. call/check: c 2. raise: r100 3. fold:f
	var actionStr string
	fmt.Scan(&actionStr)
	action := poker.NewAction(table.GetCurPlayerName(), actionStr, table.Round.CallBetSize-table.Round.PlayerBet[table.Round.CurActionPos], table.Round.PlayerBet[table.Round.CurActionPos])
	return action
}

// 直接开始
func (mi *ManualInput) WaitBettingTurn() bool {
	return true
}

// 手动的没有这个讲究
func (mi *ManualInput) RestartRound(table *poker.Table) bool {
	return false
}

func (mi *ManualInput) TraceWhichIsActive(table *poker.Table) int {
	return table.NextValidPosition(table.Round.FinalPosition)
}
