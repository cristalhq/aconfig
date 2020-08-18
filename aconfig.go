package aconfig

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

const (
	defaultValueTag = "default"
	envNameTag      = "env"
	flagNameTag     = "flag"
	usageTag        = "usage"
)

// Loader of user configuration.
type Loader struct {
	config  loaderConfig
	src     interface{}
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

	FailOnNotParsedFlags  bool
	ShouldStopOnFileError bool
	Files                 []string
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
func LoaderFor(src interface{}) *Loader {
	return &Loader{src: src}
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

// FailOnNotParsedFlags to not forget parse flags explicitly.
// Use `l.FlagSet().Parse(os.Args[1:])` in your code for this.
//
func (l *Loader) FailOnNotParsedFlags() *Loader {
	l.config.FailOnNotParsedFlags = true
	return l
}

// StopOnFileError to stop configuration loading on file error.
func (l *Loader) StopOnFileError() *Loader {
	l.config.ShouldStopOnFileError = true
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

// Build to initialize flags for a given configuration.
func (l *Loader) Build() *Loader {
	l.parseFields(l.src)
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
		flagName := l.getFlagName(field)
		l.flagSet.String(flagName, field.defaultValue, field.usage)
	}
}

// Flags returngs flag.FlagSet to create your own flags.
func (l *Loader) Flags() *flag.FlagSet {
	if !l.isBuilt {
		panic("aconfig: you must run Build method before using the loader")
	}
	return l.flagSet
}

// WalkFields iterates over configuration fields.
// Easy way to create documentation or other stuff.
func (l *Loader) WalkFields(fn func(f Field) bool) {
	if !l.isBuilt {
		panic("aconfig: you must run Build method before using the loader")
	}
	for _, f := range l.fields {
		if !fn(f) {
			return
		}
	}
}

// Load configuration into a given param.
func (l *Loader) Load(into interface{}) error {
	if !l.isBuilt {
		panic("aconfig: you must run Build method before using the loader")
	}
	// we need to get fields once more, 'cause `into` is new for us
	l.fields = getFields(into)

	if err := l.loadSources(into); err != nil {
		return fmt.Errorf("aconfig: cannot load config: %w", err)
	}
	return nil
}

// LoadWithFile configuration into a given param.
func (l *Loader) LoadWithFile(into interface{}, file string) error {
	l.config.Files = []string{file}

	return l.Load(into)
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
		if err := l.setFieldData(fd, fd.defaultValue); err != nil {
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
		switch ext {
		case ".yaml", ".yml":
			err = yaml.NewDecoder(f).Decode(dst)
		case ".json":
			err = json.NewDecoder(f).Decode(dst)
		case ".toml":
			_, err = toml.DecodeReader(f, dst)
		default:
			return fmt.Errorf("file format '%q' isn't supported", ext)
		}

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
		envName := l.getEnvName(field)
		v, ok := os.LookupEnv(envName)
		if !ok {
			continue
		}
		if err := l.setFieldData(field, v); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) loadFlags() error {
	if !l.flagSet.Parsed() {
		if l.config.FailOnNotParsedFlags {
			return errors.New("aconfig: flags must be parsed")
		}
		return nil
	}

	actualFlags := map[string]*flag.Flag{}
	l.flagSet.Visit(func(f *flag.Flag) {
		actualFlags[f.Name] = f
	})

	for _, field := range l.fields {
		flagName := l.getFlagName(field)
		flg, ok := actualFlags[flagName]
		if !ok {
			continue
		}
		if err := l.setFieldData(field, flg.Value.String()); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) getEnvName(field *fieldData) string {
	name := field.name
	if field.envName != "" {
		name = field.envName
	}
	return strings.ToUpper(l.config.EnvPrefix + strings.ReplaceAll(name, ".", "_"))
}

func (l *Loader) getFlagName(field *fieldData) string {
	name := field.name
	if field.flagName != "" {
		name = field.flagName
	}
	return strings.ToLower(l.config.FlagPrefix + name)
}

func (l *Loader) setFieldData(field *fieldData, value string) error {
	return setFieldDataHelper(field, value)
}

func getFields(x interface{}) []*fieldData {
	value := reflect.ValueOf(x)
	for value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		panic("aconfig: only struct can be passed to the loader")
	}
	return getFieldsHelper(value, nil)
}

func getFieldsHelper(valueObject reflect.Value, parent *fieldData) []*fieldData {
	typeObject := valueObject.Type()
	count := valueObject.NumField()

	fields := make([]*fieldData, 0, count)
	for i := 0; i < count; i++ {
		value := valueObject.Field(i)
		field := typeObject.Field(i)

		if !value.CanSet() {
			continue
		}

		// TODO: pointers

		fd := newFieldData(field, value, parent)

		// if just a field - add and process next, else expand struct
		if field.Type.Kind() == reflect.Struct {
			var subFieldParent *fieldData
			if field.Anonymous {
				subFieldParent = parent
			} else {
				subFieldParent = fd
			}
			fields = append(fields, getFieldsHelper(value, subFieldParent)...)
			continue
		}
		fields = append(fields, fd)
	}
	return fields
}

type fieldData struct {
	name         string
	parent       *fieldData
	field        reflect.StructField
	value        reflect.Value
	defaultValue string
	envName      string
	flagName     string
	usage        string
}

func newFieldData(field reflect.StructField, value reflect.Value, parent *fieldData) *fieldData {
	return &fieldData{
		name:         makeName(field.Name, parent),
		parent:       parent,
		value:        value,
		field:        field,
		defaultValue: field.Tag.Get(defaultValueTag),
		envName:      field.Tag.Get(envNameTag),
		flagName:     field.Tag.Get(flagNameTag),
		usage:        field.Tag.Get(usageTag),
	}
}

func newSimpleFieldData(value reflect.Value) *fieldData {
	return newFieldData(reflect.StructField{}, value, nil)
}

func makeName(name string, parent *fieldData) string {
	if parent == nil {
		return name
	}
	return parent.name + "." + name
}

func (f *fieldData) Name() string {
	return f.name
}

// DefaultValue of the field.
func (f *fieldData) DefaultValue() string {
	return f.defaultValue
}

// Usage of the field (set in `usage` tag) .
func (f *fieldData) Usage() string {
	return f.usage
}

// Tag returns a given tag for a field.
func (f *fieldData) Tag(tag string) string {
	return f.field.Tag.Get(tag)
}

func (f *fieldData) Parent() (Field, bool) {
	return f.parent, f.parent != nil
}

func setFieldDataHelper(field *fieldData, value string) error {
	// unwrap pointers
	for field.value.Type().Kind() == reflect.Ptr {
		if field.value.IsNil() {
			field.value.Set(reflect.New(field.value.Type().Elem()))
		}
		field.value = field.value.Elem()
	}

	if value == "" {
		return nil
	}

	switch kind := field.value.Type().Kind(); kind {
	case reflect.Bool:
		return setBool(field, value)

	case reflect.String:
		return setString(field, value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return setInt(field, value)

	case reflect.Int64:
		return setInt64(field, value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setUint(field, value)

	case reflect.Float32, reflect.Float64:
		return setFloat(field, value)

	case reflect.Slice:
		return setSlice(field, value)

	case reflect.Map:
		return setMap(field, value)

	default:
		return fmt.Errorf("type kind %q isn't supported", kind)
	}
}

func setBool(field *fieldData, value string) error {
	val, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	field.value.SetBool(val)
	return nil
}

func setInt(field *fieldData, value string) error {
	val, err := strconv.ParseInt(value, 0, field.value.Type().Bits())
	if err != nil {
		return err
	}
	field.value.SetInt(val)
	return nil
}

func setInt64(field *fieldData, value string) error {
	if field.field.Type == reflect.TypeOf(time.Second) {
		val, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		field.value.Set(reflect.ValueOf(val))
		return nil
	}
	return setInt(field, value)
}

func setUint(field *fieldData, value string) error {
	val, err := strconv.ParseUint(value, 0, field.value.Type().Bits())
	if err != nil {
		return err
	}
	field.value.SetUint(val)
	return nil
}

func setFloat(field *fieldData, value string) error {
	val, err := strconv.ParseFloat(value, field.value.Type().Bits())
	if err != nil {
		return err
	}
	field.value.SetFloat(val)
	return nil
}

func setString(field *fieldData, value string) error {
	field.value.SetString(value)
	return nil
}

func setSlice(field *fieldData, value string) error {
	vals := strings.Split(value, ",")
	slice := reflect.MakeSlice(field.field.Type, len(vals), len(vals))
	for i, val := range vals {
		val = strings.TrimSpace(val)

		fd := newFieldData(reflect.StructField{}, slice.Index(i), nil)
		if err := setFieldDataHelper(fd, val); err != nil {
			return fmt.Errorf("incorrect slice item %q: %w", val, err)
		}
	}
	field.value.Set(slice)
	return nil
}

func setMap(field *fieldData, value string) error {
	vals := strings.Split(value, ",")
	mapField := reflect.MakeMapWithSize(field.field.Type, len(vals))

	for _, val := range vals {
		entry := strings.SplitN(val, ":", 2)
		if len(entry) != 2 {
			return fmt.Errorf("incorrect map item: %s", val)
		}
		key := strings.TrimSpace(entry[0])
		val := strings.TrimSpace(entry[1])

		fdk := newSimpleFieldData(reflect.New(field.field.Type.Key()).Elem())
		if err := setFieldDataHelper(fdk, key); err != nil {
			return fmt.Errorf("incorrect map key %q: %w", key, err)
		}

		fdv := newSimpleFieldData(reflect.New(field.field.Type.Elem()).Elem())
		if err := setFieldDataHelper(fdv, val); err != nil {
			return fmt.Errorf("incorrect map value %q: %w", val, err)
		}
		mapField.SetMapIndex(fdk.value, fdv.value)
	}
	field.value.Set(mapField)
	return nil
}
