package roids

import (
	"log"
	"reflect"
	"testing"

	"github.com/heimdalr/dag"
)

var mockDag *dag.DAG
var graph *serviceGraph

func setupTest(tb testing.TB) func(tb testing.TB) {
	log.Println("setup test")
	mockDag = dag.NewDAG()
	graph = newServiceGraph(mockDag)

	return func(tb testing.TB) {
		graph = nil
		mockDag = nil
		log.Println("teardown test")
	}
}
func TestInstantiatingServiceGraph(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)

	if graph.dag != mockDag {
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

	err := graph.AddVertex(myService)
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

	err := graph.AddVertex(nil)
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

	err := graph.AddVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.AddVertex(myService)
	if err == nil {
		t.Errorf("Should no add duplicate vertex")
	}
	err = graph.AddVertex(&Service{SpecType: testSpec})
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

	err := graph.AddVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.AddVertex(myService2)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	err = graph.AddEdge(myService, myService2)
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

	err := graph.AddEdge(nil, myService)
	if err == nil {
		t.Errorf("Should not allow nil services!")
	}

	err = graph.AddEdge(myService, nil)
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

	err := graph.AddVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.AddVertex(myService2)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	err = graph.AddEdge(myService, myService2)
	if err != nil {
		t.Errorf("Should add edge! %s", err.Error())
	}

	err = graph.AddEdge(myService2, myService)
	if err == nil {
		t.Errorf("Should not create circular dependency")
	}
	if nerr, ok := err.(*CircularDependencyError); !ok {
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

	err := graph.AddVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}
	err = graph.AddVertex(myService2)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	err = graph.AddEdge(myService, myService2)
	if err != nil {
		t.Errorf("Should add edge! %s", err.Error())
	}

	err = graph.AddEdge(myService, myService2)
	if err == nil {
		t.Errorf("Should not create duplicate edge")
	}
	if nerr, ok := err.(*DuplicateEdgeError); !ok {
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

	err := graph.AddVertex(myService)
	if err != nil {
		t.Errorf("Should add vertex! %s", err.Error())
	}

	actualService, err := graph.GetVertex(myService.Id)
	if err != nil {
		t.Errorf("Should be able to get vertex. %s", err.Error())
	}
	if actualService != myService {
		t.Errorf("Created vertex should be the same as the one created. %s", err.Error())
	}
}
