package util

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/kbinani/screenshot"
	"github.com/otiai10/gosseract/v2"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

const WhichScreen = 0

func ImageToPNGBytes(img image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func CaptureScreenPart(rect image.Rectangle) *image.RGBA {
	bounds := screenshot.GetDisplayBounds(WhichScreen)
	if !rect.In(bounds) {
		log.Fatalf("rectangle is out of screen bounds: rectangle=%v bounds=%v", rect, bounds)
	}
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		panic(err)
	}
	return img
}

func ConvertImage(img *image.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// As the RGBA method returns color values in the range [0, 65535], we need to compare our threshold (83) after multiplying it with 257 (65535/255 = 257) to make it comparable.
			if r < 215*257 {
				r, g, b = 0, 0, 0
			}
			invertedColor := color.RGBA{
				R: uint8(^r >> 8),
				G: uint8(^g >> 8),
				B: uint8(^b >> 8),
				A: 255, // Set alpha to opaque (255)
			}
			img.Set(x, y, invertedColor)

		}
	}
}

func CheckPointIsGreen(img *image.RGBA) bool {
	_, g, _, _ := img.At(0, 0).RGBA()
	greenThreshold := uint32(180 << 8)
	if g >= greenThreshold {
		return true
	}
	return false

}

// 将所有置灰
func ConvertWePokerPoker(img *image.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// 我们将所有非白的置黑
			if r < 0xffff || g < 0xffff || b < 0xffff {
				minShade := uint8(min(min(r, g), b) >> 8)
				if minShade < 50 {
					minShade = 0
				} else if minShade > 246 {
					minShade = 255
				}
				img.Set(x, y, color.RGBA{R: minShade, G: minShade, B: minShade, A: 255})
			} else {
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // a >> 8 to scale alpha to 0-255 range
			}
		}
	}
}
func min(a, b uint32) uint32 {
	if a > b {
		return b
	}
	return a
}

func ConvertWePokerChip(img *image.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// As the RGBA method returns color values in the range [0, 65535], we need to compare our threshold (83) after multiplying it with 257 (65535/255 = 257) to make it comparable.
			if r <= 47*257 {
				r, g, b = 0, 0, 0
			}
			invertedColor := color.RGBA{
				R: uint8(^r >> 8),
				G: uint8(^g >> 8),
				B: uint8(^b >> 8),
				A: 255, // Set alpha to opaque (255)
			}
			img.Set(x, y, invertedColor)
		}
	}
}

func ConvertWePokerMyChip(img *image.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// As the RGBA method returns color values in the range [0, 65535], we need to compare our threshold (83) after multiplying it with 257 (65535/255 = 257) to make it comparable.
			if r <= 88*257 {
				r, g, b = 0, 0, 0
			}
			invertedColor := color.RGBA{
				R: uint8(^r >> 8),
				G: uint8(^g >> 8),
				B: uint8(^b >> 8),
				A: 255, // Set alpha to opaque (255)
			}
			img.Set(x, y, invertedColor)
		}
	}
}

func ContainsColor(img *image.RGBA, target color.RGBA) bool {
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Since the RGBA() function returns a color's components
			// in the range [0, 65535], we need to convert the target
			// color's components to that range by multiplying them by 257.
			if r == uint32(target.R)*257 && g == uint32(target.G)*257 && b == uint32(target.B)*257 {
				return true
			}
		}
	}

	return false
}

func SaveImage(img image.Image, path string) {
	file, _ := os.Create(path)
	defer file.Close()
	err := png.Encode(file, img)
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}

func RecognizeWePokerMyChipArea(client *gosseract.Client, area image.Rectangle, fileName string) (string, error) {
	betImg := CaptureScreenPart(area)
	ConvertWePokerMyChip(betImg)
	// Convert to image.Image to use with the imaging package
	betImg = PadImage(betImg, color.RGBA{255, 255, 255, 255})
	processed := image.Image(betImg)

	////
	// Resize srcImage to width = 800px preserving the aspect ratio.
	processed = imaging.Resize(processed, 800, 0, imaging.Lanczos)

	if fileName != "" {
		path := fmt.Sprintf("./%s.png", fileName)
		SaveImage(processed, path)
	}
	// Create a buffer to store the image data
	buf := new(bytes.Buffer)

	// Encode the image to the buffer as PNG
	err := png.Encode(buf, processed)

	err = client.SetImageFromBytes(buf.Bytes())
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	text, err := client.Text()
	if err != nil {
		log.Fatal(err)
	}
	return text, err
}

