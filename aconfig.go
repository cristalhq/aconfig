package aconfig

import (
	"encoding/json"
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
	usageTag        = "usage"
)

// Loader of user configuration.
type Loader struct {
	config  LoaderConfig
	src     interface{}
	fields  []*fieldData
	flagSet *flag.FlagSet
}

// LoaderConfig to configure configuration loader.
type LoaderConfig struct {
	SkipDefaults bool
	SkipFile     bool
	SkipEnv      bool
	SkipFlag     bool

	EnvPrefix  string
	FlagPrefix string

	Files []string
}

// NewLoader creates a new Loader based on a config.
// Zero-value config is acceptable.
func NewLoaderFor(cfg interface{}, config LoaderConfig) *Loader {
	if config.EnvPrefix != "" {
		config.EnvPrefix += "_"
	}
	if config.FlagPrefix != "" {
		config.FlagPrefix += "."
	}
	l := &Loader{config: config}
	l.preLoad(cfg)
	return l
}

// Loader creates a new Loader based on a config.
// Zero-value config is acceptable.
func LoaderFor(src interface{}) *Loader {
	return &Loader{src: src}
}

func (l *Loader) SkipDefaults() *Loader {
	l.config.SkipDefaults = true
	return l
}

func (l *Loader) SkipFiles() *Loader {
	l.config.SkipFile = true
	return l
}

func (l *Loader) SkipEnvironment() *Loader {
	l.config.SkipEnv = true
	return l
}

func (l *Loader) SkipFlags() *Loader {
	l.config.SkipFlag = true
	return l
}

func (l *Loader) WithFiles(files []string) *Loader {
	l.config.Files = files
	return l
}

func (l *Loader) WithEnvPrefix(prefix string) *Loader {
	l.config.EnvPrefix = prefix
	return l
}

func (l *Loader) WithFlagPrefix(prefix string) *Loader {
	l.config.FlagPrefix = prefix
	return l
}

func (l *Loader) Build() *Loader {
	l.preLoad(l.src)
	return l
}

func (l *Loader) preLoad(cfg interface{}) {
	fields := getFields(cfg)

	l.flagSet = flag.NewFlagSet(l.config.FlagPrefix, flag.ContinueOnError)

	for _, field := range fields {
		flagName := l.getFlagName(field.Name)
		l.flagSet.String(flagName, field.DefaultValue, field.Usage)
	}
}

func (l *Loader) Flags() *flag.FlagSet {
	return l.flagSet
}

// Load configuration into a given param.
func (l *Loader) Load(into interface{}) error {
	l.fields = getFields(into)

	if err := l.loadSources(into); err != nil {
		return fmt.Errorf("aconfig: cannot load config: %w", err)
	}
	return nil
}

