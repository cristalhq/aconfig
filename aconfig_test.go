package aconfig

import (
	"embed"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

var newParser = os.Getenv("ACONFIG_NEW") == "true"

func TestTrueSkip(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		SkipFlags:    true,
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := TestConfig{}

	if have := cfg; !reflect.DeepEqual(have, want) {
		fmt.Printf("have: %+v\n", *have.Int)
		t.Fatalf("\nhave: %+v\nwant: %+v", have, want)
	}
}

func Test_parse(t *testing.T) {
	var cfg TestConfig2

	loader := LoaderFor(&cfg, Config{
		NewParser: newParser,
		SkipEnv:   true,
		SkipFlags: true,
	})
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	// fmt.Printf("\nresult: %+v\n", cfg)
	// fmt.Printf("b: %v c: %+v\n", *cfg.B, cfg.C)
}

type TestConfig2 struct {
	A int    `default:"1"`
	B *int32 `default:"10" json:"boom_boom"`
	C *int32 `env:"ccc"`
	D string `default:"str"`
	E struct {
		Bar int    `default:"42"`
		Foo string `default:"foo"`
	}
	F  map[string]int `default:"1:20,3:4"`
	F2 map[int]string `default:"1:2,3:40"`
	G  map[string]struct {
		Baz int `default:"1234"`
	} // `default:"1:1234"`
	H  []string            `default:"ab,cd,ef"`
	H2 []int               `default:"1,2,3"`
	I  map[string][]string `default:"1:a-b,2:c-d,3:e-f"`
	J  []struct {
		Quzz int
	} //`default:"1,2,3,4"`
	Y X
	X
}
type X struct {
	Xex string `default:"XEX" env:"XEXEXE" flag:"axaxa"`
}

type LogLevel int8

func (l *LogLevel) UnmarshalText(text []byte) error {
	switch string(text) {
	case "debug":
		*l = -1
	case "info":
		*l = 0
	case "warn":
		*l = 1
	case "error":
		*l = 2
	default:
		return fmt.Errorf("unknown log level: %s", text)
	}
	return nil
}

func TestDefaults(t *testing.T) {
	// type TestConfig struct {
	// 	Str      string `default:"str-def"`
	// 	Bytes    []byte `default:"bytes-def"`
	// 	Int      *int32 `default:"123"`
	// 	HTTPPort int    `default:"8080"`
	// 	Param    int    // no default tag, so default value
	// 	ParamPtr *int   // no default tag, so default value
	// 	Sub      SubConfig
	// 	Anon     struct {
	// 		IsAnon bool `default:"true"`
	// 	}
	// 	StrSlice []string       `default:"1,2,3" usage:"just pass strings"`
	// 	Slice    []int          `default:"1,2,3" usage:"just pass elements"`
	// 	Map1     map[string]int `default:"a:1,b:2,c:3"`
	// 	Map2     map[int]string `default:"1:a,2:b,3:c"`
	// 	EmbeddedConfig
	// }

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser: newParser,
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "str-def",
		Bytes:    []byte("bytes-def"),
		Int:      int32Ptr(123),
		HTTPPort: 8080,
		Sub:      SubConfig{Float: 123.123},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{IsAnon: true},
		StrSlice:       []string{"1", "2", "3"},
		Slice:          []int{1, 2, 3},
		Map1:           map[string]int{"a": 1, "b": 2, "c": 3},
		Map2:           map[int]string{1: "a", 2: "b", 3: "c"},
		EmbeddedConfig: EmbeddedConfig{Em: "em-def"},
	}
	mustEqual(t, cfg, want)
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

		Dur time.Duration `default:"1h2m3s"`
		// Time time.Time     `default:"2000-04-05 10:20:30 +0000 UTC"`

		Level LogLevel `default:"warn"`
	}

	var cfg AllTypesConfig
	loader := LoaderFor(&cfg, Config{
		NewParser: newParser,
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

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
		// TODO: support time
		// Time :2000-04-05 10:20:30 +0000 UTC,
		Level: LogLevel(1),
	}
	mustEqual(t, cfg, want)
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
		NewParser: newParser,
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

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
	mustEqual(t, cfg, want)
}

func TestJSON(t *testing.T) {
	const filepath = "testfile.json"

	var cfg structConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
		FileSystem:   fstest.MapFS{filepath: testfile},
	})
	failIfErr(t, loader.Load())

	want := wantConfig
	mustEqual(t, cfg, want)
}

