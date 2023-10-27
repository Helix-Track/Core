package main

import "fmt"

type Celsius float64
type Fahrenheit float64

func main() {

	const (
		AbsoluteZeroC Celsius = -273.15
		FreezingC     Celsius = 0
		BoilingC      Celsius = 100
	)

	temps := []Celsius{FreezingC, FreezingC, FreezingC}

	temps[0] = AbsoluteZeroC
	temps[1] = FreezingC
	temps[2] = BoilingC

	converted := CToF(temps[0])
	fmt.Println(converted)

	converted = CToF(temps[1])
	fmt.Println(converted)

	converted = CToF(temps[2])
	fmt.Println(converted)
}

func CToF(c Celsius) Fahrenheit { return Fahrenheit(c*9/5 + 32) }
func FToC(f Fahrenheit) Celsius { return Celsius((f - 32) * 5 / 9) }
