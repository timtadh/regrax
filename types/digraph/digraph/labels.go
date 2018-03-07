package digraph

import (
	"fmt"
)

type Labels struct {
	colors map[string]int
	labels []string
}

func NewLabels() *Labels {
	return &Labels{
		colors: make(map[string]int, 1000),
		labels: make([]string, 0, 1000),
	}
}

func (c *Labels) Color(label string) int {
	if color, has := c.colors[label]; has {
		return color
	} else {
		color = len(c.labels)
		c.colors[label] = color
		c.labels = append(c.labels, label)
		return color
	}
}

func (c *Labels) Label(color int) string {
	if color < 0 || color >= len(c.labels) {
		return fmt.Sprintf("color-[%d]", color)
	}
	return c.labels[color]
}

func (c *Labels) Labels() []string {
	return c.labels
}
