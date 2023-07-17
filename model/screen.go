package model

import "image"

type Screen struct {
	MinX int
	MinY int
	MaxX int
	MaxY int
}

func (s Screen) Relative2Absolute(x, y int) image.Point {
	// 由于由外部生成的x，y，和实际screen的x，y是二倍关系（retina display），所以这里x y需要/2
	return image.Point{X: s.MinX + x/2, Y: s.MinY + 44 + y/2}
}
func (s Screen) Relative2AbsolutePoint(p image.Point) image.Point {
	return image.Point{
		X: s.MinX + 24 + p.X/2, Y: s.MinY + 24 + p.Y/2,
	}
}
func (s Screen) Relative2AbsoluteRectangle(x, y int, width, height int) image.Rectangle {
	// 由于由外部生成的x，y，和实际screen的x，y是二倍关系（retina display），所以这里x y需要/2
	return image.Rectangle{Min: image.Point{X: s.MinX + x/2, Y: s.MinY + 24 + y/2},
		Max: image.Point{X: s.MinX + (x+width)/2, Y: s.MinY + 24 + (y+height)/2}}
}
func StartAndWidthHeight2Rectangle(x, y int, width, height int) image.Rectangle {
	return image.Rectangle{Min: image.Point{X: x, Y: y}, Max: image.Point{X: x + width, Y: height + y}}
}
