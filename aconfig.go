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
	errInit error
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

	// AllowDuplicates set to true will not fail on duplicated names on fields (env, flag, etc...)
	AllowDuplicates bool

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

	// FileFlag the name of the flag that defines the path to the configuration file passed through the CLI.
	// (To make it easier to transfer the config file via flags.)
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
	Format() string
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
		names := make(map[string]bool, len(l.fields))
		for _, field := range l.fields {
			flagName := l.fullTag(l.config.FlagPrefix, field, flagNameTag)
			if flagName == "" {
				continue
			}
			if names[flagName] && !l.config.AllowDuplicates {
				l.errInit = fmt.Errorf("duplicate flag %q", flagName)
				return
			}
			names[flagName] = true
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
	if l.errInit != nil {
		return fmt.Errorf("aconfig: cannot init loader: %w", l.errInit)
	}
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
			return fmt.Errorf("loading defaults: %w", err)
		}
	}
	if !l.config.SkipFiles {
		if err := l.loadFiles(); err != nil {
			return fmt.Errorf("loading files: %w", err)
		}
	}
	if !l.config.SkipEnv {
		if err := l.loadEnvironment(); err != nil {
			return fmt.Errorf("loading environment: %w", err)
		}
	}
	if !l.config.SkipFlags {
		if err := l.loadFlags(); err != nil {
			return fmt.Errorf("loading flags: %w", err)
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

func (l *Loader) loadFiles() error {
	if l.config.FileFlag != "" {
		if err := l.loadFileFlag(); err != nil {
			return err
		}
	}

	for _, file := range l.config.Files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if l.config.FailOnFileNotFound {
				return err
			}
			continue
		}

		if err := l.loadFile(file); err != nil {
			return err
		}

		if l.config.MergeFiles {
			continue
		}
		break
	}
	return nil
}

func (l *Loader) loadFile(file string) error {
	ext := strings.ToLower(filepath.Ext(file))
	decoder, ok := l.config.FileDecoders[ext]
	if !ok {
		return fmt.Errorf("file format %q is not supported", ext)
	}

	actualFields, err := decoder.DecodeFile(file)
	if err != nil {
		return err
	}

	tag := decoder.Format()

	for _, field := range l.fields {
		name := l.fullTag("", field, tag)
		if name == "" {
			continue
		}
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
			return fmt.Errorf("unknown field in file %q: %s=%v (see AllowUnknownFields config param)", file, env, value)
		}
	}
	return nil
}

func (l *Loader) loadFileFlag() error {
	fileFlag := getActualFlag(l.config.FileFlag, l.flagSet)
	if fileFlag == nil {
		return nil
	}

	configFile := fileFlag.Value.String()
	if configFile == "" {
		return fmt.Errorf("%s should not be empty", l.config.FileFlag)
	}

	if l.config.MergeFiles {
		l.config.Files = append(l.config.Files, configFile)
	} else {
		l.config.Files = []string{configFile}
	}
	return nil
}

func (l *Loader) loadEnvironment() error {
	actualEnvs := getEnv()
	dupls := make(map[string]struct{})

	for _, field := range l.fields {
		envName := l.fullTag(l.config.EnvPrefix, field, envNameTag)
		if envName == "" {
			continue
		}

		if err := l.setField(field, envName, actualEnvs, dupls); err != nil {
			return err
		}
	}

	return l.postEnvCheck(actualEnvs, dupls)
}

func (l *Loader) postEnvCheck(values map[string]interface{}, dupls map[string]struct{}) error {
	if l.config.AllowUnknownEnvs || l.config.EnvPrefix == "" {
		return nil
	}
	for name := range dupls {
		delete(values, name)
	}
	for env, value := range values {
		if strings.HasPrefix(env, l.config.EnvPrefix) {
			return fmt.Errorf("unknown environment var %s=%v (see AllowUnknownEnvs config param)", env, value)
		}
	}
	return nil
}

func (l *Loader) loadFlags() error {
	actualFlags := getFlags(l.flagSet)
	dupls := make(map[string]struct{})

	for _, field := range l.fields {
		flagName := l.fullTag(l.config.FlagPrefix, field, flagNameTag)
		if flagName == "" {
			continue
		}

		if err := l.setField(field, flagName, actualFlags, dupls); err != nil {
			return err
		}
	}
	return l.postFlagCheck(actualFlags, dupls)
}

func (l *Loader) postFlagCheck(values map[string]interface{}, dupls map[string]struct{}) error {
	if l.config.AllowUnknownFlags || l.config.FlagPrefix == "" {
		return nil
	}
	for name := range dupls {
		delete(values, name)
	}
	for flag, value := range values {
		if strings.HasPrefix(flag, l.config.EnvPrefix) {
			return fmt.Errorf("unknown flag %s=%v (see AllowUnknownFlags config param)", flag, value)
		}
	}
	return nil
}

// TODO(cristaloleg): revisit
func (l *Loader) setField(field *fieldData, name string, values map[string]interface{}, dupls map[string]struct{}) error {
	if !l.config.AllowDuplicates {
		if _, ok := dupls[name]; ok {
			return fmt.Errorf("field %q is duplicated", name)
		}
		dupls[name] = struct{}{}
	}

	val, ok := values[name]
	if !ok {
		return nil
	}

	if err := l.setFieldData(field, val); err != nil {
		return err
	}

	field.isSet = true
	if !l.config.AllowDuplicates {
		delete(values, name)
	}
	return nil
}
