package chunks

import (
	"fmt"
	"math"
	"math/rand/v2"
	"mc2d/tiles"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/pixelgl"
)

type Chunk struct {
	tiles         [][]*tiles.Tile
	Location      pixel.Vec
	Size          pixel.Vec
	tileSize      pixel.Vec
	ChunkPosition pixel.Vec
	renderData    *pixel.Batch
}

func NewChunk(tiles [][]*tiles.Tile, Location pixel.Vec, size pixel.Vec, tileSize pixel.Vec) *Chunk {
	return &Chunk{
		tiles:    tiles,
		Location: Location,
		Size:     size,
		tileSize: tileSize,
	}
}

func (c *Chunk) Draw(target pixel.Target, spriteSheet pixel.Picture) {
	if c.renderData != nil {
		c.renderData.Draw(target)
		return
	}
	b := pixel.NewBatch(&pixel.TrianglesData{}, spriteSheet)
	loc := pixel.Vec{}
	rot := float64(0)
	for x := 0; x < int(c.Size.X); x++ {
		for y := 0; y < int(c.Size.Y); y++ {
			if x >= len(c.tiles) || y >= len(c.tiles[x]) {
				continue
			}
			tile := c.tiles[x][y]

			if tile == nil {
				continue
			}
			loc = pixel.Vec{X: float64(c.Location.X) + float64(y)*float64(c.tileSize.Y), Y: float64(c.Location.Y) + float64(x)*float64(c.tileSize.X)}
			if tiles.TileMap[tile.Meta.Id].ApplyRotation {
				rot = float64(rand.IntN(4)*int(90)) * math.Pi / 180
			} else {
				rot = 0
			}
			tile.Sprite.Draw(b, pixel.IM.Moved(loc).Rotated(loc, rot))
			// fmt.Printf("Drawing image at: %f, %f\n", x*c.tileSize.X, y*c.tileSize.Y)
		}
	}
	c.renderData = b
	b.Draw(target)
}

func (c *Chunk) Visible(win *pixelgl.Window) bool {
	return c.Location.X < float64(win.Bounds().W()) &&
		c.Location.X+c.tileSize.X*float64(c.Size.X) > 0 &&
		c.Location.Y < float64(win.Bounds().H()) &&
		c.Location.Y+c.tileSize.Y*float64(c.Size.X) > 0
}

func (c *Chunk) ValidLocation(Location pixel.Vec) bool {
	return Location.X-1 < c.Size.X || Location.Y-1 < c.Size.Y
}

func (c *Chunk) Set(Location pixel.Vec, tile *tiles.Tile) *Chunk {
	if !c.ValidLocation(Location) {
		fmt.Println("Trying to fill outside the confineds of the chunk!")
		return nil
	}
	c.tiles[int(Location.X)][int(Location.Y)] = tile
	return c
}

func (c *Chunk) Get(Location pixel.Vec) *tiles.Tile {
	if !c.ValidLocation(Location) {
		return nil
	}
	return c.tiles[int(Location.X)][int(Location.Y)]
}

func (c *Chunk) Fill(from pixel.Vec, to pixel.Vec, tile *tiles.Tile) *Chunk {
	if !c.ValidLocation(from) || !c.ValidLocation(to) {
		fmt.Println("Trying to fill outside the confineds of the chunk!")
		return nil
	}
	moveX := 1
	moveY := 1
	if to.X <= from.X {
		moveX = -1
	}
	if to.Y <= from.Y {
		moveY = -1
	}

	for x := from.X; x != to.X; x += float64(moveX) {
		for y := from.Y; y != to.Y; y += float64(moveY) {
			c.Set(pixel.Vec{X: x, Y: y}, tile)
		}
	}
	return c
}

func (c *Chunk) Reload() *Chunk {
	c.renderData = nil
	return c
}
