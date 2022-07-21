package aconfig

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

type structParser struct {
	dst       interface{}
	cfg       Config
	fields    map[string]interface{}
	flagSet   *flag.FlagSet
	envNames  map[string]struct{}
	flagNames map[string]struct{}
}

func newStructParser(cfg Config) *structParser {
	return &structParser{
		cfg:       cfg,
		flagSet:   flag.NewFlagSet(cfg.FlagPrefix, flag.ContinueOnError),
		envNames:  map[string]struct{}{},
		flagNames: map[string]struct{}{},
	}
}

type parsedField struct {
	name         string
	namefull     string
	value        interface{}
	defaultValue interface{}
	parent       *parsedField
	childs       map[string]interface{}
	tags         map[string]string
	hasChilds    bool
	isRequired   bool
	isSet        bool
}

func (pf *parsedField) String() string {
	if pf == nil {
		return "<nil-pf>"
	}
	return fmt.Sprintf("%+v", *pf)
}

func (sp *structParser) newParseField(parent *parsedField, field reflect.StructField) (*parsedField, error) {
	requiredTag := field.Tag.Get("required")
	if requiredTag != "" && requiredTag != "true" {
		panic(fmt.Sprintf("aconfig: value for 'required' tag can be only 'true' got: %q", requiredTag))
	}

	name := field.Tag.Get("name")
	if name == "" {
		name = field.Name
	}

	newName := strings.ToLower(strings.Join(splitNameByWords(name), "_"))

	env := field.Tag.Get("env")
	if env == "" {
		env = strings.ToUpper(newName)
	}

	flag := field.Tag.Get("flag")
	if flag == "" {
		flag = newName
	}

	var parentName, parentEnv, parentFlag string
	if parent != nil {
		parentName = parent.namefull + "|"

		for p := parent; p != nil; p = p.parent {
			parentEnv = p.tags["env_name"]
			if parentEnv != "-" {
				break
			}
		}
		for p := parent; p != nil; p = p.parent {
			parentFlag = p.tags["flag_name"]
			if parentFlag != "-" {
				break
			}
		}

		parentEnv += sp.cfg.envDelimiter
		parentFlag += sp.cfg.FlagDelimiter
	}

	pfield := &parsedField{
		name:     name,
		namefull: parentName + name,
		parent:   parent,
		tags: map[string]string{
			"usage":     field.Tag.Get("usage"),
			"env_name":  env,
			"env_full":  sp.cfg.EnvPrefix + parentEnv + env,
			"flag_name": flag,
			"flag_full": sp.cfg.FlagPrefix + parentFlag + flag,
		},
		isRequired: requiredTag == "true",
	}

	if !sp.cfg.SkipDefaults {
		// TODO: must be typed?
		pfield.defaultValue = field.Tag.Get("default")
	}

	if env == "-" {
		delete(pfield.tags, "env_full")
	}
	if flag == "-" {
		delete(pfield.tags, "flag_full")
	}

	if exactName, _, ok := strings.Cut(env, ",exact"); ok {
		pfield.tags["env_full"] = exactName
	}
	if exactName, _, ok := strings.Cut(flag, ",exact"); ok {
		pfield.tags["flag_full"] = exactName
	}

	if !sp.cfg.AllowDuplicates {
		name := pfield.tags["env_full"]
		if _, ok := sp.envNames[name]; ok && name != "" {
			return nil, fmt.Errorf("field %q is duplicated", name)
		}
		sp.envNames[name] = struct{}{}
	}

	if !sp.cfg.SkipFlags {
		flagName := pfield.tags["flag_full"]
		if flagName != "" {
			if _, ok := sp.flagNames[flagName]; ok && !sp.cfg.AllowDuplicates {
				return nil, fmt.Errorf("duplicate flag %q", flagName)
			}
			sp.flagNames[flagName] = struct{}{}
			// TODO: must be typed
			sp.flagSet.String(flagName, field.Tag.Get("default"), field.Tag.Get("usage"))
		}
	}

	if sp.cfg.DontGenerateTags {
		newName = name
	}
	for _, dec := range sp.cfg.FileDecoders {
		format := dec.Format()
		v := field.Tag.Get(format)
		if v == "" {
			v = newName
		}
		pfield.tags[format] = v
	}
	return pfield, nil
}

func (sp *structParser) parseStruct(x interface{}) error {
	value := reflect.ValueOf(x)
	if value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}

	fields, err := sp.parseStructHelper(nil, value, map[string]interface{}{})
	if err != nil {
		return err
	}
	sp.fields = fields

	// fmt.Printf("fields: %+v\n", fields)
	return nil
}

