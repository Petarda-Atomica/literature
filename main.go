package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path"
	"regexp"
)

func readIMG(fname string) [][]color.Color {
	// Open the image file
	inputFile, err := os.Open(fname)
	if err != nil {
		fmt.Println("⛔ERROR⛔: Couldn't open file: " + fname)
		panic(err)
	}
	defer inputFile.Close()

	// Decode the image
	img, _, err := image.Decode(inputFile)
	if err != nil {
		panic(err)
	}

	//fmt.Println(img.Bounds().Size())

	// Make an output slice
	output := make([][]color.Color, img.Bounds().Size().X)
	for i := range output {
		output[i] = make([]color.Color, img.Bounds().Size().Y)
	}

	for x := 0; x < img.Bounds().Size().X; x++ {
		for y := 0; y < img.Bounds().Size().Y; y++ {
			output[x][y] = img.At(x, y)
		}
	}

	return output
}

func saveIMG(img [][]color.Color, fname string) {
	// Convert to image
	sizex := len(img)
	sizey := len(img[0])
	image := image.NewRGBA(image.Rect(0, 0, sizex, sizey))
	for x := image.Bounds().Min.X; x < image.Bounds().Max.X-1; x++ {
		for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y-1; y++ {
			image.Set(x, y, img[x][y])
		}
	}

	// Create an output file
	outputFile, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// Encode and save the modified image as a PNG
	err = png.Encode(outputFile, image)
	if err != nil {
		panic(err)
	}
}

func handleIfError(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func newCanvas(x int, y int) [][]color.Color {
	output := make([][]color.Color, x)
	for i := range output {
		output[i] = make([]color.Color, y)
	}
	for x := 0; x < len(output); x++ {
		for y := 0; y < len(output[0]); y++ {
			output[x][y] = color.White
		}
	}
	return output
}

func drawCircle(img [][]color.Color, coordx int, coordy int, r int) {
	for x := coordx - r; x < coordx+r; x++ {
		for y := coordy - r; y < coordy+r; y++ {
			if x < 0 || y < 0 {
				continue
			}
			if x > len(img)-1 || y > len(img[0])-1 {
				continue
			}
			if math.Sqrt(math.Pow(float64(x-coordx), 2)+math.Pow(float64(y-coordy), 2)) <= float64(r) {
				img[x][y] = color.Black
			}
		}
	}
}

// ! Written by ChatGPT
func drawLine(grid [][]color.Color, x1, y1, x2, y2 int) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1

	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}

	err := dx - dy

	for {
		//grid[y1][x1] = color.RGBA{255, 0, 0, 255} // Set the cell to red (you can adjust the color)
		drawCircle(grid, y1, x1, brushSize)

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err

		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// ! Written by ChatGPT
func resizeImage(input [][]color.Color, newWidth, newHeight int) [][]color.Color {
	// Create a new RGBA image with the desired dimensions
	resizedImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Create a scale factor for resizing
	scaleX := float64(len(input)) / float64(newWidth)
	scaleY := float64(len(input[0])) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Calculate the source pixel position based on the scale factor
			srcX := int(math.Floor(float64(x) * scaleX))
			srcY := int(math.Floor(float64(y) * scaleY))

			// Get the color from the source image
			srcColor := input[srcX][srcY]

			// Set the color in the resized image
			resizedImage.Set(x, y, srcColor)
		}
	}

	// Convert the resized image to a [][]color.Color slice
	resizedSlice := make([][]color.Color, newWidth)
	for x := 0; x < newWidth; x++ {
		resizedSlice[x] = make([]color.Color, newHeight)
		for y := 0; y < newHeight; y++ {
			resizedSlice[x][y] = resizedImage.At(x, y)
		}
	}

	return resizedSlice
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

var workingDir string

