package core

// Constant to ID Static lifetimes
const StaticLifetime string = "Static"

// Constant to ID Transient lifetimes
const TransientLifetime string = "Transient"

type ConfigType int

const (
	JsonConfig ConfigType = iota
	YamlConfig
)
