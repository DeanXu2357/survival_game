package state

type LayerMask uint16

const (
	LayerNone   LayerMask = 0
	LayerPlayer LayerMask = 1 << 0
	LayerStatic LayerMask = 1 << 1
	LayerLight  LayerMask = 1 << 2
)

/*
 * Use m LayerMask instead of m *LayerMask for method receivers to avoid pointer indirection.
 * LayerMask is uint16, which is 2bytes, *LayerMask would be 8bytes on 64bit systems.
 * Pointer of LayerMast need  to be dereferenced to get the value, which adds overhead.
 * LayerMask located in stack when used as value receiver, which is faster to access than heap memory.
 * And LayerMask is immutable, it can be safely copied without worrying about unintended side effects.
 */

func (m LayerMask) Has(layer LayerMask) bool {
	return (m & layer) != 0
}

func (m LayerMask) Add(layer LayerMask) LayerMask {
	return m | layer
}

func (m LayerMask) Remove(layer LayerMask) LayerMask {
	return m &^ layer
}

func (m LayerMask) Toggle(layer LayerMask) LayerMask {
	return m ^ layer
}

func (m LayerMask) Equal(other LayerMask) bool {
	return m == other
}