func TestJSONWithOmitempty(t *testing.T) {
	const filepath = "testfile.json"

	var cfg struct {
		APIKey string `json:"b,omitempty"`
	}
	loader := LoaderFor(&cfg, Config{
		NewParser:          newParser,
		SkipDefaults:       true,
		SkipEnv:            true,
		SkipFlags:          true,
		AllowUnknownFields: true,
		Files:              []string{filepath},
		FileSystem:         fstest.MapFS{filepath: testfile},
	})
	failIfErr(t, loader.Load())
}

func TestCustomFile(t *testing.T) {
	const filepath = "custom.config"

	var cfg structConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
		FileSystem:   fstest.MapFS{filepath: testfile},
		FileDecoders: map[string]FileDecoder{
			".config": &jsonDecoder{},
		},
	})
	failIfErr(t, loader.Load())

	want := wantConfig
	mustEqual(t, cfg, want)
}

func TestFile(t *testing.T) {
	filepath := "testdata/config.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})
	failIfErr(t, loader.Load())

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
	mustEqual(t, cfg, want)
}

//go:embed testdata
var configEmbed embed.FS

func TestFileEmbed(t *testing.T) {
	filepath := "testdata/config.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
		FileSystem:   configEmbed,
	})
	failIfErr(t, loader.Load())

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
	mustEqual(t, cfg, want)
}

func TestFileMerging(t *testing.T) {
	file1 := "testdata/config1.json"
	file2 := "testdata/config2.json"
	file3 := "testdata/config3.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		MergeFiles:   true,
		Files:        []string{file1, file2, file3},
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "111",
		HTTPPort: 222,
		Sub: SubConfig{
			Float: 333.333,
		},
	}
	mustEqual(t, cfg, want)
}

func TestFileFlag(t *testing.T) {
	file1 := "testdata/config1.json"

	flags := []string{
		"-file_flag=testdata/config2.json",
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		MergeFiles:   true,
		FileFlag:     "file_flag",
		Files:        []string{file1},
		Args:         flags,
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "111",
		HTTPPort: 222,
	}
	mustEqual(t, cfg, want)
}

func TestBadFileFlag(t *testing.T) {
	flags := []string{
		"-file_flag=",
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		FileFlag:     "file_flag",
		Args:         flags,
	})
	failIfOk(t, loader.Load())
}

func TestNoFileFlagValue(t *testing.T) {
	file1 := "testdata/config1.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		FileFlag:     "file_flag",
		Files:        []string{file1},
		Args:         []string{}, // no file_flag
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "111",
		HTTPPort: 111,
	}
	mustEqual(t, cfg, want)
}

func TestEnv(t *testing.T) {
	t.Setenv("TST_STR", "str-env")
	t.Setenv("TST_BYTES", "bytes-env")
	t.Setenv("TST_INT", "121")
	t.Setenv("TST_HTTP_PORT", "3000")
	t.Setenv("TST_SUB_FLOAT", "222.333")
	t.Setenv("TST_ANON_IS_ANON", "true")
	t.Setenv("TST_EM", "em-env")
	defer os.Clearenv()

	// type TestConfig struct {
	// 	Sub SubConfig
	// }
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})
	failIfErr(t, loader.Load())

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
	mustEqual(t, cfg, want)
}

func TestFlag(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
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

	failIfErr(t, loader.Flags().Parse(flags))

	failIfErr(t, loader.Load())

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
	mustEqual(t, cfg, want)
}

func TestExactName(t *testing.T) {
	t.Setenv("STR", "str-env")
	t.Setenv("TST_STR", "bar-env")
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
		NewParser:        newParser,
		SkipDefaults:     true,
		SkipFiles:        true,
		SkipFlags:        true,
		AllowUnknownEnvs: true,
		EnvPrefix:        "TST",
	})
	failIfErr(t, loader.Load())

	want := ExactConfig{
		Foo: Foo{
			String: "str-env",
		},
		Bar: "bar-env",
	}
	mustEqual(t, cfg, want)
}

