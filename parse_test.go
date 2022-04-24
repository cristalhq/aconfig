package aconfig

import (
	"fmt"
	"testing"
)

type NewTestConfig struct {
	Str    string `default:"str-def"`
	Bytes  []byte `default:"bytes-def"`
	Int    int    `default:"8080"`
	Int64  int64  `default:"18080"`
	UInt64 uint64 `default:"80800"`
	IntPtr *int32 `default:"123"`
	Sub    SubConfig
	EmbeddedConfig
	Anon struct {
		IsAnon bool `default:"true"`
	}
	Slice      []int                        `default:"1,2,3" usage:"just pass elements"`
	StrSlice   []string                     `default:"1,2,3" usage:"just pass strings"`
	HeavySlice []SubConfig                  `default:"1,2,3" usage:"just pass strings"`
	Map1       map[string]int               `default:"a:1,b:2,c:3"`
	Map2       map[int]string               `default:"1:a,2:b,3:c"`
	Map3       map[int]EmbeddedConfig       `default:"1:a,2:b,3:c"`
	Map4       map[*structA]*EmbeddedConfig `default:"1:a,2:b,3:c"`
}

func TestNew(t *testing.T) {
	var cfg NewTestConfig
	var parser configParser
	f, err := parser.Parse(&cfg)
	if err != nil {
		t.Error(err)
	}
	printTree(f, 0)
}

func printTree(f *field, level int) {
	fmt.Printf("%*d\n", level*10, len(f.child))
	for _, f := range f.child {
		fmt.Printf("%*v\n", level*10, f.st.Name)
		if len(f.child) > 0 {
			printTree(f, level+1)
		}
	}
}

// func Test_flattenCfg(t *testing.T) {
// 	type J struct {
// 		K bool `fig:"k"`
// 	}
// 	cfg := struct {
// 		A string
// 		B struct {
// 			C []struct {
// 				D *int
// 			}
// 		}
// 		E *struct {
// 			F []string
// 		} `fig:"e"`
// 		G *struct {
// 			H int
// 		}
// 		i int
// 		J
// 	}{}
// 	cfg.B.C = []struct{ D *int }{{}, {}}
// 	cfg.E = &struct{ F []string }{}

// 	fields := flattenCfg(&cfg, "fig")
// 	if len(fields) != 10 {
// 		t.Fatalf("len(fields) == %d, expected %d", len(fields), 10)
// 	}
// 	checkField(t, fields[0], "A", "A")
// 	checkField(t, fields[1], "B", "B")
// 	checkField(t, fields[2], "C", "B.C")
// 	checkField(t, fields[3], "D", "B.C[0].D")
// 	checkField(t, fields[4], "D", "B.C[1].D")
// 	checkField(t, fields[5], "e", "e")
// 	checkField(t, fields[6], "F", "e.F")
// 	checkField(t, fields[7], "G", "G")
// 	checkField(t, fields[8], "J", "J")
// 	checkField(t, fields[9], "k", "J.k")
// }

// func Test_newStructField(t *testing.T) {
// 	cfg := struct {
// 		A int `fig:"a" default:"5" validate:"required"`
// 	}{}
// 	parent := &field{
// 		v:        reflect.ValueOf(&cfg).Elem(),
// 		t:        reflect.ValueOf(&cfg).Elem().Type(),
// 		sliceIdx: -1,
// 	}

// 	f := newStructField(parent, 0, "fig")
// 	if f.parent != parent {
// 		t.Errorf("f.parent == %p, expected %p", f.parent, f)
// 	}
// 	if f.sliceIdx != -1 {
// 		t.Errorf("f.sliceIdx == %d, expected %d", f.sliceIdx, -1)
// 	}
// 	if f.v.Kind() != reflect.Int {
// 		t.Errorf("f.v.Kind == %v, expected %v", f.v.Kind(), reflect.Int)
// 	}
// 	if f.v.Type() != reflect.TypeOf(cfg.A) {
// 		t.Errorf("f.v.Type == %v, expected %v", f.v.Kind(), reflect.TypeOf(cfg.A))
// 	}
// 	// if f.altName != "a" {
// 	// 	t.Errorf("f.altName == %s, expected %s", f.altName, "a")
// 	// }
// 	if !f.isRequired {
// 		t.Errorf("f.required == false")
// 	}
// 	// if !f.setDefault {
// 	// 	t.Errorf("f.setDefault == false")
// 	// }
// 	// if f.defaultVal != "5" {
// 	// 	t.Errorf("f.defaultVal == %s, expected %s", f.defaultVal, "5")
// 	// }
// }

