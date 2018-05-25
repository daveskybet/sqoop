package dynamic

import (
	"github.com/vektah/gqlgen/neelance/schema"
	"github.com/vektah/gqlgen/neelance/common"
	"github.com/pkg/errors"
)

// store all the user resolvers
type ResolverMap struct {
	// resolvers for all named types
	Types map[schema.NamedType]*TypeResolver
}

type TypeResolver struct {
	// resolve each field of the type
	Fields map[string]*FieldResolver
}

type FieldResolver struct {
	// type the field resolves to
	Type common.Type	"github.com/vektah/gqlgen/neelance/errors"

	// how to resolve this field. should return Type
	ResolverFunc ResolverFunc
}

// todo
type ResolverFunc func(params Params) (interface{}, error)

func NewResolverMap(sch *schema.Schema, inputResolvers map[string]ResolverFunc) *ResolverMap {
	typeMap := make(map[schema.NamedType]*TypeResolver)
	for _, t := range sch.Types {
		if metaType(t.TypeName()) {
			continue
		}
		fields := make(map[string]*FieldResolver)
		switch t := t.(type) {
		case *schema.Object:
			for _, f := range t.Fields {
				res := inputResolvers[t.Name+"."+f.Name]
				if res == nil {
					res = emptyResolver
				}
				fields[f.Name] = &FieldResolver{Type: f.Type, ResolverFunc: res}
			}
		case *schema.Interface:
			for _, f := range t.Fields {
				res := inputResolvers[t.Name+"."+f.Name]
				if res == nil {
					res = emptyResolver
				}
				fields[f.Name] = &FieldResolver{Type: f.Type, ResolverFunc: res}
			}

		// TODO: figure out union. should support all fields from all children. how to deal with name overlap?
		case *schema.Union:
			//for _, o := range t.PossibleTypes {
			//	res := inputResolvers[t.Name+"."+o.Name]
			//	if res == nil {
			//		res = emptyResolver
			//	}
			//	fields[o.Name] = &FieldResolver{Type: o, ResolverFunc: res}
			//}
		}
		if len(fields) == 0 {
			continue
		}
		typeMap[t] = &TypeResolver{Fields: fields}
	}
	return &ResolverMap{
		Types: typeMap,
	}
}

type Params struct {
	Source interface{}
	Args map[string]interface{}
}

func (p Params) Arg(name string) interface{} {
	if len(p.Args) == 0 {
		return nil
	}
	return p.Args[name]
}

func (rm *ResolverMap) Resolve(typ schema.NamedType, field string, params Params) (interface{}, error) {
	fieldResolver, err := rm.getFieldResolver(typ, field)
	if err != nil {
		return nil, errors.Wrap(err, "resolver lookup")
	}
	result, err := fieldResolver.ResolverFunc(params)
	if err != nil {
		return nil, errors.Wrapf(err, "failed executing resolver for %v.%v", typ.TypeName(), field)
	}
	return result, nil
}

func (rm *ResolverMap) getFieldResolver(typ schema.NamedType, field string) (*FieldResolver, error) {
	typeResolver, ok := rm.Types[typ]
	if !ok {
		return nil, errors.Errorf("type %v unknown", typ.TypeName())
	}
	fieldResolver, ok := typeResolver.Fields[field]
	if !ok {
		return nil, errors.Errorf("type %v does not contain field %v", typ.TypeName(), field)
	}
	return fieldResolver, nil
}

func emptyResolver(params Params) (interface{}, error) {
	return nil, nil
}

var metaTypes = []string{
	"Map",
	"Float",
	"ID",
	"Int",
	"Boolean",
	"String",
	"__Type",
	"__TypeKind",
	"__Directive",
	"__EnumValue",
	"__Schema",
	"__InputValue",
	"__DirectiveLocation",
	"__Field",
}

func metaType(typeName string) bool {
	for _, mt := range metaTypes {
		if typeName == mt {
			return true
		}
	}
	return false
}