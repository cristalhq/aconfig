package aconfig_test

import (
	"flag"
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

// Just load defaults from struct defenition.
//
func Example_Defaults() {
	loader := aconfig.NewLoader(aconfig.LoaderConfig{
		SkipFile: true,
		SkipEnv:  true,
		SkipFlag: true,
	})

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
	loader := aconfig.NewLoader(aconfig.LoaderConfig{
		SkipEnv:  true,
		SkipFlag: true,
		Files:    []string{"testdata/example_config.json"},
	})

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

	loader := aconfig.NewLoader(aconfig.LoaderConfig{
		SkipFlag:  true,
		EnvPrefix: "EXAMPLE",
		Files:     []string{"testdata/example_config.json"},
	})

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
	flag.String("ex.port", "4444", "app port")
	flag.String("ex.auth.user", "flag-user", "app user")
	flag.String("ex.auth.pass", "flag-pass", "app pass")

	loader := aconfig.NewLoader(aconfig.LoaderConfig{
		EnvPrefix:  "EXAMPLE",
		FlagPrefix: "ex",
		Files:      []string{"testdata/example_config.json"},
	})

	var cfg MyConfig
	if err := loader.Load(&cfg); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Port:      %v\n", cfg.Port)
	fmt.Printf("Auth.User: %v\n", cfg.Auth.User)
	fmt.Printf("Auth.Pass: %v\n", cfg.Auth.Pass)

	// Output:
	//
	// Port:      4444
	// Auth.User: flag-user
	// Auth.Pass: flag-pass
}
