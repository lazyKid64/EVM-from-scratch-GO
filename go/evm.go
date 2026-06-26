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

var (
	uint256 = new(big.Int).Lsh(big.NewInt(1), 256)
	int256  = new(big.Int).Lsh(big.NewInt(1), 255)
)

func toSigned(x *big.Int) *big.Int {
	result := new(big.Int).Set(x)

	if result.Cmp(int256) >= 0 {
		result.Sub(result, uint256)
	}

	return result
}
func toUnsigned(x *big.Int) *big.Int {
	result := new(big.Int).Set(x)

	if result.Sign() < 0 {
		result.Add(result, uint256)
	}

	return result
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
			stack, a = pop(stack)
			stack, b = pop(stack)
			if b.Sign() == 0 {
				stack = push(stack, big.NewInt(0))
				break
			}
			result := new(big.Int).Mod(a, b)
			stack = push(stack, result)

		case 0x08: // ADDMOD
			if len(stack) < 3 {
				return stack, false
			}
			var a, b, c *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			stack, c = pop(stack)
			if c.Sign() == 0 {
				stack = push(stack, big.NewInt(0))
				break
			}
			result := new(big.Int).Add(a, b)
			result.Mod(result, c)
			stack = push(stack, result)

		case 0x09: // MULMOD
			if len(stack) < 3 {
				return stack, false
			}
			var a, b, c *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			stack, c = pop(stack)
			if c.Sign() == 0 {
				stack = push(stack, big.NewInt(0))
				break
			}
			result := new(big.Int).Mul(a, b)
			result.Mod(result, c)
			stack = push(stack, result)

		case 0x0A: //EXP
			if len(stack) < 2 {
				return stack, false
			}
			a := stack[0]
			b := stack[1]
			result := new(big.Int).Exp(a, b, nil)
			modulus := new(big.Int).Lsh(big.NewInt(1), 256)
			result.Mod(result, modulus)
			stack = stack[2:]
			stack = push(stack, result)

		case 0x0B: // SIGNEXTEND
			if len(stack) < 2 {
				return stack, false
			}
			byteNum := stack[0]
			value := new(big.Int).Set(stack[1])

			if byteNum.Cmp(big.NewInt(31)) > 0 {
				stack = stack[2:]
				stack = push(stack, value)
				break
			}
			n := uint(byteNum.Uint64())
			bitPos := 8*n + 7
			if value.Bit(int(bitPos)) == 1 {
				mask := new(big.Int).Lsh(big.NewInt(1), 256)
				mask.Sub(mask, big.NewInt(1))
				lowerMask := new(big.Int).Lsh(big.NewInt(1), bitPos+1)
				lowerMask.Sub(lowerMask, big.NewInt(1))
				mask.Xor(mask, lowerMask)
				value.Or(value, mask)
			} else {
				mask := new(big.Int).Lsh(big.NewInt(1), bitPos+1)
				mask.Sub(mask, big.NewInt(1))
				value.And(value, mask)
			}
			stack = stack[2:]
			stack = push(stack, value)

		case 0x05: // SDIV

			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			a = toSigned(a)
			b = toSigned(b)
			if b.Sign() == 0 {
				stack = push(stack, big.NewInt(0))
				break
			}
			result := new(big.Int).Quo(a, b)
			result = toUnsigned(result)
			stack = push(stack, result)

		case 0x07: // SMOD

			if len(stack) < 2 {
				return stack, false
			}
			var x, y *big.Int
			stack, x = pop(stack)
			stack, y = pop(stack)
			x = toSigned(x)
			y = toSigned(y)
			if y.Sign() == 0 {
				stack = push(stack, big.NewInt(0))
				break
			}
			resultSMOD := new(big.Int).Rem(x, y)
			resultSMOD = toUnsigned(resultSMOD)
			stack = push(stack, resultSMOD)

		case 0x10: // LT
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := big.NewInt(0)
			if a.Cmp(b) < 0 {
				res.SetInt64(1)
			}
			stack = push(stack, res)

		case 0x11: // GT
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := big.NewInt(0)
			if a.Cmp(b) > 0 {
				res.SetInt64(1)
			}
			stack = push(stack, res)

		case 0x12: // SLT
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			a_s := toSigned(a)
			b_s := toSigned(b)
			res := big.NewInt(0)
			if a_s.Cmp(b_s) < 0 {
				res.SetInt64(1)
			}
			stack = push(stack, res)

		case 0x13: // SGT
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			a_s := toSigned(a)
			b_s := toSigned(b)
			res := big.NewInt(0)
			if a_s.Cmp(b_s) > 0 {
				res.SetInt64(1)
			}
			stack = push(stack, res)

		case 0x14: // EQ
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := big.NewInt(0)
			if a.Cmp(b) == 0 {
				res.SetInt64(1)
			}
			stack = push(stack, res)

		case 0x15: // ISZERO
			if len(stack) < 1 {
				return stack, false
			}
			var a *big.Int
			stack, a = pop(stack)
			res := big.NewInt(0)
			if a.Sign() == 0 {
				res.SetInt64(1)
			}
			stack = push(stack, res)

		case 0x16: // AND
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := new(big.Int).And(a, b)
			stack = push(stack, res)

		case 0x17: // OR
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := new(big.Int).Or(a, b)
			stack = push(stack, res)

		case 0x18: // XOR
			if len(stack) < 2 {
				return stack, false
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := new(big.Int).Xor(a, b)
			stack = push(stack, res)

		case 0x19: // NOT
			if len(stack) < 1 {
				return stack, false
			}
			var a *big.Int
			stack, a = pop(stack)
			mask := new(big.Int).Lsh(big.NewInt(1), 256)
			mask.Sub(mask, big.NewInt(1))
			res := new(big.Int).Xor(a, mask)
			stack = push(stack, res)

		case 0x1A: // BYTE
			if len(stack) < 2 {
				return stack, false
			}
			var i, x *big.Int
			stack, i = pop(stack)
			stack, x = pop(stack)
			res := big.NewInt(0)
			if i.Cmp(big.NewInt(31)) <= 0 {
				shiftAmt := (31 - uint(i.Uint64())) * 8
				res = new(big.Int).Rsh(x, shiftAmt)
				res.And(res, big.NewInt(0xFF))
			}
			stack = push(stack, res)

		case 0x1B: // SHL
			if len(stack) < 2 {
				return stack, false
			}
			var shift, value *big.Int
			stack, shift = pop(stack)
			stack, value = pop(stack)
			res := big.NewInt(0)
			if shift.Cmp(big.NewInt(256)) < 0 {
				res.Lsh(value, uint(shift.Uint64()))
				mask := new(big.Int).Lsh(big.NewInt(1), 256)
				mask.Sub(mask, big.NewInt(1))
				res.And(res, mask)
			}
			stack = push(stack, res)

		case 0x1C: // SHR
			if len(stack) < 2 {
				return stack, false
			}
			var shift, value *big.Int
			stack, shift = pop(stack)
			stack, value = pop(stack)
			res := big.NewInt(0)
			if shift.Cmp(big.NewInt(256)) < 0 {
				res.Rsh(value, uint(shift.Uint64()))
			}
			stack = push(stack, res)

		case 0x1D: // SAR
			if len(stack) < 2 {
				return stack, false
			}
			var shift, value *big.Int
			stack, shift = pop(stack)
			stack, value = pop(stack)
			res := big.NewInt(0)
			value_s := toSigned(value)
			if shift.Cmp(big.NewInt(256)) < 0 {
				res.Rsh(value_s, uint(shift.Uint64()))
				res = toUnsigned(res)
			} else {
				if value_s.Sign() < 0 {
					mask := new(big.Int).Lsh(big.NewInt(1), 256)
					res.Sub(mask, big.NewInt(1))
				}
			}
			stack = push(stack, res)

		case 0x80: // DUP1
			if len(stack) < 2 {
				return stack, false
			}
			a := stack[0]
			stack = push(stack, a)

		}
		_ = op // delete this; it's only here to make the compiler think you're already using `op`
	}
	return stack, true
}
