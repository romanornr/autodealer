package bellmanford

import (
	"math"
)

// Graph represents a graph consisting of edges and vertices
type Graph struct {
	edges    []*Edge
	vertices []uint
}

// Edge represents a weighted line between two nodes
type Edge struct {
	From, To uint
	Weight   float64
}

// NewEdge returns a pointer to a new Edge
func NewEdge(from, to uint, weight float64) *Edge {
	return &Edge{From: from, To: to, Weight: weight}
}

// NewGraph returns a graph consisting of given edges and vertices (vertices must count from 0 upwards)
func NewGraph(edges []*Edge, vertices []uint) *Graph {
	return &Graph{edges: edges, vertices: vertices}
}

// FindArbitrageLoop returns either an arbitrage loop or a nil map
func (g *Graph) FindArbitrageLoop(source uint) []uint {
	predecessors, distances := g.BellmanFord(source)
	return g.FindNegativeWeightCycle(predecessors, distances, source)
}

// BellmanFord determines the shortest path and returns the predecessors and distances
func (g *Graph) BellmanFord(source uint) ([]uint, []float64) {
	size := len(g.vertices)
	distances := make([]float64, size)
	predecessors := make([]uint, size)
	for _, v := range g.vertices {
		distances[v] = math.MaxFloat64
	}
	distances[source] = 0

	for i, changes := 0, 0; i < size-1; i, changes = i+1, 0 {
		for _, edge := range g.edges {
			if newDist := distances[edge.From] + edge.Weight; newDist < distances[edge.To] {
				distances[edge.To] = newDist
				predecessors[edge.To] = edge.From
				changes++
			}
		}
		if changes == 0 {
			break
		}
	}
	return predecessors, distances
}

// FindNegativeWeightCycle finds a negative weight cycle from predecessors and a source
func (g *Graph) FindNegativeWeightCycle(predecessors []uint, distances []float64, source uint) []uint {
	for _, edge := range g.edges {
		if distances[edge.From]+edge.Weight < distances[edge.To] {
			return arbitrageLoop(predecessors, source)
		}
	}
	return nil
}

func arbitrageLoop(predecessors []uint, source uint) []uint {
	size := len(predecessors)
	loop := make([]uint, size)
	loop[0] = source

	exists := make([]bool, size)
	exists[source] = true

	indices := make([]uint, size)

	var index, next uint
	for index, next = 1, source; ; index++ {
		next = predecessors[next]
		loop[index] = next
		if exists[next] {
			return loop[indices[next] : index+1]
		}
		indices[next] = index
		exists[next] = true
	}
}
