package inputDto

import (
	"context"
	"pokerAI/poker"
)

type InputInterface interface {
	InitPlayerWithChips(player []poker.Player, tableSize int) []poker.Player // 玩家们的初始筹码
	GetBetSize() int64                                                       // 大盲注大小
	GetDealer() int                                                          // 决定dealer的下标，从自己依次往左为0，1，2，3，4，5
	GetFlopCards() []*poker.Card                                             // flop card
	GetTurnCard() *poker.Card                                                // turn card
	GetRiverCard() *poker.Card                                               // river card
	GetSelfCard() [2]*poker.Card                                             // 发自己的牌
	Winner(table poker.Table)                                                // 判断winner
	GetTableSize() int                                                       // 确认桌子有多少个位子
	PlayerBetting(table *poker.Table, playerPos int) poker.Action
	Init()
	Betting(table *poker.Table)
}

type WithRestart interface {
	NewContext(ctx context.Context)
	RestartRound(table *poker.Table) bool // 何时重新开始牌局
	CheckExit() bool
}

type InputWithRestart interface {
	InputInterface
	WithRestart
}
