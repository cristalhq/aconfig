package aconfig_test

import (
	"fmt"
	"log"
	"os"

	"github.com/cristalhq/aconfig"
)

type MyConfig struct {
	Port int `default:"1111"`
	Auth struct {
		User string `default:"def-user"`
		Pass string `default:"def-pass"`
	}
}

func Example_NewApi() {
	loader := aconfig.LoaderFor(&MyConfig{}).
		SkipDefaults().SkipFiles().SkipEnvironment().SkipFlags().
		WithFiles([]string{"/var/opt/myapp/config.json"}).
		WithEnvPrefix("APP").
		WithFlagPrefix("app").
		Build()

	var cfg MyConfig
	if err := loader.Load(&cfg); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Port:      %v\n", cfg.Port)
	fmt.Printf("Auth.User: %q\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %q\n", cfg.Auth.Pass)

	// Output:
	//
	// Port:      0
	// Auth.User: ""
	// Auth.Pass: ""
}

// Just load defaults from struct defenition.
//
func Example_Defaults() {
	loader := aconfig.LoaderFor(&MyConfig{}).
		SkipFiles().
		SkipEnvironment().
		SkipFlags().
		Build()

	var cfg MyConfig
	if err := loader.Load(&cfg); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Port:      %v\n", cfg.Port)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// Port:      1111
	// Auth.User: def-user
	// Auth.Pass: def-pass
}

// Load defaults from struct defenition and overwrite with a file.
//
func Example_File() {
	loader := aconfig.LoaderFor(&MyConfig{}).
		SkipEnvironment().
		SkipFlags().
		WithFiles([]string{"testdata/example_config.json"}).
		Build()

	var cfg MyConfig
	if err := loader.Load(&cfg); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Port:      %v\n", cfg.Port)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// Port:      2222
	// Auth.User: json-user
	// Auth.Pass: json-pass
}

// Load defaults from struct defenition and overwrite with a file.
// And then overwrite with environment variables.
//
func Example_Env() {
	os.Setenv("EXAMPLE_PORT", "3333")
	os.Setenv("EXAMPLE_AUTH_USER", "env-user")
	os.Setenv("EXAMPLE_AUTH_PASS", "env-pass")
	defer os.Clearenv()

	loader := aconfig.LoaderFor(&MyConfig{}).
		SkipFlags().
		WithEnvPrefix("EXAMPLE").
		WithFiles([]string{"testdata/example_config.json"}).
		Build()

	var cfg MyConfig
	if err := loader.Load(&cfg); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Port:      %v\n", cfg.Port)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// Port:      3333
	// Auth.User: env-user
	// Auth.Pass: env-pass
}

// Load defaults from struct defenition and overwrite with a file.
// And then overwrite with environment variables.
// Finally read command line flags.
//
func Example_Flag() {
	loader := aconfig.LoaderFor(&MyConfig{}).
		WithEnvPrefix("EXAMPLE").
		WithFlagPrefix("ex").
		WithFiles([]string{"testdata/example_config.json"})

	flags := loader.Flags() // <- USE THIS TO DEFINE YOUR NON-CONFIG(!!) FLAGS

	flags.String("my.other.port", "1234", "debug port")

	var cfg MyConfig
	if err := loader.Load(&cfg); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Port:      %v\n", cfg.Port)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Next comment doesn't have `:` after `Output`
	// it's disabled due to additional flags passed via `go test` command
	// but it works, trust me :)

	// Output
	//
	// Port:      4444
	// Auth.User: flag-user
	// Auth.Pass: flag-pass
}
