/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 * A small application using Roids.
 */
package main

import "github.com/ShounakA/roids/needle"

type IOmegalulService interface {
	SayOmegalul()
}

type IDepService interface {
	AlsoDoSomething()
}

type ITestService interface {
	Something()
	Omegalul()
}

type IJuiceService interface {
	Juice(num uint)
}

type JuiceService struct {
	Message  string
	omegalul IOmegalulService
}

type TestService struct {
	yo       int
	Dsvc     IDepService
	omegalul IOmegalulService
}

type DepService struct {
	Do int
}

type omegalul struct {
	Message string
}

func NewJuiceService(o IOmegalulService) *JuiceService {
	return &JuiceService{
		omegalul: o,
		Message:  "Juicing...",
	}
}

func NewTestService(dService IDepService, o IOmegalulService) *TestService {
	return &TestService{
		yo:       5,
		Dsvc:     dService,
		omegalul: o,
	}
}

func NewDepService() *DepService {
	return &DepService{
		Do: 5,
	}
}

func NewOmegalul() *omegalul {
	return &omegalul{
		Message: "L OMEGALUL L",
	}
}

func (i *TestService) Something() {
	println("Something happened!!!!")
}

func (i *TestService) Omegalul() {
	i.omegalul.SayOmegalul()
}

func (ds *DepService) AlsoDoSomething() {
	println("I Also did something!!!")
}

func (i *omegalul) SayOmegalul() {
	println(i.Message)
}

func (js *JuiceService) Juice(num uint) {
	for i := uint(0); i < num; i++ {
		println(js.Message)
	}
	print("juice: ")
	js.omegalul.SayOmegalul()
}

func main() {

	_ = needle.GetRoids()

	needle.AddService(new(IOmegalulService), NewOmegalul)
	needle.AddService(new(IDepService), NewDepService)
	needle.AddService(new(IJuiceService), NewJuiceService)
	needle.AddService(new(ITestService), NewTestService)
	needle.Build()

	testService := needle.Inject[ITestService]()
	depService := needle.Inject[IDepService]()
	juiceService := needle.Inject[IJuiceService]()

	juiceService.Juice(3)
	testService.Something()
	depService.AlsoDoSomething()
	testService.Omegalul()

}
