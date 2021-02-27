package aconfig

import "reflect"

type tree struct {
	root *node
}

type node struct {
	childs []*node
}

func (t *tree) set(c map[string]interface{}) {
}

func (l *Loader) getTree(x interface{}) *tree {
	value := reflect.ValueOf(x)
	for value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}
	var t tree
	l.getTreeHelper(value, t.root)
	return &t
}

func (l *Loader) getTreeHelper(valueObject reflect.Value, parent *node) {
	typeObject := valueObject.Type()
	count := valueObject.NumField()

	// fields := make([]*fieldData, 0, count)
	for i := 0; i < count; i++ {
		value := valueObject.Field(i)
		field := typeObject.Field(i)

		if !value.CanSet() {
			continue
		}

		// fd := l.newFieldData(field, value, parent)
		var n node

		// if it's a struct - expand and process it's fields
		if field.Type.Kind() == reflect.Struct {
			var subFieldParent *node
			if field.Anonymous {
				subFieldParent = parent
			} else {
				subFieldParent = &node{}
			}
			l.getTreeHelper(value, subFieldParent)
			// fields = append(fields, l.getFieldsHelper(value, subFieldParent)...)
			continue
		}
		if field.Type.Kind() == reflect.Map {
			continue
		}
		if field.Type.Kind() == reflect.Slice {
			continue
		}
		if field.Type.Kind() == reflect.Array {
			continue
		}
		parent.childs = append(parent.childs, &n)
	}
}
