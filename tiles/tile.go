package tiles

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"
	"sync"

	"github.com/gopxl/pixel"
)

type TileData struct {
	Frame         pixel.Rect
	ApplyRotation bool
}

var TileMap = make(map[string]TileData)

type Meta struct {
	Id string
}

type Tile struct {
	Meta     Meta
	Sprite   *pixel.Sprite
	WorldPos pixel.Vec
}

func NewTile(meta Meta, texture pixel.Picture) *Tile {
	return &Tile{
		Meta:     meta,
		Sprite:   pixel.NewSprite(texture, TileMap[meta.Id].Frame),
		WorldPos: pixel.ZV,
	}
}

func drawImage(canvas *image.RGBA, img image.Image, point image.Point) {
	for x := point.X; x < img.Bounds().Dx()+point.X+1; x++ {
		for y := point.Y; y < img.Bounds().Dy()+point.Y+1; y++ {
			//fmt.Println("Drawing to X: " + strconv.Itoa(x) + ", Y:" + strconv.Itoa(y) + " from X: " + strconv.Itoa(x-point.X) + ", Y: " + strconv.Itoa(y-point.Y))
			canvas.Set(x, y, img.At(x-point.X, y-point.Y))
		}
	} //! Add prevention to drawing out of the canvas
}

type tileData struct {
	Textures      string `json:"textures"`
	ApplyRotation bool   `json:"applyRotation"`
}

type tilesJSON struct {
	Data map[string]tileData `json:"data"`
}

func LoadAndCacheTiles(wg *sync.WaitGroup, path string) (string, error) {
	defer fmt.Println("Exited function properly.")
	defer wg.Done()

	fileJson, err := os.ReadFile("resources/tiles.json")
	if err != nil {
		return "", err
	}
	var tiles tilesJSON

	e := json.Unmarshal(fileJson, &tiles)
	if e != nil {
		return "", e
	}

	if len(tiles.Data) == 0 {
		fmt.Println("No tiles found in tiles.json")
		return "", nil
	}

	var (
		width  = 256
		height = 256
		index  = 0
	)
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	fmt.Println("created canvas")
	for tileID, tileData := range tiles.Data {
		fmt.Printf("loading new...")
		file, err := os.Open(path + "/" + tileData.Textures + ".png")
		if err != nil {
			return "", err
		}
		defer file.Close()

		img, err := png.Decode(file)
		if err != nil {
			return "", err
		}

		fmt.Println(file.Name())
		drawImage(canvas, img, image.Point{X: index, Y: 0})
		index += img.Bounds().Dx()
		fmt.Println("index: " + strconv.Itoa(index))
		TileMap[tileID] = TileData{Frame: pixel.R(float64(index)-float64(img.Bounds().Dx()), float64(height), float64(index), float64(height)-float64(img.Bounds().Dy())), ApplyRotation: tileData.ApplyRotation}
		//fmt.Println("ID: " + tileID + ", min1: 16, max1: " + strconv.Itoa(height) + ", min2: " + strconv.Itoa(index) + ", max2: 16")
	}
	fmt.Println("Finished drawing to canvas.")
	outFile, err := os.Create("spritesheet.png")
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	fmt.Println("encoding file...")
	er := png.Encode(outFile, canvas)
	if er != nil {
		return "", er
	}
	fmt.Println("Sprite sheet saved:", outFile.Name())
	return outFile.Name(), nil
}

type TilesResult struct {
	Result string
	Err    error
}
