package inputDto

import (
	"context"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/kbinani/screenshot"
	"github.com/otiai10/gosseract/v2"
	"image"
	"image/color"
	"log"
	"pokerAI/gpt"
	"pokerAI/model"
	"pokerAI/poker"
	"pokerAI/util"
	"time"
)

const WePokerWidth = 546
const WePokerHeight = 875
const WePokerBetSize = 20
const WePokerTableSize = 9

// wepoker 有9个座位
type WePokerInput struct {
	ctx                 context.Context
	model.Screen        // 需要图像相关
	model.ImagePos      // 各个需要识别的位置
	GosseractClient     *gosseract.Client
	ChipGosseractClient *gosseract.Client
	ChineseGosseract    *gosseract.Client
	ActionPoint         []image.Point // 判断是否到某位玩家的行动时刻（头像周围有绿框）
	PotSizeSnapshots    []uint32
	NameAndActionArea   []image.Rectangle // 玩家名字区域，wepoker的名字区域也是用户做动作的区域，
	PotArea             image.Rectangle
	TableCardsArea      []image.Rectangle // 桌面牌
	PlayerCards         []CardImage       // 玩家牌，用于最后一步showcard比大小
}

type CardImage struct {
	SuitPoint image.Rectangle // 花色只需要一个点来确定
	CardRank  image.Rectangle // 卡片大小需要用ocr识别
}

func NewCardImage(screen model.Screen, suitX, suitY, rankX, rankY, rankWidth, rankHeight int) CardImage {
	return CardImage{
		SuitPoint: screen.Relative2AbsoluteRectangle(suitX, suitY, 2, 2),
		CardRank:  screen.Relative2AbsoluteRectangle(rankX, rankY, rankWidth, rankHeight),
	}
}

func (wpi *WePokerInput) NewContext(ctx context.Context) {
	wpi.ctx = ctx
}