// func Test_newSliceField(t *testing.T) {
// 	cfg := struct {
// 		A []struct {
// 			B int
// 		} `fig:"aaa"`
// 	}{}
// 	cfg.A = []struct {
// 		B int
// 	}{{B: 5}}

// 	parent := &field{
// 		v:        reflect.ValueOf(&cfg).Elem().Field(0),
// 		t:        reflect.ValueOf(&cfg).Elem().Field(0).Type(),
// 		st:       reflect.ValueOf(&cfg).Elem().Type().Field(0),
// 		sliceIdx: -1,
// 	}

// 	f := newSliceField(parent, 0, "fig")
// 	if f.parent != parent {
// 		t.Errorf("f.parent == %p, expected %p", f.parent, f)
// 	}
// 	if f.sliceIdx != 0 {
// 		t.Errorf("f.sliceIdx == %d, expected %d", f.sliceIdx, 0)
// 	}
// 	if f.v.Kind() != reflect.Struct {
// 		t.Errorf("f.v.Kind == %v, expected %v", f.v.Kind(), reflect.Int)
// 	}
// 	// if f.altName != "aaa" {
// 	// 	t.Errorf("f.altName == %s, expected %s", f.altName, "a")
// 	// }
// 	if f.isRequired {
// 		t.Errorf("f.required == true")
// 	}
// 	// if f.setDefault {
// 	// 	t.Errorf("f.setDefault == true")
// 	// }
// 	// if f.defaultVal != "" {
// 	// 	t.Errorf("f.defaultVal == %s, expected %s", f.defaultVal, "")
// 	// }
// }

// func Test_parseTag(t *testing.T) {
// 	t.Skip()
// 	for _, tc := range []struct {
// 		tagVal string
// 		want   structTag
// 	}{
// 		{
// 			tagVal: "",
// 			want:   structTag{},
// 		},
// 		{
// 			tagVal: `fig:"a"`,
// 			want:   structTag{altName: "a"},
// 		},
// 		{
// 			tagVal: `fig:"a,"`,
// 			want:   structTag{altName: "a"},
// 		},
// 		{
// 			tagVal: `fig:"a" default:"go"`,
// 			want:   structTag{altName: "a", setDefault: true, defaultVal: "go"},
// 		},
// 		{
// 			tagVal: `fig:"b" validate:"required"`,
// 			want:   structTag{altName: "b", required: true},
// 		},
// 		{
// 			tagVal: `fig:"b" validate:"required" default:"go"`,
// 			want:   structTag{altName: "b", required: true, setDefault: true, defaultVal: "go"},
// 		},
// 		{
// 			tagVal: `fig:"c,omitempty"`,
// 			want:   structTag{altName: "c"},
// 		},
// 	} {
// 		t.Run(tc.tagVal, func(t *testing.T) {
// 			tag := parseTag(reflect.StructTag(tc.tagVal), "fig")
// 			if !reflect.DeepEqual(tc.want, tag) {
// 				t.Fatalf("parseTag() == %+v, expected %+v", tag, tc.want)
// 			}
// 		})
// 	}
// }

// func checkField(t *testing.T, f *field, name, path string) {
// 	t.Helper()
// 	if f.name() != name {
// 		t.Errorf("f.name() == %s, expected %s", f.name(), name)
// 	}
// 	if f.path() != path {
// 		t.Errorf("f.path() == %s, expected %s", f.path(), path)
// 	}
// }
