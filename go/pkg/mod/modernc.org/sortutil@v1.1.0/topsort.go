// Copyright 2019 The sortutil Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sortutil // import "modernc.org/sortutil"

// TopologicalSortNode represents a node of a graph for TopologicalSort.
// Implementations of TopologicalSortNode must be comparable.
type TopologicalSortNode interface {
	// Edges return the list of nodes this node points to.
	Edges() []TopologicalSortNode
}

// TopologicalSort returns a reversed topological ordering of a directed
// acyclic graph or nil if graph is not a DAG.
//
// It implements the Depth-first search algorithm:
//
//	https://en.wikipedia.org/wiki/Topological_sorting#Depth-first_search
func TopologicalSort(graph []TopologicalSortNode) []TopologicalSortNode {
	l := make([]TopologicalSortNode, 0, len(graph))
	noPermanentMark := make(map[TopologicalSortNode]struct{}, len(graph))
	for _, n := range graph {
		noPermanentMark[n] = struct{}{}
	}

	temporaryMark := make(map[TopologicalSortNode]struct{}, len(graph))
	var visit func(TopologicalSortNode) bool
	visit = func(n TopologicalSortNode) bool {
		if _, ok := noPermanentMark[n]; !ok {
			return true
		}

		if _, ok := temporaryMark[n]; ok {
			return false
		}

		temporaryMark[n] = struct{}{}
		for _, m := range n.Edges() {
			visit(m)
		}

		delete(temporaryMark, n)
		delete(noPermanentMark, n)
		l = append(l, n)
		return true
	}

	for len(noPermanentMark) != 0 {
		for n := range noPermanentMark {
			if !visit(n) {
				return nil // Not a DAG
			}
		}
	}
	return l
}