func (wpi *WePokerInput) Init() {
	bounds := screenshot.GetDisplayBounds(0)
	minX, minY, maxX, maxY := bounds.Dx()-WePokerWidth, 0, bounds.Dx(), WePokerHeight
	err := util.SetChromeWindowSizeAndActivate(minX, minY, maxX, maxY)
	if err != nil {
		log.Fatal(err)
	}
	//
	fmt.Printf("screen size: %d %d %d %d", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
	// 浏览器设置位置及大小
	screen := model.Screen{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
	// hardcode 各个要识别的位置
	// 各个筹码区域
	wpi.ImagePos.Chips = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(500, 1488, 91, 25),
		screen.Relative2AbsoluteRectangle(205, 1153, 91, 25),
		screen.Relative2AbsoluteRectangle(204, 937, 91, 25),
		screen.Relative2AbsoluteRectangle(204, 719, 91, 25),
		screen.Relative2AbsoluteRectangle(384, 512, 91, 25),
		screen.Relative2AbsoluteRectangle(618, 512, 91, 27),
		screen.Relative2AbsoluteRectangle(796, 716, 91, 25),
		screen.Relative2AbsoluteRectangle(797, 937, 91, 25),
		screen.Relative2AbsoluteRectangle(796, 1150, 91, 25),
	}
	// 给出dealer的位置，通过识别在指定位置是否有D确认谁是dealer
	// 由于dealer是直接在屏幕上抓取的，所以取绝对值
	wpi.ImagePos.Dealers = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(463, 1487, 30, 28), // D0
		screen.Relative2AbsoluteRectangle(267, 1180, 30, 28), // D1
		screen.Relative2AbsoluteRectangle(267, 964, 30, 28),  // D2
		screen.Relative2AbsoluteRectangle(267, 742, 30, 28),  // D3
		screen.Relative2AbsoluteRectangle(480, 511, 30, 28),  // D4
		screen.Relative2AbsoluteRectangle(714, 508, 30, 28),  // D5
		screen.Relative2AbsoluteRectangle(789, 745, 30, 28),  // D6
		screen.Relative2AbsoluteRectangle(789, 964, 30, 28),  // D7
		screen.Relative2AbsoluteRectangle(756, 1069, 30, 28), // D8
	}
	// 每张卡占俩元素，第一个是花色，第二个是大小
	wpi.ImagePos.MyCards = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(516, 1602, 2, 2),   //第一张卡花色
		screen.Relative2AbsoluteRectangle(466, 1538, 32, 46), //第一张卡rank
		screen.Relative2AbsoluteRectangle(594, 1608, 2, 2),   // 第二张卡花色
		screen.Relative2AbsoluteRectangle(553, 1540, 32, 46), // 第二张卡rank
	}
	// 每个玩家出的筹码
	wpi.ImagePos.Betting = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(630, 1358, 60, 29), // 我的下注
		screen.Relative2AbsoluteRectangle(336, 1041, 60, 27), // P1下注
		screen.Relative2AbsoluteRectangle(336, 825, 60, 27),  // P2下注
		screen.Relative2AbsoluteRectangle(336, 606, 60, 27),  // P3下注
		screen.Relative2AbsoluteRectangle(414, 545, 60, 27),  // P4下注
		screen.Relative2AbsoluteRectangle(650, 545, 60, 27),  // P5下注
		screen.Relative2AbsoluteRectangle(694, 606, 60, 27),  // P6下注
		screen.Relative2AbsoluteRectangle(693, 826, 60, 27),  // P7下注
		screen.Relative2AbsoluteRectangle(693, 1037, 60, 27), // P8下注

	}
	// 确认目前是谁下注
	wpi.ActionPoint = []image.Point{
		//screen.Relative2AbsolutePoint(image.Point{X: 530, Y: 1371}),
		screen.Relative2AbsolutePoint(image.Point{X: 530, Y: 1371}),
		screen.Relative2AbsolutePoint(image.Point{X: 231, Y: 1035}),
		screen.Relative2AbsolutePoint(image.Point{X: 230, Y: 819}),
		screen.Relative2AbsolutePoint(image.Point{X: 231, Y: 601}),
		screen.Relative2AbsolutePoint(image.Point{X: 411, Y: 395}),
		screen.Relative2AbsolutePoint(image.Point{X: 646, Y: 395}),
		screen.Relative2AbsolutePoint(image.Point{X: 819, Y: 601}),
		screen.Relative2AbsolutePoint(image.Point{X: 824, Y: 819}),
		screen.Relative2AbsolutePoint(image.Point{X: 827, Y: 1035}),
	}

	wpi.NameAndActionArea = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(500, 1374, 91, 25),
		screen.Relative2AbsoluteRectangle(205, 1039, 91, 25),
		screen.Relative2AbsoluteRectangle(204, 823, 91, 25),
		screen.Relative2AbsoluteRectangle(204, 605, 91, 25),
		screen.Relative2AbsoluteRectangle(384, 398, 91, 25),
		screen.Relative2AbsoluteRectangle(618, 398, 91, 25),
		screen.Relative2AbsoluteRectangle(796, 606, 91, 25),
		screen.Relative2AbsoluteRectangle(797, 823, 91, 25),
		screen.Relative2AbsoluteRectangle(796, 1038, 91, 25),
	}
	// 每张卡占俩元素，第一个是花色，第二个是大小
	wpi.TableCardsArea = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(367, 946, 2, 2),   //第一张卡花色
		screen.Relative2AbsoluteRectangle(319, 864, 38, 52), //第一张卡rank
		screen.Relative2AbsoluteRectangle(460, 946, 2, 2),   //第一张卡花色
		screen.Relative2AbsoluteRectangle(413, 864, 37, 50), //第一张卡rank
		screen.Relative2AbsoluteRectangle(554, 933, 2, 2),   //第一张卡花色
		screen.Relative2AbsoluteRectangle(508, 864, 37, 50), //第一张卡rank
		screen.Relative2AbsoluteRectangle(650, 946, 2, 2),   //第一张卡花色
		screen.Relative2AbsoluteRectangle(601, 864, 37, 50), //第一张卡rank
		screen.Relative2AbsoluteRectangle(744, 946, 2, 2),   //第一张卡花色
		screen.Relative2AbsoluteRectangle(695, 864, 37, 50), //第一张卡rank
	}

	//wpi.PlayerCards = []CardImage{
	//	NewCardImage(screen),
	//}
	wpi.PotSizeSnapshots = make([]uint32, wpi.GetTableSize())

	wpi.PotArea = screen.Relative2AbsoluteRectangle(491, 602, 110, 28)
	// 用于确认是否到我下注
	wpi.GosseractClient = gosseract.NewClient()
	wpi.GosseractClient.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK_VERT_TEXT)
	wpi.GosseractClient.SetWhitelist("0123456789AJQK")
	//wpi.GosseractClient.SetVariable("oem", "0")
	wpi.ChineseGosseract = gosseract.NewClient()
	wpi.ChineseGosseract.SetLanguage("chi_sim", "eng")

	wpi.ChipGosseractClient = gosseract.NewClient()
	wpi.ChipGosseractClient.SetWhitelist("0123456789")
	// 整张图留档做debug
	wholeImg := util.CaptureScreenPart(image.Rectangle{Min: image.Point{X: screen.MinX, Y: screen.MinY}, Max: image.Point{X: screen.MaxX, Y: screen.MaxY}})
	path := fmt.Sprintf("./whole.png")
	util.SaveImage(wholeImg, path)
}
func (wpi *WePokerInput) GetBetSize() (betSize int64) {
	return WePokerBetSize
}
func (wpi *WePokerInput) GetTableSize() int {
	return WePokerTableSize
}

