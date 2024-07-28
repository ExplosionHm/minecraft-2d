package ui

import (
	"fmt"
	"image"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
	"github.com/gopxl/pixel"
)

type UIManager struct {
}

/*
Anything that contains a texture is considered an element
*/

type Element struct {
	Rect pixel.Rect
	Path string
}

type Instruction struct {
	Align struct {
		Container int
		Positon   string
	}
	Factory struct {
		Type    int
		Matrix  []int
		Element Element
		Align   string
	}
}

func Init() {
	file, err := os.Open("resources/ui/gameplay.yml")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Create a new YAML decoder
	var data map[string]interface{}
	dec := yaml.NewDecoder(file)

	// Decode the YAML content into the data variable
	if err := dec.Decode(&data); err != nil {
		fmt.Println(err)
		return
	}
	h, _ := Parse(data)
	fmt.Println(h)
}

var Errors = []string{"Expected string.", "Expected number.", "Expected object."}

func Parse(scene map[string]interface{}) (Instruction, error) {
	var (
		instruction Instruction
	)

	parseElement := func(payload interface{}) Element {
		e := Element{}
		for key, value := range payload.(map[string]interface{}) {
			_, isString := value.(string)
			//_, isObject := value.(map[string]interface{})
			switch key {
			case "texture":
				if !isString {
					panic(key + Errors[0])
				}
				e.Path = value.(string)
			default:
				fmt.Println("skipped element key: ", key)
			}
		}
		e.Rect = pixel.R(0, 0, 0, 0) //Left undefined until render process
		if e.Path == "" {
			panic("Element requires texture id")
		}
		return e
	}
	parseFactory := func(instruction *Instruction, factory interface{}) {
		for key, value := range factory.(map[string]interface{}) {
			_, isString := value.(string)
			_, isObject := value.(map[string]interface{})
			switch key {
			case "type":
				if !isString {
					panic(key + Errors[0])
				}
				instruction.Factory.Type = 0 //! Add proper enum system for the type of factory
			case "payload":
				if !isObject {
					panic(key + Errors[2])
				}
				instruction.Factory.Element = parseElement(value)
			case "matrix":
				if !isString {
					panic(key + Errors[0])
				}
				index := 0
				output := []int{0, 0}
				temp := ""
				for i, v := range value.(string) + " " {
					if v == '*' || i >= len(value.(string)) {
						integer, err := strconv.Atoi(temp)
						if err != nil {
							panic(err)
						}
						output[index] = integer
						index++
						temp = ""
						continue
					}
					temp += string(v)
				}
				instruction.Factory.Matrix = output
			case "align":
				if !isString {
					panic(key + Errors[0])
				}
				instruction.Factory.Align = value.(string)
			}
		}
	}
	for key, value := range scene["scene"].(map[string]interface{}) {
		_, isString := value.(string)
		_, isObject := value.(map[string]interface{})
		switch key {
		case "align":
			if !isString {
				panic(key + Errors[0])
			}
			instruction.Align.Positon = value.(string)
			instruction.Align.Container = 0
		case "factory":
			if !isObject {
				panic(key + Errors[2])
			}
			parseFactory(&instruction, value)
		default:
			fmt.Println("skipped element key: ", key)
		}
		fmt.Println(key, isString)
	}
	return instruction, nil
}

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

func Draw(instructions Instruction) []*pixel.Sprite {

	var (
		output []*pixel.Sprite
		//align string
	)

	img, err := loadPicture(instructions.Factory.Element.Path)
	if err != nil {
		panic(err)
	}
	//align = instructions.Factory.Align
	instructions.Factory.Element.Rect = img.Bounds()
	for x := 0; x < int(instructions.Factory.Element.Rect.W())*instructions.Factory.Matrix[0]; x += int(instructions.Factory.Element.Rect.W()) {
		for y := 0; y < int(instructions.Factory.Element.Rect.H())*instructions.Factory.Matrix[1]; y += int(instructions.Factory.Element.Rect.H()) {
			pixel.NewSprite(img, instructions.Factory.Element.Rect)

		}
	}
	return output
}
