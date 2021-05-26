package aconfig

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{
		Str:      "str-def",
		Bytes:    []byte("bytes-def"),
		Int:      int32Ptr(123),
		HTTPPort: 8080,
		Sub: SubConfig{
			Float: 123.123,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
		StrSlice: []string{"1", "2", "3"},
		Slice:    []int{1, 2, 3},
		Map1:     map[string]int{"a": 1, "b": 2, "c": 3},
		Map2:     map[int]string{1: "a", 2: "b", 3: "c"},
		EmbeddedConfig: EmbeddedConfig{
			Em: "em-def",
		},
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestDefaults_AllTypes(t *testing.T) {
	type AllTypesConfig struct {
		Bool   bool   `default:"true"`
		String string `default:"str"`

		Int   int   `default:"1"`
		Int8  int8  `default:"12"`
		Int16 int16 `default:"123"`
		Int32 int32 `default:"13"`
		Int64 int64 `default:"23"`

		Uint   uint   `default:"1234"`
		Uint8  uint8  `default:"124"`
		Uint16 uint16 `default:"134"`
		Uint32 uint32 `default:"234"`
		Uint64 uint64 `default:"24"`

		Float32 float32 `default:"1234.213"`
		Float64 float64 `default:"1234.234"`

		Dur  time.Duration `default:"1h2m3s"`
		Time time.Time     `default:"2000-04-05 10:20:30 +0000 UTC"`
	}

	var cfg AllTypesConfig
	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := AllTypesConfig{
		Bool:    true,
		String:  "str",
		Int:     1,
		Int8:    12,
		Int16:   123,
		Int32:   13,
		Int64:   23,
		Uint:    1234,
		Uint8:   124,
		Uint16:  134,
		Uint32:  234,
		Uint64:  24,
		Float32: 1234.213,
		Float64: 1234.234,
		Dur:     time.Hour + 2*time.Minute + 3*time.Second,
		// TODO
		// Time :2000-04-05 10:20:30 +0000 UTC,
	}

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestDefaults_OtherNumberFormats(t *testing.T) {
	type OtherNumberFormats struct {
		Int    int   `default:"0b111"`
		Int8   int8  `default:"0o123"`
		Int8x2 int8  `default:"0123"`
		Int16  int16 `default:"0x123"`

		Uint   uint   `default:"0b111"`
		Uint8  uint8  `default:"0o123"`
		Uint16 uint16 `default:"0123"`
		Uint32 uint32 `default:"0x123"`
	}

	var cfg OtherNumberFormats
	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := OtherNumberFormats{
		Int:    7,
		Int8:   83,
		Int8x2: 83,
		Int16:  291,

		Uint:   7,
		Uint8:  83,
		Uint16: 83,
		Uint32: 291,
	}

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestJSON(t *testing.T) {
	filepath := createTestFile(t)

	var cfg structConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := wantConfig
	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestCustomFile(t *testing.T) {
	filepath := createTestFile(t, "custom.config")

	var cfg structConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
		FileDecoders: map[string]FileDecoder{
			".config": &jsonDecoder{},
		},
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := wantConfig
	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFile(t *testing.T) {
	filepath := "testdata/config.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{
		Str:      "str-json",
		Bytes:    []byte("Ynl0ZXMtanNvbg=="),
		Int:      int32Ptr(101),
		HTTPPort: 65000,
		Sub: SubConfig{
			Float: 999.111,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFileMerging(t *testing.T) {
	file1 := "testdata/config1.json"
	file2 := "testdata/config2.json"
	file3 := "testdata/config3.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		MergeFiles:   true,
		Files:        []string{file1, file2, file3},
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{
		Str:      "111",
		HTTPPort: 222,
		Sub: SubConfig{
			Float: 333.333,
		},
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFileFlag(t *testing.T) {
	file1 := "testdata/config1.json"

	flags := []string{
		"-file_flag=testdata/config2.json",
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		MergeFiles:   true,
		FileFlag:     "file_flag",
		Files:        []string{file1},
		Args:         flags,
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{
		Str:      "111",
		HTTPPort: 222,
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestBadFileFlag(t *testing.T) {
	flags := []string{
		"-file_flag=",
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		FileFlag:     "file_flag",
		Args:         flags,
	})
	if err := loader.Load(); err == nil {
		t.Fatal("should be an error")
	}
}

func TestNoFileFlagValue(t *testing.T) {
	file1 := "testdata/config1.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		FileFlag:     "file_flag",
		Files:        []string{file1},
		Args:         []string{}, // no file_flag
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{
		Str:      "111",
		HTTPPort: 111,
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestEnv(t *testing.T) {
	setEnv(t, "TST_STR", "str-env")
	setEnv(t, "TST_BYTES", "bytes-env")
	setEnv(t, "TST_INT", "121")
	setEnv(t, "TST_HTTP_PORT", "3000")
	setEnv(t, "TST_SUB_FLOAT", "222.333")
	setEnv(t, "TST_ANON_IS_ANON", "true")
	setEnv(t, "TST_EM", "em-env")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{
		Str:      "str-env",
		Bytes:    []byte("bytes-env"),
		Int:      int32Ptr(121),
		HTTPPort: 3000,
		Sub: SubConfig{
			Float: 222.333,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
		EmbeddedConfig: EmbeddedConfig{
			Em: "em-env",
		},
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFlag(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	flags := []string{
		"-tst.str=str-flag",
		"-tst.bytes=Ynl0ZXMtZmxhZw==",
		"-tst.int=1001",
		"-tst.http_port=30000",
		"-tst.sub.float=123.321",
		"-tst.anon.is_anon=true",
		"-tst.em=em-flag",
	}

	if err := loader.Flags().Parse(flags); err != nil {
		t.Fatal(err)
	}

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{
		Str:      "str-flag",
		Bytes:    []byte("Ynl0ZXMtZmxhZw=="),
		Int:      int32Ptr(1001),
		HTTPPort: 30000,
		Sub: SubConfig{
			Float: 123.321,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
		EmbeddedConfig: EmbeddedConfig{
			Em: "em-flag",
		},
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestExactName(t *testing.T) {
	setEnv(t, "STR", "str-env")
	setEnv(t, "TST_STR", "bar-env")
	defer os.Clearenv()

	type Foo struct {
		String string `env:"STR,exact"`
	}
	type ExactConfig struct {
		Foo Foo
		Bar string `env:"STR"`
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipDefaults:     true,
		SkipFiles:        true,
		SkipFlags:        true,
		AllowUnknownEnvs: true,
		EnvPrefix:        "TST",
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := ExactConfig{
		Foo: Foo{
			String: "str-env",
		},
		Bar: "bar-env",
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSkipName(t *testing.T) {
	setEnv(t, "STR", "str-env")
	setEnv(t, "BAR", "bar-env")
	defer os.Clearenv()

	type Foo struct {
		String string `env:"STR"`
	}
	type ExactConfig struct {
		Foo Foo    `env:"-"`
		Bar string `default:"def" env:"-"`
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipFlags: true,
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := ExactConfig{
		Foo: Foo{
			String: "str-env",
		},
		Bar: "def",
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestDuplicatedName(t *testing.T) {
	setEnv(t, "FOO_BAR", "str-env")
	defer os.Clearenv()

	type Foo struct {
		Bar string
	}
	type ExactConfig struct {
		Foo    Foo
		FooBar string
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipFlags:       true,
		AllowDuplicates: true,
	})
	if err := loader.Load(); err != nil {
		t.Error(err)
	}

	want := ExactConfig{
		Foo: Foo{
			Bar: "str-env",
		},
		FooBar: "str-env",
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFailOnDuplicatedName(t *testing.T) {
	type Foo struct {
		Bar string
	}
	type ExactConfig struct {
		Foo    Foo
		FooBar string
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipFlags: true,
	})

	err := loader.Load()
	if !strings.Contains(err.Error(), "is duplicated") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUsage(t *testing.T) {
	loader := LoaderFor(&EmbeddedConfig{}, Config{})

	var builder strings.Builder
	flags := loader.Flags()
	flags.SetOutput(&builder)
	flags.PrintDefaults()

	got := builder.String()
	want := `  -em string
    	use... em...field. (default "em-def")
`

	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestBadDefauts(t *testing.T) {
	f := func(cfg interface{}) {
		t.Helper()

		loader := LoaderFor(cfg, Config{
			SkipFiles: true,
			SkipEnv:   true,
			SkipFlags: true,
		})
		if err := loader.Load(); err == nil {
			t.Fatal(err)
		}
	}

	f(&struct {
		Bool bool `default:"omg"`
	}{})

	f(&struct {
		Int int `default:"1a"`
	}{})

	f(&struct {
		Int8 int8 `default:"12a"`
	}{})

	f(&struct {
		Int16 int16 `default:"123a"`
	}{})

	f(&struct {
		Int32 int32 `default:"13a"`
	}{})

	f(&struct {
		Int64 int64 `default:"23a"`
	}{})

	f(&struct {
		Uint uint `default:"1234a"`
	}{})

	f(&struct {
		Uint8 uint8 `default:"124a"`
	}{})

	f(&struct {
		Uint16 uint16 `default:"134a"`
	}{})

	f(&struct {
		Uint32 uint32 `default:"234a"`
	}{})

	f(&struct {
		Uint64 uint64 `default:"24a"`
	}{})

	f(&struct {
		Float32 float32 `default:"1234x213"`
	}{})

	f(&struct {
		Float64 float64 `default:"1234x234"`
	}{})

	f(&struct {
		Dur time.Duration `default:"1h_2m3s"`
	}{})

	f(&struct {
		Slice []int `default:"1,a,2"`
	}{})

	f(&struct {
		Map map[string]int `default:"1:a,2:2"`
	}{})

	f(&struct {
		Map map[int]string `default:"a:1"`
	}{})

	f(&struct {
		Map map[int]string `default:"a1"`
	}{})

	f(&struct {
		Array [2]string `default:"a1"`
	}{})
}

func TestBadFiles(t *testing.T) {
	f := func(filepath string) {
		var cfg TestConfig
		loader := LoaderFor(&cfg, Config{
			SkipDefaults:       true,
			SkipEnv:            true,
			SkipFlags:          true,
			FailOnFileNotFound: true,
			Files:              []string{filepath},
		})
		if err := loader.Load(); err == nil {
			t.Fatal(err)
		}
	}

	t.Run("no_such_file.json", func(t *testing.T) {
		f("no_such_file.json")
	})

	t.Run("bad_config.json", func(t *testing.T) {
		filepath := t.TempDir() + "unknown.ext"
		file, err := os.Create(filepath)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		_, err = file.WriteString(`{almost": "json`)
		if err != nil {
			t.Fatal(err)
		}

		f(filepath)
	})

	t.Run("unknown.ext", func(t *testing.T) {
		filepath := t.TempDir() + "unknown.ext"
		file, err := os.Create(filepath)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		f(filepath)
	})
}

func TestFailOnFileNotFound(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := LoaderFor(&TestConfig{}, Config{
			SkipDefaults:       true,
			SkipEnv:            true,
			SkipFlags:          true,
			FailOnFileNotFound: false,
			Files:              []string{filepath},
		})

		if err := loader.Load(); err != nil {
			t.Fatal(err)
		}
	}

	f("testdata/config.json")
	f("testdata/not_found.json")
}

func TestBadEnvs(t *testing.T) {
	setEnv(t, "TST_HTTP_PORT", "30a00")
	defer os.Clearenv()

	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})

	if err := loader.Load(); err == nil {
		t.Fatal(err)
	}
}

func TestBadFlags(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	if err := loader.Flags().Parse([]string{"-tst.param=10a01"}); err != nil {
		t.Fatal(err)
	}
	if err := loader.Load(); err == nil {
		t.Fatal(err)
	}
}

func TestUnknownFields(t *testing.T) {
	filepath := "testdata/unknown_fields.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})

	err := loader.Load()
	if err == nil {
		t.Fatal("must not be nil")
	}
	if !strings.Contains(err.Error(), "unknown field in file") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownEnvs(t *testing.T) {
	setEnv(t, "TST_STR", "defined")
	setEnv(t, "TST_UNKNOWN", "42")
	setEnv(t, "JUST_ENV", "JUST_VALUE")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})

	err := loader.Load()
	if err == nil {
		t.Fatal("must not be nil")
	}
	if !strings.Contains(err.Error(), "unknown environment var") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownEnvsWithEmptyPrefix(t *testing.T) {
	setEnv(t, "STR", "defined")
	setEnv(t, "UNKNOWN", "42")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
	})

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownFlags(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	flags := []string{
		"-tst.str=str-flag",
		"-tst.unknown=1001",
		"-just_env=just_value",
	}

	// just for tests
	flagSet := loader.Flags()
	flagSet.SetOutput(ioutil.Discard)

	// define flag with a loader's prefix which is unknown
	flagSet.Int("tst.unknown", 42, "")
	flagSet.String("just_env", "just_def", "")

	if err := flagSet.Parse(flags); err != nil {
		t.Fatal(err)
	}

	err := loader.Load()
	if err == nil {
		t.Fatal("must not be nil")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownFlagsWithEmptyPrefix(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
	})

	flags := []string{
		"-str=str-flag",
		"-unknown=1001",
	}

	// just for tests
	flagSet := loader.Flags()
	flagSet.SetOutput(ioutil.Discard)

	// define flag with a loader's prefix which is unknown
	flagSet.Int("unknown", 42, "")

	if err := flagSet.Parse(flags); err != nil {
		t.Fatal(err)
	}

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}
}

// flag.FlagSet already fails on undefined flag
func TestUnknownFlagsStdlib(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	flags := []string{
		"-tst.str=str-flag",
		"-tst.unknown=1001",
	}

	// just for tests
	flagSet := loader.Flags()
	flagSet.SetOutput(ioutil.Discard)

	if err := flagSet.Parse(flags); err == nil {
		t.Fatal("must not be nil")
	}
}

func TestCustomNames(t *testing.T) {
	type TestConfig struct {
		A int `default:"-1" env:"ONE"`
		B int `default:"-1" flag:"two"`
		C int `default:"-1" env:"three" flag:"four"`
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		Args: []string{"-two=2", "-four=4"},
	})

	setEnv(t, "ONE", "1")
	setEnv(t, "three", "3")
	defer os.Clearenv()

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	if want := 1; cfg.A != want {
		t.Errorf("got %#v, want %#v", cfg.A, want)
	}
	if want := 2; cfg.B != want {
		t.Errorf("got %#v, want %#v", cfg.B, want)
	}
	if want := 4; cfg.C != want {
		t.Errorf("got %#v, want %#v", cfg.C, want)
	}
}

func TestDontGenerateTags(t *testing.T) {
	type testConfig struct {
		A        string `json:"aaa"`
		B        string `yaml:"aaa" toml:"bbb"`
		DooDoo   string
		HTTPPort int    `yaml:"port"`
		D        string `env:"aaa"`
		E        string `flag:"aaa"`
	}

	want := map[string]string{
		"A::json":        "aaa",
		"B::yaml":        "aaa",
		"C::toml":        "c",
		"DooDoo::toml":   "DooDoo",
		"DooDoo::flag":   "doo_doo",
		"HTTPPort::flag": "http_port",
		"HTTPPort::json": "HTTPPort",
		"HTTPPort::yaml": "port",
		"D::env":         "aaa",
		"E::flag":        "aaa",
		"E::json":        "E",
	}
	cfg := Config{
		DontGenerateTags: true,
	}
	LoaderFor(&testConfig{}, cfg).WalkFields(func(f Field) bool {
		for _, tag := range []string{"json", "yaml", "toml", "env", "flag"} {
			k := f.Name() + "::" + tag
			if v, ok := want[k]; ok && v != f.Tag(tag) {
				t.Fatalf("%v: got %v, want %v", tag, f.Tag(tag), v)
				return false
			}
		}
		return true
	})
}

func TestWalkFields(t *testing.T) {
	type TestConfig struct {
		A int `default:"-1" env:"one" marco:"polo"`
		B struct {
			C int `default:"-1" flag:"two" usage:"pretty simple usage duh" json:"kek" yaml:"lel" toml:"mde"`
			D struct {
				E int `default:"-1" env:"three" json:"kek" yaml:"lel" toml:"mde"`
			}
		}
	}

	fields := []struct {
		Name         string
		ParentName   string
		DefaultValue string
		EnvName      string
		FlagName     string
		Usage        string
	}{
		{
			Name:         "A",
			EnvName:      "one",
			DefaultValue: "-1",
		},
		{
			Name:         "B.C",
			ParentName:   "B",
			FlagName:     "two",
			DefaultValue: "-1",
			Usage:        "pretty simple usage duh",
		},
		{
			Name:         "B.D.E",
			ParentName:   "B.D",
			EnvName:      "three",
			DefaultValue: "-1",
		},
	}

	i := 0

	LoaderFor(&TestConfig{}, Config{}).WalkFields(func(f Field) bool {
		wantFields := fields[i]
		if f.Name() != wantFields.Name {
			t.Fatalf("got name %v, want %v", f.Name(), wantFields.Name)
		}

		if parent, ok := f.Parent(); ok && parent.Name() != wantFields.ParentName {
			t.Fatalf("got name %v, want %v", parent.Name(), wantFields.ParentName)
		}
		if f.Tag("default") != wantFields.DefaultValue {
			t.Fatalf("got default %#v, want %#v", f.Tag("default"), wantFields.DefaultValue)
		}
		if f.Tag("usage") != wantFields.Usage {
			t.Fatalf("got usage %#v, want %#v", f.Tag("usage"), wantFields.Usage)
		}
		i++
		return true
	})

	if want := 3; i != want {
		t.Fatalf("got %v, want %v", i, want)
	}

	i = 0
	LoaderFor(&TestConfig{}, Config{}).WalkFields(func(f Field) bool {
		if i > 0 {
			return false
		}
		if got := f.Tag("marco"); got != "polo" {
			t.Fatalf("got %v, want %v", got, "polo")
		}
		i++
		return true
	})
	if i != 1 {
		t.Fatal()
	}
}

func TestDontFillFlagsIfDisabled(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipFlags: true,
		Args:      []string{},
	})
	if err := loader.Load(); err != nil {
		t.Error(err)
	}

	if flags := loader.Flags().NFlag(); flags != 0 {
		t.Errorf("want empty, got %v", flags)
	}
}

func TestPassNonStructs(t *testing.T) {
	f := func(cfg interface{}) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()

		_ = LoaderFor(cfg, Config{})
	}

	f(nil)
	f(map[string]string{})
	f([]string{})
	f([4]string{})
	f(func() {})
}

func TestBadRequiredTag(t *testing.T) {
	type TestConfig struct {
		Field string `required:"boom"`
	}

	f := func(cfg interface{}) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()

		_ = LoaderFor(cfg, Config{})
	}

	f(&TestConfig{})
}

func setEnv(t *testing.T, key, value string) {
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
}

func int32Ptr(a int32) *int32 {
	return &a
}

func createTestFile(t *testing.T, name ...string) string {
	t.Helper()
	if len(name) > 1 {
		t.Fatal()
	}

	dir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	filepath := dir + "/testfile.json"
	if len(name) == 1 {
		filepath = dir + name[0]
	}

	f, err := os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(testfileContent)
	if err != nil {
		t.Fatal(err)
	}
	return filepath
}

type TestConfig struct {
	Str      string `default:"str-def"`
	Bytes    []byte `default:"bytes-def"`
	Int      *int32 `default:"123"`
	HTTPPort int    `default:"8080"`
	Param    int    // no default tag, so default value
	Sub      SubConfig
	Anon     struct {
		IsAnon bool `default:"true"`
	}

	StrSlice []string       `default:"1,2,3" usage:"just pass strings"`
	Slice    []int          `default:"1,2,3" usage:"just pass elements"`
	Map1     map[string]int `default:"a:1,b:2,c:3"`
	Map2     map[int]string `default:"1:a,2:b,3:c"`

	EmbeddedConfig
}

type EmbeddedConfig struct {
	Em string `default:"em-def" usage:"use... em...field."`
}

type SubConfig struct {
	Float float64 `default:"123.123"`
}

type structConfig struct {
	A string
	C int
	E float64
	B []byte
	I *int32
	J *int64
	Y structY

	AA structA `json:"A"`
	StructM

	M interface{} `json:"M"`

	P *structP `json:"P"`
}

type structY struct {
	X string
	Z []int
	A structD
}

type structA struct {
	X  string  `json:"x"`
	BB structB `json:"B"`
}

type structB struct {
	CC structC  `json:"C"`
	DD []string `json:"D"`
}

type structC struct {
	MM string `json:"m"`
	BB []byte `json:"b"`
}

type structD struct {
	I bool
}

type StructM struct {
	M string
}

type structP struct {
	P string `json:"P"`
}

const testfileContent = `{
    "a": "b",
    "c": 10,
    "e": 123.456,
    "b": "abc",
    "i": 42,
	"j": 420,

    "y": {
        "x": "y",
		"z": [1, "2", "3"],
		"a": {
			"i": true
		}
    },

    "A": {
        "x": "y",
        "B": {
            "C": {
                "m": "n",
                "b": "boo"
            },
            "D": ["x", "y", "z"]
        }
	},

	"m": "n",

	"M":["q", "w"],
	
	"P": {
		"P": "r"
	}
}
`

var wantConfig = func() structConfig {
	i := int32(42)
	j := int64(420)
	mInterface := make([]interface{}, 2)
	for iI, vI := range []string{"q", "w"} {
		mInterface[iI] = vI
	}

	return structConfig{
		A: "b",
		C: 10,
		E: 123.456,
		B: []byte("abc"),
		I: &i,
		J: &j,
		Y: structY{
			X: "y",
			Z: []int{1, 2, 3},
			A: structD{
				I: true,
			},
		},
		AA: structA{
			X: "y",
			BB: structB{
				CC: structC{
					MM: "n",
					BB: []byte("boo"),
				},
				DD: []string{"x", "y", "z"},
			},
		},
		StructM: StructM{
			M: "n",
		},
		M: mInterface,
		P: &structP{
			P: "r",
		},
	}
}()
