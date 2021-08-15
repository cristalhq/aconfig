package aconfig

import (
	"reflect"
	"strconv"
	"time"
)

func (l *Loader) createFlags() error {
	for _, field := range l.fields {
		flagName := l.fullTag(l.config.FlagPrefix, field, flagNameTag)
		if flagName == "" {
			continue
		}

		// unwrap pointers
		fd := field.value
		for fd.Type().Kind() == reflect.Ptr {
			if fd.IsNil() {
				fd.Set(reflect.New(fd.Type().Elem()))
			}
			fd = fd.Elem()
		}

		value := field.Tag(defaultValueTag)
		usage := field.Tag(usageTag)

		switch fd.Type().Kind() {
		case reflect.Bool:
			if value == "" {
				value = "false"
			}
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			l.flagSet.Bool(flagName, b, usage)

		case reflect.String:
			l.flagSet.String(flagName, value, usage)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			if value == "" {
				value = "0"
			}

			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			println(v)
			l.flagSet.Int(flagName, int(v), usage)

		case reflect.Int64:
			if field.field.Type == reflect.TypeOf(time.Second) {
				if value == "" {
					value = "0s"
				}
				d, err := time.ParseDuration(value)
				if err != nil {
					return err
				}
				l.flagSet.Duration(flagName, d, usage)
			} else {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return err
				}
				l.flagSet.Int64(flagName, v, usage)
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if value == "" {
				value = "0"
			}
			v, err := strconv.ParseUint(value, 10, field.field.Type.Bits())
			if err != nil {
				return err
			}
			l.flagSet.Uint64(flagName, v, usage)

		case reflect.Float32, reflect.Float64:
			if value == "" {
				value = "0"
			}
			v, err := strconv.ParseFloat(value, field.field.Type.Bits())
			if err != nil {
				return err
			}
			l.flagSet.Float64(flagName, v, usage)

		case reflect.Slice, reflect.Array:
			if field.field.Type.Elem().Kind() == reflect.Uint8 {
				l.flagSet.String(flagName, value, usage)
			}

		case reflect.Map:
			// process fields
		case reflect.Struct:
			// process fields
		}
	}
	return nil
}
