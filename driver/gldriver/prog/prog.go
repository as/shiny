package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	"github.com/as/shiny/screen"
	"github.com/as/ui"
	"github.com/as/shiny/event/paint"
)

func main() {
	dev, err := ui.Init(&screen.NewWindowOptions{Width: 1024, Height: 768})
	if err != nil {
		panic(err)
	}
	win := dev.Window()
	D := screen.Dev
	buf, _ := dev.NewBuffer(image.Pt(512, 512))
	red := image.NewUniform(color.RGBA{255, 0, 0, 255})
	blue := image.NewUniform(color.RGBA{0, 0, 255, 255})
	draw.Draw(buf.RGBA(), buf.RGBA().Bounds(), blue, image.ZP, draw.Src)
	for {
		select {
		case m := <-D.Mouse:
			r := image.ZR.Inset(-4).Add(image.Pt(int(m.X), int(m.Y)))
			draw.Draw(buf.RGBA(), r, red, image.ZP, draw.Src)
			select {
			case D.Paint <- paint.Event{}:
				log.Println("painted")
			default:
				log.Println("miss")
			}
		case <-D.Key:
			log.Println("key")
		case <-D.Lifecycle:
			log.Println("life")
		case <-D.Paint:
			log.Println("paint")
			win.Upload(image.ZP, buf, buf.Bounds())
			win.Publish()
		case <-D.Size:
			log.Println("size")
			buf, _ = dev.NewBuffer(image.Pt(512, 512))
			draw.Draw(buf.RGBA(), buf.RGBA().Bounds(), blue, image.ZP, draw.Src)

		}
	}

}
