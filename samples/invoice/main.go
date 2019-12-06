package main

import (
	"fmt"
	"strconv"

	"github.com/unidoc/unipdf/v3/creator"
)

func main() {
	// Instantiate new PDF creator
	c := creator.New()

	// Create a new PDF page and select it for editing
	c.NewPage()

	// Create new invoice and populate it with data
	invoice := createInvoice(c, "logo.png")

	// Write invoice to page
	checkErr(c.Draw(invoice))

	// Write output file.
	// Alternative is writing to a Writer interface by using c.Write
	checkErr(c.WriteToFile("simple_invoice.pdf"))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func createInvoice(c *creator.Creator, logoPath string) *creator.Invoice {
	// Create an instance of Logo used as a header for the invoice
	// If the image is not stored localy, you can use NewImageFromData to generate it from byte array
	logo, err := c.NewImageFromFile(logoPath)
	checkErr(err)

	// Create a new invoice
	invoice := c.NewInvoice()

	// Set invoice logo
	invoice.SetLogo(logo)

	// Set invoice information
	invoice.SetNumber("0001")
	invoice.SetDate("28/07/2016")
	invoice.SetDueDate("28/07/2016")
	invoice.AddInfo("Payment terms", "Due on receipt")
	invoice.AddInfo("Paid", "No")

	// Set invoice addresses
	invoice.SetSellerAddress(&creator.InvoiceAddress{
		Name:    "John Doe",
		Street:  "8 Elm Street",
		City:    "Cambridge",
		Zip:     "CB14DH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "johndoe@email.com",
	})

	invoice.SetBuyerAddress(&creator.InvoiceAddress{
		Name:    "Jane Doe",
		Street:  "9 Elm Street",
		City:    "London",
		Zip:     "LB15FH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "janedoe@email.com",
	})

	// Add products to invoice
	for i := 1; i < 6; i++ {
		invoice.AddLine(
			fmt.Sprintf("Test product #%d", i),
			"1",
			strconv.Itoa((i-1)*7),
			strconv.Itoa((i+4)*3),
		)
	}

	// Set invoice totals
	invoice.SetSubtotal("$100.00")
	invoice.AddTotalLine("Tax (10%)", "$10.00")
	invoice.AddTotalLine("Shipping", "$5.00")
	invoice.SetTotal("$115.00")

	// Set invoice content sections
	invoice.SetNotes("Notes", "Thank you for your business.")
	invoice.SetTerms("Terms and conditions", "Full refund for 60 days after purchase.")

	return invoice
}
