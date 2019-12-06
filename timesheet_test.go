package sheet2inv

import (
	"reflect"
	"testing"
	"time"
)

func mustParse(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return t
}

func Test_lotusTime(t *testing.T) {
	type args struct {
		datetime float64
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"9 am", args{1.41667}, mustParse(time.Parse("2006-01-02 15:04:05", "1900-01-02 10:00:00"))},
		{"3 mar 60, 1:30:45", args{21978.063020833}, mustParse(time.Parse("2006-01-02 15:04:05", "1960-03-05 01:30:45"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lotusTime(tt.args.datetime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lotusTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
