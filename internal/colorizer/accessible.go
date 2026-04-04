// Package colorizer provides an accessible color scheme for dyff output.
// The default dyff palette uses red for removals and green for additions,
// which are indistinguishable for people with red-green color vision
// deficiency (the most common form, affecting ~8% of males). This package
// provides a blue/orange palette from the Okabe-Ito set, which is
// distinguishable across all common forms of color blindness.
//
// dyff also renders +/-/± symbols alongside color, so tritanopes and
// people with achromatopsia retain non-color differentiation regardless
// of palette choice.
package colorizer

import (
	"fmt"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"
	"github.com/homeport/dyff/pkg/dyff"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// Accessible implements dyff.Colorizer using an Okabe-Ito blue/orange
// palette in place of the default red/green one.
type Accessible struct{}

// compile-time check that Accessible satisfies the interface
var _ dyff.Colorizer = (*Accessible)(nil)

var (
	additionBlue   = hex("#0072B2") // Okabe-Ito blue   — additions
	removalOrange  = hex("#D55E00") // Okabe-Ito orange  — removals
	modifyYellow   = hex("#F0E442") // Okabe-Ito yellow  — modifications
	lightBlue      = hex("#56B4E9") // Okabe-Ito sky blue — light additions
	lightOrange    = hex("#E8956D") // lighter orange     — light removals
)

func hex(s string) colorful.Color {
	c, _ := colorful.Hex(s)
	return c
}

func colored(color colorful.Color, text string) string {
	return bunt.Style(text, bunt.EachLine(), bunt.Foreground(color))
}

func coloredf(color colorful.Color, format string, a ...interface{}) string {
	var text string
	if len(a) == 0 {
		text = format
	} else {
		text = fmt.Sprintf(format, a...)
	}
	return bunt.Style(text, bunt.EachLine(), bunt.Foreground(color))
}

func (c *Accessible) Green(text string) string             { return colored(additionBlue, text) }
func (c *Accessible) Greenf(f string, a ...interface{}) string { return coloredf(additionBlue, f, a...) }
func (c *Accessible) Red(text string) string               { return colored(removalOrange, text) }
func (c *Accessible) Redf(f string, a ...interface{}) string   { return coloredf(removalOrange, f, a...) }
func (c *Accessible) Yellowf(f string, a ...interface{}) string { return coloredf(modifyYellow, f, a...) }
func (c *Accessible) LightGreen(text string) string        { return colored(lightBlue, text) }
func (c *Accessible) LightRed(text string) string          { return colored(lightOrange, text) }

func (c *Accessible) DimGray(text string) string {
	return colored(bunt.DimGray, text)
}

func (c *Accessible) Bold(text string) string {
	return bunt.Style(text, bunt.EachLine(), bunt.Bold())
}

func (c *Accessible) Italic(text string) string {
	return bunt.Style(text, bunt.EachLine(), bunt.Italic())
}

func (c *Accessible) BoldGreen(text string) string { return c.Bold(c.Green(text)) }
func (c *Accessible) BoldRed(text string) string   { return c.Bold(c.Red(text)) }

func (c *Accessible) StylizeHeader(header string) string {
	return bunt.Style(
		header,
		bunt.ForegroundFunc(func(x int, _ int, _ rune) *colorful.Color {
			switch {
			case x < 7:
				blue := additionBlue
				return &blue
			case x < 13:
				yellow := modifyYellow
				return &yellow
			case x < 21:
				orange := removalOrange
				return &orange
			}
			return nil
		}),
	)
}

func (c *Accessible) YAMLInGreenishColors(input interface{}, useIndentLines bool) (string, error) {
	return neat.NewOutputProcessor(useIndentLines, true, &map[string]colorful.Color{
		"keyColor":           additionBlue,
		"indentLineColor":    hex("#003D60"),
		"scalarDefaultColor": lightBlue,
		"boolColor":          lightBlue,
		"floatColor":         lightBlue,
		"intColor":           lightBlue,
		"multiLineTextColor": hex("#2D8BBF"),
		"nullColor":          hex("#4A9FCC"),
		"emptyStructures":    hex("#7AC5E0"),
		"dashColor":          additionBlue,
	}).ToYAML(input)
}

func (c *Accessible) YAMLInRedishColors(input interface{}, useIndentLines bool) (string, error) {
	return neat.NewOutputProcessor(useIndentLines, true, &map[string]colorful.Color{
		"keyColor":           removalOrange,
		"indentLineColor":    hex("#6B2F00"),
		"scalarDefaultColor": lightOrange,
		"boolColor":          lightOrange,
		"floatColor":         lightOrange,
		"intColor":           lightOrange,
		"multiLineTextColor": hex("#C07840"),
		"nullColor":          hex("#D4884A"),
		"emptyStructures":    hex("#E8A87C"),
		"dashColor":          removalOrange,
	}).ToYAML(input)
}
