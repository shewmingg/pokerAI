package poker

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Card struct {
	Value int
	Suit  string
}

func NewCardFromRecognition(suitStr string, value string) *Card {
	//suitStr = strings.ToLower(suitStr)
	//var suit string
	//if strings.ContainsRune(suitStr, '+') {
	//	suit = Diamonds
	//} else if strings.ContainsRune(suitStr, '&') {
	//	suit = Clubs
	//} else if strings.ContainsRune(suitStr, 'v') {
	//	suit = Hearts
	//} else if strings.ContainsRune(suitStr, 'a') {
	//	suit = Spades
	//}
	var val int
	value = strings.ToLower(value)
	switch value {
	case "j":
		val = 11
	case "q":
		val = 12
	case "k":
		val = 13
	case "a":
		val = 1
	case "t":
		fallthrough
	case "in":
		val = 10
	default:
		var err error
		val, err = strconv.Atoi(value)
		if err != nil {
			log.Fatal(err)
		}
	}
	return &Card{Suit: suitStr, Value: val}
}

// 解析string，生成card
// 支持格式 s4（spades 4） d13（diamonds K）
func NewCard(s string) *Card {
	if len(s) > 3 {
		log.Fatal("invalid card format")
	}
	s = strings.ToLower(s)
	var suit string
	switch s[0] {
	case 's':
		fallthrough
	case 'a':
		suit = Spades
	case 'd':
		fallthrough
	case '+':
		suit = Diamonds
	case 'h':
		fallthrough
	case 'j':
		fallthrough
	case 'v':
		suit = Hearts
	case 'c':
		fallthrough
	case '&':
		suit = Clubs
	}
	var val int
	switch s[1:] {
	case "j":
		val = 11
	case "q":
		val = 12
	case "k":
		val = 13
	case "a":
		val = 1
	default:
		var err error
		val, err = strconv.Atoi(s[1:])
		if err != nil {
			log.Fatal(err)
		}
	}
	return &Card{Suit: suit, Value: val}
}

type Deck struct {
	Ptr   int // 当前发到哪张牌
	Cards []Card
}

func SuitFromString2Figure(suit string) rune {
	switch suit {
	case Diamonds:
		return ShowDiamonds
	case Clubs:
		return ShowClubs
	case Hearts:
		return ShowHearts
	case Spades:
		return ShowSpades
	}
	return '\U0001F0A0'
}

const Diamonds = "Diamonds"
const Clubs = "Clubs"
const Hearts = "Hearts"
const Spades = "Spades"
const ShowClubs = '\u2663'
const ShowDiamonds = '\u2666'
const ShowHearts = '\u2665'
const ShowSpades = '\u2660'

func InputCard(suit string, value int) {

}

func NewDeck() (deck Deck) {
	suits := []string{Diamonds, Clubs, Hearts, Spades}
	for _, suit := range suits {
		for value := 2; value <= 14; value++ {
			deck.Cards = append(deck.Cards, Card{Suit: suit, Value: value})
		}
	}
	return
}
func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range d.Cards {
		j := r.Intn(i + 1)
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	}
	d.Ptr = 0
}

// 发牌
func (d *Deck) Deal() (card Card) {
	card = d.Cards[d.Ptr]
	d.Ptr += 1
	return
}

func (c *Card) String() string {
	str := string(SuitFromString2Figure(c.Suit))
	if c.Value == 0 {
		return str
	} else {
		var value string
		switch c.Value {
		case 13:
			value = "K"
		case 12:
			value = "Q"
		case 11:
			value = "J"
		case 1:
			value = "A"
		default:
			value = strconv.Itoa(c.Value)
		}
		return str + value
	}
}

func (c *Card) Equals(other *Card) bool {
	if c == nil || other == nil {
		return false
	}
	if c.Value == other.Value && c.Suit == other.Suit {
		return true
	}
	return false
}
