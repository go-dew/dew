package dew

import (
	"sort"
	"strings"
)

type node struct {
	// prefix is the common prefix we ignore.
	prefix string

	// handler on the leaf node.
	handler handlerEntry

	// child nodes should be stored in-order for iteration,
	// in groups of the node type.
	children nodes

	// first byte of the prefix.
	label byte
}

type handlerEntry struct {
	handler *handler
	op      OpType
}

func (h *handlerEntry) insert(op OpType, hh *handler) {
	if h.handler != nil {
		panic("dew: multiple handlers registered for the same command")
	}
	h.handler = hh
	h.op = op
}

func (h *handlerEntry) isEmpty() bool {
	return h.handler == nil
}

type handler struct {
	// handler is the function to call.
	handler any
	// mux is the mux that the handler belongs to.
	mux *Mux
}

func (n *node) insert(op OpType, key string, handler *handler) *node {
	var parent *node
	search := key

	for {
		// Handle key exhaustion
		if len(search) == 0 {
			n.setHandler(op, handler)
			return n
		}

		// We're going to be searching for a wild node next,
		// in this case, we need to get the tail
		var label = search[0]

		// Look for the edge to attach to
		parent = n
		n = n.getEdge(label)

		// No edge, create one
		if n == nil {
			// handler leaf node
			hn := parent.addChild(&node{label: label, prefix: search})
			hn.setHandler(op, handler)
			return hn
		}

		// Determine the longest prefix of the search key on match.
		commonPrefixLen := longestPrefixLen(search, n.prefix)
		if commonPrefixLen == len(n.prefix) {
			// the common prefix is as long as the current node's prefix we're attempting to insert.
			// keep the search going.
			search = search[commonPrefixLen:]
			continue
		}

		// Split the node
		child := &node{
			prefix: search[:commonPrefixLen],
		}
		parent.replaceChild(search[0], child)

		// Restore the existing node
		n.label = n.prefix[commonPrefixLen]
		n.prefix = n.prefix[commonPrefixLen:]
		child.addChild(n)

		// If the new key is a subset, set the method/handler on this node and finish.
		search = search[commonPrefixLen:]

		// Create a new edge for the node
		subchild := &node{
			label:  search[0],
			prefix: search,
		}
		hn := child.addChild(subchild)
		hn.setHandler(op, handler)
		return hn
	}
}

func (n *node) setHandler(op OpType, h *handler) {
	n.handler.insert(op, h)
}

// Recursive edge traversal by checking all nodeTyp groups along the way.
// It's like searching through a multi-dimensional radix trie.
func (n *node) findRoute(op OpType, key string) *node {
	nn := n
	search := key

	var xn *node
	xsearch := search

	var label byte
	if search != "" {
		label = search[0]
	}

	xn = nn.children.findEdge(label)
	if xn == nil || !strings.HasPrefix(xsearch, xn.prefix) {
		return nil
	}
	xsearch = xsearch[len(xn.prefix):]

	if len(xsearch) == 0 && xn.isLeaf() {
		if xn.handler.handler != nil && xn.handler.op == op {
			return xn
		}
	}

	// recursively find the next node.
	return xn.findRoute(op, xsearch)
}

func (n *node) isLeaf() bool {
	return !n.handler.isEmpty()
}

func (n *node) getEdge(label byte) *node {
	nds := n.children
	for i := 0; i < len(nds); i++ {
		if nds[i].label == label {
			return nds[i]
		}
	}
	return nil
}

// addChild append the new `child` node to the tree.
func (n *node) addChild(child *node) *node {
	n.children = append(n.children, child)
	n.children.Sort()
	return child
}

func (n *node) replaceChild(label byte, child *node) {
	for i := 0; i < len(n.children); i++ {
		if n.children[i].label == label {
			n.children[i] = child
			n.children[i].label = label
			return
		}
	}
	panic("dew: replacing missing child")
}

type nodes []*node

// Sort the list of nodes by label
func (ns nodes) Sort()              { sort.Sort(ns) }
func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Less(i, j int) bool { return ns[i].label < ns[j].label }

func (ns nodes) findEdge(label byte) *node {
	if len(ns) == 0 {
		return nil
	}
	num := len(ns)
	idx := 0
	i, j := 0, num-1
	for i <= j {
		idx = i + (j-i)/2
		if label > ns[idx].label {
			i = idx + 1
		} else if label < ns[idx].label {
			j = idx - 1
		} else {
			i = num // breaks cond
		}
	}
	if ns[idx].label != label {
		return nil
	}
	return ns[idx]
}

// longestPrefixLen finds the length of the shared prefix
// of two strings
func longestPrefixLen(k1, k2 string) int {
	maxL := len(k1)
	if l := len(k2); l < maxL {
		maxL = l
	}
	var i int
	for i = 0; i < maxL; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	return i
}
