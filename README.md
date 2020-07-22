# aconfig

[![Build Status][build-img]][build-url]
[![GoDoc][doc-img]][doc-url]
[![Go Report Card][reportcard-img]][reportcard-url]
[![Coverage][coverage-img]][coverage-url]

Simple, opinionated and helpful config loader. Defaults, files, environemnt, flags - doesn't matter. Everything will be scanned.

## Features

* Simple API.
* Automates a lot of things.
* Opinionated.
* Supports different sources:
  * defaults in code
  * files (JSON, YAML, TOML)
  * environemtn variables
  * command flags  
* Dependency-free (except file parsers)

## Install

Go version 1.14+

```
go get github.com/cristalhq/aconfig
```

## Example

```go
type MyConfig struct {
	Port int `default:"port"`
	Auth struct {
		User string `default:"admin"`
		Pass stirng `default:"github"`
	}
}

loader := aconfig.NewLoader(aconfig.Config{
	UseDefaults: true,
	UseFile:     true,
	UseEnv:      true,
	UseFlag:     true,
	Files:       []string{"file.json", "ouch.yaml"},
	FlagPrefix:  "app",
	EnvPrefix:   "APP",
})

var cfg MyConfig
if err := loader.Load(&cfg); err != nil {
	panic(err)
}

// configuration fields will be loaded from:
//
// 1. defaults set in structure tags (see structure defenition)
// 2. loaded from files `file.json` if not `ouch.yaml`
// 3. than corresponding environment variables with prefix `APP`
// 4. and command-line flags if they are
```

Also see examples: TODO

## Documentation

See [these docs][doc-url].

## License

[MIT License](LICENSE).

[build-img]: https://github.com/cristalhq/aconfig/workflows/build/badge.svg
[build-url]: https://github.com/cristalhq/aconfig/actions
[doc-img]: https://godoc.org/github.com/cristalhq/aconfig?status.svg
[doc-url]: https://pkg.go.dev/github.com/cristalhq/aconfig
[reportcard-img]: https://goreportcard.com/badge/cristalhq/aconfig
[reportcard-url]: https://goreportcard.com/report/cristalhq/aconfig
[coverage-img]: https://codecov.io/gh/cristalhq/aconfig/branch/master/graph/badge.svg
[coverage-url]: https://codecov.io/gh/cristalhq/aconfig
