// Package aconfig provides simple but still powerful config loader.
//
// It can read configuration from different sources, like defaults, files, environment variables, console flag parameters.
//
// Defaults are defined in structure tags (`default` tag). For files JSON, YAML, TOML and .Env are supported.
//
// Environment variables and flag parameters can have an optional prefix to separate them from other entries.
//
// Also, aconfig is dependency-free, file decoders are used as separate modules (submodules to be exact) and are added to your go.mod only when used.
//
// Loader configuration (`Config` type) has different ways to configure loader, to skip some sources, define prefixes, fail on unknown params.
//
package aconfig
