package aconfig

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultValueTag = "default"
	usageTag        = "usage"
	jsonNameTag     = "json"
	yamlNameTag     = "yaml"
	tomlNameTag     = "toml"
	hclNameTag      = "hcl"
	envNameTag      = "env"
	flagNameTag     = "flag"
)

// Loader of user configuration.
type Loader struct {
	config  Config
	dst     interface{}
	fields  []*fieldData
	flagSet *flag.FlagSet
}

// Config to configure configuration loader.
type Config struct {
	SkipDefaults bool // SkipDefaults set to true will not load config from 'default' tag.
	SkipFiles    bool // SkipFiles set to true will not load config from files.
	SkipEnv      bool // SkipEnv set to true will not load config from environment variables.
	SkipFlags    bool // SkipFlags set to true will not load config from flag parameters.

	EnvPrefix  string // EnvPrefix for environment variables.
	FlagPrefix string // FlagPrefix for flag parameters.

	// envDelimiter for environment variables. Is always "_" due to env-var format.
	// Also unexported cause there is no sense to change it.
	envDelimiter string

	FlagDelimiter string // FlagDelimiter for flag parameters. If not set - default is ".".

	// AllFieldsRequired set to true will fail config loading if one of the fields was not set.
	// File, environment, flag must provide a value for the field.
	// If default is set and this option is enabled (or required tag is set) there will be an error.
	AllFieldRequired bool

	// AllowUnknownFields set to true will not fail on unknown fields in files.
	AllowUnknownFields bool

	// AllowUnknownEnvs set to true will not fail on unknown environment variables ().
	// When false error is returned only when EnvPrefix isn't empty.
	AllowUnknownEnvs bool

	// AllowUnknownFlags set to true will not fail on unknown flag parameters ().
	// When false error is returned only when FlagPrefix isn't empty.
	AllowUnknownFlags bool

	// DontGenerateTags disables tag generation for JSON, YAML, TOML file formats.
	DontGenerateTags bool

	// FailOnFileNotFound will stop Loader on a first not found file from Files field in this structure.
	FailOnFileNotFound bool

	// MergeFiles set to true will collect all the entries from all the given files.
	// Easy wat to cobine base.yaml with prod.yaml
	MergeFiles bool

	// FileFlag to make easier pass file with a config via flags.
	FileFlag string

	// Files from which config should be loaded.
	Files []string

	// Args hold the command-line arguments from which flags will be parsed.
	// By default is nil and then os.Args will be used.
	// Unless loader.Flags() will be explicitly parsed by the user.
	Args []string

	// FileDecoders to enable other than JSON file formats and prevent additional dependencies.
	// Add required submodules to the go.mod and register them in this field.
	// Example:
	//	FileDecoders: map[string]aconfig.FileDecoder{
	//		".yaml": aconfigyaml.New(),
	//		".toml": aconfigtoml.New(),
	//		".env": aconfigdotenv.New(),
	// 	}
	FileDecoders map[string]FileDecoder
}

// FileDecoder is used to read config from files. See aconfig submodules.
type FileDecoder interface {
	DecodeFile(filename string) (map[string]interface{}, error)
}

// Field of the user configuration structure.
// Done as an interface to export less things in lib.
type Field interface {
	// Name of the field.
	Name() string

	// Tag returns a given tag for a field.
	Tag(tag string) string

	// Parent of the current node.
	Parent() (Field, bool)
}

// LoaderFor creates a new Loader based on a given configuration structure.
// Supports only non-nil structures.
func LoaderFor(dst interface{}, cfg Config) *Loader {
	assertStruct(dst)

	l := &Loader{
		dst:    dst,
		config: cfg,
	}
	l.init()
	return l
}

func (l *Loader) init() {
	l.config.envDelimiter = "_"

	if l.config.FlagDelimiter == "" {
		l.config.FlagDelimiter = "."
	}

	if l.config.EnvPrefix != "" {
		l.config.EnvPrefix += l.config.envDelimiter
	}
	if l.config.FlagPrefix != "" {
		l.config.FlagPrefix += l.config.FlagDelimiter
	}

	if _, ok := l.config.FileDecoders[".json"]; !ok {
		if l.config.FileDecoders == nil {
			l.config.FileDecoders = map[string]FileDecoder{}
		}
		l.config.FileDecoders[".json"] = &jsonDecoder{}
	}

	if l.config.Args == nil {
		l.config.Args = os.Args[1:]
	}

	l.fields = l.getFields(l.dst)

	l.flagSet = flag.NewFlagSet(l.config.FlagPrefix, flag.ContinueOnError)
	if !l.config.SkipFlags {
		for _, field := range l.fields {
			flagName := l.config.FlagPrefix + l.fullTag(field, flagNameTag)
			l.flagSet.String(flagName, field.Tag(defaultValueTag), field.Tag(usageTag))
		}
	}
	if l.config.FileFlag != "" {
		// TODO: should be prefixed ?
		l.flagSet.String(l.config.FileFlag, "", "config file param")
	}
}

