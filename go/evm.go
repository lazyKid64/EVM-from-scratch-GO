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

func push( stack []*big.Int, value *big.Int) []*big.Int {
	return append(stack, value)
}

func pop(stack []*big.Int) ([]*big.Int, *big.Int) {
	if len(stack) == 0 {
		return stack, nil
	}
	return stack[:len(stack)-1], stack[len(stack)-1]
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
		 stack = push(stack , new(big.Int).SetInt64(0))
		
		case 0x60: // PUSH1
			if pc >= len(code) {
				return stack, false
			}
			stack = push(stack, new(big.Int).SetInt64(int64(code[pc])))
		
	    case 0x61: // PUSH2
		   if pc +1 >= len(code) {
			   return stack, false
		}
		    value := new(big.Int).SetBytes(code[pc : pc+2])
		    stack = push(stack, value)
		    pc +=2

	    case 0x63: //PUSH4
		   if pc +3 >= len(code) {
			   return stack, false
		}
		    value := new(big.Int).SetBytes(code[pc : pc+4])
		    stack = push(stack, value)
		    pc +=4
	
        case 0x65 : // PUSH6
	       if pc +5 >= len(code) {
		       return stack, false
	    }
	        value := new(big.Int).SetBytes(code[pc : pc+6])
	        stack = push(stack, value)
	        pc +=6
		
		case 0x69: // PUSH10
			if pc +9 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+10])
			stack = push(stack, value)
			pc += 10	
		
		case 0x6A: // PUSH11
			if pc +10 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+11])
			stack = push(stack, value)
			pc += 11

		case 0x7F: // PUSH32
			if pc +31 >= len(code) {
				return stack, false
			}
			value := new(big.Int).SetBytes(code[pc : pc+32])
			stack = push(stack, value)
			pc += 32	
		
		case 	
		}
		_ = op // delete this; it's only here to make the compiler think you're already using `op`
	}
	return stack, true
}
