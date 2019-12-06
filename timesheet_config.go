package sheet2inv

import (
	"errors"
	"strings"
	"time"

	"github.com/rusq/sheet2inv/forms"
	"github.com/shopspring/decimal"
)

// TimesheetConfig is the configuration of the Invoice output.
type TimesheetConfig struct {
	Spreadsheet *Spreadsheet
	Values      *InvoiceValues `yaml:"invoice"`
}

// InvoiceParameters contains invoice parameters.
type InvoiceValues struct {
	Rate            decimal.Decimal `yaml:"hourly_rate"`
	Tax             decimal.Decimal `yaml:"tax_rate"`
	Shipping        decimal.Decimal
	UsePrevMonth    bool                `yaml:"use_previous_month"` // if defined, date fields are ignored
	PrevMonthDueDay int                 `yaml:"due_day"`            // day of the month for due date
	InvoiceFields   forms.InvoiceFields `yaml:"invoice_fields"`
	IssueSummary    map[string]string   `yaml:"issue_summary,omitempty"`
}

// Spreadsheet is the source spreadsheet parameters.
type Spreadsheet struct {
	ID      string
	Range   string
	Columns Columns
}

// Columns is the column index within the spreadsheet.
type Columns struct {
	TimeStart   string `yaml:"time_start"`
	TimeEnd     string `yaml:"time_end"`
	Invoice     string
	Description string
	Issue       string

	// calculated column indexes
	start int
	end   int
	inv   int
	descr int
	issue int
}

func (ci *Columns) resolve() (err error) {
	defer func() {
		if r := recover(); r != nil {
			msg, ok := r.(string)
			if !ok {
				panic(r)
			}
			err = errors.New(msg)
		}
	}()

	ci.start = ci.char2int(ci.TimeStart)
	ci.end = ci.char2int(ci.TimeEnd)
	ci.inv = ci.char2int(ci.Invoice)
	ci.descr = ci.char2int(ci.Description)
	ci.issue = ci.char2int(ci.Issue)

	return
}

func (*Columns) char2int(char string) int {
	if len(char) == 0 {
		panic("empty column value")
	}
	char = strings.ToUpper(char)
	if char[0] < 'A' || 'Z' < char[0] {
		panic("character out of range")
	}
	return int(char[0] - 'A')
}

func (v *InvoiceValues) adjustDates() error {
	if !v.UsePrevMonth {
		return nil
	}
	if v.PrevMonthDueDay < 1 || 31 < v.PrevMonthDueDay {
		return errors.New("invalid prev_month_due_day")
	}
	today := time.Now()
	thisMoStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := thisMoStart.Add(-24 * time.Hour)
	start := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(today.Year(), today.Month(), v.PrevMonthDueDay, 0, 0, 0, 0, time.UTC)

	v.InvoiceFields.Date = today
	v.InvoiceFields.PeriodStart = start
	v.InvoiceFields.PeriodEnd = end
	v.InvoiceFields.Due = due

	return nil
}
