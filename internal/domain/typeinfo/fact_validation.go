package typeinfo

import (
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain"
	"reflect"
)

// Walk scalar facts and references uniformly so new attribute/member locations
// cannot silently bypass scope/evidence validation.
func validateMetadata(v reflect.Value, scope TypeScope, location string) error {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return validateMetadata(v.Elem(), scope, location)
	}
	if v.CanInterface() {
		if f, ok := v.Interface().(interface{ Validate(TypeScope) error }); ok {
			if err := f.Validate(scope); err != nil {
				return fmt.Errorf("%s: %w", location, err)
			}
			return nil
		}
		switch item := v.Interface().(type) {
		case SymbolReference:
			if err := item.Validate(); err != nil {
				return fmt.Errorf("%s: %w", location, err)
			}
			return nil
		case domain.SourceRange:
			if err := item.Validate(); err != nil {
				return fmt.Errorf("%s: %w", location, err)
			}
			return nil
		case OccurrenceLink:
			if item.ID == "" {
				return fmt.Errorf("%s: occurrence requires identity", location)
			}
		case AttributeUse:
			if item.ID == "" {
				return fmt.Errorf("%s: attribute requires identity", location)
			}
		case AttributeValue:
			switch item.Kind {
			case AttributeLiteral:
				if item.Literal.EffectiveStatus() != FactKnown || item.Expression.Value != nil {
					return fmt.Errorf("%s: literal argument requires only a literal payload", location)
				}
			case AttributeExpression:
				if item.Expression.EffectiveStatus() != FactKnown || item.Literal.Value != nil {
					return fmt.Errorf("%s: expression argument requires only an expression payload", location)
				}
			case AttributeRedacted:
				if item.Expression.Value != nil || item.Literal.Value != nil || item.RedactionReason == "" {
					return fmt.Errorf("%s: redacted argument cannot retain a payload and requires a reason", location)
				}
			default:
				return fmt.Errorf("%s: invalid attribute argument kind", location)
			}
		case ArrayShape:
			if item.Rank.Value != nil && (*item.Rank.Value < 1 || len(item.Dimensions) != *item.Rank.Value) {
				return fmt.Errorf("%s: array rank must match dimensions", location)
			}
		}
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := validateMetadata(v.Field(i), scope, location+"."+v.Type().Field(i).Name); err != nil {
				return err
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := validateMetadata(v.Index(i), scope, fmt.Sprintf("%s[%d]", location, i)); err != nil {
				return err
			}
		}
	}
	return nil
}
