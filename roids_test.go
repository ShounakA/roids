package roids_test

import (
	"testing"

	"github.com/ShounakA/roids"
)

type (
	testInterface interface {
		DoSomethingBob() string
	}

	dependedService interface {
		PlanSomething() string
	}

	iCycleService interface {
		WontWorkBeforeTo() string
	}

	ibCycleService interface {
		WontWorkBeforeMain()
	}

	iToCycleService interface {
		WontWorkBeforeB() string
	}
)

type (
	testObject struct {
		something      string
		dependedObject dependedService
	}

	dependedObject struct {
		plan string
	}

	cycleService struct {
		cycle iToCycleService
	}

	toCycleService struct {
		bCycle ibCycleService
	}

	bCycleService struct {
		ogCycle iCycleService
	}
)

func newTestObject(service dependedService) *testObject {
	return &testObject{
		something:      "Testing add",
		dependedObject: service,
	}
}

func newDependedObject() *dependedObject {
	return &dependedObject{
		plan: "Drive",
	}
}

func newCycle(toCycle iToCycleService) *cycleService {
	return &cycleService{
		cycle: toCycle,
	}
}

func newToCycle(bCycle ibCycleService) *toCycleService {
	return &toCycleService{
		bCycle: bCycle,
	}
}

func newBCycle(mainCycle iCycleService) *bCycleService {
	return &bCycleService{
		ogCycle: mainCycle,
	}
}

func (obj *testObject) DoSomethingBob() string {
	obj.dependedObject.PlanSomething()
	println(obj.something)
	return obj.something
}

func (obj *dependedObject) PlanSomething() string {
	println(obj.plan)
	return obj.plan
}

func (obj *cycleService) WontWorkBeforeTo() string {
	return "to"
}

func (obj *toCycleService) WontWorkBeforeB() string {
	return "b"
}

func (obj *bCycleService) WontWorkBeforeMain() {
	return
}

func TestGetRoids(t *testing.T) {
	firstRoids := roids.GetRoids()
	secondRoids := roids.GetRoids()
	if firstRoids != secondRoids {
		t.Error("Both roids should be the same instance.")
	}
}

func TestAddStaticService(t *testing.T) {

	_ = roids.GetRoids()

	err := roids.AddStaticService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}
	err = roids.AddStaticService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}

	roids.UNSAFE_Clear()
}

func TestAddStaticService_IncorrectOrder(t *testing.T) {

	_ = roids.GetRoids()

	err := roids.AddStaticService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependency", err.Error())
	}
	err = roids.AddStaticService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Added dependency with first define the service.", err.Error())
	}
	roids.UNSAFE_Clear()

}

func TestAddStaticService_CircularDependency(t *testing.T) {
	_ = roids.GetRoids()

	err := roids.AddStaticService(new(iCycleService), newCycle)
	if err != nil {
		t.Error("Should be able to add simple out of order dependencies.", err.Error())
	}
	err = roids.AddStaticService(new(iToCycleService), newToCycle)
	if err != nil {
		t.Error("Should be able to add simple out of order dependencies.", err.Error())
	}
	err = roids.AddStaticService(new(ibCycleService), newBCycle)
	if err == nil {
		t.Error("Should catch circular dependency here!!")
	}
	roids.UNSAFE_Clear()
}

func TestAddStaticService_InvalidInterface(t *testing.T) {
	_ = roids.GetRoids()

	err := roids.AddStaticService(new(iCycleService), newBCycle)
	if err == nil {
		t.Error("Should catch that impl does not match spec.", err.Error())
	}
	if nerr, ok := err.(*roids.ServiceError); !ok {
		t.Errorf("%s should be ServiceError", nerr.Error())
	}
	roids.UNSAFE_Clear()
}

func TestAddStaticService_NotAConstructor(t *testing.T) {
	_ = roids.GetRoids()

	err := roids.AddStaticService(new(iCycleService), &cycleService{})
	if err == nil {
		t.Error("Should catch that impl does not match spec.", err.Error())
	}
	if nerr, ok := err.(*roids.InjectorError); !ok {
		t.Error("Unexpected error returned.", nerr.Error())
	}
	roids.UNSAFE_Clear()
}

func TestInject_Transient_Branch(t *testing.T) {

	_ = roids.GetRoids()

	err := roids.AddTransientService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}
	err = roids.AddStaticService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}

	roids.Build()

	testService := roids.Inject[testInterface]()
	if testService.DoSomethingBob() != "Testing add" {
		t.Error("Did not inject service correctly.")
	}
	testService2 := roids.Inject[testInterface]()
	if testService == testService2 {
		t.Error("No transient :(")
	}

	roids.UNSAFE_Clear()
}

func TestInject_Transient_Leaf(t *testing.T) {

	_ = roids.GetRoids()

	err := roids.AddStaticService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}
	err = roids.AddTransientService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}

	roids.Build()

	testService := roids.Inject[testInterface]()
	if testService.DoSomethingBob() != "Testing add" {
		t.Error("Did not inject service correctly.")
	}
	testService2 := roids.Inject[testInterface]()
	if testService != testService2 {
		t.Error("Services should be the same, otherwise its not static.")
	}
	depService := roids.Inject[dependedService]()
	if depService.PlanSomething() != "Drive" {
		t.Error("Did not inject correctly.")
	}
	depService2 := roids.Inject[dependedService]()
	if depService == depService2 {
		t.Error("Services should not be the same, otherwise they are not transient.")
	}

	roids.UNSAFE_Clear()
}

func TestInject_Static(t *testing.T) {

	_ = roids.GetRoids()

	err := roids.AddStaticService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}
	err = roids.AddStaticService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}

	roids.Build()

	testService := roids.Inject[testInterface]()
	if testService.DoSomethingBob() != "Testing add" {
		t.Error("Did not inject service correctly.")
	}
	testService2 := roids.Inject[testInterface]()
	if testService != testService2 {
		t.Error("Services should be the same, otherwise its not static.")
	}

	roids.UNSAFE_Clear()
}

func TestAddTransientService(t *testing.T) {

	_ = roids.GetRoids()

	err := roids.AddTransientService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}
	err = roids.AddTransientService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.", err.Error())
	}

	roids.UNSAFE_Clear()
}
