package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/rand"
	"mc2d/chunks"
	"mc2d/tiles"
	"mc2d/ui"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/imdraw"
	"github.com/gopxl/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

type Camera struct {
	location     pixel.Vec
	zoomModifier float64
}

func newCamera(location pixel.Vec, zoomModifier float64) *Camera {
	return &Camera{
		location:     location,
		zoomModifier: zoomModifier,
	}
}

type Player struct {
	id       uint32
	name     string
	location pixel.Vec
	velocity pixel.Vec
}

func newPlayer(name string, location pixel.Vec) *Player {
	id := rand.Uint32() * 4294967295
	return &Player{
		id:       id,
		name:     name,
		location: location,
		velocity: pixel.Vec{X: 0, Y: 0},
	}
}

func main() {
	pixelgl.Run(run)
}

func run() {
	icon, err := loadPicture("icon.png")
	if err != nil {
		panic(err)
	}
	icons := []pixel.Picture{
		icon,
	}
	cfg := pixelgl.WindowConfig{
		Title:     "Minecraft 2D",
		Bounds:    pixel.R(0, 0, 1024, 768),
		Resizable: true,
		Maximized: true,
		Icon:      icons,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	game(win)
}

func align(location pixel.Vec, spacing pixel.Vec) pixel.Vec {
	alignedX := math.Round(location.X/spacing.X) * spacing.X
	alignedY := math.Round(location.Y/spacing.Y) * spacing.Y

	return pixel.Vec{X: alignedX, Y: alignedY}
}

func renderChunks(wg *sync.WaitGroup, win *pixelgl.Window, chunks []*chunks.Chunk, spriteSheet pixel.Picture) {
	defer wg.Done()
	skipped := 0
	for _, chunk := range chunks {
		if !chunk.Visible(win) {
			skipped++
			continue
		}
		chunk.Draw(win, spriteSheet)
	}
}

func game(win *pixelgl.Window) {
	var wg sync.WaitGroup
	wg.Add(1)
	resultChan := make(chan string)
	go func() {
		result, err := tiles.LoadAndCacheTiles(&wg, "resources/textures/tiles")
		if err != nil {
			panic(err)
		}
		resultChan <- result
	}()
	wg.Wait()
	result := <-resultChan
	tileSheet, err := loadPicture(result)
	if err != nil {
		panic(err)
	}
	/*
		TILE SHEET LOADED
	*/

	ui.Init()
	entities := make([][]*tiles.Tile, 16) //! TEMP

	for i := 0; i < 16; i++ {
		entities[i] = make([]*tiles.Tile, 16) //! TEMP
	}
	chunkList := []*chunks.Chunk{} //! TEMP
	//chunk := newChunk(entities, win.Bounds().Center(), pixel.Vec{X: 8, Y: 8}, pixel.Vec{X: 16, Y: 16})
	//basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	//basicTxt := text.New(win.Bounds().Center(), basicAtlas)
	chunkCount := 0
	startTime := time.Now()
	fps := 0
	cps := 0
	var (
		camPos       = pixel.ZV
		camSpeed     = 500.0
		camZoom      = 1.0
		camZoomSpeed = 1.2
	)

	var (
		debug = false
	)
	//basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	//basicTxt := text.New(pixel.V(100, 500), basicAtlas)
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)
		if win.Pressed(pixelgl.MouseButtonLeft) {
			/* for _, chunk := range chunks {
				chunk.location = win.MousePosition().Sub(win.MousePreviousPosition()).Add(chunk.location)
			} */
			//chunk.location = chunk.location.Add(pixel.Vec{X: 0, Y: 1})
			camPos = win.MousePosition().Sub(win.MousePreviousPosition()).Add(camPos)
		} else if win.Pressed(pixelgl.KeyR) {
			chunkList = []*chunks.Chunk{}
			chunkCount = 0
			//chunk.location = win.Bounds().Center()
			//chunk.Fill(pixel.Vec{X: 0, Y: 0}, pixel.Vec{X: 8, Y: 8}, pixel.NewSprite(image, pixel.R(0, 0, image.Bounds().H(), image.Bounds().W())))
		} else if win.JustPressed(pixelgl.KeySpace) {
			genChunk := chunks.NewChunk(entities, pixel.Vec{X: rand.Float64() * win.Bounds().W(), Y: rand.Float64() * win.Bounds().H()}, pixel.Vec{X: 16, Y: 16}, pixel.Vec{X: 16, Y: 16})

			genChunk.Fill(pixel.Vec{X: 0, Y: 0}, pixel.Vec{X: 16, Y: 16}, tiles.NewTile(tiles.Meta{Id: "minecraft:stone"}, tileSheet))
			genChunk.Set(pixel.Vec{X: 7, Y: 8}, tiles.NewTile(tiles.Meta{Id: "minecraft:dirt"}, tileSheet))
			genChunk.Set(pixel.Vec{X: 8, Y: 8}, tiles.NewTile(tiles.Meta{Id: "minecraft:grass"}, tileSheet))
			genChunk.Set(pixel.Vec{X: 7, Y: 7}, tiles.NewTile(tiles.Meta{Id: "minecraft:dirt"}, tileSheet))
			genChunk.Set(pixel.Vec{X: 8, Y: 7}, tiles.NewTile(tiles.Meta{Id: "minecraft:grass"}, tileSheet))
			genChunk.Location = align(genChunk.Location, pixel.Vec{X: 256, Y: 256})
			chunkList = append(chunkList, genChunk)
			chunkCount++
		}
		if win.JustPressed(pixelgl.KeyF3) {
			debug = !debug
		}
		if win.Pressed(pixelgl.KeyD) {
			fuck := rand.New(rand.NewSource(time.Now().UnixNano()))
			chunkList[fuck.Intn(len(chunkList))].Set(pixel.Vec{X: float64(fuck.Intn(16)), Y: float64(fuck.Intn(16))}, tiles.NewTile(tiles.Meta{Id: "minecraft:dirt"}, tileSheet)).Reload()
		}
		if win.Pressed(pixelgl.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyUp) {
			camPos.Y += camSpeed * dt
		}

		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		if win.JustReleased(pixelgl.MouseButtonRight) {
			cps++
		}

		if win.JustReleased(pixelgl.MouseButtonLeft) {
			cps++
		}
		img := image.NewRGBA(image.Rect(0, 0, 16, 16))
		for y := img.Bounds().Dy(); y > 0; y-- { // Start from the bottom to avoid index out of range error
			for x := img.Bounds().Dx(); x > 0; x-- {
				// Set the color of the pixel at (x, y) to blue
				pixelColor := color.RGBA{0, 0, 255, 255} // Blue color with full opacity
				img.Set(x, y, pixelColor)
			}
		}
		var pixels []uint8
		for y := img.Bounds().Dy(); y > 0; y-- {
			for x := img.Bounds().Dx(); x > 0; x-- {
				r, g, b, a := img.At(x, y).RGBA()
				alphaPremultipliedRed := int(float64(r) * float64(a) / 255)
				alphaPremultipliedGreen := int(float64(g) * float64(a) / 255)
				alphaPremultipliedBlue := int(float64(b) * float64(a) / 255)
				pixels = append(pixels, uint8(alphaPremultipliedBlue), uint8(alphaPremultipliedGreen), uint8(alphaPremultipliedRed), uint8(a))
			}
		}
		win.Canvas().SetPixels(pixels)
		//basicTxt.Orig = mouse
		if time.Since(startTime).Seconds() >= 1 {
			win.SetTitle("FPS: " + strconv.Itoa(fps) + " CPS: " + strconv.Itoa(cps) + " Chunks: " + strconv.Itoa(chunkCount) + " DeltaTime: " + fmt.Sprintf("%f", dt))
			startTime = time.Now() // Reset the start time
			fps = 0
			cps = 0
		}

		win.Clear(colornames.Skyblue)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			renderChunks(&wg, win, chunkList, tileSheet)
		}()
		wg.Wait()
		if debug {
			b := pixel.NewBatch(&pixel.TrianglesData{}, nil)
			for _, c := range chunkList {
				imd := imdraw.New(nil)

				var (
					p1 = pixel.V(c.Location.X, c.Location.Y)
					p2 = pixel.V(c.Location.X, c.Location.Y+((c.Size.Y-1)*16))
					p3 = pixel.V(c.Location.X+((c.Size.X-1)*16), c.Location.Y+((c.Size.Y-1)*16))
					p4 = pixel.V(c.Location.X+((c.Size.X-1)*16), c.Location.Y)
				)
				imd.Color = colornames.Red
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(p1, p2, p3, p4, p1)
				imd.Line(5)
				imd.Draw(b)
			}
			b.Draw(win)
		}
		win.Update()
		fps++
	}
}
