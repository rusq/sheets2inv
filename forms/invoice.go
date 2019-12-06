package forms

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// measurements
const (
	// items offset
	addressOffsetY = 17.0 // 17 mm from top

	// formatting
	defCellHeight = 6.0 // 6 mm
	lineSizeMul   = 1.5 // 1.5 * normal line

	// drawing
	lThick = 6.0
	lThin  = 0.2

	// image
	imgW = 18.00 //mm
)

//dimensions
const (
	entrySz    = 4 // 4 columns in the entry table: description, qty, price, total
	totalSz    = 2 // 2 values in total row, i.e. SUBTOTAL, 42
	keyValueSz = 2 // key: value pair array size.
)

// table columns
var (
	tabCol   = []string{"#", "DESCRIPTION", "QTY (hrs)", "UNIT PRICE", "TOTAL"}
	tabColPc = []float64{0.05, 0.55, 0.15, 0.15, 0.10}
)

// InvoiceFields contains all fields needed for the invoice.
type InvoiceFields struct {
	Date time.Time `yaml:"invoice_date"`
	Due  time.Time `yaml:"due_date"`

	PeriodStart time.Time `yaml:"period_start"`
	PeriodEnd   time.Time `yaml:"period_end"`

	Bank    string
	Account string

	Address Address
	BillTo  Address `yaml:"bill_to"`

	Image string

	Remarks string // note at the end of the invoice

}

// InvoiceForm is the form
type InvoiceForm struct {
	pdf       *gofpdf.Fpdf
	invoiceID string

	m Margins
	s Style

	//
	// +-----------+
	// |     _     |
	// |     ↑     |
	// |    w|     |
	// | |<--+-->| |
	// |     |     |
	// |    h|     |
	// |     ↓     |
	// |     -     |
	// +-----------+ <- maxX, maxY

	maxX, maxY float64 // page max X, max Y
	w, h       float64 // text area width, height

	// page elements
	account   [][keyValueSz]string
	entries   [][entrySz]string
	subTotals [][totalSz]string
	total     [totalSz]string
}

// NewInvoice creates an invoice form.
func NewInvoice(invoiceID string, sizeStr string, m *Margins, s *Style) *InvoiceForm {
	if invoiceID == "" {
		invoiceID = time.Now().Format(defInvoiceNoFmt)
	}
	if sizeStr == "" {
		sizeStr = PgA4
	}
	if m == nil {
		m = &defMargins
	}
	if s == nil {
		s = &defStyle
	}
	pdf := gofpdf.New("P", "mm", sizeStr, "")
	pdf.SetMargins(m.Left, m.Top, m.Right)
	pdf.SetAutoPageBreak(true, m.Bottom)
	pdf.AddPage()

	maxX, maxY := pdf.GetPageSize()

	inv := &InvoiceForm{
		pdf:       pdf,
		invoiceID: invoiceID,
		m:         *m, // margins
		s:         *s, // style

		// dimensions
		maxX: maxX,
		maxY: maxY,
		w:    maxX - (m.Left + m.Right),
		h:    maxY - (m.Bottom + m.Top),

		// page elements
		account:   make([][keyValueSz]string, 0),
		entries:   make([][entrySz]string, 0),
		subTotals: make([][totalSz]string, 0),
	}

	inv.line(inv.s.AccentColor, lThick, 0, 0, inv.maxY, 0)
	inv.line(inv.s.AccentColor, lThick, 0, inv.maxY, inv.maxX, inv.maxY)

	return inv
}

// Fpdf returns underlying Fpdf instance.
func (f *InvoiceForm) Fpdf() *gofpdf.Fpdf {
	return f.pdf
}

func (f *InvoiceForm) margins() Margins {
	return f.m
}

// line draws a line from (x0,y0) to (x1,y1)
func (f *InvoiceForm) line(col color.RGBA, width, x0, y0, x1, y1 float64) {
	f.pdf.SetDrawColor(toRGB(col))
	f.pdf.MoveTo(x0, y0)
	f.pdf.LineTo(x1, y1)
	f.pdf.SetLineWidth(width)
	f.pdf.DrawPath("DF")
}

