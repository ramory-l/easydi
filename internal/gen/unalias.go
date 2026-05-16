package gen

import "go/types"

// unaliasDeep resolves type aliases to their underlying named/structural
// target, recursively, including aliases nested inside pointers, slices,
// arrays, maps and channels. This makes the emitter render an alias root such
// as `type Config = *cfg.Config` (declared in the generated package) as its
// right-hand side `*cfg.Config`, so the generated file never qualifies — and
// thus never imports — its own output package.
//
// Limitation: alias type arguments embedded inside a generic instantiation
// (a *types.Named with TypeArgs, e.g. a root `*Wrapper[di.Config]`) are not
// descended into and remain aliased. Generic roots are out of scope for
// v0.3.0; revisit if/when generics support lands.
func unaliasDeep(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	t = types.Unalias(t)
	switch x := t.(type) {
	case *types.Pointer:
		e := unaliasDeep(x.Elem())
		if e != x.Elem() {
			return types.NewPointer(e)
		}
		return x
	case *types.Slice:
		e := unaliasDeep(x.Elem())
		if e != x.Elem() {
			return types.NewSlice(e)
		}
		return x
	case *types.Array:
		e := unaliasDeep(x.Elem())
		if e != x.Elem() {
			return types.NewArray(e, x.Len())
		}
		return x
	case *types.Map:
		k := unaliasDeep(x.Key())
		v := unaliasDeep(x.Elem())
		if k != x.Key() || v != x.Elem() {
			return types.NewMap(k, v)
		}
		return x
	case *types.Chan:
		e := unaliasDeep(x.Elem())
		if e != x.Elem() {
			return types.NewChan(x.Dir(), e)
		}
		return x
	default:
		return t
	}
}
