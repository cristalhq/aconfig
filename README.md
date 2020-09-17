# aconfig

[![Build Status][build-img]][build-url]
[![GoDoc][pkg-img]][pkg-url]
[![Go Report Card][reportcard-img]][reportcard-url]
[![Coverage][coverage-img]][coverage-url]

Simple, useful and opinionated config loader.

## Rationale

There are more than 2000 repositories on Github regarding configuration in Go. I was looking for a simple configuration loader that will automate a lot of things for me. Idea was to load config from 4 common places: defaults (in the code), config files, environment variables, command-line flags. This library works with all of them.

## Features

* Simple API.
* Automates a lot of things.
  * config params names
  * config loading
  * generates documentation
* Opinionated.
* Supports different sources:
  * defaults in code
  * files (JSON, YAML, TOML)
  * environment variables
  * command-line flags
* Dependency-free (except file parsers).
* Walk over configuration fields.

## Install

Go version 1.14+

```
go get github.com/cristalhq/aconfig
```

## Example

```go
type MyConfig struct {
	Port int `default:"1111" usage:"just give a number"`
	Auth struct {
		User string `default:"def-user"`
		Pass string `default:"def-pass"`
	}
	Pass string `default:"" env:"SECRET" flag:"sec_ret"`
}

var cfg MyConfig
loader := aconfig.LoaderFor(&cfg).
	// feel free to skip some steps :)
	// SkipDefaults().
	// SkipFiles().
	// SkipEnvironment().
	// SkipFlags().
	WithFiles([]string{"file.json", "ouch.yaml"}).
	WithEnvPrefix("APP").
	WithFlagPrefix("app").
	Build()

flagSet := loader.Flags() // Important: use exactly this flags to define your own

if err := loader.Load(); err != nil {
	panic(err)
}

// configuration fields will be loaded from (in order):
//
// 1. defaults set in structure tags (see structure defenition)
// 2. loaded from files `file.json` if not `ouch.yaml` will be used
// 3. from corresponding environment variables with the prefix `APP_`
// 4. command-line flags with the prefix `app.` if they are 
```

Also see examples: [examples_test.go](https://github.com/cristalhq/aconfig/blob/master/example_test.go).

Integration with `spf13/cobra` [playground](https://play.golang.org/p/OsCR8qTCN0H).

## Documentation

See here: [pkg.go.dev][pkg-url].

## License

[MIT License](LICENSE).

[build-img]: https://github.com/cristalhq/aconfig/workflows/build/badge.svg
[build-url]: https://github.com/cristalhq/aconfig/actions
[pkg-img]: https://pkg.go.dev/badge/cristalhq/aconfig
[pkg-url]: https://pkg.go.dev/github.com/cristalhq/aconfig
[reportcard-img]: https://goreportcard.com/badge/cristalhq/aconfig
[reportcard-url]: https://goreportcard.com/report/cristalhq/aconfig
[coverage-img]: https://codecov.io/gh/cristalhq/aconfig/branch/master/graph/badge.svg
[coverage-url]: https://codecov.io/gh/cristalhq/aconfig
