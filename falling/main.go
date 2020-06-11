package main

import (
	"image"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/ByteArena/box2d"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	// Prepare for simulation. Typically we use a time step of 1/60 of a
	// second (60Hz) and 10 iterations. This provides a high quality simulation
	// in most game scenarios.
	velocityIterations = 8
	positionIterations = 3

	camZoomSpeed = 1.2
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

func loadSprites() []*pixel.Sprite {
	spritesheet, err := loadPicture("falling/trees.png")
	if err != nil {
		panic(err)
	}
	var sprites []*pixel.Sprite
	for x := spritesheet.Bounds().Min.X; x < spritesheet.Bounds().Max.X; x += 32 {
		for y := spritesheet.Bounds().Min.Y; y < spritesheet.Bounds().Max.Y; y += 32 {
			sprites = append(sprites, pixel.NewSprite(spritesheet, pixel.R(x, y, x+32, y+32)))
		}
	}
	return sprites
}

func generateTree(world *box2d.B2World) *box2d.B2Body {
	x := rand.Float64()*200 - 100
	y := rand.Float64()*80 + 8
	bodyDef := box2d.MakeB2BodyDef()
	bodyDef.Type = box2d.B2BodyType.B2_dynamicBody
	bodyDef.Position.Set(x, y)
	bodyDef.LinearDamping = 0.02
	body := world.CreateBody(&bodyDef)
	dynamicBox := box2d.MakeB2CircleShape()
	dynamicBox.SetRadius(1)
	fixtureDef := box2d.MakeB2FixtureDef()
	fixtureDef.Shape = &dynamicBox
	fixtureDef.Density = 1
	fixtureDef.Friction = 1
	fixtureDef.Restitution = 0.4
	body.CreateFixtureFromDef(&fixtureDef)
	return body
}

func createWorld() (*box2d.B2World, *imdraw.IMDraw) {
	// Define the gravity vector.
	gravity := box2d.MakeB2Vec2(0.0, -10.0)

	// Construct a world object, which will hold and simulate the rigid bodies.
	world := box2d.MakeB2World(gravity)

	// Create the ground in the physics model
	groundBodyDef := box2d.MakeB2BodyDef()
	groundBodyDef.Position.Set(0, 0)
	groundBody := world.CreateBody(&groundBodyDef)

	groundTriangle := box2d.MakeB2PolygonShape()
	vertices := []box2d.B2Vec2{
		box2d.MakeB2Vec2(10, 1),
		box2d.MakeB2Vec2(0, 10),
		box2d.MakeB2Vec2(-10, 1),
	}
	groundTriangle.Set(vertices, len(vertices))
	groundBody.CreateFixture(&groundTriangle, 0.0)

	groundBase := box2d.MakeB2PolygonShape()
	groundBase.SetAsBox(50, 1)
	groundBody.CreateFixture(&groundBase, 0.0)

	// Draw the ground directly
	imd := imdraw.New(nil)
	imd.Color = colornames.Sandybrown
	for _, v := range vertices {
		imd.Push(pixel.V(v.X, v.Y).Scaled(32))
	}
	imd.Polygon(0)
	imd.Push(
		pixel.V(50, 1).Scaled(32),
		pixel.V(-50, -1).Scaled(32),
	)
	imd.Rectangle(0)

	return &world, imd
}

func sim() {

	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Create a world
	sprites := loadSprites()
	world, drawableWorld := createWorld()

	// Generate random trees
	trees := []*box2d.B2Body{}
	for i := 0; i < 800; i++ {
		// Use the middle tree sprite since it's big and fills circular physics body nicely
		trees = append(trees, generateTree(world))
	}

	camZoom := 0.4
	camPos := pixel.V(1024/2, 0)
	lastTime := time.Now()
	treeSprite := sprites[4] // Big tree that fills the physics body nicely
	for !win.Closed() {

		// We're v-synced so calculate the time elapsed since the last frame and step the simulation that far
		currentTime := time.Now()
		dt := currentTime.Sub(lastTime)
		lastTime = currentTime
		world.Step(dt.Seconds(), velocityIterations, positionIterations)

		// Check the mouse wheel to determine camera position
		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)
		cam := pixel.IM.Scaled(pixel.ZV, camZoom).Moved(camPos)
		if win.Pressed(pixelgl.MouseButtonLeft) {
			camPos = win.MousePosition()
		}
		win.SetMatrix(cam)

		// Draw the world and trees
		win.Clear(colornames.Whitesmoke)
		drawableWorld.Draw(win)
		for _, tree := range trees {

			// Physics X and Y which are in metres
			x := tree.GetPosition().X
			y := tree.GetPosition().Y

			// Determine the position on screen by scaling so that we get 32 pixels to the metre
			pos := pixel.V(x, y).Scaled(32)

			// Draw a tree sprite for this physics body
			treeSprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 2).Moved(pos))

		}
		win.Update()

	}

}

func main() {
	pixelgl.Run(sim)
}
