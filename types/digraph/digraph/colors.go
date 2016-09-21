package digraph

type ColorStore struct {
	colors map[string]int
	labels []string
}

func (c *ColorStore) Color(label string) int {
	if color, has := c.colors[label]; has {
		return color
	} else {
		color = len(c.labels)
		c.colors[label] = color
		c.labels = append(label)
		return color
	}
}

func (c *ColorStore) Label(color int) string {
	if color >= len(c.labels) {
		return "---no-label---"
	}
	return c.labels[color]
}