func (f *InvoiceForm) cell(cw, ch float64, str, borderStr string, ln int, alignStr string, fill bool, link int, linkStr string) {
	if cw == 0 {
		cw = f.pdf.GetStringWidth(str)
	}
	if ch == 0 {
		_, ch = f.pdf.GetFontSize()
	}
	f.pdf.CellFormat(cw, ch, str, borderStr, ln, alignStr, fill, link, linkStr)
}

// writeAt returns the writer at X with cell width of cw x ch
func (f *InvoiceForm) writeAt(x, cw, ch float64, borderStr string, ln int, alignStr string, fill bool, link int, linkStr string) func(str string) {
	return func(str string) {
		f.pdf.SetX(x)
		f.cell(cw, ch, str, borderStr, ln, alignStr, fill, link, linkStr)
	}
}

func (f *InvoiceForm) heading5(x, y, lineSz float64, str string) {
	var (
		pt     = f.s.Heading.Size
		offset = 0.0
	)
	var cy = f.pdf.PointConvert(pt)
	oldColor := fromRGB(f.pdf.GetTextColor())
	f.pdf.SetXY(x, y)
	f.pdf.SetFont(f.s.Heading.font())
	f.pdf.SetTextColor(f.s.Heading.fg())
	f.cell(0, cy, str, "0", 0, "L", false, 0, "")
	f.pdf.Ln(-1)
	if lineSz == 0.0 {
		lineSz = f.pdf.GetStringWidth(str)
	}
	f.line(cBlack, lThin, x, f.pdf.GetY()-offset, x+lineSz, f.pdf.GetY()-offset)
	f.pdf.Ln(cy + lThin)
	f.pdf.SetTextColor(toRGB(oldColor))
}

func (f *InvoiceForm) printAddrAt(x, y float64, addr Address) {
	var writeln = f.writeAt(x, 0.0, f.pdf.PointConvert(f.s.Address.Size)*f.s.LineSpacing, "", 1, "L", false, 0, "")

	f.pdf.SetXY(x, y)
	f.pdf.SetFont(f.s.Address.FamilyStr, "B", f.s.Address.Size)
	f.pdf.SetTextColor(toRGB(f.s.Address.Foreground))
	if addr.Name != "" {
		writeln(addr.Name)
	}
	if addr.Organisation != "" {
		writeln(addr.Organisation)
	}
	f.pdf.SetFont("", "", 0)
	writeln(nvljoin([]string{addr.Floor, addr.Street}, delim))
	writeln(strings.Join([]string{addr.Suburb, addr.Town, addr.Postcode}, delim))
	f.pdf.Ln(f.pdf.PointConvert(f.s.Address.Size+1) / 2)
	writeln(addr.Phone)
	writeln(addr.Email)
}

// keyValuesAt prints formatted key values val at x,y.
func (f *InvoiceForm) keyValuesAt(x, y float64, val [][keyValueSz]string, attr [keyValueSz]string, align [keyValueSz]string) {
	_, cellH := f.pdf.GetFontSize()
	cellH *= f.s.LineSpacing
	colsSz := [keyValueSz]float64{}
	for row := range val {
		for col := range val[row] {
			curSz := colsSz[col]
			if newSz := f.pdf.GetStringWidth(val[row][col]); newSz > curSz {
				colsSz[col] = newSz
			}
		}
	}
	f.pdf.SetY(y)
	for row := range val {
		f.pdf.SetX(x)
		for col, str := range val[row] {
			f.pdf.SetFont("", attr[col], 0)
			f.cell(colsSz[col], cellH, str, "", 0, align[col], false, 0, "")
		}
		f.pdf.Ln(-1)
	}
}

func (f *InvoiceForm) title(x, y float64) (lastX, lastY float64) {
	f.pdf.MoveTo(x, y)
	f.pdf.SetFont(f.s.Title.font())
	f.pdf.SetTextColor(f.s.Title.fg())
	// CellFormat(width, height, text, border, position after, align, fill, link, linkStr)
	f.cell(190, 7, "INVOICE", "0", 0, "LM", false, 0, "")
	f.pdf.Ln(-1)
	// f.line(cBlack, lThin, x, f.pdf.GetY(), f.w+x, f.pdf.GetY())
	return f.pdf.GetXY()
}

