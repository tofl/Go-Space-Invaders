package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image"
	_ "image/png"
	"os"
	"time"
	"math/rand"
)

type bullet struct {
	x, y float64
}

type alien struct {
	x, y float64
}

type alienBullet struct {
	x, y float64
}

var shipX float64 = 0.0
var bullets []bullet
var aliens [][]alien
var alienBullets []alienBullet
var timeLatestWave time.Time
var timeLatestAlienBulletShot time.Time
var gameLost bool
var nbAliensShot int

func newText(textToPrint string, v pixel.Vec) *text.Text {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	returnText := text.New(v, atlas)
	fmt.Fprint(returnText, textToPrint)

	return returnText
}

func addAlienRows(nbRows int) {
	if len(aliens) > 0 {
		for alienRowKey, _ := range aliens {
			for alienKey, _ := range aliens[alienRowKey] {
				aliens[alienRowKey][alienKey].y = aliens[alienRowKey][alienKey].y - 100
			}
		}
	}

	x, y := 100.0, 668.0
	for a := 0; a < nbRows; a++ {
		var alienRow []alien
		for b := 0; b < 10; b++ {
			alienRow = append(alienRow, alien{x: x, y: y})
			x += 100
		}
		aliens = append(aliens, alienRow)
		x = 100
		y -= 100
	}
}