func TestSkipName(t *testing.T) {
	t.Setenv("STR", "str-env")
	t.Setenv("BAR", "bar-env")
	defer os.Clearenv()

	type Foo struct {
		String string `default:"str" env:"STR"`
	}
	type ExactConfig struct {
		Foo Foo    `env:"-"`
		Bar string `default:"def" env:"-"`
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		NewParser: newParser,
		SkipFiles: true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	want := ExactConfig{
		Foo: Foo{
			String: "str-env",
		},
		Bar: "def",
	}
	mustEqual(t, cfg, want)
}

func TestDuplicatedName(t *testing.T) {
	t.Setenv("FOO_BAR", "str-env")
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
		NewParser:       newParser,
		SkipFlags:       true,
		AllowDuplicates: true,
	})
	failIfErr(t, loader.Load())

	want := ExactConfig{
		Foo: Foo{
			Bar: "str-env",
		},
		FooBar: "str-env",
	}
	mustEqual(t, cfg, want)
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
		NewParser: newParser,
		SkipFlags: true,
	})

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "is duplicated") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestFailOnDuplicatedFlag(t *testing.T) {
	type Foo struct {
		Bar string `flag:"yes"`
		Baz string `flag:"yes"`
	}

	err := LoaderFor(&Foo{}, Config{NewParser: newParser}).Load()
	failIfOk(t, err)

	want := `init loader: duplicate flag "yes"`
	mustEqual(t, err.Error(), want)
}

func TestUsage(t *testing.T) {
	loader := LoaderFor(&EmbeddedConfig{}, Config{
		NewParser: newParser,
	})

	var builder strings.Builder
	flags := loader.Flags()
	flags.SetOutput(&builder)
	flags.PrintDefaults()

	have := builder.String()
	want := `  -em string
    	use... em...field. (default "em-def")
`
	mustEqual(t, have, want)
}

func TestBadDefauts(t *testing.T) {
	f := func(cfg any) {
		t.Helper()

		loader := LoaderFor(cfg, Config{
			NewParser: newParser,
			SkipFiles: true,
			SkipEnv:   true,
			SkipFlags: true,
		})
		failIfOk(t, loader.Load())
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
		Map map[string]int `default:"1:a;2:2"`
	}{})

	f(&struct {
		Map map[int]string `default:"a:1;"`
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
		t.Helper()
		t.Run(filepath, func(t *testing.T) {
			t.Helper()
			var cfg TestConfig
			loader := LoaderFor(&cfg, Config{
				NewParser:          newParser,
				SkipDefaults:       true,
				SkipEnv:            true,
				SkipFlags:          true,
				FailOnFileNotFound: true,
				Files:              []string{filepath},
				FileSystem: fstest.MapFS{
					"bad_config.json": &fstest.MapFile{Data: []byte(`{almost": "json`)},
					"unknown.ext":     &fstest.MapFile{},
				},
			})
			failIfOk(t, loader.Load())
		})
	}

	f("no_such_file.json")
	f("bad_config.json")
	f("unknown.ext")
}

func TestFailOnFileNotFound(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := LoaderFor(&TestConfig{}, Config{
			NewParser:          newParser,
			SkipDefaults:       true,
			SkipEnv:            true,
			SkipFlags:          true,
			FailOnFileNotFound: false,
			Files:              []string{filepath},
		})

		failIfErr(t, loader.Load())
	}

	f("testdata/config.json")
	f("testdata/not_found.json")
}

func TestBadEnvs(t *testing.T) {
	t.Setenv("TST_HTTP_PORT", "30a00")
	defer os.Clearenv()

	loader := LoaderFor(&TestConfig{}, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})

	failIfOk(t, loader.Load())
}

func TestBadFlags(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	args := []string{"-tst.param=10a01"}

	failIfErr(t, loader.Flags().Parse(args))
	failIfOk(t, loader.Load())
}

func TestUnknownFields(t *testing.T) {
	filepath := "testdata/unknown_fields.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "unknown field in file") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownEnvs(t *testing.T) {
	t.Setenv("TST_STR", "defined")
	t.Setenv("TST_UNKNOWN", "42")
	t.Setenv("JUST_ENV", "JUST_VALUE")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "unknown environment var") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownEnvsWithEmptyPrefix(t *testing.T) {
	t.Setenv("STR", "defined")
	t.Setenv("UNKNOWN", "42")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
	})

	failIfErr(t, loader.Load())
}

