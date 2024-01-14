package needle_test

import (
	"testing"

	"github.com/ShounakA/roids/needle"
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

func (obj *bCycleService) WontWorkBeforeMain() string {
	return "main"
}

func TestGetNeedle(t *testing.T) {
	roids := needle.GetRoids()
	secondRoids := needle.GetRoids()
	if roids != secondRoids {
		t.Error("Both roids should be the same instance.")
	}
}

func TestAddService(t *testing.T) {

	_ = needle.GetRoids()

	err := needle.AddService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.")
	}
	err = needle.AddService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependencies.")
	}

	needle.UNSAFE_Clear()
}

func TestAddService_IncorrectOrder(t *testing.T) {

	_ = needle.GetRoids()

	err := needle.AddService(new(dependedService), newDependedObject)
	if err != nil {
		t.Error("Should be able to add simple dependency", err.Error())
	}
	err = needle.AddService(new(testInterface), newTestObject)
	if err != nil {
		t.Error("Added dependency with first define the service.", err.Error())
	}
	needle.UNSAFE_Clear()

}

func TestAddService_CircularDependency(t *testing.T) {
	_ = needle.GetRoids()

	err := needle.AddService(new(iCycleService), newCycle)
	if err != nil {
		t.Error("Should be able to add simple out of order dependencies.", err.Error())
	}
	err = needle.AddService(new(iToCycleService), newToCycle)
	if err != nil {
		t.Error("Should be able to add simple out of order dependencies.", err.Error())
	}
	err = needle.AddService(new(ibCycleService), newBCycle)
	if err == nil {
		t.Error("Should catch circular dependency here!!")
	}

	needle.UNSAFE_Clear()
}
