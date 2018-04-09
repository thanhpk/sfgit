package main

import (
	. "github.com/thanhpk/log"
)

func main() {
	Log("thanh")
	WithStack("thanh2")
	A()
}

func A() {
	B()
}

func B() {
	WithStack("nested")
}