func (wpi *WePokerInput) GetDealer() int {
	dealerRGB := color.RGBA{
		R: 250, G: 250, B: 250,
	}
	for {
		// 把dealer位置块都识别出来，确认哪个是dealer
		for i := 0; i < len(wpi.ImagePos.Dealers); i++ {
			dealerImg := util.CaptureScreenPart(wpi.ImagePos.Dealers[i])
			path := fmt.Sprintf("./dealer%d.png", i)
			util.SaveImage(dealerImg, path)
			isDealer := util.ContainsColor(dealerImg, dealerRGB)
			if isDealer {
				return i
			}
		}
	}
}
func (wpi *WePokerInput) GetTurnCard() *poker.Card {
	return wpi.GetCardWithRepeat(wpi.TableCardsArea[6], wpi.TableCardsArea[7], "turn")

}
func (wpi *WePokerInput) GetRiverCard() *poker.Card {
	return wpi.GetCardWithRepeat(wpi.TableCardsArea[8], wpi.TableCardsArea[9], "river")
}

func (wpi *WePokerInput) GetFlopCards() (cards []*poker.Card) {
	return []*poker.Card{
		wpi.GetCardWithRepeat(wpi.TableCardsArea[0], wpi.TableCardsArea[1], "flop1"),
		wpi.GetCardWithRepeat(wpi.TableCardsArea[2], wpi.TableCardsArea[3], "flop2"),
		wpi.GetCardWithRepeat(wpi.TableCardsArea[4], wpi.TableCardsArea[5], "flop3"),
	}
}

func (wpi *WePokerInput) GetSelfCard() (cards [2]*poker.Card) {
	return [2]*poker.Card{
		wpi.GetCardWithRepeat(wpi.ImagePos.MyCards[0], wpi.ImagePos.MyCards[1], "mycard1"),
		wpi.GetCardWithRepeat(wpi.ImagePos.MyCards[2], wpi.ImagePos.MyCards[3], "mycard2"),
	}
}