// LoadWithFiles configuration into a given param.
func (l *Loader) LoadWithFiles(into interface{}, files []string) error {
	l.fields = getFields(into)
	l.config.Files = files

	if err := l.loadSources(into); err != nil {
		return fmt.Errorf("aconfig: cannot load config: %w", err)
	}
	return nil
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
		if err := l.setFieldData(fd, fd.DefaultValue); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) loadFromFile(dst interface{}) error {
	for _, file := range l.config.Files {
		f, err := os.Open(file)
		if err != nil {
			return err
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
		if err != nil {
			return fmt.Errorf("file parsing error: %w", err)
		}
		break
	}
	return nil
}

func (l *Loader) loadEnvironment() error {
	for _, field := range l.fields {
		envName := l.getEnvName(field.Name)
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
		if err := l.flagSet.Parse(os.Args[1:]); err != nil {
			return err
		}
	}

	actualFlags := map[string]*flag.Flag{}
	l.flagSet.Visit(func(f *flag.Flag) {
		actualFlags[f.Name] = f
	})

	for _, field := range l.fields {
		flagName := l.getFlagName(field.Name)
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

func (l *Loader) getEnvName(name string) string {
	return strings.ToUpper(l.config.EnvPrefix + strings.ReplaceAll(name, ".", "_"))
}

func (l *Loader) getFlagName(name string) string {
	return strings.ToLower(l.config.FlagPrefix + name)
}

func (l *Loader) setFieldData(field *fieldData, value string) error {
	return setFieldDataHelper(field, value)
}

func getFields(x interface{}) []*fieldData {
	// TODO: check not struct
	valueObject := reflect.ValueOf(x).Elem()
	return getFieldsHelper(valueObject, nil)
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
	Name         string
	Field        reflect.StructField
	Value        reflect.Value
	DefaultValue string
	Usage        string
}

func newFieldData(field reflect.StructField, value reflect.Value, parent *fieldData) *fieldData {
	return &fieldData{
		Name:         makaName(field.Name, parent),
		Value:        value,
		Field:        field,
		DefaultValue: field.Tag.Get(defaultValueTag),
		Usage:        field.Tag.Get(usageTag),
	}
}

func newSimpleFieldData(value reflect.Value) *fieldData {
	return newFieldData(reflect.StructField{}, value, nil)
}

func makaName(name string, parent *fieldData) string {
	if parent == nil {
		return name
	}
	return parent.Name + "." + name
}

func setFieldDataHelper(field *fieldData, value string) error {
	// unwrap pointers
	for field.Value.Type().Kind() == reflect.Ptr {
		if field.Value.IsNil() {
			field.Value.Set(reflect.New(field.Value.Type().Elem()))
		}
		field.Value = field.Value.Elem()
	}

	if value == "" {
		return nil
	}

	switch kind := field.Value.Type().Kind(); kind {
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
	field.Value.SetBool(val)
	return nil
}

func setInt(field *fieldData, value string) error {
	val, err := strconv.ParseInt(value, 0, field.Value.Type().Bits())
	if err != nil {
		return err
	}
	field.Value.SetInt(val)
	return nil
}

func setInt64(field *fieldData, value string) error {
	if field.Field.Type == reflect.TypeOf(time.Second) {
		val, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		field.Value.Set(reflect.ValueOf(val))
		return nil
	}
	return setInt(field, value)
}

func setUint(field *fieldData, value string) error {
	val, err := strconv.ParseUint(value, 0, field.Value.Type().Bits())
	if err != nil {
		return err
	}
	field.Value.SetUint(val)
	return nil
}

func setFloat(field *fieldData, value string) error {
	val, err := strconv.ParseFloat(value, field.Value.Type().Bits())
	if err != nil {
		return err
	}
	field.Value.SetFloat(val)
	return nil
}

func setString(field *fieldData, value string) error {
	field.Value.SetString(value)
	return nil
}

func setSlice(field *fieldData, value string) error {
	vals := strings.Split(value, ",")
	slice := reflect.MakeSlice(field.Field.Type, len(vals), len(vals))
	for i, val := range vals {
		val = strings.TrimSpace(val)

		fd := newFieldData(reflect.StructField{}, slice.Index(i), nil)
		if err := setFieldDataHelper(fd, val); err != nil {
			return fmt.Errorf("incorrect slice item %q: %w", val, err)
		}
	}
	field.Value.Set(slice)
	return nil
}

func setMap(field *fieldData, value string) error {
	vals := strings.Split(value, ",")
	mapField := reflect.MakeMapWithSize(field.Field.Type, len(vals))

	for _, val := range vals {
		entry := strings.SplitN(val, ":", 2)
		if len(entry) != 2 {
			return fmt.Errorf("incorrect map item: %s", val)
		}
		key := strings.TrimSpace(entry[0])
		val := strings.TrimSpace(entry[1])

		fdk := newSimpleFieldData(reflect.New(field.Field.Type.Key()).Elem())
		if err := setFieldDataHelper(fdk, key); err != nil {
			return fmt.Errorf("incorrect map key %q: %w", key, err)
		}

		fdv := newSimpleFieldData(reflect.New(field.Field.Type.Elem()).Elem())
		if err := setFieldDataHelper(fdv, val); err != nil {
			return fmt.Errorf("incorrect map value %q: %w", val, err)
		}
		mapField.SetMapIndex(fdk.Value, fdv.Value)
	}
	field.Value.Set(mapField)
	return nil
}
