package color

import (
	"hash/fnv"

	"github.com/fatih/color"
)

var (
	// GroupColor has LogGroupName and color mapping
	GroupColor = groupColorMap{}
	// StreamColor has LogStreamName and color mapping
	StreamColor = streamColorMap{}

	groupColorList = []*color.Color{
		color.New(color.FgHiCyan),
		color.New(color.FgHiGreen),
		color.New(color.FgHiMagenta),
		color.New(color.FgHiYellow),
		color.New(color.FgHiBlue),
		color.New(color.FgHiRed),
	}

	streamColorList = []*color.Color{
		color.New(color.FgCyan),
		color.New(color.FgGreen),
		color.New(color.FgMagenta),
		color.New(color.FgYellow),
		color.New(color.FgBlue),
		color.New(color.FgRed),
	}
)

func determineColor(name string) uint32 {
	hash := fnv.New32()
	hash.Write([]byte(name))
	return hash.Sum32() % uint32(len(groupColorList))
}

type groupColorMap map[string]*color.Color

func (cm groupColorMap) Get(name string) (c *color.Color) {
	var ok bool
	if c, ok = cm[name]; !ok {
		c = groupColorList[determineColor(name)]
		cm[name] = c
	}
	return c
}

type streamColorMap map[string]*color.Color

func (cm streamColorMap) Get(name string) (c *color.Color) {
	var ok bool
	if c, ok = cm[name]; !ok {
		c = streamColorList[determineColor(name)]
		cm[name] = c
	}
	return c
}