func (wpi *WePokerInput) GetCardWithRepeat(suitRec, rankRec image.Rectangle, identifier string) *poker.Card {
	var suit, rank string
	for suit == "" {
		select {
		case <-wpi.ctx.Done():
			return nil
		default:
			suit = Color2Suit(suitRec, identifier+"suit")
		}
	}
	time.Sleep(150 * time.Millisecond)
	log.Println(identifier + suit)
	for rank == "" {
		rank, _ = util.RecognizeWePokerPokerArea(wpi.GosseractClient, rankRec, identifier+"rank")
	}
	log.Println(identifier + rank)

	return poker.NewCardFromRecognition(suit, rank)

}

func (wpi *WePokerInput) InitPlayerWithChips(players []poker.Player, tableSize int) []poker.Player {
	for i := 0; i < len(wpi.ImagePos.Chips); i++ {
		//name, _ := util.RecognizeArea(wpi.ChineseGosseract, wpi.NameAndActionArea[i], fmt.Sprintf("name%d", i))
		//log.Printf("name%d: %s\n", i, name)

		text, _ := util.RecognizeWePokerMyChipArea(wpi.ChipGosseractClient, wpi.ImagePos.Chips[i], fmt.Sprintf("chip%d", i))
		log.Printf("P%d ocr: %s  ", i, text)
		intNum, err := util.TrimNumber(text)
		// 说明没人
		if err != nil {
			intNum = 0
		}
		inTable := true
		if intNum == 0 {
			inTable = false
		}
		players = append(players, poker.Player{InTable: inTable, Chip: intNum})
		fmt.Printf("P%d has %d chips\n", i, intNum)
	}
	return players
}

// 每个轮次的下注
func (wpi *WePokerInput) Winner(table poker.Table) {
	fmt.Println("input who's winning")
	// TODO
	var num int
	fmt.Print("winner num: ")
	fmt.Scan(&num)
	//for i:=0;i<num;i++ {
	//	fmt.Scan()
	//}
}

func (wpi *WePokerInput) PlayerBetting(table *poker.Table, playerPos int) poker.Action {
	text, _ := util.RecognizeWePokerChipArea(wpi.ChipGosseractClient, wpi.ImagePos.Betting[playerPos], fmt.Sprintf("bet%d", playerPos))
	log.Printf("P%d ocr: %s\n", playerPos, text)
	intNum, _ := util.ParseNumber(text)
	var ac string
	// 需要关注大盲位翻前的情况，当playerBet=betSize=intNum
	if intNum < table.Round.CallBetSize {
		ac = poker.Fold
		intNum = 0
	} else if (intNum == 0 && table.Round.CallBetSize == 0) || (table.Round.PlayerBet[playerPos] == intNum) {
		ac = poker.Check
	} else if intNum == table.Round.CallBetSize {
		ac = poker.Call
	} else if intNum > table.Round.CallBetSize {
		ac = poker.Raise
	}

	return poker.Action{Player: table.GetCurPlayerName(), Action: ac, Chip: intNum}
}

func (wpi *WePokerInput) RestartRound(table *poker.Table) bool {
	cards := wpi.GetSelfCard()
	// 卡片与上次不同就换局了
	if cards[0].Equals(table.MyCard[0]) && cards[1].Equals(table.MyCard[1]) {
		return false
	}
	return true
}

func Color2Suit(rectangle image.Rectangle, fileName string) string {
	img := util.CaptureScreenPart(rectangle)
	if fileName != "" {
		path := fmt.Sprintf("./%s.png", fileName)
		util.SaveImage(img, path)
	}

	// 不是绿框，继续hold and recognize
	r, g, b, _ := img.At(0, 0).RGBA()
	colorAtPoint := color.RGBA{uint8(r / 257), uint8(g / 257), uint8(b / 257), 0}
	//fmt.Println(colorAtPoint)
	hearts := color.RGBA{226, 15, 32, 0}
	clubs := color.RGBA{28, 148, 37, 0}
	spades := color.RGBA{0, 0, 0, 0}
	diamonds := color.RGBA{13, 92, 187, 0}
	if compareRGB(colorAtPoint, hearts) {
		return poker.Hearts
	} else if compareRGB(colorAtPoint, clubs) {
		return poker.Clubs
	} else if compareRGB(colorAtPoint, spades) {
		return poker.Spades
	} else if compareRGB(colorAtPoint, diamonds) {
		return poker.Diamonds
	}
	return ""
}