func main() {
	// Get working dir
	var err error
	workingDir, err = os.Getwd()
	handleIfError(err)

	// Find text mappings
	rawMappings, err := os.ReadFile("dataMappings.json")
	handleIfError(err)
	var Mappings map[string][]string
	handleIfError(json.Unmarshal(rawMappings, &Mappings))

	// Determine needed letters
	rawPrompt, err := os.ReadFile("prompt.txt")
	handleIfError(err)
	if len(string(rawPrompt)) <= 0 {
		fmt.Println("⚠️WARNING⚠️: Prompt file is empty, process terminated")
		return
	}
	letters := []string{}
	for i := 0; i < len(string(rawPrompt)); i++ {
		thisLetter := string(string(rawPrompt)[i])
		if thisLetter == " " {
			letters = append(letters, "++")
			continue
		} else if thisLetter == "\t" {
			for i := 0; i < tabSize; i++ {
				letters = append(letters, "++")
			}
		}
		var thisProcessedLetter string
		if len(Mappings[thisLetter]) != 0 {
			thisProcessedLetter = Mappings[thisLetter][rand.Intn(len(Mappings[thisLetter]))]
		} else {
			fmt.Println("⚠️WARNING⚠️: Undefined symbol found: " + thisLetter)
			thisProcessedLetter = "++"
		}
		letters = append(letters, thisProcessedLetter)
	}
	fmt.Println(letters)

	// Make a new canvas for our letters
	img := newCanvas(len(letters)*(idealLetterSize+letterSpacing), canvasY)

	// Fill the canvas with letters
	cursor := 0
	var prevConnection []int
	for i := 0; i < len(letters); i++ {
		// Put in a space
		if letters[i] == "++" {
			cursor += spaceSize
			prevConnection = nil
			continue
		}

		//* Put in a letter
		// Set apropriate size
		upper := true
		letterName := letters[i] + ".png"
		if !('A' <= letters[i][0] && letters[i][0] <= 'Z') {
			letterName = "_" + letterName
			upper = false
		}
		photo := readIMG(path.Join(workingDir, "/computerData/", letterName))
		if upper || (letters[i][0] == 'b' || letters[i][0] == 'd' || letters[i][0] == 'f' || letters[i][0] == 'h' || letters[i][0] == 'k' || letters[i][0] == 'l' || letters[i][0] == 't' || (regexp.MustCompile(`^[[:punct:]]+$`).MatchString(string(letters[i][0])) && letters[i][0] != '.' && letters[i][0] != ',') || letters[i] == "question" || regexp.MustCompile(`^\d+$`).MatchString(string(letters[i][0])) || false) {
			photo = resizeImage(photo, len(photo), capitalSize)
		} else {
			photo = resizeImage(photo, len(photo), lowerSize)
		}
		// Place letter
		var startPosition, endPosition []int
		for x := 0; x < len(photo); x++ {
			for y := 0; y < len(photo[0]); y++ {
				// Place pixel
				if x == len(photo)-1 || y == len(photo[0])-1 {
					continue
				}
				img[x+cursor][canvasY-len(photo[0])+y] = photo[x][y]
				// Find some important coords
				if photo[x][y] == startColor {
					startPosition = []int{x + cursor, canvasY - len(photo[0]) + y}
				} else if photo[x][y] == endColor {
					endPosition = []int{x + cursor, canvasY - len(photo[0]) + y}
				}
			}
		}
		// Connect letters
		if prevConnection == nil || startPosition == nil {
			// No connection to be made
		} else {
			drawLine(img, prevConnection[1], prevConnection[0], startPosition[1], startPosition[0])
		}
		prevConnection = endPosition

		cursor += len(photo)
		cursor += letterSpacing
	}

	// Trim out empty space
	img = img[:cursor]

	saveIMG(img, "canvasy.png")
}

const spaceSize = 400
const letterSpacing = 0
const tabSize = 3
const capitalSize = 250
const lowerSize = 150

const idealLetterSize = 250
const canvasY = 500
const brushSize = 7

// var transparency = color.RGBA{R: 0, G: 0, B: 0, A: 0}
var startColor = color.RGBA{R: 63, G: 72, B: 204, A: 255}
var endColor = color.RGBA{R: 34, G: 177, B: 76, A: 255}