func (sp *structParser) parseStructHelper(parent *parsedField, structValue reflect.Value, res map[string]interface{}) (map[string]interface{}, error) {
	count := structValue.NumField()
	structType := structValue.Type()

	for i := 0; i < count; i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)
		fieldType := fieldValue.Type()
		if !fieldValue.CanSet() {
			continue
		}

		defaultTagValue := field.Tag.Get("default")
		pfield, err := sp.newParseField(parent, field)
		if err != nil {
			return nil, err
		}

		// do not set defaultValue for struct or pointer type without a default value
		// if fieldType.Kind() == reflect.Struct ||
		// 	(fieldType.Kind() == reflect.Pointer && defaultTagValue == "") {
		// 	pfield.defaultValue = nil
		// }

		if fieldType.Kind() == reflect.Pointer {
			fieldValue = fieldValue.Elem()
			fieldValue = reflect.New(fieldType)
			fieldType = fieldValue.Type()
		}

		value := fieldValue.Interface() // to have 'value' of type field

		// if !sp.cfg.SkipDefaults {
		// 	pv := fieldValue.Addr().Interface()
		// 	if v, ok := pv.(encoding.TextUnmarshaler); ok {
		// 		value = defaultTagValue
		// 		err := v.UnmarshalText([]byte(fmt.Sprint(value)))
		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 	}
		// 	pfield.value =
		// 	res[pfield.name] = pfield
		// 	continue
		// }

		switch fieldType.Kind() {
		// case reflect.Array:
		// TODO: same as slice + check len?

		case reflect.Interface:
			// TODO: just assign?

		case reflect.Struct:
			pfield.hasChilds = true

			param := map[string]interface{}{}
			parent := pfield
			if field.Anonymous {
				pfield.hasChilds = false
				param = res
				parent = pfield.parent
			}

			values, err := sp.parseStructHelper(parent, fieldValue, param)
			if err != nil {
				return nil, err
			}
			// fmt.Printf("field: %+v got: %+v\n\n", pfield.name, values)

			value = values

		case reflect.Slice, reflect.Array:
			if isPrimitive(field.Type.Elem()) {
				// byte-slice case
				if field.Type.Elem().Kind() == reflect.Uint8 {
					value = []byte(defaultTagValue)
				} else {
					values := []interface{}{}
					if defaultTagValue != "" && strings.Index(defaultTagValue, ",") == -1 {
						return nil, fmt.Errorf("incorrect default tag value for slice/array: %v", defaultTagValue)
					}
					for _, val := range strings.Split(defaultTagValue, ",") {
						values = append(values, val)
					}
					value = values
				}
			} else {
				pfield.hasChilds = true
				// TODO: if value is struct - parse
				// value = parseSlice(fieldValue, map[string]interface{}{})
			}

			// if !sp.cfg.SkipDefaults {
			// 	pfield.value = value
			// }

		case reflect.Map:
			// if isPrimitive(field.Type.Elem()) {
			values := map[string]interface{}{}
			parts := strings.Split(defaultTagValue, ",")
			if defaultTagValue != "" && strings.Index(defaultTagValue, ",") == -1 {
				return nil, fmt.Errorf("incorrect default tag value for map: %v", defaultTagValue)
			}

			if len(parts) > 1 {
				for _, entry := range parts {
					// fmt.Printf("parts: %+v\n", parts)
					entries := strings.SplitN(entry, ":", 2)
					if len(entries) != 2 {
						return nil, fmt.Errorf("want 2 parts got %d (%s)", len(entries), entries)
					}
					// TODO: convert entry[1] to a primitive?
					values[entries[0]] = entries[1]
				}
			}
			value = values
			// } else {
			// 	pfield.hasChilds = true
			// }

		default:
			// TODO: do not set pointer
			if fieldType.Kind() == reflect.Pointer && defaultTagValue == "" {
				// skip
				value = nil
			} else {
				// TODO: when WeaklyTypedInput will be false use decodePrimitive(...)
				if !sp.cfg.SkipDefaults {
					value = defaultTagValue
					if fieldType == reflect.TypeOf(time.Second) {
						val, err := time.ParseDuration(defaultTagValue)
						if err != nil {
							return nil, err
						}
						value = val
					}
				}

			}
		}

		// we should not overwrite struct because there are childs
		if sp.cfg.SkipDefaults && fieldType.Kind() != reflect.Struct {
			pfield.value = fieldValue.Interface()
		} else {
			pfield.value = value
		}

		// fmt.Printf("def: %v %T '%+v'\n", fieldType.String(), value, value)
		res[pfield.name] = pfield
	}
	return res, nil
}

var fieldType = reflect.TypeOf(&parsedField{})

var hook = mapstructure.DecodeHookFuncType(func(from, to reflect.Type, data interface{}) (interface{}, error) {
	if from != fieldType {
		// fmt.Printf("hook: got %T (%+v) when %s\n", i, i, to.String())
		return data, nil
	}
	field := data.(*parsedField)

	ifaceTo := reflect.New(to).Interface()
	if unmarshaller, ok := ifaceTo.(encoding.TextUnmarshaler); ok {
		// TODO: only string can be here?
		b := []byte(field.value.(string))
		err := unmarshaller.UnmarshalText(b)
		return unmarshaller, err
	}
	// fmt.Printf("hook: when %s do '%+v' // %+v\n\n", to.String(), field.value, field)
	return field.value, nil
})