func compareRGB(c1, c2 color.RGBA) bool {
	return c1.R == c2.R && c1.G == c2.G && c1.B == c2.B
}

// 查看哪个人正在做动作
func (wpi *WePokerInput) TraceWhichIsActive(table *poker.Table) int {
	for {
		select {
		case <-wpi.ctx.Done():
			return -1
		default:
			for i := 0; i < len(table.Players); i++ {
				// 在桌上，我们就给查一下他是否在做动作
				if table.Players[i].InTable {
					actionRectangle := image.Rectangle{Min: wpi.ActionPoint[i], Max: image.Point{wpi.ActionPoint[i].X + 12, wpi.ActionPoint[i].Y + 12}}
					// 识别到玩家的周围有绿框 && 底池数变了时
					actionImg := util.CaptureScreenPart(actionRectangle)
					util.SaveImage(actionImg, fmt.Sprintf("./action%d.png", i))
					// 是绿框, 开始等待其变为原样
					if util.CheckPointIsGreen(actionImg) {
						// 循环直到绿框消失，说明这个玩家已经做出了选择
						for {
							select {
							case <-wpi.ctx.Done():
								return -1
							default:
								actionImg = util.CaptureScreenPart(actionRectangle)
								time.Sleep(time.Millisecond * 350)
								if !util.CheckPointIsGreen(actionImg) {
									return i
								}
							}
						}
					}
				}
			}

		}
	}
}
func genGptPrompt(table *poker.Table) {
	gpt := gpt.NewGpt()
	prompt := gpt.GetPrompt(table)
	log.Println(prompt)
	err := clipboard.WriteAll(prompt)
	if err != nil {
		fmt.Println("Failed to copy to clipboard:", err)
		return
	}
}

func (wpi *WePokerInput) Betting(table *poker.Table) {
	fmt.Printf("%s started\n", table.Round.WhichRound)
	// 一直betting直到（自己需要行动/结束）才停
	// img_recog 时应该只在"我"要做行动时才开始做识别
	// 等待到我要行动时再开始做动作 注：此处会hang住，有循环
	//input.WaitBettingTurn()
	if table.Round.CurActionPos == 0 {
		genGptPrompt(table)
	}
	firstTime := true
	for {
		activePlayerPos := wpi.TraceWhichIsActive(table)
		log.Printf("active: %d\n", activePlayerPos)
		// 可以一直执行到正在举行动作的人
		for {
			select {
			case <-wpi.ctx.Done():
				return
			default:
				action := wpi.PlayerBetting(table, table.Round.CurActionPos)
				fmt.Println(action)
				if firstTime {
					action.Immune2Chip = true
				}
				action.Execute(table)
				// 到最后一家了，结束本轮
				if table.Round.CurActionPos == table.Round.FinalPosition {
					table.ShowSituation()
					fmt.Printf("%s finished\n", table.Round.WhichRound)
					return
				}
				table.Round.CurActionPos = table.NextValidPosition(table.Round.CurActionPos)
				// 到自己的时候就该gpt出马了
				if table.Round.CurActionPos == 0 {
					genGptPrompt(table)
				}
			}
			if table.Round.CurActionPos == table.NextValidPosition(activePlayerPos) {
				break
			}
		}
		firstTime = false
	}

}

func (wpi *WePokerInput) CheckExit() bool {
	select {
	case <-wpi.ctx.Done():
		fmt.Println("exit")
		return true
	default:
		return false
	}
}