func TestUnknownFlags(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		NewParser:    newParser,
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
	flagSet.SetOutput(io.Discard)

	// define flag with a loader's prefix which is unknown
	flagSet.Int("tst.unknown", 42, "")
	flagSet.String("just_env", "just_def", "")

	failIfErr(t, flagSet.Parse(flags))

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownFlagsWithEmptyPrefix(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		NewParser:    newParser,
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
	flagSet.SetOutput(io.Discard)

	// define flag with a loader's prefix which is unknown
	flagSet.Int("unknown", 42, "")

	failIfErr(t, flagSet.Parse(flags))
	failIfErr(t, loader.Load())
}

// flag.FlagSet already fails on undefined flag.
func TestUnknownFlagsStdlib(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		NewParser:    newParser,
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
	flagSet.SetOutput(io.Discard)

	failIfOk(t, flagSet.Parse(flags))
}

func TestCustomEnvsAndArgs(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		Envs:         []string{"PARAM=2"},
		Args:         []string{"-str=4"},
	})

	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:   "4",
		Param: 2,
	}
	mustEqual(t, cfg, want)
}

func TestCustomNames(t *testing.T) {
	type TestConfig struct {
		A int `default:"-1" env:"ONE"`
		B int `default:"-1" flag:"two"`
		C int `default:"-1" env:"three" flag:"four"`
	}

	t.Setenv("ONE", "1")
	t.Setenv("three", "3")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		NewParser: newParser,
		Args:      []string{"-two=2", "-four=4"},
	})

	failIfErr(t, loader.Load())

	mustEqual(t, cfg.A, 1)
	mustEqual(t, cfg.B, 2)
	mustEqual(t, cfg.C, 4)
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
		NewParser:        newParser,
	}
	LoaderFor(&testConfig{}, cfg).WalkFields(func(f Field) bool {
		for _, tag := range []string{"json", "yaml", "env", "flag"} {
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
	if newParser {
		t.Skip()
	}
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

	LoaderFor(&TestConfig{}, Config{NewParser: newParser}).WalkFields(func(f Field) bool {
		wantFields := fields[i]
		mustEqual(t, f.Name(), wantFields.Name)
		mustEqual(t, f.Name(), wantFields.Name)
		if parent, ok := f.Parent(); ok {
			mustEqual(t, parent.Name(), wantFields.ParentName)
		}
		mustEqual(t, f.Tag("default"), wantFields.DefaultValue)
		mustEqual(t, f.Tag("usage"), wantFields.Usage)
		i++
		return true
	})

	mustEqual(t, i, 3)

	i = 0
	LoaderFor(&TestConfig{}, Config{NewParser: newParser}).WalkFields(func(f Field) bool {
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
		NewParser: newParser,
		SkipFlags: true,
		Args:      []string{},
	})
	failIfErr(t, loader.Load())

	if flags := loader.Flags().NFlag(); flags != 0 {
		t.Errorf("want empty, got %v", flags)
	}
}

func TestPassBadStructs(t *testing.T) {
	f := func(cfg any) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()

		_ = LoaderFor(cfg, Config{
			NewParser: newParser,
		})
	}

	f(nil)
	f(map[string]string{})
	f([]string{})
	f([4]string{})
	f(func() {})

	type S struct {
		Foo int
	}
	f(S{})
}

func TestBadRequiredTag(t *testing.T) {
	type TestConfig struct {
		Field string `required:"boom"`
	}

	f := func(cfg any) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()

		_ = LoaderFor(cfg, Config{
			NewParser: newParser,
		})
	}

	f(&TestConfig{})
}

func TestMissingFieldWithRequiredTag(t *testing.T) {
	cfg := struct {
		Field1 string `required:"true"`
	}{}
	loader := LoaderFor(&cfg, Config{
		SkipFlags: true,
	})

	err := loader.Load()
	want := "load config: fields required but not set: Field1"

	if have := err.Error(); have != want {
		t.Fatalf("got %v, want %v", err, want)
	}
}

func TestMissingFieldsWithRequiredTag(t *testing.T) {
	cfg := struct {
		Field1 string `required:"true"`
		Field2 string `required:"true"`
	}{}
	loader := LoaderFor(&cfg, Config{
		SkipFlags: true,
	})

	err := loader.Load()
	want := "load config: fields required but not set: Field1,Field2"

	if have := err.Error(); have != want {
		t.Fatalf("got %v, want %v", err, want)
	}
}

func int32Ptr(a int32) *int32 {
	return &a
}

