package aconfig

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	raw := map[string]interface{}{
		"a": "b",
		"c": 10.0,
		"d": map[string]interface{}{
			"x": "y",
			"z": []interface{}{"1", "2", "3"},
		},
		"A": map[string]interface{}{
			"B": map[string]interface{}{
				"C": map[string]interface{}{
					"a": "b", // TODO: empty
				},
			},
		},
	}
	res := map[string]interface{}{}
	for key, value := range raw {
		flatten("", key, value, res)
	}

	t.Log(res)
}

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
		{"", args{"MarshalJSONStruct"}, []string{"Marshal", "JSON", "Struct"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitNameByWords(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitNameByWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONDecoder(t *testing.T) {
	got, err := (&jsonDecoder{}).DecodeFile("testfile.json")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{
		"A.B.C.a": "b",
		"A.B.D":   []interface{}{"x", "y", "z"},
		"a":       "b",
		"c":       "10",
		"d.x":     "y",
		"d.z":     []interface{}{"1", "2", "3"},
	}

	if len(want) != len(got) {
		t.Fatalf("got %v, want %v", len(got), len(want))
	}
	for k, v := range want {
		if kv, ok := got[k]; !ok || !reflect.DeepEqual(kv, v) {
			t.Errorf("for %v got %v, want %v %T %T", k, kv, v, kv, v)
		}
	}
}
