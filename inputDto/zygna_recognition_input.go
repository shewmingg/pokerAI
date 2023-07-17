package inputDto

import (
	"fmt"
	"github.com/kbinani/screenshot"
	"github.com/otiai10/gosseract/v2"
	"image"
	"image/color"
	"log"
	"pokerAI/model"
	"pokerAI/poker"
	"pokerAI/util"
	"strconv"
)

const GameWidth = 950
const GameHeight = 800
const DealerR = 203
const DealerG = 178
const DealerB = 187

type ZygnaRecognitionInput struct {
	model.Screen    // 需要图像相关
	model.ImagePos  // 各个需要识别的位置
	GosseractClient *gosseract.Client
	ManualInput
}

// TODO 方案比对
// 要把整个页面拷下来再去抠图做识别 优势：无需多次调用抠图
// 一小块一小块的抠 优势：内存耗费小 先用这个吧

// 如果用图像识别做input，需要将浏览器调起，这里用chrome打开game
func (img *ZygnaRecognitionInput) Init() {
	bounds := screenshot.GetDisplayBounds(0)
	minX, minY, maxX, maxY := bounds.Dx()-GameWidth, 0, bounds.Dx(), GameHeight
	err := util.SetChromeWindowSizeAndActivate(minX, minY, maxX, maxY)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("screen size: %d %d %d %d", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
	// 浏览器设置位置及大小
	screen := model.Screen{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
	// hardcode 各个要识别的位置
	// 各个筹码区域
	img.ImagePos.Chips = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(816, 1427, 297, 57),
		screen.Relative2AbsoluteRectangle(318, 1226, 228, 42),
		screen.Relative2AbsoluteRectangle(444, 764, 228, 42),
		screen.Relative2AbsoluteRectangle(1245, 764, 228, 42),
		screen.Relative2AbsoluteRectangle(1362, 1226, 228, 42),
	}
	// 给出5个dealer的位置，通过识别在指定位置是否有D确认谁是dealer
	// 由于dealer是直接在屏幕上抓取的，所以取绝对值
	img.ImagePos.Dealers = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(825, 1038, 45, 45),
		screen.Relative2AbsoluteRectangle(639, 1029, 48, 45),
		screen.Relative2AbsoluteRectangle(675, 756, 48, 48),
		screen.Relative2AbsoluteRectangle(1191, 756, 42, 45),
		screen.Relative2AbsoluteRectangle(1236, 1026, 48, 48),
	}
	// 每张卡占俩元素，第一个是花色，第二个是大小
	img.ImagePos.MyCards = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(837, 1347, 57, 54), //第一张卡花色
		screen.Relative2AbsoluteRectangle(822, 1290, 66, 60), //第一张卡rank
		screen.Relative2AbsoluteRectangle(948, 1329, 57, 54), // 第二张卡花色
		screen.Relative2AbsoluteRectangle(960, 1275, 66, 60), // 第二张卡rank
	}
	// 每个玩家出的筹码
	img.ImagePos.Betting = []image.Rectangle{
		screen.Relative2AbsoluteRectangle(867, 1008, 192, 63), // 我的下注
		screen.Relative2AbsoluteRectangle(432, 945, 183, 45),  // P1下注
		screen.Relative2AbsoluteRectangle(456, 822, 183, 45),  // P2下注
		screen.Relative2AbsoluteRectangle(1275, 824, 183, 45), // P3下注
		screen.Relative2AbsoluteRectangle(1305, 945, 183, 45), // P4下注
	}
	// 用于确认是否到我下注
	img.ImagePos.MyActionArea = screen.Relative2AbsoluteRectangle(1758, 1374, 87, 45) // 区域 Raise or Bet
	img.GosseractClient = gosseract.NewClient()
	// 整张图留档做debug
	wholeImg := util.CaptureScreenPart(image.Rectangle{Min: image.Point{X: screen.MinX, Y: screen.MinY}, Max: image.Point{X: screen.MaxX, Y: screen.MaxY}})
	path := fmt.Sprintf("./whole.png")
	util.SaveImage(wholeImg, path)

}