type TestConfig struct {
	Str      string `default:"str-def"`
	Bytes    []byte `default:"bytes-def"`
	Int      *int32 `default:"123"`
	HTTPPort int    `default:"8080"`
	Param    int    // no default tag, so default value
	// ParamPtr *int   // no default tag, so default value
	Sub  SubConfig
	Anon struct {
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

	MM any `json:"MM"`

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

var testfile = &fstest.MapFile{Data: []byte(`{
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

	"MM":["q", "w"],
	
	"P": {
		"P": "r"
	}
}
`)}

var wantConfig = func() structConfig {
	i := int32(42)
	j := int64(420)
	mInterface := make([]any, 2)
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
		MM: mInterface,
		P: &structP{
			P: "r",
		},
	}
}()

type ConfigTest struct {
	VCenter ConfigVCenter `json:"vcenter" env:"VCENTER"`
}

type ConfigVCenter struct {
	User        string                  `json:"user" env:"USER"`
	Password    string                  `json:"password" env:"PASSWORD"`
	Port        string                  `json:"port" env:"PORT"`
	Datacenters []ConfigVCenterDCRegion `json:"datacenters" env:"-"`
}

type ConfigVCenterDCRegion struct {
	Region    string            `json:"region"`
	Addresses []ConfigVCenterDC `json:"addresses"`
}

type ConfigVCenterDC struct {
	Zone       string `json:"zone"`
	Address    string `json:"address"`
	Datacenter string `json:"datacenter"`
}

func TestSliceStructs(t *testing.T) {
	var cfg ConfigTest
	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{"testdata/complex.json"},
	})

	failIfErr(t, loader.Load())

	want := ConfigTest{
		VCenter: ConfigVCenter{
			User:     "user-test",
			Password: "pass-test",
			Port:     "8080",
			Datacenters: []ConfigVCenterDCRegion{
				{
					Region: "region-test",
					Addresses: []ConfigVCenterDC{
						{
							Zone:       "zone-test",
							Address:    "address-test",
							Datacenter: "datacenter-test",
						},
					},
				},
			},
		},
	}
	mustEqual(t, cfg, want)
}

func TestJSONMap(t *testing.T) {
	type TestConfig struct {
		Options map[string]float64
	}
	var cfg TestConfig

	loader := LoaderFor(&cfg, Config{
		NewParser:    newParser,
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{"testdata/toy.json"},
	})

	failIfErr(t, loader.Load())

	want := TestConfig{
		Options: map[string]float64{
			"foo": 0.4,
			"bar": 0.25,
		},
	}

	mustEqual(t, cfg, want)
}

func TestBad(t *testing.T) {
	t.Skip("probably too picky")

	type TestConfig struct {
		Params url.Values
	}
	var cfg TestConfig
	t.Setenv("PARAMS", "foo:bar")

	p, err := url.ParseQuery("foo=bar")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("have: %+v\n", p)

	loader := LoaderFor(&cfg, Config{
		NewParser: newParser,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Params: p,
	}
	mustEqual(t, cfg, want)
}

func TestFileConfigFlagDelim(t *testing.T) {
	type TestConfig struct {
		Options struct {
			Foo float64
			Bar float64
		}
	}
	var cfg TestConfig

	loader := LoaderFor(&cfg, Config{
		NewParser:     newParser,
		SkipDefaults:  true,
		SkipEnv:       true,
		SkipFlags:     true,
		FlagDelimiter: "_",
		Files:         []string{"testdata/toy.json"},
	})

	failIfErr(t, loader.Load())

	want := TestConfig{Options: struct {
		Foo float64
		Bar float64
	}{0.4, 0.25}}

	mustEqual(t, cfg, want)
}

func TestSliceOfStructsWithSliceOfPrimitives(t *testing.T) {
	type TestService struct {
		Name     string
		Strings  []string
		Integers []int
		Booleans []bool
	}

	type TestConfig struct {
		Services []TestService
	}
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{"testdata/slice-struct-primitive-slice.json"},
	})

	failIfErr(t, loader.Load())

	want := TestConfig{
		Services: []TestService{
			{
				Name:     "service1",
				Strings:  []string{"string1", "string2"},
				Integers: []int{1, 2},
				Booleans: []bool{true, false},
			},
		},
	}
	mustEqual(t, cfg, want)
}

func failIfOk(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fatal("must be non-nil")
	}
}

func failIfErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatal(err)
	}
}

func mustEqual(tb testing.TB, got, want any) {
	tb.Helper()
	if !reflect.DeepEqual(got, want) {
		tb.Fatalf("\nhave %+v\nwant %+v", got, want)
	}
}
