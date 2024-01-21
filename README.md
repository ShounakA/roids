<div align="center">
	<h1>
		Roids: Dependency Injection
	</h1>

	[![Roids Main Build](https://github.com/ShounakA/roids/actions/workflows/build-test.yml/badge.svg)](https://github.com/ShounakA/roids/actions/workflows/build-test.yml)

	Roids is a simple dependency injection container that you can use to share and pass services into your application.

	[Features](#features) | 
	[Getting Started](#get-roids) | 
	[Usage](#usage) | 
	[Building](#building-roids) | 
	[Enhancements](#enhancements)
</div>

## Features

- Simple setup
  - global container instance
  - automatic and manual injecting
  - out of order configuration
  - error handling only on setup
- Constructor-like dependency injection
- 2 dependency lifetimes: 
  - Static: Created once and shared. Lives for life of container.
  - Transient: Created everytime it is injected. Lives for life of the dependency using it, or life of the last pointer referencing it.
- Http-Framework agnostic
  
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

    // Instantiate the one and only roids container
	_ = roids.GetRoids()

    // Add your services
	roids.AddTransientService(new(IOmegalulService), NewOmegalul)
	roids.AddTransientService(new(IDepService), NewDepService)
	roids.AddStaticService(new(IJuiceService), NewJuiceService)
	roids.AddStaticService(new(ITestService), NewTestService)
	
	// Build your needle, to instantiate your services
	roids.Build()

    // Inject your instantiated services with configured implementations anywhere in your app
	testService := roids.Inject[ITestService]()
	depService := roids.Inject[IDepService]()
	juiceService := roids.Inject[IJuiceService]()

	// Do stuff
	juiceService.Juice(3)
	testService.Something()
	depService.AlsoDoSomething()
	testService.Omegalul()
}
```

## Building `roids`

### Prerequisites
 - Golang 1.21.1 +

### Build module
```bash
	go build
```

### Test module
```bash
	go test
```
### Run example application
```bash
	go run testapp/main.go
```

## Enhancements

- Internal logging
- Startup/Cleanup actions for services