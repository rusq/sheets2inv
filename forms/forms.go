package forms

import (
	"image/color"
)

// literal constants
const (
	// Fonts
	helvetica = "Helvetica"
	times     = "Times"
	courier   = "Courier"

	PgA4     = "A4"
	PgLetter = "Letter"

	dateFmt = "02/01/2006"
	delim   = ", "

	// invoice specific
	defInvoiceNoFmt = "20060102"
)

// colors
var (
	cBlack     = color.RGBA{0, 0, 0, 0}
	cLightGray = color.RGBA{191, 191, 191, 0}
	cWhite     = color.RGBA{255, 255, 255, 0}

	cLightBlue = color.RGBA{75, 118, 210, 0}
	cBlue      = color.RGBA{36, 55, 97, 0}

	cTeal = color.RGBA{73, 123, 158, 0}

	cDefault = color.RGBA{255, 255, 255, 255}
)

type Address struct {
	Name         string
	Organisation string
	Floor        string
	Street       string
	Suburb       string
	Town         string
	Postcode     string
	Country      string

	Phone string
	Email string
}

// Style is a form style
type Style struct {
	Title   Font
	Heading Font
	Body    Font
	Address Font

	TableHead    Font
	TableRowOdd  Font
	TableRowEven Font

	AccentColor color.RGBA

	LineSpacing float64
}

// Font represents a font
type Font struct {
	FamilyStr  string
	StyleStr   string
	Size       float64 //size in pt
	Foreground color.RGBA
	Background color.RGBA
}

// Margins define page margins
type Margins struct {
	Left, Top, Right, Bottom float64
}

var defMargins = Margins{25.7, 25.7, 25.7, 25.7}
var defStyle = Style{
	Title:   Font{helvetica, "", 16, cLightGray, cDefault},
	Heading: Font{helvetica, "", 10, cTeal, cDefault},
	Body:    Font{helvetica, "", 10, cBlack, cDefault},
	Address: Font{helvetica, "", 10, cBlack, cDefault},

	TableHead:    Font{helvetica, "", 10, cWhite, cTeal},
	TableRowOdd:  Font{helvetica, "", 9, cBlack, cDefault},
	TableRowEven: Font{helvetica, "", 9, cBlack, cDefault},

	AccentColor: cTeal,

	LineSpacing: 1.5,
}

func toRGB(c color.RGBA) (r, g, b int) {
	return int(c.R), int(c.G), int(c.B)
}

func fromRGB(r int, g int, b int) color.RGBA {
	return color.RGBA{uint8(r), uint8(g), uint8(b), 0}
}

// font returns measurements for gofpdf.SetFont()
func (f Font) font() (family, style string, size float64) {
	return f.FamilyStr, f.StyleStr, f.Size
}

// fg returns the foreground colour
func (f Font) fg() (r int, g int, b int) {
	return toRGB(f.Foreground)
}

// bg returns the background colour
func (f Font) bg() (r int, g int, b int) {
	return toRGB(f.Background)
}
