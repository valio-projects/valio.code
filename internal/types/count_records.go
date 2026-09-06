package types

import (
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
	"reflect"
)

func Counts(d typeinfo.TypeDescriptor) MemberCounts {
	c := MemberCounts{Fields: len(d.Fields), Properties: len(d.Properties), Methods: len(d.Methods), Constructors: len(d.Constructors), GenericParameters: len(d.GenericParameters), Attributes: len(d.Attributes), Constants: len(d.Constants)}
	for _, g := range d.GenericParameters {
		c.Attributes += len(g.Attributes)
	}
	parameter := func(p typeinfo.ParameterDescriptor) { c.Parameters++; c.Attributes += len(p.Attributes) }
	method := func(m typeinfo.MethodDescriptor) {
		c.Attributes += len(m.Attributes)
		c.GenericParameters += len(m.GenericParameters)
		for _, g := range m.GenericParameters {
			c.Attributes += len(g.Attributes)
		}
		for _, p := range m.Parameters {
			parameter(p)
		}
		c.Returns += len(m.Returns)
		for _, r := range m.Returns {
			c.Attributes += len(r.Attributes)
		}
	}
	for _, f := range d.Fields {
		c.Attributes += len(f.Attributes)
	}
	for _, p := range d.Properties {
		c.Attributes += len(p.Attributes)
		for _, parameterValue := range p.Parameters {
			parameter(parameterValue)
		}
		if p.Getter != nil {
			method(*p.Getter)
		}
		if p.Setter != nil {
			method(*p.Setter)
		}
	}
	for _, methods := range [][]typeinfo.MethodDescriptor{d.Methods, d.Constructors} {
		for _, m := range methods {
			method(m)
		}
	}
	for _, m := range d.Constants {
		c.Attributes += len(m.Attributes)
		c.Occurrences += len(m.Occurrences)
	}
	if d.Enum != nil {
		c.EnumMembers = len(d.Enum.Members)
		for _, m := range d.Enum.Members {
			c.Attributes += len(m.Attributes)
			c.Occurrences += len(m.Occurrences)
		}
	}
	c.Attributes = countAttributeRecords(reflect.ValueOf(d))
	return c
}
func countAttributeRecords(v reflect.Value) int {
	if !v.IsValid() {
		return 0
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return 0
		}
		return countAttributeRecords(v.Elem())
	}
	count := 0
	if v.CanInterface() {
		if _, ok := v.Interface().(typeinfo.AttributeUse); ok {
			count++
		}
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			count += countAttributeRecords(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			count += countAttributeRecords(v.Index(i))
		}
	}
	return count
}