func (img *ZygnaRecognitionInput) InitPlayerWithChips(players []poker.Player, tableSize int) []poker.Player {
	for i := 0; i < len(img.ImagePos.Chips); i++ {
		chipImg := util.CaptureScreenPart(img.ImagePos.Chips[i])
		path := fmt.Sprintf("./chip%dbefore.png", i)
		util.SaveImage(chipImg, path)
		util.ConvertImage(chipImg)

		path = fmt.Sprintf("./chip%d.png", i)
		util.SaveImage(chipImg, path)

		pngBytes, err := util.ImageToPNGBytes(chipImg)
		if err != nil {
			log.Fatal(err)
		}
		err = img.GosseractClient.SetImageFromBytes(pngBytes)
		if err != nil {
			log.Fatal(err)
		}
		text, err := img.GosseractClient.Text()
		fmt.Printf("P%d ocr: %s  ", i, text)
		if err != nil {
			log.Fatal(err)
		}
		intNum, err := util.ParseNumber(text)
		if err != nil {
			log.Fatal(err)
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

func (img *ZygnaRecognitionInput) GetBetSize() int64 {
	return 10_000
}

// dealer 太不规则了，通过比较RGB色值确定是否包含D
func (img *ZygnaRecognitionInput) GetDealer() int {
	dealerRGB := color.RGBA{
		R: DealerR, G: DealerG, B: DealerB,
	}
	// 把dealer位置块都识别出来，确认哪个是dealer
	for i := 0; i < len(img.ImagePos.Dealers); i++ {
		dealerImg := util.CaptureScreenPart(img.ImagePos.Dealers[i])
		path := fmt.Sprintf("./dealer%d.png", i)
		util.SaveImage(dealerImg, path)
		isDealer := util.ContainsColor(dealerImg, dealerRGB)
		if isDealer {
			fmt.Printf("dealer is %d\n", i)
			return i
		}
	}
	log.Fatal("unable to find dealer")
	return 0
}

func (img *ZygnaRecognitionInput) GetFlopCards() []*poker.Card {
	return img.ManualInput.GetFlopCards()
}

func (img *ZygnaRecognitionInput) GetTurnCard() *poker.Card {
	//TODO implement me
	return img.ManualInput.GetTurnCard()
}

func (img *ZygnaRecognitionInput) GetRiverCard() *poker.Card {
	//TODO implement me
	return img.ManualInput.GetRiverCard()
}

// 用ocr识别: ocr 识别不了花色，需要有其他方式来做，先用个笨方法吧：
func (img *ZygnaRecognitionInput) GetSelfCard() [2]*poker.Card {
	return img.ManualInput.GetSelfCard()
	tmpResult := [4]string{}
	for i := 0; i < 4; i++ {
		cardImg := util.CaptureScreenPart(img.ImagePos.MyCards[i])
		path := fmt.Sprintf("./card%d.png", i)
		util.SaveImage(cardImg, path)

		pngBytes, err := util.ImageToPNGBytes(cardImg)
		if err != nil {
			log.Fatal(err)
		}
		err = img.GosseractClient.SetImageFromBytes(pngBytes)
		if err != nil {
			log.Fatal(err)
		}
		text, err := img.GosseractClient.Text()
		tmpResult[i] = text
		fmt.Println(text)
		if err != nil {
			log.Fatal(err)
		}
	}
	value1, _ := strconv.Atoi(tmpResult[1])
	value2, _ := strconv.Atoi(tmpResult[3])
	res := [2]*poker.Card{{Suit: tmpResult[0], Value: value1}, {Suit: tmpResult[2], Value: value2}}

	return res
}

func (img *ZygnaRecognitionInput) Betting(table poker.Table) {
	//TODO implement me
	panic("implement me")
}

func (img *ZygnaRecognitionInput) Winner(table poker.Table) {
	//TODO implement me
	panic("implement me")
}

func (img *ZygnaRecognitionInput) GetTableSize() int {
	//TODO implement me
	return 5
}

// 玩家下注
func (img *ZygnaRecognitionInput) PlayerBetting(table *poker.Table, playerPos int) poker.Action {
	betImg := util.CaptureScreenPart(img.ImagePos.Betting[playerPos])
	path := fmt.Sprintf("./bet%d.png", playerPos)
	util.SaveImage(betImg, path)

	pngBytes, err := util.ImageToPNGBytes(betImg)
	if err != nil {
		log.Fatal(err)
	}
	err = img.GosseractClient.SetImageFromBytes(pngBytes)
	if err != nil {
		log.Fatal(err)
	}
	text, err := img.GosseractClient.Text()
	fmt.Printf("Player%d betting: %s\n", playerPos, text)
	if err != nil {
		log.Fatal(err)
	}
	intNum, _ := util.ParseNumber(text)
	// 如果是翻前，需要识别到哪俩是大小盲
	if table.Round.WhichRound == poker.Preflop {

	}
	var ac string
	if intNum == table.Round.CallBetSize {
		ac = poker.Call
	} else if intNum == 0 {
		ac = poker.Check
	} else if intNum > table.Round.CallBetSize {
		ac = poker.Raise
	}

	return poker.Action{Player: table.GetCurPlayerName(), Action: ac, Chip: intNum}
}

// 是否开始betting
func (img *ZygnaRecognitionInput) WaitBettingTurn() bool {
	// 直到识别到Me有Bet或Raise时才开始
	actionImg := util.CaptureScreenPart(img.ImagePos.MyActionArea)
	path := fmt.Sprintf("./myaction.png")
	util.SaveImage(actionImg, path)
	pngBytes, err := util.ImageToPNGBytes(actionImg)
	if err != nil {
		log.Fatal(err)
	}
	err = img.GosseractClient.SetImageFromBytes(pngBytes)
	if err != nil {
		log.Fatal(err)
	}
	text, err := img.GosseractClient.Text()
	fmt.Printf("action area: %s \n", text)
	if err != nil {
		log.Fatal(err)
	}
	if text == "Raise" || text == "Bet" {
		return true
	}

	return false
}

func (img *ZygnaRecognitionInput) RestartRound(table *poker.Table) bool {
	newDealer := img.GetDealer()
	// 发现dealer变化就意味着换局了
	if newDealer != table.Dealer {
		return true
	}
	return false
}
