package tea

type HitMsg *Layer

type Layer struct {
	// The name of the layer.
	Name     string
	X, Y     int
	Z        int
	Content  string
	Layers   []Layer
	Hittable bool
}

type Layers struct {
	layers []Layer
}

func (l *Layers) Hit(x, y int) *Layer {
	for _, layer := range l.layers {
		if x >= layer.X && x < layer.X+len(layer.Content) && y == layer.Y {
			return &layer
		}
	}
	return nil
}

type View struct {
	Layers
	*Cursor
}

func (m Model) View() View {
	if !m.hideCursor {
		v.Cursor = m.cursor
	}
	layers := []Layer{
		{
			Name:    "main",
			X:       0,
			Y:       0,
			Z:       0,
			Content: m.Buffer.String(),
		},
		{
			Name:    "box",
			X:       m.box.X,
			Y:       m.box.Y,
			Z:       1,
			Content: m.box.Content,
		},
	}
	if layer := v.layers.Hit(m.lastX, m.lastY); layer != nil {
		// Drag and do this...
	}
	return layers
}
