package util

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
)

func TrimNumber(numStr string) (int64, error) {
	digitStr := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, numStr)
	return strconv.ParseInt(digitStr, 10, 64)
}
func ParseNumber(numStr string) (int64, error) {
	// delete useless symbols
	numStr = strings.Replace(numStr, "$", "", -1) // replace dollar sign
	numStr = strings.Replace(numStr, ",", "", -1) // replace comma
	numStr = strings.Replace(numStr, ".", "", -1) // replace comma
	numStr = strings.Replace(numStr, "_", "", -1) // replace comma
	numStr = strings.Replace(numStr, "+", "", -1) // replace comma
	numStr = strings.Replace(numStr, "¢", "", -1) // replace comma
	numStr = strings.Replace(numStr, " ", "", -1) // replace comma

	numStr = strings.ToLower(numStr)
	multiplier := int64(1)

	if strings.HasSuffix(numStr, "k") {
		multiplier = 1_000
		numStr = strings.TrimSuffix(numStr, "k")
	} else if strings.HasSuffix(numStr, "m") {
		multiplier = 1_000_000
		numStr = strings.TrimSuffix(numStr, "m")
	}

	base, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, nil
	}

	return int64(base * float64(multiplier)), nil
}
func SetChromeWindowSizeAndActivate(left, top, right, bottom int) error {
	cmd := exec.Command("osascript",
		"-e", `tell application "Google Chrome" to activate`,
		"-e", `tell application "Google Chrome" to set bounds of front window to {`+
			fmt.Sprintf("%d, %d, %d, %d", left, top, right, bottom)+
			`}`)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("osascript error: %v", stderr.String())
	}
	return err
}
