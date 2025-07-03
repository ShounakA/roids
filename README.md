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

![Roids Logo](/assets/roids.jpeg)

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

### Configuration File
You can add your own configuration file. 
The name of the file can be anything, but must follow a certain interface (shape). Only JSON and Yaml files are currently supported.

```javascript
{
	// the "roids" field is currently required. 
    "roids": {
        "version": "0.4.0"
    },
    "app": {
		// this should match the shape of the struct you want to add.
        "message": "Test from JSON",
        "somenumber": 420,
        "somearray": [
            "this is a test",
            "this is a another test",
            "here is another"
        ],
        "complextype": {
            "somenumber2": 12.4,
            "otherarray": [
                123, 456, 987
            ]
        }
    }
}
```

```yaml
# roids field is required.
roids:
  version: 0.4.0
# this should match the shape of the struct you want to add.
app:
  message: "Test from YAML"
  somenumber: 69
  somearray:
  - "this is a test"
  - "this is a another test"
  - "here is another"
  complextype:
    somenumber2: 12.4
    otherarray:
    - 123
    - 456
    - 987

```

```golang

type TestConfig struct {
	Message string `json:"message"`
	SomeNumber int `json:"somenumber"`
	SomeArray []string `json:"somearray"`
	ComplexType struct {
		SomeNumber2 float64 `json:"somenumber2"`
		OtherArray []float64 `json:"otherarray"`
	} `json:"complextype"`
}

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
	
	// [Optional] add a configuration file. You can inject this config anywhere in your app.
	err := roids.AddConfigurationBuilder[TestConfig]("./roids.settings.json", core.JsonConfig)
	if err != nil {
		e.Logger.Fatal("Could not read configuration file.")
	}

    // Add your servicess
	roids.AddStaticService(new(HelloWorld), func(configAdapter config.IConfiguration[TestConfig]) {
		config := configAdapter.Config()
		return &Hw {
			message: config.Message
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