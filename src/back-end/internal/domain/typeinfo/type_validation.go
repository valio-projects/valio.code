package typeinfo

import (
	"fmt"
	"reflect"
)

// Validate verifies descriptor identity, scope, member uniqueness, and evidence consistency.
func (t TypeDescriptor) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("type descriptor requires identity")
	}
	if err := t.Scope.Validate(); err != nil {
		return err
	}
	if t.Name.EffectiveStatus() != FactKnown || t.Name.Value == nil || *t.Name.Value == "" {
		return fmt.Errorf("type descriptor requires a known nonempty name")
	}
	if err := validateMetadata(reflect.ValueOf(t), t.Scope, "type"); err != nil {
		return err
	}
	ids := map[string]bool{}
	add := func(id string) error {
		if id == "" {
			return fmt.Errorf("member requires identity")
		}
		if ids[id] {
			return fmt.Errorf("duplicate member identity %q", id)
		}
		ids[id] = true
		return nil
	}
	for _, f := range t.Fields {
		if err := add(f.ID); err != nil {
			return err
		}
	}
	var method func(MethodDescriptor) error
	method = func(m MethodDescriptor) error {
		if err := add(m.ID); err != nil {
			return err
		}
		if m.Receiver != nil {
			if err := add(m.Receiver.ID); err != nil {
				return err
			}
		}
		for _, p := range m.Parameters {
			if err := add(p.ID); err != nil {
				return err
			}
		}
		for _, r := range m.Returns {
			if err := add(r.ID); err != nil {
				return err
			}
		}
		for _, g := range m.GenericParameters {
			if err := add(g.ID); err != nil {
				return err
			}
		}
		return nil
	}
	for _, p := range t.Properties {
		if err := add(p.ID); err != nil {
			return err
		}
		for _, parameter := range p.Parameters {
			if err := add(parameter.ID); err != nil {
				return err
			}
		}
		if p.Getter != nil {
			if err := method(*p.Getter); err != nil {
				return err
			}
		}
		if p.Setter != nil {
			if err := method(*p.Setter); err != nil {
				return err
			}
		}
	}
	for _, methods := range [][]MethodDescriptor{t.Methods, t.Constructors} {
		for _, m := range methods {
			if err := method(m); err != nil {
				return err
			}
		}
	}
	for _, g := range t.GenericParameters {
		if err := add(g.ID); err != nil {
			return err
		}
	}
	memberGroups := [][]EnumMember{t.Constants}
	if t.Enum != nil {
		memberGroups = append(memberGroups, t.Enum.Members)
	}
	occurrences := map[string]bool{}
	for _, members := range memberGroups {
		for _, m := range members {
			if err := add(m.ID); err != nil {
				return err
			}
			for _, o := range m.Occurrences {
				if occurrences[o.ID] {
					return fmt.Errorf("duplicate constant occurrence identity %q", o.ID)
				}
				occurrences[o.ID] = true
			}
		}
	}
	fieldIDs := map[string]bool{}
	for _, f := range t.Fields {
		fieldIDs[f.ID] = true
	}
	physicalKnown := t.Layout.SizeBytes.EffectiveStatus() == FactKnown || t.Layout.AlignmentBytes.EffectiveStatus() == FactKnown || t.Layout.PackingBytes.EffectiveStatus() == FactKnown
	for _, f := range t.Layout.Fields {
		if !fieldIDs[f.FieldID] {
			return fmt.Errorf("layout refers to unknown field %q", f.FieldID)
		}
		physicalKnown = physicalKnown || f.OffsetBytes.EffectiveStatus() == FactKnown || f.BitOffset.EffectiveStatus() == FactKnown || f.BitWidth.EffectiveStatus() == FactKnown
	}
	if physicalKnown {
		for _, f := range []Fact[string]{t.Layout.TargetArchitecture, t.Layout.ABI, t.Layout.Compiler, t.Layout.CompilerVersion} {
			if f.EffectiveStatus() != FactKnown || f.Value == nil || *f.Value == "" {
				return fmt.Errorf("physical layout facts require known architecture, ABI, compiler and compiler version")
			}
		}
	}
	return nil
}