func (sp *structParser) apply(x interface{}) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           x,
		DecodeHook:       hook,
		WeaklyTypedInput: true, // TODO: temp fix?
	})
	if err != nil {
		panic(fmt.Sprintf("aconfig: BUG with mapstructure.NewDecoder: %v", err))
	}

	if err := dec.Decode(sp.fields); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func (sp *structParser) applyLevel(tag string, values map[string]interface{}) error {
	if err := sp.applyLevelHelper2(sp.fields, tag, values); err != nil {
		return err
	}

	if !sp.cfg.AllowUnknownFields {
		for env, value := range values {
			return fmt.Errorf("unknown field in file %q: %s=%v (see AllowUnknownFields config param)", "file", env, value)
		}
	}
	return nil
}

func (sp *structParser) applyLevelHelper2(fields map[string]interface{}, tag string, values map[string]interface{}) error {
	for _, field := range fields {
		pfield, ok := field.(*parsedField)
		if !ok {
			fmt.Printf("wat in level %T (%+v)\n", field, field)
			continue
		}
		tagValue, ok := pfield.tags[tag]
		if !ok {
			continue
		}
		value, ok := values[tagValue]
		if !ok {
			continue
		}

		switch value := value.(type) {
		case map[string]interface{}:
			if pfield.hasChilds {
				pfieldValue, ok := pfield.value.(map[string]interface{})
				if !ok {
					fmt.Printf("ouch %T (%+v)\n", pfield.value, pfield.value)
					continue
				}
				err := sp.applyLevelHelper2(pfieldValue, tag, value)
				if err != nil {
					return err
				}
			} else {
				pfield.value = value
			}
		default:
			pfield.value = value
		}

		delete(values, tagValue)
	}
	return nil
}

func (sp *structParser) applyLevelHelper(fields map[string]interface{}, tag string, values map[string]interface{}) error {
	for _, v := range fields {
		field, ok := v.(*parsedField)
		if !ok {
			// fmt.Printf("got type %T (%v)\n", v, v)
			continue
		}

		want := field.tags[tag]
		value, ok := values[want]
		if !ok {
			continue
		}
		vval, ok := value.(map[string]interface{})

		// TODO: can be only for leaf nodes?
		if !ok {
			// fmt.Printf("got val %T (%v)\n", val, val)
			field.value = value
			continue
		}

		// fmt.Printf("got map: %+v %T\n", vval, vval)

		// no struct in childs - simple apply, mapstructure will take care
		if field.childs == nil {
			// TODO: reencode values?
			field.value = vval
		} else {
			if err := sp.applyLevelHelper(field.childs, tag, vval); err != nil {
				return err
			}
		}
	}
	return nil
}

func (sp *structParser) applyFlat(tag string, values map[string]interface{}) error {
	allowUnknown := true
	prefix := ""

	switch tag {
	case "env":
		allowUnknown, prefix = sp.cfg.AllowUnknownEnvs, sp.cfg.EnvPrefix
	case "flag":
		allowUnknown, prefix = sp.cfg.AllowUnknownFlags, sp.cfg.FlagPrefix
	}

	dupls := map[string]struct{}{}

	if err := sp.applyFlatHelper(sp.fields, tag, values); err != nil {
		return err
	}

	if allowUnknown || prefix == "" {
		return nil
	}

	for name := range dupls {
		delete(values, name)
	}
	for key, value := range values {
		if strings.HasPrefix(key, prefix) {
			return fmt.Errorf("unknown %s %s=%v (see AllowUnknownXXX config param)", tag, key, value)
		}
	}
	return nil
}

func (sp *structParser) applyFlatHelper(fields map[string]interface{}, tag string, values map[string]interface{}) error {
	for _, field := range fields {
		pfield, ok := field.(*parsedField)
		if !ok {
			fmt.Printf("wat in flat %T (%+v)\n", field, field)
			continue
		}

		tagValue, ok := pfield.tags[tag+"_full"]
		if !ok {
			continue
		}
		value, ok := values[tagValue]
		if !ok {
			if !pfield.hasChilds {
				continue
			}
			if err := sp.applyFlatHelper(pfield.value.(map[string]interface{}), tag, values); err != nil {
				return err
			}
			continue
		}

		pfield.value = value
		if !sp.cfg.AllowDuplicates {
			delete(values, tagValue)
		}
	}
	return nil
}

func isPrimitive(v reflect.Type) bool {
	return v.Kind() < reflect.Array || v.Kind() == reflect.String
}
