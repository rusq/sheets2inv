package sheet2inv

import (
	"testing"
)

func TestColumns_resolve(t *testing.T) {
	type fields struct {
		TimeStart   string
		TimeEnd     string
		Invoice     string
		Description string
		Issue       string
		start       int
		end         int
		inv         int
		descr       int
		issue       int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"in range",
			fields{
				TimeStart:   "A",
				TimeEnd:     "B",
				Invoice:     "C",
				Description: "D",
				Issue:       "Z",
			},
			false,
		},
		{"out of range",
			fields{
				TimeStart:   "A",
				TimeEnd:     "B",
				Invoice:     "C",
				Description: "D",
				Issue:       "3",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &Columns{
				TimeStart:   tt.fields.TimeStart,
				TimeEnd:     tt.fields.TimeEnd,
				Invoice:     tt.fields.Invoice,
				Description: tt.fields.Description,
				Issue:       tt.fields.Issue,
				start:       tt.fields.start,
				end:         tt.fields.end,
				inv:         tt.fields.inv,
				descr:       tt.fields.descr,
				issue:       tt.fields.issue,
			}
			if err := ci.resolve(); (err != nil) != tt.wantErr {
				t.Errorf("Columns.resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestColumns_char2int(t *testing.T) {
	type args struct {
		char string
	}
	tests := []struct {
		name      string
		args      args
		want      int
		wantPanic bool
	}{
		{"IN range", args{"A"}, 0, false},
		{"IN range", args{"a"}, 0, false},
		{"IN range", args{"Z"}, 25, false},
		{"IN range", args{"z"}, 25, false},
		{"1+ chars", args{"aa"}, 0, false},
		{"out of range", args{"5"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("unexpected panic: %v", r)
				}
			}()
			c := &Columns{}
			if got := c.char2int(tt.args.char); got != tt.want {
				t.Errorf("Columns.char2int() = %v, want %v", got, tt.want)
			}
		})
	}
}
