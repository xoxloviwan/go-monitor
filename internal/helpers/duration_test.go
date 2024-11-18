package helpers

import (
	"testing"
	"time"
)

func TestDuration_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		d       *Duration
		args    args
		wantErr bool
		value   time.Duration
	}{
		{
			name: "valid duration",
			d:    &Duration{},
			args: args{
				b: []byte(`"2m"`),
			},
			value:   time.Minute * 2,
			wantErr: false,
		},
		{
			name: "invalid duration",
			d:    &Duration{},
			args: args{
				b: []byte(`"2ч"`),
			},
			wantErr: true,
		},
		{
			name: "invalid data",
			d:    &Duration{},
			args: args{
				b: []byte(``),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Duration.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.d.Duration != tt.value {
				t.Errorf("Duration.UnmarshalJSON() = %v, want %v", tt.d.Duration, tt.value)
			}
		})
	}
}
