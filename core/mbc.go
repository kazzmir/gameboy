package core

import (
    "fmt"
)

type MBC interface {
}

type MBC0 struct {
}

type MBC1 struct {
}

var _ MBC = &MBC0{}
var _ MBC = &MBC1{}

func MakeMBC(mbcType uint8) (MBC, error) {
    switch mbcType {
        case 0:
            return &MBC0{}, nil
        case 1:
            return &MBC1{}, nil
        default:
            return nil, fmt.Errorf("Unknown MBC type")
    }
}
