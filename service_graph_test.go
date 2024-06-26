package roids

import (
	"reflect"
	"testing"

	"github.com/ShounakA/roids/core"
)

var mockDag2 *core.AcyclicGraph
var graph *serviceGraph

func setupTest(tb testing.TB) func(tb testing.TB) {
	mockDag2 = core.NewGraph()

	graph = newServiceGraph(mockDag2)

	return func(tb testing.TB) {
		graph = nil
		mockDag2 = nil
	}
}
func TestInstantiatingServiceGraph(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)

	if graph.dag != mockDag2 {
		t.Errorf("Did no instantiate service graph correctly.")
	}
}

func TestAddVertex(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)
	expectedOrder := 1
	specValue := 5

	testSpec := reflect.TypeOf(specValue)
	myService := &Service{SpecType: testSpec}

	err := graph.addVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	actualOrder := graph.dag.GetOrder()
	if expectedOrder != actualOrder {
		t.Errorf("Did not add the vertex correctly.")
	}

	if myService.Id == "" {
		t.Errorf("Did no mutate the services ID field.")
	}
}

func TestAddVertex_InvalidParameters(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)
	expectedOrder := 0

	err := graph.addVertex(nil)
	if err == nil {
		t.Errorf("Should not allow nil")
	}

	if graph.dag.GetOrder() != expectedOrder {
		t.Errorf("Should not add nil to service graph")
	}
}

func TestAddVertex_DuplicateService(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)
	expectedOrder := 1
	specValue := 5

	testSpec := reflect.TypeOf(specValue)
	myService := &Service{SpecType: testSpec}

	err := graph.addVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.addVertex(myService)
	if err == nil {
		t.Errorf("Should no add duplicate vertex")
	}
	err = graph.addVertex(&Service{SpecType: testSpec})
	if err == nil {
		t.Errorf("Should no add duplicate vertex")
	}
	actualOrder := graph.dag.GetOrder()
	if expectedOrder != actualOrder {
		t.Errorf("Did not add the vertex correctly.")
	}
}

func TestAddEdge(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)
	expectedSize := 1
	specValue := 5

	testSpec := reflect.TypeOf(specValue)
	myService := &Service{SpecType: testSpec}

	testSpec2 := reflect.TypeOf(int64(5))
	myService2 := &Service{SpecType: testSpec2}

	err := graph.addVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.addVertex(myService2)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	err = graph.addEdge(myService, myService2)
	if err != nil {
		t.Errorf("Should add edge! %s", err.Error())
	}

	if graph.dag.GetSize() != expectedSize {
		t.Errorf("Should have added edge to dag")
	}
}

func TestAddEdge_InvalidParameters(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)
	testSpec := reflect.TypeOf(4)
	myService := &Service{SpecType: testSpec}

	err := graph.addEdge(nil, myService)
	if err == nil {
		t.Errorf("Should not allow nil services!")
	}

	err = graph.addEdge(myService, nil)
	if err == nil {
		t.Errorf("Should not allow nil services!")
	}

}

func TestAddEdge_CircularDeps(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)
	expectedSize := 1
	specValue := 5

	testSpec := reflect.TypeOf(specValue)
	myService := &Service{SpecType: testSpec}

	testSpec2 := reflect.TypeOf(int64(5))
	myService2 := &Service{SpecType: testSpec2}

	err := graph.addVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.addVertex(myService2)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	err = graph.addEdge(myService, myService2)
	if err != nil {
		t.Errorf("Should add edge! %s", err.Error())
	}

	err = graph.addEdge(myService2, myService)
	if err == nil {
		t.Errorf("Should not create circular dependency")
	}
	if nerr, ok := err.(*core.CircularDependencyError); !ok {
		t.Error("Unexpected error returned.", nerr.Error())
	}

	if graph.dag.GetSize() != expectedSize {
		t.Errorf("Should have added edge to dag")
	}
}

func TestAddEdge_DuplicateEdge(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)
	expectedSize := 1
	specValue := 5

	testSpec := reflect.TypeOf(specValue)
	myService := &Service{SpecType: testSpec}

	testSpec2 := reflect.TypeOf(int64(5))
	myService2 := &Service{SpecType: testSpec2}

	err := graph.addVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.addVertex(myService2)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	err = graph.addEdge(myService, myService2)
	if err != nil {
		t.Errorf("Should add edge! %s", err.Error())
	}

	err = graph.addEdge(myService, myService2)
	if err == nil {
		t.Errorf("Should not create duplicate edge")
	}
	if nerr, ok := err.(*core.DuplicateEdgeError); !ok {
		t.Error("Unexpected error returned.", nerr.Error())
	}

	if graph.dag.GetSize() != expectedSize {
		t.Errorf("Should have added edge to dag")
	}
}

func TestGetVertex(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)

	specValue := 5

	testSpec := reflect.TypeOf(specValue)
	myService := &Service{SpecType: testSpec}

	err := graph.addVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	actualService, err := graph.getVertex(myService.Id)
	if err != nil {
		t.Errorf("Should be able to get vertex. %s", err.Error())
	}
	if actualService != myService {
		t.Errorf("Created vertex should be the same as the one created. %s", err.Error())
	}
}
