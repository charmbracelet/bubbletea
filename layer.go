package tea

import (
	"image"
	"sort"
	"strings"

	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/x/ansi"
)

// Layers is a collection of Layer objects that can be rendered together.
type Layers []*Layer

// NewLayers creates a new Layers collection.
func NewLayers(layers ...*Layer) (l Layers) {
	l.AddLayers(append(Layers{}, layers...)...)
	return l
}

// AddLayers adds child layers to the Layers collection.
// It sorts the layers by z-index after adding them.
func (l *Layers) AddLayers(layers ...*Layer) *Layers {
	*l = append(*l, layers...)
	sortLayers(*l, false)
	return l
}

// InBounds returns true if the point is within the bounds of the Canvas.
func (l Layers) InBounds(x, y int) bool {
	return image.Pt(x, y).In(l.Bounds())
}

// Hit returns the [Layer.ID] at the given point. If no Layer is found,
// nil is returned.
func (l Layers) Hit(x, y int) *Layer {
	for i := len(l) - 1; i >= 0; i-- {
		if l[i].InBounds(x, y) {
			return l[i].Hit(x, y)
		}
	}
	return nil
}

func (l Layers) Len() int           { return len(l) }
func (l Layers) Less(i, j int) bool { return l[i].zIndex < l[j].zIndex }
func (l Layers) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// Bounds returns the bounds of all the layers combined.
func (l Layers) Bounds() image.Rectangle {
	// Figure out the size of the canvas
	x0, y0, x1, y1 := 0, 0, 0, 0
	for _, layer := range l {
		if layer.GetX() < x0 {
			x0 = layer.GetX()
		}
		if layer.GetY() < y0 {
			y0 = layer.GetY()
		}
		if layer.GetX()+layer.GetWidth() > x1 {
			x1 = layer.GetX() + layer.GetWidth()
		}
		if layer.GetY()+layer.GetHeight() > y1 {
			y1 = layer.GetY() + layer.GetHeight()
		}
	}

	// Adjust the size of the canvas if it's negative
	if x0 < 0 {
		x1 -= x0
		x0 = 0
	}
	if y0 < 0 {
		y1 -= y0
		y0 = 0
	}

	// Create a buffer with the size of the canvas
	width, height := x1-x0, y1-y0
	return image.Rect(x0, y0, x0+width, y0+height)
}

// Layer represents a window layer that can be composed with other layers.
type Layer struct {
	rect     image.Rectangle
	zIndex   int
	children []*Layer
	id       string
	content  string
}

// NewLayer creates a new Layer with the given content. It calculates the size
// based on the widest line and the number of lines in the content.
func NewLayer(content string) (l *Layer) {
	l = new(Layer)
	l.content = content
	var width int
	height := strings.Count(content, "\n") + 1
	for _, line := range strings.Split(content, "\n") {
		lineWidth := ansi.StringWidth(line)
		if lineWidth > width {
			width = lineWidth
		}
	}
	l.rect = image.Rect(0, 0, width, height)
	return l
}

// InBounds returns true if the point is within the bounds of the Layer.
func (l *Layer) InBounds(x, y int) bool {
	return image.Pt(x, y).In(l.Bounds())
}

// Bounds returns the bounds of the Layer.
func (l *Layer) Bounds() image.Rectangle {
	return l.rect
}

// Hit returns the [Layer.ID] at the given point. If no Layer is found,
// returns nil is returned.
func (l *Layer) Hit(x, y int) *Layer {
	// Reverse the order of the layers so that the top-most layer is checked
	// first.
	for i := len(l.children) - 1; i >= 0; i-- {
		if l.children[i].InBounds(x, y) {
			return l.children[i].Hit(x, y)
		}
	}

	if image.Pt(x, y).In(l.Bounds()) {
		return l
	}

	return nil
}

// ID sets the ID of the Layer. The ID can be used to identify the Layer when
// performing hit tests.
func (l *Layer) ID(id string) *Layer {
	l.id = id
	return l
}

// GetID returns the ID of the Layer.
func (l *Layer) GetID() string {
	return l.id
}

// X sets the x-coordinate of the Layer.
func (l *Layer) X(x int) *Layer {
	l.rect = l.rect.Add(image.Pt(x, 0))
	return l
}

// Y sets the y-coordinate of the Layer.
func (l *Layer) Y(y int) *Layer {
	l.rect = l.rect.Add(image.Pt(0, y))
	return l
}

// Z sets the z-index of the Layer.
func (l *Layer) Z(z int) *Layer {
	l.zIndex = z
	return l
}

// GetX returns the x-coordinate of the Layer.
func (l *Layer) GetX() int {
	return l.rect.Min.X
}

// GetY returns the y-coordinate of the Layer.
func (l *Layer) GetY() int {
	return l.rect.Min.Y
}

// GetZ returns the z-index of the Layer.
func (l *Layer) GetZ() int {
	return l.zIndex
}

// Width sets the width of the Layer.
func (l *Layer) Width(width int) *Layer {
	l.rect.Max.X = l.rect.Min.X + width
	return l
}

// Height sets the height of the Layer.
func (l *Layer) Height(height int) *Layer {
	l.rect.Max.Y = l.rect.Min.Y + height
	return l
}

// GetWidth returns the width of the Layer.
func (l *Layer) GetWidth() int {
	return l.rect.Dx()
}

// GetHeight returns the height of the Layer.
func (l *Layer) GetHeight() int {
	return l.rect.Dy()
}

// AddLayers adds child layers to the Layer.
func (l *Layer) AddLayers(layers ...*Layer) *Layer {
	// Make children relative to the parent
	for _, child := range layers {
		child.rect = child.rect.Add(l.rect.Min)
		child.zIndex += l.zIndex
	}
	l.children = append(l.children, layers...)
	sortLayers(l.children, false)
	return l
}

// SetContent sets the content of the Layer.
func (l *Layer) SetContent(content string) *Layer {
	l.content = content
	return l
}

// Content returns the content of the Layer.
func (l *Layer) Content() string {
	return l.content
}

// Get returns the Layer with the given ID. If the ID is not found, it returns
// nil.
func (l *Layer) Get(id string) *Layer {
	if l.id == id {
		return l
	}
	for _, child := range l.children {
		if child.id == id {
			return child
		}
	}
	return nil
}

// sortLayers sorts the layers by z-index, from lowest to highest.
func sortLayers(ls []*Layer, reverse bool) {
	if reverse {
		sort.Stable(sort.Reverse(Layers(ls)))
	} else {
		sort.Stable(Layers(ls))
	}
}

// renderLayers renders the layers into the given buffer using the specified width method.
func renderLayers(l []*Layer, buf *tv.Buffer, method ansi.Method) {
	// Render the view layers into the buffer.
	for _, layer := range l {
		if layer == nil {
			continue
		}
		area := layer.Bounds()
		buf.ClearArea(area)
		ss := tv.NewStyledString(method, layer.Content())
		ss.RenderComponent(buf, area) //nolint:errcheck,gosec
	}
}
