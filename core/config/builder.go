package config

func Create[T any]() *T {
	return new(T)
}

type IConfiguration[T any] interface {
	Config() T
}

type RoidsConfiguration[T any] struct {
	Roids roidsSetting `json:"roids" yaml:"roids"`
	App   T            `json:"app" yaml:"app"`
}

type roidsSetting struct {
	Version string `json:"version" yaml:"version"`
}

func (b *RoidsConfiguration[T]) Config() T {
	return b.App
}
