package aconfig

import (
	"fmt"
	"reflect"
	"strings"
)

type configParser struct {
}

func (cf *configParser) Parse(cfg interface{}) (*field, error) {
	root := &field{
		child:    make([]*field, 0),
		v:        reflect.ValueOf(cfg).Elem(),
		t:        reflect.ValueOf(cfg).Elem().Type(),
		sliceIdx: -1,
	}
	if err := cf.parseHelper(root); err != nil {
		return nil, err
	}
	return root, nil
}

func (cf *configParser) parseHelper(f *field) error {
	for f.v.Kind() == reflect.Ptr && !f.v.IsNil() {
		f.v = f.v.Elem()
		f.t = f.v.Type()
	}

	switch f.v.Kind() {
	case reflect.Struct:
		for i := 0; i < f.t.NumField(); i++ {
			unexported := f.t.Field(i).PkgPath != ""
			embedded := f.t.Field(i).Anonymous
			if (unexported && !embedded) || !f.v.CanSet() {
				continue
			}
			child := newStructField(f, i)
			f.child = append(f.child, child)
			cf.parseHelper(child)
		}

	case reflect.Slice, reflect.Array:
		switch f.t.Elem().Kind() {
		case reflect.Struct, reflect.Slice, reflect.Array, reflect.Ptr, reflect.Interface:
			for i := 0; i < f.v.Len(); i++ {
				child := newSliceField(f, i)
				cf.parseHelper(child)
			}
		}
		// case reflect.Map:
		// 	panic("unimplemented")
	}
	return nil
}

// field is a settable field of a config object.
type field struct {
	child    []*field
	fullname string
	parent   *field

	v        reflect.Value
	t        reflect.Type
	st       reflect.StructField
	sliceIdx int // >=0 if this field is a member of a slice.

	defaultValue string
	isSet        bool
	isRequired   bool
	tags         map[string]string
}

func newStructField(parent *field, idx int) *field {
	val := parent.v.Field(idx)
	stf := parent.t.Field(idx)

	f := &field{
		child:        make([]*field, 0),
		parent:       parent,
		v:            val,
		t:            val.Type(),
		st:           stf,
		defaultValue: stf.Tag.Get(defaultValueTag),
		sliceIdx:     -1,
		isSet:        false,
	}
	parseTags(f)
	return f
}

// func newMapField(parent *field) *field {
// 	val := parent.v.Field(idx)
// 	stf := parent.t.Key()
// 	f := &field{
// 		child:        make([]*field, 0),
// 		parent:       parent,
// 		v:            val,
// 		t:            val.Type(),
// 		st:           stf,
// 		defaultValue: stf.Tag.Get(defaultValueTag),
// 		sliceIdx:     -1,
// 		isSet:        false,
// 	}
// 	parseTags(f)
// 	return f
// }

func newSliceField(parent *field, idx int) *field {
	val := parent.v.Index(idx)
	stf := parent.st
	f := &field{
		child:        make([]*field, 0),
		parent:       parent,
		v:            val,
		t:            val.Type(),
		st:           stf,
		defaultValue: stf.Tag.Get(defaultValueTag),
		sliceIdx:     idx,
		isSet:        false,
	}
	parseTags(f)
	return f
}

// // parseTag parses a fields struct tags into a more easy to use structTag.
// // key is the key of the struct tag which contains the field's alt name.
// func parseTag(tag reflect.StructTag, key string) structTag {
// 	var st structTag
// 	if val, ok := tag.Lookup(key); ok {
// 		i := strings.Index(val, ",")
// 		if i == -1 {
// 			i = len(val)
// 		}
// 		st.altName = val[:i]
// 	}

// 	if val := tag.Get("required"); val == "true" {
// 		st.required = true
// 	}

// 	if val, ok := tag.Lookup("default"); ok {
// 		st.setDefault = true
// 		st.defaultVal = val
// 	}
// 	return st
// }

func parseTags(f *field) {
	tags := f.st.Tag
	if val := tags.Get("required"); val == "true" {
		f.isRequired = true
	}
	f.tags = map[string]string{
		usageTag: tags.Get(usageTag),
		// envNameTag:  l.makeTagValue(field, envNameTag, words),
		// flagNameTag: l.makeTagValue(field, flagNameTag, words),
	}
}

// name is the name of the field. if the field contains an alt name
// in the struct struct that name is used, else  it falls back to
// the field's name as defined in the struct.
// if this field is a slice field, then its name is simply its
// index in the slice.
func (f *field) name() string {
	if f.sliceIdx >= 0 {
		return fmt.Sprintf("[%d]", f.sliceIdx)
	}
	// if f.altName != "" {
	// 	return f.altName
	// }
	return f.st.Name
}

// path is a dot separated path consisting of all the names of
// the field's ancestors starting from the topmost parent all the
// way down to the field itself.
func (f *field) path() (path string) {
	var visit func(f *field)
	visit = func(f *field) {
		if f.parent != nil {
			visit(f.parent)
		}
		path += f.name()
		// if it's a slice/array we don't want a dot before the slice indexer
		// e.g. we want A[0].B instead of A.[0].B
		if f.t.Kind() != reflect.Slice && f.t.Kind() != reflect.Array {
			path += "."
		}
	}
	visit(f)
	return strings.Trim(path, ".")
}
