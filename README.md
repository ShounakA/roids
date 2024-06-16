<!-- markdownlint-configure-file {
  "MD013": {
    "code_blocks": false,
    "tables": false
  },
  "MD033": false,
  "MD041": false
} -->

<div align="center">

<hr />

# Roids: Dependency Injection

[Features](#features) •
[Getting Started](#get-roids) •
[Usage](#usage) •
[Building](#building-roids) •
[Enhancements](#enhancements)

<br/>

![Roids Logo](/assets/roids.png)

Roids is a simple dependency injection container that you can use to share and pass services into your application.
</div>

<br/>

[![Roids Main Build](https://github.com/ShounakA/roids/actions/workflows/build-test.yml/badge.svg)](https://github.com/ShounakA/roids/actions/workflows/build-test.yml)


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

type interface HelloWorld {
	SayHello() string
}

type struct Hw {
	message string
}

func (hw *Hw) SayHello() {
	println("hello,", hw.message)
	return hw.message
}

func main() {

    // Instantiate the one and only roids container
	_ = roids.GetRoids()

    // Add your servicess
	roids.AddStaticService(new(HelloWorld), func() {
		return &Hw {
			message: "chad"
		}
	})
	
	// Build your needle, to instantiate your services
	roids.Build()


	// Do stuff
	e.GET("/", func(c echo.Context) error {
		// Inject your instantiated services with configured implementations anywhere in your app
		helloService := roids.Inject[HelloWorld]()

		return c.String(http.StatusOK, helloService.SayHello())
	})
	e.Logger.Fatal(e.Start(":1323"))
s
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