package aconfig

import (
	"reflect"
	"testing"
)

func Test_splitNameByWords(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"", args{""}, []string{}},
		{"", args{"str"}, []string{"str"}},
		{"", args{"apikey"}, []string{"apikey"}},
		{"", args{"apiKey"}, []string{"api", "Key"}},
		{"", args{"ApiKey"}, []string{"Api", "Key"}},
		{"", args{"APIKey"}, []string{"API", "Key"}},
		{"", args{"Type2"}, []string{"Type", "2"}},
		{"", args{"Type∆"}, []string{"Type", "∆"}},
		{"", args{"MarshalJSONStruct"}, []string{"Marshal", "JSON", "Struct"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := splitNameByWords(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitNameByWords() = %v, want %v", got, tt.want)
			}
		})
	}
}
