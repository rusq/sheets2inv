package forms

import (
	"image/color"
	"testing"
)

func Test_toRGB(t *testing.T) {
	type args struct {
		col color.RGBA
	}
	tests := []struct {
		name  string
		args  args
		wantR int
		wantG int
		wantB int
	}{
		{"1", args{cLightBlue}, int(cLightBlue.R), int(cLightBlue.G), int(cLightBlue.B)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotG, gotB := toRGB(tt.args.col)
			if gotR != tt.wantR {
				t.Errorf("toRGB() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotG != tt.wantG {
				t.Errorf("toRGB() gotG = %v, want %v", gotG, tt.wantG)
			}
			if gotB != tt.wantB {
				t.Errorf("toRGB() gotB = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}