func RecognizeWePokerChipArea(client *gosseract.Client, area image.Rectangle, fileName string) (string, error) {
	betImg := CaptureScreenPart(area)
	// Create a buffer to store the image data
	buf := new(bytes.Buffer)
	// direct

	ConvertWePokerChip(betImg)

	// Convert to image.Image to use with the imaging package
	betImg = PadImage(betImg, color.RGBA{255, 255, 255, 255})
	processed := image.Image(betImg)

	////
	// Resize srcImage to width = 800px preserving the aspect ratio.
	processed = imaging.Resize(processed, 800, 0, imaging.Lanczos)

	if fileName != "" {
		path := fmt.Sprintf("./%s.png", fileName)
		SaveImage(processed, path)
	}

	// Encode the image to the buffer as PNG
	err := png.Encode(buf, processed)

	err = client.SetImageFromBytes(buf.Bytes())
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	text, err := client.Text()
	if err != nil {
		log.Fatal(err)
	}
	return text, err
}

// PadImage adds a 2-pixel padding around the input image and returns a new image.
func PadImage(img *image.RGBA, padColor color.RGBA) *image.RGBA {
	// Get the dimensions of the input image
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Create a new image of the necessary size, filled with the padding color
	newImg := image.NewRGBA(image.Rect(0, 0, width+4, height+4))
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{padColor}, image.Point{}, draw.Src)

	// Draw the input image onto the new image, offset by the padding size
	draw.Draw(newImg, image.Rect(2, 2, width+2, height+2), img, bounds.Min, draw.Src)

	// Return the new image
	return newImg
}

func RecognizeWePokerPokerArea(client *gosseract.Client, area image.Rectangle, fileName string) (string, error) {
	betImg := CaptureScreenPart(area)
	ConvertWePokerPoker(betImg)
	processed := PadImage(betImg, color.RGBA{255, 255, 255, 255})

	//// Convert to image.Image to use with the imaging package
	//srcImage := image.Image(betImg)
	//
	//// Resize srcImage to width = 800px preserving the aspect ratio.
	//processed := imaging.Resize(srcImage, 800, 0, imaging.Lanczos)

	if fileName != "" {
		path := fmt.Sprintf("./%s.png", fileName)
		SaveImage(processed, path)
	}
	// Create a buffer to store the image data
	buf := new(bytes.Buffer)

	// Encode the image to the buffer as PNG
	err := png.Encode(buf, processed)

	err = client.SetImageFromBytes(buf.Bytes())
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	text, err := client.Text()
	if err != nil {
		log.Fatal(err)
	}
	return text, err
}

func RecognizeArea(client *gosseract.Client, area image.Rectangle, fileName string) (string, error) {
	betImg := CaptureScreenPart(area)
	if fileName != "" {
		path := fmt.Sprintf("./%s.png", fileName)
		SaveImage(betImg, path)
	}

	pngBytes, err := ImageToPNGBytes(betImg)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	err = client.SetImageFromBytes(pngBytes)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	text, err := client.Text()
	if err != nil {
		log.Fatal(err)
	}
	return text, err
}

func HashRGBA(rgba *image.RGBA) uint32 {
	// Create the FNV-1a hash object
	hash := fnv.New32a()

	// Convert the RGBA image to a byte slice
	byteSlice := rgbaToByteSlice(rgba)

	// Write the byte slice to the hash object
	hash.Write(byteSlice)

	// Get the hash value as an integer
	hashValue := hash.Sum32()

	return hashValue
}

func rgbaToByteSlice(rgba *image.RGBA) []byte {
	// Calculate the total number of bytes in the image
	numBytes := 4 * rgba.Rect.Dx() * rgba.Rect.Dy()

	// Create a byte slice with the required capacity
	byteSlice := make([]byte, numBytes)

	// Copy the pixel data to the byte slice
	copy(byteSlice, rgba.Pix)

	return byteSlice
}
