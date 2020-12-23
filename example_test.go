package aconfig_test

import (
	"fmt"
	"log"
	"os"

	"github.com/cristalhq/aconfig"
)

type MyConfig struct {
	HTTPPort int `default:"1111" usage:"just a number"`
	Auth     struct {
		User string `default:"def-user" usage:"your user"`
		Pass string `default:"def-pass" usage:"make it strong"`
	}
}

func Example_SimpleUsage() {
	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{"/var/opt/myapp/config.json"},
		EnvPrefix:    "APP",
		FlagPrefix:   "app",
	})
	if err := loader.Load(); err != nil {
		log.Panic(err)
	}

	fmt.Printf("HTTPPort:  %v\n", cfg.HTTPPort)
	fmt.Printf("Auth.User: %q\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %q\n", cfg.Auth.Pass)

	// Output:
	//
	// HTTPPort:  0
	// Auth.User: ""
	// Auth.Pass: ""
}

func Example_WalkFields() {
	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	loader.WalkFields(func(f aconfig.Field) bool {
		fmt.Printf("%v: %q %q %q %q\n", f.Name(), f.Tag("env"), f.Tag("flag"), f.Tag("default"), f.Tag("usage"))
		return true
	})

	// Output:
	// HTTPPort: "HTTP_PORT" "http_port" "1111" "just a number"
	// Auth.User: "USER" "user" "def-user" "your user"
	// Auth.Pass: "PASS" "pass" "def-pass" "make it strong"
}

// Just load defaults from struct defenition.
//
func Example_Defaults() {
	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	if err := loader.Load(); err != nil {
		log.Panic(err)
	}

	fmt.Printf("HTTPPort:  %v\n", cfg.HTTPPort)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// HTTPPort:  1111
	// Auth.User: def-user
	// Auth.Pass: def-pass
}

// Load defaults from struct defenition and overwrite with a file.
//
func Example_File() {
	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipEnv:   true,
		SkipFlags: true,
		Files:     []string{"testdata/example_config.json"},
	})
	if err := loader.Load(); err != nil {
		log.Panic(err)
	}

	fmt.Printf("HTTPPort:  %v\n", cfg.HTTPPort)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// HTTPPort:  2222
	// Auth.User: json-user
	// Auth.Pass: json-pass
}

// Load defaults from struct defenition and overwrite with a file.
// And then overwrite with environment variables.
//
func Example_Env() {
	os.Setenv("EXAMPLE_HTTP_PORT", "3333")
	os.Setenv("EXAMPLE_AUTH_USER", "env-user")
	os.Setenv("EXAMPLE_AUTH_PASS", "env-pass")
	defer os.Clearenv()

	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFlags: true,
		EnvPrefix: "EXAMPLE",
		Files:     []string{"testdata/example_config.json"},
	})
	if err := loader.Load(); err != nil {
		log.Panic(err)
	}

	fmt.Printf("HTTPPort:  %v\n", cfg.HTTPPort)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// HTTPPort:  3333
	// Auth.User: env-user
	// Auth.Pass: env-pass
}

// Load defaults from struct defenition and overwrite with a file.
// And then overwrite with environment variables.
// Finally read command line flags.
//
func Example_Flag() {
	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		FlagPrefix: "ex",
		Files:      []string{"testdata/example_config.json"},
	})

	flags := loader.Flags() // <- IMPORTANT: use this to define your non-config flags
	flags.String("my.other.port", "1234", "debug port")

	// IMPORTANT: next statement is made only to hack flag params
	// to make test example work
	// feel free to remove it completely during copy-paste :)
	os.Args = append([]string{}, os.Args[0],
		"-ex.http_port=4444",
		"-ex.auth.user=flag-user",
		"-ex.auth.pass=flag-pass",
	)

	if err := loader.Load(); err != nil {
		log.Panic(err)
	}

	fmt.Printf("HTTPPort:  %v\n", cfg.HTTPPort)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// HTTPPort:  4444
	// Auth.User: flag-user
	// Auth.Pass: flag-pass
}
