package webui

import (
	"testing"

	"github.com/chuccp/http2smtp/service"
)

// The click order reaches the page as shape keys (rendered as icons) and translated
// labels (used for aria-label/title); the text prompt remains as a fallback.
func TestCaptchaViewKeepsShapeOrder(t *testing.T) {
	p := &Pages{}
	view := p.captchaView("en", &service.CaptchaData{
		ID:     "id",
		Image:  "data:image/png;base64,AA==",
		Shapes: []string{"star", "circle", "square"},
		Width:  300,
		Height: 200,
	})
	if len(view.Shapes) != 3 || view.Shapes[0] != "star" || view.Shapes[1] != "circle" || view.Shapes[2] != "square" {
		t.Errorf("shape order = %v", view.Shapes)
	}
	if len(view.Labels) != 3 || view.Labels[0] != "Star" || view.Labels[1] != "Circle" || view.Labels[2] != "Square" {
		t.Errorf("labels = %v", view.Labels)
	}
	if view.Prompt != "Star → Circle → Square" {
		t.Errorf("prompt fallback = %q", view.Prompt)
	}
}