// Generate generates a PDF file
func (f *InvoiceForm) Generate(filename string, fields *InvoiceFields) error {
	val := *fields

	// HEADER
	f.title(f.m.Left, f.m.Top)

	f.imageAt(f.maxX-f.m.Right-imgW, f.m.Top, imgW, 0, val.Image)

	// invoice details
	f.invoiceData(f.maxX-f.m.Right-(f.w*0.18), f.m.Top+addressOffsetY+defCellHeight, f.invoiceID, val.Date, val.Due)

	// Address
	x, y := f.address(f.m.Left, f.m.Top+addressOffsetY, val.Address)
	// PAYMENT ACCOUNT DETAILS at the same Y as BILL TO
	f.accountDetails(f.maxX/2, y)
	// BILL TO
	x, y = f.billToAddress(x, y, val.BillTo)

	//Time period
	x, y = f.timePeriod(f.m.Left, y+defCellHeight, val.PeriodStart, val.PeriodEnd)

	// Table
	_, tabY := f.table(x, y)

	// Remarks
	f.remarks(f.m.Left, tabY, f.w*(tabColPc[1]), val.Remarks)

	return f.pdf.OutputFileAndClose(filename)
}

func (f *InvoiceForm) imageAt(x, y, w, h float64, filename string) (lastX, lastY float64) {
	f.pdf.ImageOptions(filename, x, y, w, h, false, gofpdf.ImageOptions{}, 0, "")
	return f.pdf.GetXY()
}

func (f *InvoiceForm) invoiceData(x, y float64, id string, date, due time.Time) (lastX, lastY float64) {
	f.pdf.SetTextColor(f.s.Body.fg())
	f.pdf.SetFont(f.s.Body.font())
	f.keyValuesAt(x, y,
		[][keyValueSz]string{
			{"No.:", id},
			{"Date:", date.Format(dateFmt)},
			{"", ""},
			{"Due: ", due.Format(dateFmt)},
		},
		[keyValueSz]string{"", "B"}, [keyValueSz]string{"R", "L"},
	)
	return f.pdf.GetXY()
}

func (f *InvoiceForm) address(x, y float64, addr Address) (lastX, lastY float64) {
	f.printAddrAt(x, y, addr)
	f.pdf.Ln(-1)
	return f.pdf.GetXY()
}

func (f *InvoiceForm) billToAddress(x, y float64, addr Address) (lastX, lastY float64) {
	f.heading5(x, y, 60.0, "BILL TO")
	f.printAddrAt(x, f.pdf.GetY(), addr)
	return f.pdf.GetXY()
}

func (f *InvoiceForm) accountDetails(x, y float64) (lastX, lastY float64) {
	f.heading5(x, y, 60.0, "PAYMENT ACCOUNT DETAILS")
	f.keyValuesAt(x, f.pdf.GetY(),
		f.account,
		[keyValueSz]string{"", "B"}, [keyValueSz]string{"L", "L"},
	)
	return f.pdf.GetXY()
}

func (f *InvoiceForm) timePeriod(x, y float64, periodStart, periodEnd time.Time) (lastX, lastY float64) {
	f.pdf.SetFont(f.s.Body.font())
	period := fmt.Sprintf("%s - %s",
		periodStart.Format(dateFmt),
		periodEnd.Format(dateFmt),
	)
	f.keyValuesAt(x, y,
		[][keyValueSz]string{
			{"Time period: ", period},
		},
		[keyValueSz]string{"", "B"}, [keyValueSz]string{"L", "L"})
	f.pdf.Ln(0.5)
	return f.pdf.GetXY()
}

