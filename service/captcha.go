package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	mathrand "math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chuccp/go-web-frame/core"
)

const (
	captchaWidth  = 300
	captchaHeight = 200
	captchaTTL    = 5 * time.Minute
	// captchaMaxChallenges bounds the in-memory store; challenges are short-lived, so a
	// flood at worst forces older pages to refresh.
	captchaMaxChallenges = 1024
	// captchaTolerance is the click radius in image pixels (the client scales its
	// coordinates back to the 300x200 canvas before sending them).
	captchaTolerance = 30
)

// captchaShapeKeys is the vocabulary shared with the UI, which translates each key for
// the prompt. The order is only the source pool; every challenge shuffles it.
var captchaShapeKeys = []string{"circle", "triangle", "square", "diamond", "star"}

var captchaPalette = []color.RGBA{
	{47, 111, 237, 255},
	{214, 69, 69, 255},
	{45, 138, 86, 255},
	{156, 84, 200, 255},
	{214, 138, 35, 255},
}

// CaptchaData is one challenge handed to the login page: the rendered shapes, the order
// their names must be clicked in, and the id the answer is checked against.
type CaptchaData struct {
	ID     string
	Image  string
	Shapes []string // shape keys in required click order
	Width  int
	Height int
}

type captchaTarget struct {
	x, y int
}

type captchaChallenge struct {
	targets []captchaTarget
	// binding ties the challenge to the username/password pair it was requested for,
	// so a solved challenge cannot be reused for different credentials.
	binding []byte
	expires time.Time
}

// CaptchaService issues and checks the login's click captcha. Challenges live in memory
// and are one-shot: an attempt consumes the id whether it succeeds or fails, so a solved
// challenge cannot be replayed. A restart only invalidates challenges that were on screen;
// the page fetches a new one on reload.
type CaptchaService struct {
	mu         sync.Mutex
	challenges map[string]*captchaChallenge
	now        func() time.Time
	key        []byte
}

func (s *CaptchaService) Init(*core.Context) error {
	s.challenges = make(map[string]*captchaChallenge)
	s.now = time.Now
	s.key = make([]byte, 32)
	if _, err := rand.Read(s.key); err != nil {
		return err
	}
	return nil
}

// bindingFor hashes the credentials with a per-process key. Only the digest is kept with
// the challenge; the password itself is never stored.
func (s *CaptchaService) bindingFor(username, password string) []byte {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(username))
	mac.Write([]byte{0})
	mac.Write([]byte(password))
	return mac.Sum(nil)
}

// New renders a challenge bound to the credentials supplied with it.
func (s *CaptchaService) New(username, password string) (*CaptchaData, error) {
	targets, shapes, img, err := buildCaptcha()
	if err != nil {
		return nil, err
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	id := hex.EncodeToString(raw)
	binding := s.bindingFor(username, password)

	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.challenges) >= captchaMaxChallenges {
		s.pruneLocked(now)
	}
	s.challenges[id] = &captchaChallenge{targets: targets, binding: binding, expires: now.Add(captchaTTL)}
	return &CaptchaData{
		ID:     id,
		Image:  img,
		Shapes: shapes,
		Width:  captchaWidth,
		Height: captchaHeight,
	}, nil
}

func (s *CaptchaService) pruneLocked(now time.Time) {
	for id, challenge := range s.challenges {
		if now.After(challenge.expires) {
			delete(s.challenges, id)
		}
	}
	if len(s.challenges) >= captchaMaxChallenges {
		// Still full of live challenges: start over rather than grow without bound.
		s.challenges = make(map[string]*captchaChallenge)
	}
}

// VerifyForm validates the credentials and the comma-separated click lists submitted by
// a login form.
func (s *CaptchaService) VerifyForm(id, username, password, xs, ys string) bool {
	return s.Verify(id, username, password, parseCaptchaPoints(xs), parseCaptchaPoints(ys))
}

// Verify reports whether the challenge was issued for these credentials and the clicks
// match the stored targets in order. The challenge is consumed either way, which is what
// makes a solved one unreplayable.
func (s *CaptchaService) Verify(id, username, password string, xs, ys []int) bool {
	if id == "" {
		return false
	}
	s.mu.Lock()
	challenge, ok := s.challenges[id]
	delete(s.challenges, id)
	s.mu.Unlock()
	if !ok {
		return false
	}
	if s.now().After(challenge.expires) {
		return false
	}
	if !hmac.Equal(challenge.binding, s.bindingFor(username, password)) {
		return false
	}
	if len(xs) != len(challenge.targets) || len(ys) != len(challenge.targets) {
		return false
	}
	for i, target := range challenge.targets {
		dx, dy := xs[i]-target.x, ys[i]-target.y
		if dx*dx+dy*dy > captchaTolerance*captchaTolerance {
			return false
		}
	}
	return true
}

func parseCaptchaPoints(raw string) []int {
	parts := strings.Split(raw, ",")
	points := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil
		}
		points = append(points, n)
	}
	return points
}

