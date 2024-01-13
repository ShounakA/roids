# Inject Roids

Roids is a simple dependency injection container that you can use to share and pass services into your application.

[![Roids Main Build](https://github.com/ShounakA/roids/actions/workflows/build-test.yml/badge.svg)](https://github.com/ShounakA/roids/actions/workflows/build-test.yml)

## Get Roids
```
go get github.com/ShounakA/roids
```

## Features

- Simple setup
- Constructor-like dependency injection
- Http-Framework agnostic

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

    // Instantiate the one and only needle
	_ = needle.GetRoids()

    // Add your services
	needle.AddService(new(IOmegalulService), NewOmegalul)
	needle.AddService(new(IDepService), NewDepService)
	needle.AddService(new(IJuiceService), NewJuiceService)
	needle.AddService(new(ITestService), NewTestService)
	
	// Build your needle, to instantiate your services
	needle.Build()

    // Inject your instantiated services with configured implementations anywhere in your app
	testService := needle.Inject[ITestService]()
	depService := needle.Inject[IDepService]()
	juiceService := needle.Inject[IJuiceService]()

	// Do stuff
	juiceService.Juice(3)
	testService.Something()
	depService.AlsoDoSomething()
	testService.Omegalul()
}
```
