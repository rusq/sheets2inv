package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/rusq/sheet2inv"
	"github.com/rusq/sheet2inv/bugtracker"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/yaml.v3"
)

var debug = (os.Getenv("DEBUG") != "")

// flags
var (
	credFile    = flag.String("creds", "sheets-credentials.json", "credentials `filename`")
	ticketCreds = flag.String("ticket-cred", "ticket-creds.json", "ticketing system credentials `filename`")
	export      = flag.String("export", "", "export timesheet to `file` in yaml or json format, use \"-\" for stdout")
	cfgFile     = flag.String("f", "fields.yaml", "timesheet config `file`")

	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func invoiceFromArgs() string {
	args := flag.Args()
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

func main() {
	flag.Parse()

	invoiceNo := invoiceFromArgs()

	b, err := ioutil.ReadFile(*credFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	cfg, err := sheet2inv.NewConfigFromFile(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	timesheet, err := sheet2inv.NewFromSheets(srv, cfg, invoiceNo)

	jira, err := bugtracker.JiraFromFile(*ticketCreds)
	if err != nil {
		log.Fatal(err)
	}

	if *export != "" {
		saveTo(*export, timesheet)
	}

	invoices := timesheet.Invoices(jira)
	for no := range invoices.Invoices {
		if err := invoices.Get(no).ToPDF(fmt.Sprintf("invoice-%s.pdf", no)); err != nil {
			log.Fatal(err)
		}
	}

	if debug {
		// debugging output
		if data, err := yaml.Marshal(invoices); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(string(data))
		}
	}

	if err := timesheet.SaveConfig(*cfgFile); err != nil {
		log.Fatal(err)
	}

	// profile
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

}

func saveTo(filename string, timesheet *sheet2inv.Timesheet) error {
	var output io.Writer
	if *export == "-" {
		output = os.Stdout
	} else {
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		output = f
	}

	if err := timesheet.MarshalTo(output); err != nil {
		return err
	}
	return nil
}