func (f *InvoiceForm) table(x, y float64) (lastX, lastY float64) {

	const tabCellH = defCellHeight

	f.pdf.SetXY(x, y)
	fillColor := fromRGB(f.pdf.GetFillColor())
	f.pdf.SetFont(f.s.Heading.font())
	f.pdf.SetFillColor(f.s.TableHead.bg())
	f.pdf.SetTextColor(f.s.TableHead.fg())
	for i := range tabCol {
		f.pdf.CellFormat(f.w*tabColPc[i], tabCellH, tabCol[i], "TB", 0, "C", true, 0, "")
	}
	f.pdf.SetFillColor(toRGB(fillColor))
	f.pdf.Ln(-1)
	f.pdf.SetFont(f.s.TableRowOdd.font())
	f.pdf.SetTextColor(f.s.TableRowOdd.fg())
	idx := 1
	for _, entry := range f.entries {
		y := f.pdf.GetY()
		f.pdf.SetX(f.m.Left + f.w*tabColPc[0])
		f.pdf.MultiCell(f.w*tabColPc[1], tabCellH, entry[0], "B", "L", false)
		Δy := f.pdf.GetY() - y
		f.pdf.SetXY(f.m.Left+f.w*(tabColPc[0]+tabColPc[1]), y) // reset line
		// numbers
		for col := 1; col < len(entry); col++ {
			f.pdf.CellFormat(f.w*tabColPc[col+1], Δy, entry[col], "B", 0, "R", false, 0, "")
		}
		// item #
		f.pdf.SetX(f.m.Left)
		f.pdf.CellFormat(f.w*tabColPc[0], Δy, fmt.Sprintf("%d", idx), "B", 1, "R", false, 0, "")
		idx++
	}

	lastX, lastY = f.pdf.GetXY()

	// Totals
	// totalLine: ch - cell height
	var totalLine = func(ch float64, sz float64, name, value string) {
		f.pdf.SetX(f.m.Left + f.w*(tabColPc[0]+tabColPc[1]))
		f.pdf.SetFont(f.s.TableRowOdd.FamilyStr, "B", sz)
		f.cell(f.w*(tabColPc[2]+tabColPc[3]), ch, name, "B", 0, "L", false, 0, "")
		f.pdf.SetFont("", "", 0)
		f.cell(f.w*tabColPc[4], ch, value, "B", 0, "R", false, 0, "")
		f.pdf.Ln(-1)
	}
	for _, subtotal := range f.subTotals {
		totalLine(tabCellH, 0, subtotal[0], subtotal[1])
	}
	totalLine(tabCellH+2, f.s.TableRowOdd.Size+2, f.total[0], f.total[1])

	return lastX, lastY
}

func (f *InvoiceForm) remarks(x, y, cw float64, remark string) (lastX, lastY float64) {
	const remCh = defCellHeight
	f.pdf.SetXY(x, y)
	f.pdf.Ln(-1)
	f.pdf.SetFont(f.s.Body.font())
	f.pdf.SetFont("", "B", 0)
	f.cell(cw, remCh, "Remarks/Payment instructions", "0", 1, "L", false, 0, "")
	f.pdf.SetFont("", "", 0) // cancel bold
	f.pdf.MultiCell(f.w*(tabColPc[1]), remCh, remark, "", "", false)
	return f.pdf.GetXY()
}

// AddEntry adds an entry to the Invoice item list.
func (f *InvoiceForm) AddEntry(description, qty, price, total string) *InvoiceForm {
	f.entries = append(f.entries, [entrySz]string{description, qty, price, total})
	return f
}

// AddSubTotal adds a total to the invoice form
func (f *InvoiceForm) AddSubTotal(name, value string) *InvoiceForm {
	f.subTotals = append(f.subTotals, [totalSz]string{name, value})
	return f
}

// SetTotal sets the total line name and value.
func (f *InvoiceForm) SetTotal(name, value string) *InvoiceForm {
	if name == "" {
		name = "BALANCE DUE"
	}
	f.total = [totalSz]string{name, value}
	return f
}

// AddAccountDetail adds the account information row.
func (f *InvoiceForm) AddAccountDetail(name, value string) *InvoiceForm {
	f.account = append(f.account, [keyValueSz]string{name + " ", value})
	return f
}

func nvljoin(ss []string, delim string) string {
	var buf strings.Builder
	for _, s := range ss {
		if s != "" {
			empty := (buf.Len() == 0)
			if !empty {
				buf.WriteString(delim)
			}
			buf.WriteString(s)
		}
	}
	return buf.String()
}
