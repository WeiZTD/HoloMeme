package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// Game implements ebiten.Game interface.
type Game struct {
	count  int
	player *audio.Player
}

const (
	screenWidth  = 1280
	screenHeight = 720

	instructionsfontDPI = 72

	instructionsText = `←: Kalm
→: Gotta Go Fast
Space: Switch character`

	sampleRate = 44100
)

var (
	//go:embed resources/* html/index.html
	f embed.FS

	normalFont font.Face

	debugMode      = false
	debugKey       = ebiten.KeyF4
	cursorPosition string

	spriteX         = 1
	spriteY         = 1
	frameWidth      = 750
	frameHeight     = 720
	tickPerFrame    = 6
	animScale       = 0.8
	currentChar     *ebiten.Image
	ameImage        *ebiten.Image
	kfcImage        *ebiten.Image
	backgroundImage *ebiten.Image

	audioContext    = audio.NewContext(sampleRate)
	backgroundMusic []byte

	speedUpAnimKey  = ebiten.KeyRight
	slowDownAnimKey = ebiten.KeyLeft
	changeCharaKey  = ebiten.KeySpace
)

func init() {
	b, err := f.ReadFile("resources/font/BalsamiqSans-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(b)
	if err != nil {
		log.Fatal(err)
	}
	normalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    26,
		DPI:     instructionsfontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	b, err = f.ReadFile("resources/music/shootingStars.ogg")
	if err != nil {
		log.Fatal(err)
	}
	backgroundMusic = b

	b, err = f.ReadFile("resources/images/spaceBG.png")
	if err != nil {
		log.Fatal(err)
	}
	backGroundB, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	backgroundImage = ebiten.NewImageFromImage(image.Image(backGroundB))

	b, err = f.ReadFile("resources/images/ameSprite.png")
	if err != nil {
		log.Fatal(err)
	}
	ameImageB, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	ameImage = ebiten.NewImageFromImage(image.Image(ameImageB))

	b, err = f.ReadFile("resources/images/kfcSprite.png")
	if err != nil {
		log.Fatal(err)
	}
	kfcImageB, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	kfcImage = ebiten.NewImageFromImage(image.Image(kfcImageB))
	currentChar = ameImage
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	if debugMode {
		cursorPosition = getCursorPosition()
	}

	switch {
	case inpututil.IsKeyJustPressed(speedUpAnimKey):
		if tickPerFrame > 1 {
			tickPerFrame--
		}
	case inpututil.IsKeyJustPressed(slowDownAnimKey):
		if tickPerFrame < 8 {
			tickPerFrame++
		}
	case inpututil.IsKeyJustPressed(changeCharaKey):
		switch currentChar {
		case ameImage:
			currentChar = kfcImage
		case kfcImage:
			currentChar = ameImage
		}
	//30 frames
	case inpututil.KeyPressDuration(debugKey) == 30:
		debugMode = !debugMode
	}

	if g.player != nil {
		g.count++
		return nil
	}
	oggS, err := vorbis.Decode(audioContext, bytes.NewReader(backgroundMusic))
	if err != nil {
		return err
	}

	s := audio.NewInfiniteLoop(oggS, 8*4*sampleRate)

	g.player, err = audio.NewPlayer(audioContext, s)
	if err != nil {
		return err
	}

	g.player.Play()
	g.count++
	return nil
}

func getCursorPosition() string {
	x, y := ebiten.CursorPosition()
	return fmt.Sprintf("X:%v,Y:%v", x, y)

}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(backgroundImage, nil)

	if g.count%int(tickPerFrame) == 0 {
		g.count = 24
		spriteX++
	}
	if spriteX > 6 {
		spriteX = 1
		spriteY++
	}
	if spriteY > 4 {
		spriteY = 1
		animScale = 0.8
	}
	sx, sy := spriteX*frameWidth, spriteY*frameHeight
	subImage := currentChar.SubImage(image.Rect(sx-frameWidth, sy-frameHeight, sx, sy))
	op := &ebiten.DrawImageOptions{}
	x, y := ebiten.CursorPosition()
	op.GeoM.Scale(animScale, animScale)
	op.GeoM.Translate(float64(x-frameWidth*int(animScale)), float64(y-frameHeight*int(animScale)))
	screen.DrawImage(subImage.(*ebiten.Image), op)
	animScale -= 0.0033

	text.Draw(screen, instructionsText, normalFont, 8, 75, color.White)

	if debugMode {
		ebitenutil.DebugPrintAt(screen, cursorPosition, 0, 0)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS：%.2f", ebiten.CurrentFPS()), 0, 15)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS：%.2f", ebiten.CurrentTPS()), 0, 30)
	}
}

// Layout accepts a native outside size in device-independent pixels and returns the game's logical
// screen size. On desktops, the outside is a window or a monitor (fullscreen mode)
//
// Even though the outside size and the screen size differ, the rendering scale is automatically
// adjusted to fit with the outside.
//
// You can return a fixed screen size if you don't care, or you can also return a calculated screen
// size adjusted with the given outside size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(`word.exe`)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
