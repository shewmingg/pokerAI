package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"pokerAI/inputDto"
	"pokerAI/poker"
	"sync"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)

	var wg sync.WaitGroup
	routineNum := 1
	// determine which input will be used, (strategy pattern
	input := &inputDto.WePokerInput{}
	input.Init()
	_, restart := interface{}(input).(inputDto.WithRestart)
	if restart {
		// 如果有重启机制就要监听并等待重启
		for {
			ctx, cancel := context.WithCancel(context.Background())
			input.NewContext(ctx)
			table := poker.NewTable()
			table.TableSize = input.GetTableSize()
			table.MyCard = input.GetSelfCard()
			fmt.Printf("%s %s\n", table.MyCard[0], table.MyCard[1])
			wg.Add(1)
			go routine(ctx, &wg, routineNum, &table, input, cancel)

			wg.Add(1)
			go monitor(ctx, cancel, &wg, input, &table)

			wg.Wait()

			// After first routine is cancelled, start it again
			routineNum++
		}
	} else {
		for {
			table := poker.NewTable()
			table.TableSize = input.GetTableSize()
			table.MyCard = input.GetSelfCard()
			NewGame(&table, input)
		}
	}

}

func preflopSetup(table *poker.Table, input inputDto.InputInterface) {
	// when wepoker initially recognize whole table, player's chip is already deducted
	// in this situation, immune to player's chip deduction
	immune2Chip := false
	if _, ok := input.(*inputDto.WePokerInput); ok {
		immune2Chip = true
	}

	ac := poker.Action{
		Action:      poker.SmallBlind,
		Chip:        table.BetSize / 2,
		Player:      table.GetCurPlayerName(),
		Immune2Chip: immune2Chip,
	}
	ac.Execute(table)
	fmt.Printf("%s auto small blind %d\n", table.GetCurPlayerName(), table.BetSize/2)

	table.Round.CurActionPos = table.NextValidPosition(table.Round.CurActionPos)
	ac = poker.Action{
		Action:      poker.BigBlind,
		Chip:        table.BetSize,
		Player:      table.GetCurPlayerName(),
		Immune2Chip: immune2Chip,
	}
	ac.Execute(table)
	fmt.Printf("%s auto big blind %d\n", table.GetCurPlayerName(), table.BetSize)
	table.Round.CurActionPos = table.NextValidPosition(table.Round.CurActionPos)
}

func NewGame(table *poker.Table, input inputDto.InputInterface) {
	restartImp, restart := interface{}(input).(inputDto.WithRestart)
	table.Dealer = input.GetDealer()
	if restart && restartImp.CheckExit() {
		return
	}
	fmt.Printf("dealer is %d\n", table.Dealer)
	table.Players = input.InitPlayerWithChips(table.Players, table.TableSize)
	if restart && restartImp.CheckExit() {
		return
	}
	table.BetSize = input.GetBetSize()
	fmt.Printf("cards: %s %s\n", table.MyCard[0], table.MyCard[1])
	table.Round = poker.NewRound(poker.Preflop, table.PreviousValidPosition(table.NextValidPosition(table.Dealer)), table.NextValidPosition(table.Dealer), table.TableSize)
	preflopSetup(table, input)
	input.Betting(table)
	if restart && restartImp.CheckExit() {
		return
	}
	flopCards := input.GetFlopCards()
	if restart && restartImp.CheckExit() {
		wrapUp(*table)
		return
	}
	table.TableCard[0] = flopCards[0]
	table.TableCard[1] = flopCards[1]
	table.TableCard[2] = flopCards[2]
	table.Round = poker.NewRound(poker.Flop, table.PreviousValidPosition(table.NextValidPosition(table.Dealer)), table.NextValidPosition(table.Dealer), table.TableSize)
	input.Betting(table)
	if restart && restartImp.CheckExit() {
		wrapUp(*table)
		return
	}

	table.TableCard[3] = input.GetTurnCard()
	if restart && restartImp.CheckExit() {
		wrapUp(*table)
		return
	}
	table.ShowTableCards()

	table.Round = poker.NewRound(poker.Turn, table.PreviousValidPosition(table.NextValidPosition(table.Dealer)), table.NextValidPosition(table.Dealer), table.TableSize)
	input.Betting(table)
	if restart && restartImp.CheckExit() {
		wrapUp(*table)
		return
	}

	table.TableCard[4] = input.GetRiverCard()
	if restart && restartImp.CheckExit() {
		wrapUp(*table)
		return
	}
	table.ShowTableCards()

	table.Round = poker.NewRound(poker.River, table.PreviousValidPosition(table.NextValidPosition(table.Dealer)), table.NextValidPosition(table.Dealer), table.TableSize)
	input.Betting(table)
	if restart && restartImp.CheckExit() {
		wrapUp(*table)
		return
	}
	// todo 判断winner
	wrapUp(*table)
	return
}

func routine(ctx context.Context, wg *sync.WaitGroup, routineNum int, table *poker.Table, inputInterface inputDto.InputWithRestart, cancel context.CancelFunc) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Routine %d restarting\n", routineNum)
			return
		default:
			fmt.Printf("Routine %d is working\n", routineNum)
			NewGame(table, inputInterface)
			cancel()
			return
		}
	}
}

// 观察什么时候要开启新一轮
func monitor(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, input inputDto.InputWithRestart, table *poker.Table) {
	defer wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 重新开启的条件
			if input.RestartRound(table) {
				// If restart condition is met, send a signal on restartCh.
				fmt.Println("cancel")
				cancel()
				return
			}
		}
	}
}

// based on table, gen table situation and write to file, used for player identification
func wrapUp(table poker.Table) {

}
