// Package colorizer provides color schemes for dyff output.
// The default scheme uses a blue/orange palette from the Okabe-Ito set,
// which is distinguishable across all common forms of color blindness
// (including red-green color vision deficiency, the most common form,
// affecting ~8% of males).
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

// OkabeIto implements dyff.Colorizer using a blue/orange/yellow palette
// from the Okabe-Ito set, which is the default color scheme.
type OkabeIto struct{}

// compile-time check that OkabeIto satisfies the interface
var _ dyff.Colorizer = (*OkabeIto)(nil)

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

func (c *OkabeIto) Green(text string) string             { return colored(additionBlue, text) }
func (c *OkabeIto) Greenf(f string, a ...interface{}) string { return coloredf(additionBlue, f, a...) }
func (c *OkabeIto) Red(text string) string               { return colored(removalOrange, text) }
func (c *OkabeIto) Redf(f string, a ...interface{}) string   { return coloredf(removalOrange, f, a...) }
func (c *OkabeIto) Yellowf(f string, a ...interface{}) string { return coloredf(modifyYellow, f, a...) }
func (c *OkabeIto) LightGreen(text string) string        { return colored(lightBlue, text) }
func (c *OkabeIto) LightRed(text string) string          { return colored(lightOrange, text) }

func (c *OkabeIto) DimGray(text string) string {
	return colored(bunt.DimGray, text)
}

func (c *OkabeIto) Bold(text string) string {
	return bunt.Style(text, bunt.EachLine(), bunt.Bold())
}

func (c *OkabeIto) Italic(text string) string {
	return bunt.Style(text, bunt.EachLine(), bunt.Italic())
}

func (c *OkabeIto) BoldGreen(text string) string { return c.Bold(c.Green(text)) }
func (c *OkabeIto) BoldRed(text string) string   { return c.Bold(c.Red(text)) }

func (c *OkabeIto) StylizeHeader(header string) string {
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

func (c *OkabeIto) YAMLInGreenishColors(input interface{}, useIndentLines bool) (string, error) {
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

func (c *OkabeIto) YAMLInRedishColors(input interface{}, useIndentLines bool) (string, error) {
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