func run() {
	config := pixelgl.WindowConfig{
		Title: "Space Invaders",
		Bounds: pixel.R(0, 0, 1124, 768),
		VSync: true,
	}

	win, err := pixelgl.NewWindow(config)
	if err != nil {
		panic(err)
	}
	win.SetSmooth(true)

	// Ship sprite
	shipFile, err := loadPicture("images/ship.png")
	if err != nil {
		panic(err)
	}
	shipSprite := pixel.NewSprite(shipFile, shipFile.Bounds())

	// Bullet sprite
	bulletFile, err := loadPicture("images/bullet.png")
	if err != nil {
		panic(err)
	}
	bulletSprite := pixel.NewSprite(bulletFile, bulletFile.Bounds())

	// Alien sprite
	alienFile, err := loadPicture("images/alien.png")
	if err != nil {
		panic(err)
	}
	alienSprite := pixel.NewSprite(alienFile, alienFile.Bounds())

	// Alien bullet sprite
	alienBulletFile, err := loadPicture("images/alienBullet.png")
	if err != nil {
		panic(err)
	}
	alienBulletSprite := pixel.NewSprite(alienBulletFile, alienBulletFile.Bounds()) // TODO

	// Initialize the lines of aliens
	addAlienRows(2)


	timeLatestWave = time.Now()
	timeLatestAlienBulletShot = time.Now()

	gameOverText := newText("Game over!", pixel.V(1124/2-170, 500))
	// var scoreText *text.Text
	scoreText := newText("Your score: 0", pixel.V(1124/2-170, 450))
	newGameText := newText("Press Enter to continue.", pixel.V(1124/2-170, 400))


	for !win.Closed() {
		win.Clear(colornames.Black)

		if gameLost {
			// Game over message
			gameOverText.Draw(win, pixel.IM.Scaled(gameOverText.Orig, 3))
			scoreText.Draw(win, pixel.IM.Scaled(scoreText.Orig, 2))
			newGameText.Draw(win, pixel.IM.Scaled(newGameText.Orig, 2))

			if win.JustPressed(pixelgl.KeyEnter) {
				scoreText = newText("Your score: 0", pixel.V(1124/2-170, 450))
				bullets = []bullet{}
				gameLost = false
				aliens = [][]alien{}
				alienBullets = []alienBullet{}
				nbAliensShot = 0
				addAlienRows(2)
			}

			win.Update()
			continue
		}

		// Move the ship
		if win.Pressed(pixelgl.KeyRight) && shipX < 550 {
			shipX += 7
		}
		if win.Pressed(pixelgl.KeyLeft) && shipX > -550 {
			shipX -= 7
		}

		// Shoot a bullet
		if win.JustPressed(pixelgl.KeySpace) {
			bullets = append(bullets, bullet{
				x: 1124/2 + shipX,
				y: 90,
			})
		}

		// Draw the ship sprite
		shipPos := pixel.IM
		shipPos = shipPos.ScaledXY(pixel.ZV, pixel.V(0.5, 0.5))
		shipPos = shipPos.Moved(pixel.V(win.Bounds().Center().X + shipX, 50))
		shipSprite.Draw(win, shipPos)

		// Draw the bullets
		for key, bullet := range bullets {
			bullets[key].y += 8

			bulletSprite.Draw(
				win,
				pixel.IM.
					ScaledXY(pixel.ZV, pixel.V(0.5, 0.5)).
					Moved(pixel.V(bullet.x, bullet.y)),
			)
		}

		// Draw the aliens
		for _, alienRow := range aliens {
			for _, alien := range alienRow {
				alienSprite.Draw(
					win,
					pixel.IM.
						ScaledXY(pixel.ZV, pixel.V(0.3, 0.3)).
						Moved(pixel.V(alien.x, alien.y)),
				)
			}
		}

		// Draw the alien bullets
		for alienBulletsKey, alienBullet := range alienBullets {
			alienBulletSprite.Draw(
				win,
				pixel.IM.
					Moved(pixel.V(alienBullet.x, alienBullet.y)),
			)
			alienBullets[alienBulletsKey].y -= 3
		}

		// Remove an alien bullet when it leaves the screen, lose game when it hits the ship
		for alienBulletsKey, alienBullet := range alienBullets {
			if alienBullet.y < 2 {
				alienBullets = append(alienBullets[:alienBulletsKey], alienBullets[alienBulletsKey+1:]...)
			}

			// Lose game if an alien bullet hits the ship
			shipPosition := win.Bounds().Center().X + shipX
			if alienBullet.y < 100 && alienBullet.x >= shipPosition - 60 && alienBullet.x < shipPosition + 40 {
				gameLost = true
			}
		}

		// or hit an alien
		for bulletKey, bullet := range bullets {
			// Delete the bullets as they leave the screen
			if bullet.y > win.Bounds().Max.Y - 20 {
				bullets = append(bullets[:bulletKey], bullets[bulletKey+1:]...)
			}

			for alienRowKey, aliensRow := range aliens {
				for alienKey, alien := range aliensRow {
					if alien.x - 40 < bullet.x && alien.x + 50 > bullet.x && alien.y < bullet.y && alien.y + 20 < bullet.y {
						// Delete the aliens when they are hit by a bullet
						alienRow := aliens[alienRowKey]
						alienRow = append(alienRow[:alienKey], alienRow[alienKey+1:]...)
						aliens[alienRowKey] = alienRow

						// If there are no more aliens in the row, delete the row
						if len(aliens[alienRowKey]) == 0 {
							aliens = append(aliens[:alienRowKey], aliens[alienRowKey+1:]...)
						}

						nbAliensShot++
						scoreText = newText(fmt.Sprintf("Your score: %v", nbAliensShot), pixel.V(1124/2-170, 450))

						// Delete the bullet when it hits an alien
						bullets = append(bullets[:bulletKey], bullets[bulletKey+1:]...)
					}
				}
			}
		}

		// Add a row of aliens every 5 seconds
		if time.Since(timeLatestWave).Milliseconds() >= 5000 {
			addAlienRows(1)
			timeLatestWave = time.Now()
		}

		// Aliens shoot bullets
		if time.Since(timeLatestAlienBulletShot).Milliseconds() > 750 && len(aliens) > 0 {
			// Pick a random alien that shoots a bullet
			randomAliensRowKey := rand.Intn(len(aliens))
			randomAlienKey := rand.Intn(len(aliens[randomAliensRowKey]))
			selectedAlien := aliens[randomAliensRowKey][randomAlienKey]
			alienBullets = append(alienBullets, alienBullet{x: selectedAlien.x, y: selectedAlien.y})

			timeLatestAlienBulletShot = time.Now()
		}

		// Lose game if an alien row gets to close to the ship
		for _, aliensRow := range aliens {
			for _, alien := range aliensRow {
				if alien.y < 100 {
					// Lose game
					gameLost = true
				}
			}
		}

		win.Update()
	}
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

func main() {
	pixelgl.Run(run)
}
