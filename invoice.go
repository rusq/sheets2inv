package sheet2inv

import (
	"log"
	"strings"
	"time"

	"github.com/rusq/sheet2inv/bugtracker"

	"github.com/shopspring/decimal"
)

// Invoices contains all invoices
type Invoices struct {
	Invoices map[string]*Invoice
	ticketer bugtracker.Ticketer
	cfg      *InvoiceValues
}

type Invoice struct {
	InvoiceID string
	Entries   map[string]InvoiceEntry

	values *InvoiceValues

	Total decimal.Decimal // invoice total

	tsIssues map[string][]*TsEntry
	ticketer bugtracker.Ticketer
}

type InvoiceEntry struct {
	Issue    string
	Details  []string
	Summary  string
	Duration time.Duration
	Rate     decimal.Decimal
	Total    decimal.Decimal
}

func NewInvoices(values *InvoiceValues, ticketer bugtracker.Ticketer) *Invoices {
	return &Invoices{Invoices: make(map[string]*Invoice), ticketer: ticketer, cfg: values}
}

func (invs *Invoices) Append(e *TsEntry) {
	invoice, ok := invs.Invoices[e.Invoice]
	if !ok {
		invs.Invoices[e.Invoice] = NewInvoice(invs.cfg, e.Invoice, invs.ticketer).Add(e)
	} else {
		invoice = invoice.Add(e)
	}
}

// Get returns the invoice by invoiceID
func (invs *Invoices) Get(invoiceID string) *Invoice {
	return invs.Invoices[invoiceID]
}

// NewInvoice creates the new invoice
func NewInvoice(values *InvoiceValues, id string, ticketer bugtracker.Ticketer) *Invoice {
	return &Invoice{values: values, InvoiceID: id, ticketer: ticketer}
}

// Add adds item to the issue.
func (i *Invoice) Add(e *TsEntry) *Invoice {
	if i.tsIssues == nil {
		i.tsIssues = make(map[string][]*TsEntry, 1)
	}
	for _, task := range e.Items {
		if _, ok := i.tsIssues[task.Issue]; !ok {
			i.tsIssues[task.Issue] = []*TsEntry{e}
		} else {
			i.tsIssues[task.Issue] = append(i.tsIssues[task.Issue], e)
		}
	}
	i.Recalculate()
	return i
}

// Recalculate recalculates all fields.  If IssueSummaryFunc is provided
// it is called to fetch the summary from the Bugtracking system
func (i *Invoice) Recalculate() *Invoice {
	if i.Entries == nil {
		i.Entries = make(map[string]InvoiceEntry)
	}
	i.Total = decimal.New(0, 0)

	for issue, tsEntries := range i.tsIssues {
		entry := InvoiceEntry{
			Total: decimal.New(0, 0),
		}

		for _, tsEntry := range tsEntries {
			text := make([]string, 0, len(tsEntry.Items))

			for _, item := range tsEntry.Recalculate().Items {
				if item.Issue != issue {
					continue
				}
				entry.Duration += item.duration
				text = append(text, item.Description)
			}
			entry.Issue = issue
			entry.Details = append(entry.Details, strings.Join(text, "; "))
			entry.Rate = tsEntry.rate.Mul(tsEntry.multiplier)
			entry.Total = entry.Total.Add(entry.Rate.Mul(decimal.NewFromFloat(tsEntry.Duration.Hours())))

			if i.ticketer != nil {
				var err error
				entry.Summary, err = i.ticketer.Name(entry.Issue)
				if err != nil {
					log.Println(issue + " " + err.Error())
				}
			}
		}

		i.Entries[issue] = entry
		i.Total = i.Total.Add(entry.Total)
	}

	return i
}
