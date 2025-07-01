package roids_test

import (
	"log"
	"testing"

	"github.com/ShounakA/roids"
	"github.com/ShounakA/roids/core"
	"github.com/ShounakA/roids/core/config"
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

	myInterface interface {
		SameShape() string
	}

	myInterfacePart2 interface {
		SameShape() string
	}
)

type shape struct {
	test string
	dep  testInterface
}

type shapePart2 struct {
	test string
	dep  testInterface
}

func newShape(dep testInterface) *shape {
	return &shape{
		test: "test",
		dep:  dep,
	}
}

func newShapePart2(dep testInterface) *shapePart2 {
	return &shapePart2{
		test: "testset",
		dep:  dep,
	}
}

func (s *shape) SameShape() string {
	return s.test + s.dep.DoSomethingBob()
}

func (s *shapePart2) SameShape() string {
	return s.test + s.dep.DoSomethingBob()
}

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

type SqliteProvider struct {
	db string
}

type IDbProvider interface {
	Close() error
}

func NewSqliteProvider() *SqliteProvider {
	return &SqliteProvider{
		db: "test",
	}
}

func (s *SqliteProvider) Close() error {
	return nil
}

type TodoRepository struct {
	db       IDbProvider
	MemCache ICache
}

type ITodoRepository interface {
	DoStuff() error
}

func (t *TodoRepository) DoStuff() error {
	return nil
}

// NewTodoRepository creates a new TodoRepository
// It requires a database provider and a cache
func NewTodoRepository(db IDbProvider, memCache ICache) *TodoRepository {
	return &TodoRepository{
		db:       db,
		MemCache: memCache,
	}
}

type ICache interface {
	Delete(k string)
}

type MyCache struct {
	test string
}

func NewCache() *MyCache {
	return &MyCache{
		test: "test",
	}
}

func (c *MyCache) Delete(k string) {
	return
}

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

type testConfigInjectedService struct {
	Message string
}

type iTestConfigInjectedService interface {
	ShowInjectedMessage() string
}

func newTestConfigInjectedService(config config.IConfiguration[TestConfig]) *testConfigInjectedService {
	injectedMessage := config.Config().Message
	return &testConfigInjectedService{
		Message: injectedMessage,
	}
}

func (inj *testConfigInjectedService) ShowInjectedMessage() string {
	return inj.Message
}

