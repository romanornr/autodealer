package bellmanford

import (
	"github.com/sirupsen/logrus"
	"math"
	"testing"
)

func _newGraph() *Graph {
	return &Graph{
		vertices: []uint{0, 1, 2, 3, 4, 5},
		edges: []*Edge{
			&Edge{To: 1, From: 0, Weight: -math.Log(1.380)},
			&Edge{To: 2, From: 1, Weight: -math.Log(3.08)},
			&Edge{To: 3, From: 2, Weight: -math.Log(15.120)},
			&Edge{To: 4, From: 3, Weight: -math.Log(0.012)},
			&Edge{To: 0, From: 4, Weight: -math.Log(1.30)},
			&Edge{To: 5, From: 4, Weight: -math.Log(0.57)}},
	}
}

func BenchmarkNewGraph(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = _newGraph()
	}
}

func BenchmarkBellmanFord(b *testing.B) {
	g := _newGraph()
	var source uint = 1
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g.BellmanFord(source)
	}
}

func BenchmarkFindNegativeWeightCycle(b *testing.B) {
	g := _newGraph()
	var source uint = 1
	predecessors, distances := g.BellmanFord(source)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g.FindNegativeWeightCycle(predecessors, distances, source)
	}
}

func BenchmarkArbitrageLoop(b *testing.B) {
	g := _newGraph()
	var source uint = 1
	predecessors, _ := g.BellmanFord(source)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		arbitrageLoop(predecessors, source)
	}
}

func BenchmarkFindArbitrageLoop(b *testing.B) {
	g := _newGraph()
	var source uint = 1
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g.FindArbitrageLoop(source)
	}
}

func TestFullSequence(t *testing.T) {
	//results := map[uint][]uint{
	//	0: []uint{0, 4, 3, 2, 1, 0},
	//	1: []uint{1, 0, 4, 3, 2, 1},
	//	2: []uint{2, 1, 0, 4, 3, 2},
	//	3: []uint{3, 2, 1, 0, 4, 3},
	//	4: []uint{4, 3, 2, 1, 0, 4},
	//}

	// VIA: 0
	// BTC: 1
	// EUR: 3
	// USD: 4
	// EUR: 5
	results := map[uint][]uint{
		0: []uint{0, 4, 3, 2, 1, 0},
		1: []uint{1, 0, 4, 3, 2, 1},
		2: []uint{2, 1, 0, 4, 3, 2},
		3: []uint{3, 2, 1, 0, 4, 3},
		4: []uint{4, 3, 2, 1, 0, 4},
	}

	for source, res := range results {
		g := _newGraph()
		loop := g.FindArbitrageLoop(source)
		if len(loop) != len(res) {
			t.Fatalf("loops have different lengths (%d != %d)", loop, res)
		}
		logrus.Printf("%v", loop)
		for i, v := range loop {
			if res[i] != v {
				t.Fatalf("incorrect arbitrage loop (%v != %v; source is %d)\n", loop, res, source)
			}
		}
	}
}

func _getPythonGraph() *Graph {
	return &Graph{edges: []*Edge{
		&Edge{From: 0, To: 1, Weight: -4.582438665548869},
		&Edge{From: 0, To: 2, Weight: 0.2981813979749493},
		&Edge{From: 0, To: 3, Weight: 4.838300943835368},
		&Edge{From: 1, To: 0, Weight: 4.585249918552961},
		&Edge{From: 1, To: 2, Weight: 4.836396313495658},
		&Edge{From: 1, To: 3, Weight: 9.375215015166416},
		&Edge{From: 2, To: 0, Weight: -0.3751523503802663},
		&Edge{From: 2, To: 1, Weight: -5.004605689846387},
		&Edge{From: 2, To: 3, Weight: 4.362953685292599},
		&Edge{From: 3, To: 0, Weight: -4.6488526240960395},
		&Edge{From: 3, To: 1, Weight: -9.277409346383422},
		&Edge{From: 3, To: 2, Weight: -4.344533438603351},
	}, vertices: []uint{0, 1, 2, 3}}
}

func TestWithPythonDataSource(t *testing.T) {
	results := map[uint][]uint{
		0: []uint{2, 1, 2},
		1: []uint{1, 2, 1},
		2: []uint{2, 1, 2},
		3: []uint{2, 1, 2},
	}
	for source, res := range results {
		g := _getPythonGraph()
		loop := g.FindArbitrageLoop(source)
		if len(loop) != len(res) {
			t.Fatalf("loops have different lengths (%d != %d)", loop, res)
		}
		for i, v := range loop {
			if res[i] != v {
				t.Fatalf("incorrect arbitrage loop (%v != %v; source is %d)\n", loop, res, source)
			}
		}
	}
}
