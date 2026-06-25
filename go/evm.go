// Package evm is an **incomplete** implementation of the Ethereum Virtual
// Machine for the "EVM From Scratch" course:
// https://github.com/w1nt3r-eth/evm-from-scratch
//
// To work on EVM From Scratch In Go:
//
// - Install Golang: https://golang.org/doc/install
// - Go to the `go` directory: `cd go`
// - Edit `evm.go` (this file!), see TODO below
// - Run `go test ./...` to run the tests
package evm

import (
	"math/big"
)

func push(stack []*big.Int, value *big.Int) []*big.Int {
	return append([]*big.Int{value}, stack...)
}

func pop(stack []*big.Int) ([]*big.Int, *big.Int) {
	if len(stack) == 0 {
		return stack, nil
	}
	value := stack[0]
	stack = stack[1:]
	return stack, value
}

// Run runs the EVM code and returns the stack and a success indicator.
func Evm(code []byte) ([]*big.Int, bool) {
	var stack []*big.Int
	pc := 0

	for pc < len(code) {
		op := code[pc]
		pc++

		// TODO: Implement the EVM here!
		switch op {
		case 0x00: // STOP
			return stack, true

		case 0x5f: // PUSH0
			stack = push(stack, new(big.Int).SetInt64(0))

		case 0x60: // PUSH1
			if pc >= len(code) {
				return stack, false
			}
			stack = push(stack, new(big.Int).SetInt64(int64(code[pc])))
			pc++

		case 0x61: // PUSH2
			if pc+1 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+2])
			stack = push(stack, value)
			pc += 2

		case 0x63: //PUSH4
			if pc+3 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+4])
			stack = push(stack, value)
			pc += 4

		case 0x65: // PUSH6
			if pc+5 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+6])
			stack = push(stack, value)
			pc += 6

		case 0x69: // PUSH10
			if pc+9 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+10])
			stack = push(stack, value)
			pc += 10

		case 0x6A: // PUSH11
			if pc+10 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+11])
			stack = push(stack, value)
			pc += 11

		case 0x7F: // PUSH32
			if pc+31 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+32])
			stack = push(stack, value)
			pc += 32

		case 0x50: // POP {
			if len(stack) == 0 {
				return stack, false
			}
			stack, _ = pop(stack)

		case 0x01: // ADD
			if len(stack) < 2 {
				return stack, false
			}
			a := stack[0]
			b := stack[1]
			result := new(big.Int).Add(a, b)
			modulus := new(big.Int).Lsh(big.NewInt(1), 256)
			result.Mod(result, modulus)
			stack = stack[2:]
			stack = push(stack, result)

		case 0x02: // MUL
			if len(stack) < 2 {
				return stack, false
			}
			a := stack[0]
			b := stack[1]
			result := new(big.Int).Mul(a, b)
			modulus := new(big.Int).Lsh(big.NewInt(1), 256)
			result.Mod(result, modulus)
			stack = stack[2:]
			stack = push(stack, result)

		case 0x03: // SUB
			if len(stack) < 2 {
				return stack, false
			}
			a := stack[0]
			b := stack[1]
			result := new(big.Int).Sub(a, b)
			modulus := new(big.Int).Lsh(big.NewInt(1), 256)
			result.Mod(result, modulus)
			stack = stack[2:]
			stack = push(stack, result)

		 case 0x04: // DIV
		 	if len(stack) < 2 {
		 		return stack, false
		 	}
		 	a := stack[0]
		 	b := stack[1]
		 	var result *big.Int
		 	if b.Sign() == 0 {
		 		result = big.NewInt(0)
		 	} else {
		 		result = new(big.Int).Div(a, b)
		 	}
		 	stack = stack[2:]
		 	stack = push(stack, result)

		case 0x06: // MOD
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			// Pop dividend (top of stack)
			stack, a = pop(stack)
			// Pop divisor (second element)
			stack, b = pop(stack)
			// EVM: x % 0 = 0
			if b.Sign() == 0 {
				stack = push(stack, big.NewInt(0))
				break
			}
			result := new(big.Int).Mod(a, b)
			stack = push(stack, result)
		}
		_ = op // delete this; it's only here to make the compiler think you're already using `op`
	}
	return stack, true
}
