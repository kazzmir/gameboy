package core

type MBC interface {
}

type MBC0 struct {
}

type MBC1 struct {
}

var _ MBC = &MBC0{}
var _ MBC = &MBC1{}
