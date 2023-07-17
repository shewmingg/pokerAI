package model

import (
	"github.com/otiai10/gosseract/v2"
	"image"
	"log"
	"pokerAI/util"
)

func Recognize(client *gosseract.Client, img *image.RGBA) string {
	pngBytes, err := util.ImageToPNGBytes(img)
	if err != nil {
		log.Fatal(err)
	}
	err = client.SetImageFromBytes(pngBytes)
	if err != nil {
		log.Fatal(err)
	}
	text, err := client.Text()
	if err != nil {
		log.Fatal(err)
	}
	return text
}