type TestConfig struct {
	Message     string   `json:"message"`
	SomeNumber  int      `json:"somenumber"`
	SomeArray   []string `json:"somearray"`
	ComplexType struct {
		SomeNumber2 float64   `json:"somenumber2"`
		OtherArray  []float64 `json:"otherarray"`
	} `json:"complextype"`
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
	if nerr, ok := err.(*core.ServiceError); !ok {
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
	if nerr, ok := err.(*core.InjectorError); !ok {
		t.Error("Unexpected error returned.", nerr.Error())
	}
	roids.UNSAFE_Clear()
}

func TestInject_Transient_Branch(t *testing.T) {

	_ = roids.GetRoids()

	err := roids.AddTransientService(new(ITodoRepository), NewTodoRepository)
	if err != nil {
		log.Fatal("Did not bind service.", err.Error())
		return
	}
	err = roids.AddStaticService(new(IDbProvider), NewSqliteProvider)
	if err != nil {
		log.Fatal("Did not bind service.", err.Error())
		return
	}
	err = roids.AddStaticService(new(ICache), NewCache)
	if err != nil {
		log.Fatal("Did not bind service.", err.Error())
		return
	}
	roids.Build()
	todoRepo := roids.Inject[ITodoRepository]()
	todoRepo.DoStuff()
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

func TestAddConfigurationBuilder_JSON(t *testing.T) {
	_ = roids.GetRoids()

	err := roids.AddConfigurationBuilder[TestConfig]("./roids.settings.json", core.JsonConfig)
	if err != nil {
		t.Error("Should add configuration with no errors")
	}

	roids.Build()
	cfg := roids.Inject[config.IConfiguration[TestConfig]]()
	msg := cfg.Config().Message
	if msg != "Test from JSON" {
		t.Error("Should add configuration file.")
	}

	num := cfg.Config().SomeNumber
	if num != 420 {
		t.Error("Should deseriallize numbers correctly.")
	}

	somearray := cfg.Config().SomeArray
	if len(somearray) != 3 {
		t.Error("Should deseriallize arrays correctly.")
	}

	complextype := cfg.Config().ComplexType
	if complextype.SomeNumber2 != 12.4 {
		t.Error("Should deseriallize complex types")
	}

	if len(complextype.OtherArray) != 3 {
		t.Error("Should deseriallize complex type arrays")
	}

	expectedArray := []float64{123, 456, 987}
	for i, v := range complextype.OtherArray {
		if expectedArray[i] != v {
			t.Errorf("Expected element %f. Got %f", expectedArray[i], v)
		}
	}

	roids.UNSAFE_Clear()
}

func TestAddConfigurationBuilder_YAML(t *testing.T) {
	_ = roids.GetRoids()

	err := roids.AddConfigurationBuilder[TestConfig]("./roids.settings.yaml", core.YamlConfig)
	if err != nil {
		t.Error("Should add configuration with no errors")
	}

	roids.Build()
	cfg := roids.Inject[config.IConfiguration[TestConfig]]()
	msg := cfg.Config().Message
	if msg != "Test from YAML" {
		t.Errorf("Should add configuration file. Got %s", msg)
	}

	num := cfg.Config().SomeNumber
	if num != 69 {
		t.Error("Should deseriallize numbers correctly.")
	}

	somearray := cfg.Config().SomeArray
	if len(somearray) != 3 {
		t.Error("Should deseriallize arrays correctly.")
	}

	complextype := cfg.Config().ComplexType
	if complextype.SomeNumber2 != 12.4 {
		t.Error("Should deseriallize complex types")
	}

	if len(complextype.OtherArray) != 3 {
		t.Error("Should deseriallize complex type arrays")
	}

	expectedArray := []float64{123, 456, 987}
	for i, v := range complextype.OtherArray {
		if expectedArray[i] != v {
			t.Errorf("Expected element %f. Got %f", expectedArray[i], v)
		}
	}

	roids.UNSAFE_Clear()
}

func TestInjectIConfigurationIntoStaticService(t *testing.T) {
	_ = roids.GetRoids()

	err := roids.AddConfigurationBuilder[TestConfig]("./roids.settings.yaml", core.YamlConfig)
	if err != nil {
		t.Error("Should add configuration with no errors")
	}

	roids.AddStaticService(new(iTestConfigInjectedService), newTestConfigInjectedService)

	roids.Build()

	cfg := roids.Inject[config.IConfiguration[TestConfig]]()
	injSvc := roids.Inject[iTestConfigInjectedService]()

	msg := cfg.Config().Message
	actualMsg := injSvc.ShowInjectedMessage()
	if msg != actualMsg {
		t.Errorf("Should have injected config into service %s. Got %s", msg, actualMsg)
	}

	roids.UNSAFE_Clear()
}

func TestInjectionOfSameShape(t *testing.T) {
	_ = roids.GetRoids()

	if err := roids.AddStaticService(new(dependedService), newDependedObject); err != nil {
		t.Errorf("Should be able to add simple dependencies. %s", err.Error())
	}

	if err := roids.AddStaticService(new(testInterface), newTestObject); err != nil {
		t.Errorf("Should add configuration with no errors: %s", err.Error())
	}

	if err := roids.AddStaticService(new(myInterface), newShape); err != nil {
		t.Errorf("Should add configuration with no errors: %s", err.Error())
	}

	if err := roids.AddStaticService(new(myInterfacePart2), newShapePart2); err != nil {
		t.Errorf("Should add configuration with no errors %s", err.Error())
	}

	roids.Build()

	shape := roids.Inject[myInterface]()
	if s := shape.SameShape(); s != "testTesting add" {
		t.Errorf("Expected 'testTesting add' but got %s", s)
	}
	shapePart2 := roids.Inject[myInterfacePart2]()
	if s := shapePart2.SameShape(); s != "testsetTesting add" {
		t.Errorf("Expected 'testsetTesting add' but got %s", s)
	}

	roids.UNSAFE_Clear()
}