// buildCaptcha picks exactly three shapes (their order is the prompt order) and places
// them on the canvas at well-separated anchors. What the prompt names is what the image
// shows — no decoys.
func buildCaptcha() ([]captchaTarget, []string, string, error) {
	keys := append([]string(nil), captchaShapeKeys...)
	mathrand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	targets := append([]string(nil), keys[:3]...)
	mathrand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })

	drawKeys := append([]string(nil), targets...)
	mathrand.Shuffle(len(drawKeys), func(i, j int) { drawKeys[i], drawKeys[j] = drawKeys[j], drawKeys[i] })

	anchors := []image.Point{
		{X: 70, Y: 65}, {X: 230, Y: 65}, {X: 150, Y: 145},
	}
	positions := make(map[string]image.Point, len(drawKeys))
	for i, key := range drawKeys {
		p := anchors[i]
		p.X += mathrand.Intn(15) - 7
		p.Y += mathrand.Intn(13) - 6
		positions[key] = p
	}

	expected := make([]captchaTarget, 0, len(targets))
	for _, key := range targets {
		p := positions[key]
		expected = append(expected, captchaTarget{x: p.X, y: p.Y})
	}

	img, err := drawCaptchaImage(drawKeys, positions)
	if err != nil {
		return nil, nil, "", err
	}
	return expected, targets, img, nil
}

func drawCaptchaImage(keys []string, positions map[string]image.Point) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, captchaWidth, captchaHeight))
	background := color.RGBA{248, 249, 252, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: background}, image.Point{}, draw.Src)

	for i := 0; i < 90; i++ {
		g := uint8(190 + mathrand.Intn(50))
		img.Set(mathrand.Intn(captchaWidth), mathrand.Intn(captchaHeight), color.RGBA{g, g, g, 255})
	}
	for i := 0; i < 3; i++ {
		drawCaptchaLine(img,
			mathrand.Intn(captchaWidth), mathrand.Intn(captchaHeight),
			mathrand.Intn(captchaWidth), mathrand.Intn(captchaHeight),
			color.RGBA{224, 228, 235, 255})
	}
	for _, key := range keys {
		p := positions[key]
		size := 20 + mathrand.Intn(5)
		colour := captchaPalette[mathrand.Intn(len(captchaPalette))]
		if key == "circle" {
			fillCircle(img, p.X, p.Y, size, colour)
		} else {
			fillPolygon(img, shapePoints(key, p.X, p.Y, size), colour)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func shapePoints(key string, cx, cy, size int) []image.Point {
	switch key {
	case "triangle":
		return []image.Point{{X: cx, Y: cy - size}, {X: cx + size, Y: cy + size}, {X: cx - size, Y: cy + size}}
	case "square":
		return []image.Point{
			{X: cx - size, Y: cy - size}, {X: cx + size, Y: cy - size},
			{X: cx + size, Y: cy + size}, {X: cx - size, Y: cy + size},
		}
	case "diamond":
		return []image.Point{
			{X: cx, Y: cy - size}, {X: cx + size, Y: cy},
			{X: cx, Y: cy + size}, {X: cx - size, Y: cy},
		}
	case "star":
		points := make([]image.Point, 0, 10)
		for i := 0; i < 10; i++ {
			radius := size
			if i%2 == 1 {
				radius = size * 2 / 5
			}
			angle := -math.Pi/2 + float64(i)*math.Pi/5
			points = append(points, image.Point{
				X: cx + int(math.Cos(angle)*float64(radius)),
				Y: cy + int(math.Sin(angle)*float64(radius)),
			})
		}
		return points
	}
	return nil
}

func fillCircle(img *image.RGBA, cx, cy, radius int, colour color.RGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		dy := y - cy
		half := int(math.Sqrt(float64(radius*radius - dy*dy)))
		for x := cx - half; x <= cx+half; x++ {
			img.Set(x, y, colour)
		}
	}
}

// fillPolygon is a small even-odd scanline fill, enough for the triangle, square, diamond
// and star.
func fillPolygon(img *image.RGBA, points []image.Point, colour color.RGBA) {
	if len(points) < 3 {
		return
	}
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	for y := minY; y <= maxY; y++ {
		crossings := make([]int, 0, len(points))
		for i := range points {
			p1, p2 := points[i], points[(i+1)%len(points)]
			if p1.Y == p2.Y {
				continue
			}
			if (p1.Y <= y && p2.Y > y) || (p2.Y <= y && p1.Y > y) {
				crossings = append(crossings, p1.X+(y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y))
			}
		}
		sort.Ints(crossings)
		for i := 0; i+1 < len(crossings); i += 2 {
			for x := crossings[i]; x <= crossings[i+1]; x++ {
				img.Set(x, y, colour)
			}
		}
	}
}

func drawCaptchaLine(img *image.RGBA, x1, y1, x2, y2 int, colour color.RGBA) {
	dx := absInt(x2 - x1)
	dy := absInt(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		img.Set(x1, y1, colour)
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

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
