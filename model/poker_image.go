package model

import "image"

// 存下图像
type Image struct {
	Chips        []*image.RGBA
	Dealer       []image.RGBA
	MyCards      []image.RGBA
	TableCards   []image.RGBA
	Betting      []image.RGBA // 每个人的下注
	MyActionArea image.RGBA   // 用于确认是否到我下注
}

type ImagePos struct {
	Chips        []image.Rectangle
	Dealers      []image.Rectangle
	MyCards      []image.Rectangle
	TableCards   []image.RGBA
	Betting      []image.Rectangle
	MyActionArea image.Rectangle // 用于确认是否到我下注
}
