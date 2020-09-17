package aconfig_test

import (
	"fmt"
	"log"
	"os"

	"github.com/cristalhq/aconfig"
)

type MyConfig struct {
	HTTPPort int `default:"1111"`
	Auth     struct {
		User string `default:"def-user"`
		Pass string `default:"def-pass"`
	}
}

func Example_NewApi() {
	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg).
		SkipDefaults().SkipFiles().SkipEnvironment().SkipFlags().
		WithFiles([]string{"/var/opt/myapp/config.json"}).
		WithEnvPrefix("APP").
		WithFlagPrefix("app").
		Build()

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

// Just load defaults from struct defenition.
//
func Example_Defaults() {
	var cfg MyConfig
	loader := aconfig.LoaderFor(&cfg).
		SkipFiles().
		SkipEnvironment().
		SkipFlags().
		Build()

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
	loader := aconfig.LoaderFor(&cfg).
		SkipEnvironment().
		SkipFlags().
		WithFiles([]string{"testdata/example_config.json"}).
		Build()

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
	loader := aconfig.LoaderFor(&cfg).
		SkipFlags().
		WithEnvPrefix("EXAMPLE").
		WithFiles([]string{"testdata/example_config.json"}).
		Build()

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
	loader := aconfig.LoaderFor(&cfg).
		WithFlagPrefix("ex").
		WithFiles([]string{"testdata/example_config.json"}).
		Build()

	flags := loader.Flags() // <- IMPORTANT: use this to define your non-config flags
	flags.String("my.other.port", "1234", "debug port")

	// IMPORTANT: next statement is made only to hack flag params
	// to make test example work
	// feel free to remove it completely during copy-paste :)
	os.Args = append([]string{}, os.Args[0],
		"-ex.http_port=4444",
		"-ex.auth.user=flag-user",
		"-ex.auth.pass=flag-pass")

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
