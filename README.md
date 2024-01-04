# Inject Roids

Roids is a dependency injection container that you can use to share and pass dependencies into your services.

[![Go](https://github.com/ShounakA/roids/actions/workflows/build-test.yml/badge.svg)](https://github.com/ShounakA/roids/actions/workflows/build-test.yml)

## Get Roids

```
go get github.com/ShounakA/roids
```

## Usage

```golang
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

func main() {

    // Instantiate one and only needle
	_ = needle.GetNeedle()

    // Add your services
	needle.AddService(new(IOmegalulService), NewOmegalul)
	needle.AddService(new(IDepService), NewDepService)
	needle.AddService(new(IJuiceService), NewJuiceService)
	needle.AddService(new(ITestService), NewTestService)
	needle.Build()

    // Inject your instantiated services with configured implementations anywhere in your app
	testService := needle.Inject[ITestService]()
	depService := needle.Inject[IDepService]()
	juiceService := needle.Inject[IJuiceService]()

	juiceService.Juice(3)
	testService.Something()
	depService.AlsoDoSomething()
	testService.Omegalul()
}
```