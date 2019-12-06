package sheet2inv

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/rusq/sheet2inv/forms"
	"github.com/shopspring/decimal"
)

// ToPDF generates a pdf file from the invoice.
func (i *Invoice) ToPDF(filename string) error {
	f := forms.NewInvoice(i.InvoiceID, forms.PgLetter, nil, nil)
	// sorting entries
	order := make([]string, 0, len(i.Entries))
	for k := range i.Entries {
		order = append(order, k)
	}
	sort.Strings(order)

	for _, key := range order {
		entry := i.Entries[key]
		summary := fmt.Sprintf("%s: %s", entry.Issue, i.summary(&entry))
		duration := strconv.FormatFloat(entry.Duration.Hours(), 'f', 2, 64)
		rate := entry.Rate.StringFixedBank(2)
		total := entry.Total.StringFixedBank(2)
		f.AddEntry(summary, duration, rate, total)
	}

	f.AddSubTotal("SUBTOTAL", i.Total.StringFixedBank(2)).
		AddSubTotal(fmt.Sprintf("TAX (%s%%)", i.values.Tax.String()), i.values.Tax.Mul(decimal.New(100, 0)).StringFixedBank(2)).
		AddSubTotal("SHIPPING", i.values.Shipping.StringFixedBank(2)).
		SetTotal("BALANCE DUE", "$ "+i.Total.Add(i.Total.Mul(i.values.Tax)).StringFixedBank(2))

	f.AddAccountDetail("Bank", i.values.InvoiceFields.Bank).AddAccountDetail("Account No.", i.values.InvoiceFields.Account)

	return f.Generate(filename, &i.values.InvoiceFields)
}

func (i *Invoice) summary(entry *InvoiceEntry) string {
	val := i.values.IssueSummary
	if entry.Summary == "" && val[entry.Issue] == "" {
		buf := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter summary for issue %q:", entry.Issue)
		entry.Summary, _ = buf.ReadString('\n')
		entry.Summary = strings.Trim(entry.Summary, "\n")
	}
	if val[entry.Issue] == "" {
		// allows to override ticketing system values
		val[entry.Issue] = entry.Summary
	}
	return val[entry.Issue]
}
