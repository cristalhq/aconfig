package aconfig

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultValueTag = "default"
	usageTag        = "usage"
	envNameTag      = "env"
	flagNameTag     = "flag"
)

// Loader of user configuration.
type Loader struct {
	config  loaderConfig
	dst     interface{}
	fields  []*fieldData
	flagSet *flag.FlagSet
	isBuilt bool
}

// loaderConfig to configure configuration loader.
type loaderConfig struct {
	SkipDefaults bool
	SkipFile     bool
	SkipEnv      bool
	SkipFlag     bool

	EnvPrefix  string
	FlagPrefix string

	ShouldStopOnFileError bool
	Files                 []string
	FileDecoders          map[string]FileDecoder
}

// FileDecoder is used to read config from files. See aconfig submodules.
type FileDecoder interface {
	DecodeFile(filename string, dst interface{}) error
}

// Field of the user configuration structure.
// Done as an interface to export less things in lib.
type Field interface {
	// Name of the field.
	Name() string

	// DefaultValue of the field.
	DefaultValue() string

	// Usage of the field (set in `usage` tag).
	Usage() string

	// Tag returns a given tag for a field.
	Tag(tag string) string

	// Parent of the current node.
	Parent() (Field, bool)
}

// LoaderFor creates a new Loader based on a given configuration structure.
func LoaderFor(dst interface{}) *Loader {
	return &Loader{
		dst: dst,
		config: loaderConfig{
			FileDecoders: map[string]FileDecoder{},
		},
	}
}

// SkipDefaults if you don't want to use them.
func (l *Loader) SkipDefaults() *Loader {
	l.config.SkipDefaults = true
	return l
}

// SkipFiles if you don't want to use them.
func (l *Loader) SkipFiles() *Loader {
	l.config.SkipFile = true
	return l
}

// SkipEnvironment if you don't want to use it.
func (l *Loader) SkipEnvironment() *Loader {
	l.config.SkipEnv = true
	return l
}

// SkipFlags if you don't want to use them.
func (l *Loader) SkipFlags() *Loader {
	l.config.SkipFlag = true
	return l
}

// WithFileDecoder to decode files with a given extension (".json" is already added).
func (l *Loader) WithFileDecoder(ext string, decoder FileDecoder) *Loader {
	l.config.FileDecoders[ext] = decoder
	return l
}

// WithFiles for a configuration.
func (l *Loader) WithFiles(files []string) *Loader {
	l.config.Files = files
	return l
}

// WithEnvPrefix to specify environment prefix.
func (l *Loader) WithEnvPrefix(prefix string) *Loader {
	l.config.EnvPrefix = prefix
	if l.config.EnvPrefix != "" {
		l.config.EnvPrefix += "_"
	}
	return l
}

// WithFlagPrefix to specify command-line flags prefix.
func (l *Loader) WithFlagPrefix(prefix string) *Loader {
	l.config.FlagPrefix = prefix
	if l.config.FlagPrefix != "" {
		l.config.FlagPrefix += "."
	}
	return l
}

// StopOnFileError to stop configuration loading on file error.
func (l *Loader) StopOnFileError() *Loader {
	l.config.ShouldStopOnFileError = true
	return l
}

// Build to initialize flags for a given configuration.
func (l *Loader) Build() *Loader {
	_, ok := l.config.FileDecoders[".json"]
	if !ok {
		l.config.FileDecoders[".json"] = &jsonDecoder{}
	}

	l.parseFields(l.dst)
	l.isBuilt = true
	return l
}

func (l *Loader) parseFields(cfg interface{}) {
	l.flagSet = flag.NewFlagSet(l.config.FlagPrefix, flag.ContinueOnError)
	l.fields = getFields(cfg)

	if l.config.SkipFlag {
		return
	}
	for _, field := range l.fields {
		flagName := l.config.FlagPrefix + field.flagName
		l.flagSet.String(flagName, field.defaultValue, field.usage)
	}
}

// Flags returngs flag.FlagSet to create your own flags.
func (l *Loader) Flags() *flag.FlagSet {
	l.assertBuilt()
	return l.flagSet
}

// WalkFields iterates over configuration fields.
// Easy way to create documentation or other stuff.
func (l *Loader) WalkFields(fn func(f Field) bool) {
	l.assertBuilt()
	for _, f := range l.fields {
		if !fn(f) {
			return
		}
	}
}

// Load configuration into a given param.
func (l *Loader) Load() error {
	l.assertBuilt()
	if err := l.loadSources(l.dst); err != nil {
		return fmt.Errorf("aconfig: cannot load config: %w", err)
	}
	return nil
}

// LoadWithFile configuration into a given param.
func (l *Loader) LoadWithFile(file string) error {
	l.config.Files = []string{file}
	return l.Load()
}

func (l *Loader) assertBuilt() {
	if !l.isBuilt {
		panic("aconfig: you must run Build method before using the loader")
	}
}

func (l *Loader) loadSources(into interface{}) error {
	if !l.config.SkipDefaults {
		if err := l.loadDefaults(); err != nil {
			return err
		}
	}
	if !l.config.SkipFile {
		if err := l.loadFromFile(into); err != nil {
			return err
		}
	}
	if !l.config.SkipEnv {
		if err := l.loadEnvironment(); err != nil {
			return err
		}
	}
	if !l.config.SkipFlag {
		if err := l.loadFlags(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) loadDefaults() error {
	for _, fd := range l.fields {
		if err := setFieldData(fd, fd.defaultValue); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) loadFromFile(dst interface{}) error {
	for _, file := range l.config.Files {
		f, err := os.Open(file)
		if err != nil {
			if l.config.ShouldStopOnFileError {
				return err
			}
			continue
		}
		defer func() { _ = f.Close() }()

		ext := strings.ToLower(filepath.Ext(file))
		d, ok := l.config.FileDecoders[ext]
		if !ok {
			return fmt.Errorf("file format '%q' isn't supported", ext)
		}

		err = d.DecodeFile(file, dst)
		if err == nil {
			return nil
		}
		if l.config.ShouldStopOnFileError {
			return fmt.Errorf("file parsing error: %w", err)
		}
	}
	return nil
}

func (l *Loader) loadEnvironment() error {
	for _, field := range l.fields {
		envName := l.config.EnvPrefix + field.envName
		v, ok := os.LookupEnv(envName)
		if !ok {
			continue
		}
		if err := setFieldData(field, v); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) loadFlags() error {
	if !l.flagSet.Parsed() {
		if err := l.flagSet.Parse(os.Args[1:]); err != nil {
			return err
		}
	}

	actualFlags := map[string]*flag.Flag{}
	l.flagSet.Visit(func(f *flag.Flag) {
		actualFlags[f.Name] = f
	})

	for _, field := range l.fields {
		flagName := l.config.FlagPrefix + field.flagName
		flg, ok := actualFlags[flagName]
		if !ok {
			continue
		}
		if err := setFieldData(field, flg.Value.String()); err != nil {
			return err
		}
	}
	return nil
}

type jsonDecoder struct{}

// DecodeFile implements FileDecoder.
func (d *jsonDecoder) DecodeFile(filename string, dst interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return json.NewDecoder(f).Decode(dst)
}
