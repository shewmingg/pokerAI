package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pokerAI/poker"
	"strings"
)

type Gpt struct {
	History []string
}

func NewGpt() Gpt {
	return Gpt{
		History: []string{systemMessage},
	}
}

const apiKey = "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"
const modelId = "gpt-3.5-turbo"
const maxTokens = 60
const systemMessage = "You are a texas No-limit hold 'em probability helper Bot. You need to evaluate situations and instruct my betting. If raise is suggested, name the appropriate raising chips. Only suggested action is needed, no explanation."

type requestData struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseData struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// func (g *Gpt) GetEvaluationPrompt
func (g *Gpt) GetPrompt(t *poker.Table) (str string) {
	// table situation
	background := "You are an domain expect in texas No-limit hold 'em. I am an exploitative GTO player. " +
		"You need to evaluate situations and instruct my betting." +
		"If raise is suggested, name the appropriate raising chips.I have no knowledge of other players."
	str = background + fmt.Sprintf("there's %d seats on the table. I'm P0, starting from my seat and going clockwise, seats are named P1, P2, P3, P4, P5, P6, P7, P8 respectively. ",
		t.TableSize)
	// chip situation
	str += fmt.Sprintf("Current active players' chip situation as follows: ")
	for i := 0; i < len(t.Players); i++ {
		if t.Players[i].InTable {
			str += fmt.Sprintf("%s has %d chips. ", t.GetPlayerName(i), t.Players[i].Chip)
		}
	}
	// BetSize situation
	str += fmt.Sprintf("bet size is %d with no anti.", t.BetSize)

	// dealer situation
	str += fmt.Sprintf("This round %s is the dealer. ", t.GetPlayerName(t.Dealer))

	rounds := []string{poker.Preflop, poker.Flop, poker.Turn, poker.River}
	for _, v := range rounds {
		str += fmt.Sprintf("%s betting as follows: ", v)
		actions := t.Actions[v]
		for i := 0; i < len(actions); i++ {
			str += fmt.Sprintf("%s, ", actions[i])
		}
		if v == t.Round.WhichRound {
			break
		}
	}

	//// pot size
	str += fmt.Sprintf("Current pot size is %d chips. ", t.Pot)
	switch t.Round.WhichRound {
	case poker.Preflop:
		break
	case poker.River:
		str += fmt.Sprintf("river card is %s. ", t.TableCard[4])
		fallthrough
	case poker.Turn:
		str += fmt.Sprintf("turn card is %s. ", t.TableCard[3])
		fallthrough
	case poker.Flop:
		str += fmt.Sprintf("flop cards are: %s %s %s. ", t.TableCard[0], t.TableCard[1], t.TableCard[2])
	}
	// my card situation
	str += fmt.Sprintf("I hold %s and %s.", t.MyCard[0], t.MyCard[1])
	str += fmt.Sprintf("Now, it's back to me (P0) to act again. What should I do?")
	return
}

func (g *Gpt) GenerateResponse(promptText string) (string, error) {
	// Construct prompt by joining conversation history with newline characters
	requestBody := requestData{
		Model:    modelId,
		Messages: encodeAsMessage(promptText),
		//Temperature: 0.3,
		MaxTokens: maxTokens,
	}

	// Marshal request body to JSON
	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Send POST request to OpenAI GPT API
	url := "https://api.openai.com/v1/chat/completions"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyJson))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	log.Printf("send request:%v\n", request)
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	log.Printf("response:%v\n", response)
	// Check for HTTP error status codes
	if response.StatusCode != 200 {
		var errorResponse struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(response.Body).Decode(&errorResponse)
		if err != nil {
			return "", fmt.Errorf("API error: status %d", response.StatusCode)
		}
		return "", fmt.Errorf("API error: %s", errorResponse.Error)
	}

	// Parse response JSON
	var responseBody responseData
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return "", err
	}

	// Extract response text from API response
	responseText := responseBody.Choices[0].Message.Content
	responseText = strings.Trim(responseText, "\n")

	return responseText, nil
}

func encodeAsMessage(prompt string) []Message {
	message := Message{Role: "user", Content: prompt}
	res := []Message{message}
	return res
}
