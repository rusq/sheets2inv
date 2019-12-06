package sheet2inv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/rusq/sheet2inv/bugtracker"
	"github.com/shopspring/decimal"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/yaml.v3"
)

// indices
const (
	idxTimeStart = iota
	idxTimeEnd
	idxInvoice
	idxDescription
	idxIssue
)

const defMultiplier = 1.0

const (
	datetimeRenderOption = "SERIAL_NUMBER"
	valueRenderOption    = "UNFORMATTED_VALUE"
)

// Timesheet is the timesheet.
type Timesheet struct {
	Entries []*TsEntry

	config *TimesheetConfig
}

// TsEntry is a timesheet entry
type TsEntry struct {
	Invoice string
	Start   time.Time
	End     time.Time
	Items   []Item

	// calculated
	Duration time.Duration // total entry duration

	rate       decimal.Decimal // hourly rate for this item
	multiplier decimal.Decimal // multiplier
}

// Item is an paricular task/issue/ticket entry within one timesheet Item
// group.
type Item struct {
	Issue       string `json:",omitempty" yaml:",omitempty"`
	Description string `json:",omitempty" yaml:",omitempty"`

	duration time.Duration
}

func asString(v interface{}) string {
	return fmt.Sprint(v)
}

// NewFromSheets creates timesheet from Google Sheets.
func NewFromSheets(srv *sheets.Service, cfg *TimesheetConfig, invoiceID string) (*Timesheet, error) {
	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	resp, err := srv.Spreadsheets.Values.
		Get(cfg.Spreadsheet.ID, cfg.Spreadsheet.Range).
		ValueRenderOption(valueRenderOption).
		DateTimeRenderOption(datetimeRenderOption).
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	timesheet := New(cfg)

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for i, row := range resp.Values {
			if invoiceID != "" && asString(value(row, idxInvoice)) != invoiceID {
				continue
			}
			if err := timesheet.AddRow(row); err != nil {
				return nil, fmt.Errorf("row: %d: %s", i, err)
			}
		}
	}
	return timesheet, nil
}

// New creates a new timesheet with the rate.
func New(cfg *TimesheetConfig) *Timesheet {
	return &Timesheet{config: cfg}
}

// MarshalTo marshalls the timesheet to provided writer.
func (ts *Timesheet) MarshalTo(output io.Writer) error {
	return ts.ToYAML(output)
}

// Add adds item to the timesheet.  If the item misses start date, it appends
// the description
func (ts *Timesheet) Add(e *TsEntry) *Timesheet {
	// for flexibility, maybe include support in having these as columns
	// in GSheets.
	e.rate = ts.config.Values.Rate
	e.multiplier = decimal.NewFromFloat(defMultiplier)

	e.Recalculate()

	if !(e.Start.IsZero() && e.End.IsZero()) {
		return ts.append(e)
	}
	if e.Start.IsZero() {
		return ts.update(e, len(ts.Entries)-1)
	}
	return ts
}

// append appends an entry to timesheet.
func (ts *Timesheet) append(e *TsEntry) *Timesheet {
	ts.Entries = append(ts.Entries, e)
	return ts
}

// update updates tasks and end date on an entry in the timesheet.  If
// invoice id is different, will update it and print warning message.
func (ts *Timesheet) update(e *TsEntry, index int) *Timesheet {
	item := ts.Entries[index]

	if !e.End.IsZero() {
		item.End = e.End
	}
	if item.Invoice != e.Invoice {
		log.Printf("warning, different invoices within one entry: %v and %v", item, e)
		item.Invoice = e.Invoice
	}

	item.Items = append(item.Items, e.Items...)

	return ts
}

// AddRow parses row and adds resulting item to the timesheet.
func (ts *Timesheet) AddRow(row []interface{}) error {
	item, err := ts.parse(row)
	if err != nil {
		return err
	}
	ts.Add(item)
	return nil
}

// Recalculate recalculates duration and price for a single entry.
func (e *TsEntry) Recalculate() *TsEntry {
	if !(e.Start.IsZero() && e.End.IsZero()) {
		e.Duration = e.End.Sub(e.Start)
	}
	// tasks duration
	dur := time.Duration(e.Duration.Nanoseconds()/int64(len(e.Items))) * time.Nanosecond

	for i := range e.Items {
		e.Items[i].duration = dur
	}

	return e
}

// parse parses a google sheet row.
func (ts *Timesheet) parse(row []interface{}) (*TsEntry, error) {
	tr := TsEntry{
		Items: make([]Item, 1),
	}

	cols := ts.config.Spreadsheet.Columns

	if fStart, ok := value(row, cols.start).(float64); ok {
		tr.Start = lotusTime(fStart)
	}
	if fEnd, ok := value(row, cols.end).(float64); ok {
		tr.End = lotusTime(fEnd)
	}

	tr.Invoice = asString(value(row, cols.inv))
	tr.Items[0] = Item{
		Issue:       asString(value(row, cols.issue)),
		Description: asString(value(row, cols.descr)),
	}

	return &tr, nil
}

// ToYAML exports the timesheet to yaml format.
func (ts *Timesheet) ToYAML(w io.Writer) error {
	return ts.export(w, yaml.Marshal)
}

// ToJSON exports the timesheet to json format.
func (ts *Timesheet) ToJSON(w io.Writer) error {
	return ts.export(w, json.Marshal)
}

func value(row []interface{}, idx int) interface{} {
	if len(row) <= idx {
		return ""
	}
	return row[idx]
}

func lotusTime(datetime float64) time.Time {
	days := math.Trunc(datetime)
	seconds := time.Duration(math.Round((datetime-days)*24*60*60))*time.Second +
		time.Duration(days*24)*time.Hour

	ret, _ := time.Parse("2006-01-02", "1900-01-01")
	return ret.Add(seconds)
}

func (ts *Timesheet) export(w io.Writer, marshaller func(in interface{}) ([]byte, error)) error {
	buf := bufio.NewWriter(w)
	defer buf.Flush()

	data, err := marshaller(ts)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(buf, string(data))
	return err
}

// Invoices returns the invoices structure populated with timesheets.
func (ts *Timesheet) Invoices(ticketer bugtracker.Ticketer) *Invoices {
	invs := NewInvoices(ts.config.Values, ticketer)
	for i := range ts.Entries {
		invs.Append(ts.Entries[i])
	}
	return invs
}

// NewConfigFromFile loads timesheet config from file.
func NewConfigFromFile(filename string) (*TimesheetConfig, error) {
	return readConfig(filename)
}

func readConfig(filename string) (*TimesheetConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg TimesheetConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Spreadsheet.Columns.resolve(); err != nil {
		return nil, err
	}
	if err := cfg.Values.adjustDates(); err != nil {
		return nil, err
	}

	if cfg.Values.IssueSummary == nil {
		cfg.Values.IssueSummary = make(map[string]string)
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to a file on disk.
func (ts *Timesheet) SaveConfig(filename string) error {
	return saveConfig(filename, ts.config)
}

func saveConfig(filename string, cfg *TimesheetConfig) error {
	// saving config
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing config: %s", err)
	}
	return nil
}