// Flags returngs flag.FlagSet to create your own flags.
// FlagSet name is Config.FlagPrefix and error handling is set to ContinueOnError.
func (l *Loader) Flags() *flag.FlagSet {
	return l.flagSet
}

// WalkFields iterates over configuration fields.
// Easy way to create documentation or user-friendly help.
func (l *Loader) WalkFields(fn func(f Field) bool) {
	for _, f := range l.fields {
		if !fn(f) {
			return
		}
	}
}

// Load configuration into a given param.
func (l *Loader) Load() error {
	if err := l.loadConfig(); err != nil {
		return fmt.Errorf("aconfig: cannot load config: %w", err)
	}
	return nil
}

func (l *Loader) loadConfig() error {
	if err := l.parseFlags(); err != nil {
		return err
	}
	if err := l.loadSources(); err != nil {
		return err
	}
	if err := l.checkRequired(); err != nil {
		return err
	}
	return nil
}

func (l *Loader) parseFlags() error {
	// TODO: too simple?
	if l.flagSet.Parsed() || l.config.SkipFlags {
		return nil
	}
	return l.flagSet.Parse(l.config.Args)
}

func (l *Loader) loadSources() error {
	if !l.config.SkipDefaults {
		if err := l.loadDefaults(); err != nil {
			return err
		}
	}
	if !l.config.SkipFiles {
		if err := l.loadFromFile(); err != nil {
			return err
		}
	}
	if !l.config.SkipEnv {
		if err := l.loadEnvironment(); err != nil {
			return err
		}
	}
	if !l.config.SkipFlags {
		if err := l.loadFlags(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) checkRequired() error {
	for _, field := range l.fields {
		if field.isSet {
			continue
		}
		if field.isRequired || l.config.AllFieldRequired {
			return fmt.Errorf("field %s was not set but it is required", field.name)
		}
	}
	return nil
}

func (l *Loader) loadDefaults() error {
	for _, field := range l.fields {
		defaultValue := field.Tag(defaultValueTag)
		if err := l.setFieldData(field, defaultValue); err != nil {
			return err
		}
		field.isSet = (defaultValue != "")
	}
	return nil
}

func (l *Loader) loadFromFile() error {
	if l.config.FileFlag != "" {
		flag := l.flagSet.Lookup(l.config.FileFlag)
		if flag != nil {
			configFile := flag.Value.String()
			if l.config.MergeFiles {
				l.config.Files = append(l.config.Files, configFile)
			} else {
				l.config.Files = []string{configFile}
			}
		}
	}

	for _, file := range l.config.Files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if l.config.FailOnFileNotFound {
				return err
			}
			continue
		}

		ext := strings.ToLower(filepath.Ext(file))
		decoder, ok := l.config.FileDecoders[ext]
		if !ok {
			return fmt.Errorf("file format '%q' isn't supported", ext)
		}

		actualFields, err := decoder.DecodeFile(file)
		if err != nil {
			return err
		}

		tag := ext[1:]

		for _, field := range l.fields {
			name := l.fullTag(field, tag)
			value, ok := actualFields[name]
			if !ok {
				actualFields = find(actualFields, name)
				value, ok = actualFields[name]
				if !ok {
					continue
				}
			}

			if err := l.setFieldData(field, value); err != nil {
				return err
			}
			field.isSet = true
			delete(actualFields, name)
		}

		if !l.config.AllowUnknownFields {
			for env, value := range actualFields {
				return fmt.Errorf("unknown field in file %q: %s=%s (see AllowUnknownFields config param)", file, env, value)
			}
		}

		if l.config.MergeFiles {
			continue
		}
		return nil
	}
	return nil
}

func (l *Loader) loadEnvironment() error {
	actualEnvs := getEnv()

	for _, field := range l.fields {
		envName := l.config.EnvPrefix + l.fullTag(field, envNameTag)

		if err := l.setField(field, envName, actualEnvs); err != nil {
			return err
		}
	}

	if !l.config.AllowUnknownEnvs && l.config.EnvPrefix != "" {
		for env, value := range actualEnvs {
			if strings.HasPrefix(env, l.config.EnvPrefix) {
				return fmt.Errorf("unknown environment var %s=%s (see AllowUnknownEnvs config param)", env, value)
			}
		}
	}
	return nil
}

func (l *Loader) loadFlags() error {
	actualFlags := getFlags(l.flagSet)

	for _, field := range l.fields {
		flagName := l.config.FlagPrefix + l.fullTag(field, flagNameTag)

		if err := l.setField(field, flagName, actualFlags); err != nil {
			return err
		}
	}

	if !l.config.AllowUnknownFlags && l.config.FlagPrefix != "" {
		for flag, value := range actualFlags {
			if strings.HasPrefix(flag, l.config.FlagPrefix) {
				return fmt.Errorf("unknown flag %s=%s (see AllowUnknownFlags config param)", flag, value)
			}
		}
	}
	return nil
}

func (l *Loader) setField(field *fieldData, name string, values map[string]interface{}) error {
	val, ok := values[name]
	if !ok {
		return nil
	}

	if err := l.setFieldData(field, val); err != nil {
		return err
	}

	field.isSet = true
	delete(values, name)
	return nil
}
