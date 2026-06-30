package goap

import "math/bits"

type Symbol uint64

type SymbolSet uint64

func (s SymbolSet) Has(sym Symbol) bool { return s&SymbolSet(sym) != 0 }

func (s SymbolSet) With(sym Symbol) SymbolSet { return s | SymbolSet(sym) }

func (s SymbolSet) Without(sym Symbol) SymbolSet { return s &^ SymbolSet(sym) }

func (s SymbolSet) Union(o SymbolSet) SymbolSet { return s | o }

func (s SymbolSet) Contains(o SymbolSet) bool { return s&o == o }

func (s SymbolSet) Apply(set, clear SymbolSet) SymbolSet { return (s &^ clear) | set }

func (s SymbolSet) Missing(o SymbolSet) SymbolSet { return o &^ s }

func (s SymbolSet) Count() int { return bits.OnesCount64(uint64(s)) }
