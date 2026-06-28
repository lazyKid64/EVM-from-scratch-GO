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
	"fmt"
	"math/big"
	"golang.org/x/crypto/sha3"
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

func expandMemory(memory []byte, offset int, size int) []byte {
	if offset+size > len(memory) {
		newSize := ((offset + size + 31) / 32) * 32
		newMem := make([]byte, newSize)
		copy(newMem, memory)
		return newMem
	}
	return memory
}
type Tx struct {
	To       *big.Int
	From     *big.Int
	Origin   *big.Int
	GasPrice *big.Int
	Value    *big.Int
	Data     []byte
	IsStatic bool
}
type Block struct {
    Basefee    *big.Int
    Coinbase   *big.Int
    Timestamp  *big.Int
    Number     *big.Int
    Difficulty *big.Int
    GasLimit   *big.Int
    ChainId    *big.Int
}

type Account struct {
	Balance *big.Int
	Code    []byte
	Storage map[string]*big.Int
}

type State map[string]Account


type Log struct {
	Address *big.Int
	Data    []byte
	Topics  []*big.Int
}

// Run runs the EVM code and returns the stack and a success indicator.
func Evm(code []byte, tx Tx , block Block, state State) ([]*big.Int, bool, []Log, []byte) {

	var stack []*big.Int
	var memory []byte
	var logs []Log
	var lastReturnData []byte
	pc := 0

	validJumpDests := make(map[int]bool)
	for i := 0; i < len(code); {
		op := code[i]
		if op == 0x5b {
			validJumpDests[i] = true
		}
		if op >= 0x60 && op <= 0x7F {
			i += int(op - 0x60 + 1) + 1
		} else {
			i++
		}
	}

	for pc < len(code) {
		op := code[pc]
		pc++

		// TODO: Implement the EVM here!
		switch op {
		case 0x00: // STOP
			return stack, true, logs, nil

		case 0x5f: // PUSH0
			stack = push(stack, new(big.Int).SetInt64(0))

		case 0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f: // PUSH1 to PUSH32
			size := int(op - 0x60 + 1)
			if pc+size-1 >= len(code) {
				return stack, false, logs, nil
			}
			value := new(big.Int).SetBytes(code[pc : pc+size])
			stack = push(stack, value)
			pc += size


		case 0x50: // POP {
			if len(stack) == 0 {
				return stack, false, logs, nil
			}
			stack, _ = pop(stack)

		case 0x01: // ADD
			if len(stack) < 2 {
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := new(big.Int).And(a, b)
			stack = push(stack, res)

		case 0x17: // OR
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := new(big.Int).Or(a, b)
			stack = push(stack, res)

		case 0x18: // XOR
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var a, b *big.Int
			stack, a = pop(stack)
			stack, b = pop(stack)
			res := new(big.Int).Xor(a, b)
			stack = push(stack, res)

		case 0x19: // NOT
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var a *big.Int
			stack, a = pop(stack)
			mask := new(big.Int).Lsh(big.NewInt(1), 256)
			mask.Sub(mask, big.NewInt(1))
			res := new(big.Int).Xor(a, mask)
			stack = push(stack, res)

		case 0x1A: // BYTE
			if len(stack) < 2 {
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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
				return stack, false, logs, nil
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

		case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f: //DUP1 to DUP16
			n := int(op - 0x80 + 1)
			if len(stack) < n {
				return stack, false, logs, nil
			}
			value := new(big.Int).Set(stack[n-1])
			stack = push(stack, value)
		
		case 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e, 0x9f: //SWAP1 to SWAP16
			n := int(op - 0x90 + 1)
			if len(stack) < n+1 {
				return stack, false, logs, nil
			}
			stack[0], stack[n] = stack[n], stack[0]

		case 0x5b : // JUMPDEST

		case 0x56 : // JUMP
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var dest *big.Int
			stack, dest = pop(stack)
			destPC := int(dest.Uint64())
			if !validJumpDests[destPC] {
				return stack, false, logs, nil 
			}
			pc = destPC

		case 0x57 : // JUMPI
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var dest, cond *big.Int
			stack, dest = pop(stack)
			stack, cond = pop(stack)
			if cond.Sign() != 0 {
				destPC := int(dest.Uint64())
				if !validJumpDests[destPC] {
					return stack, false, logs, nil
				}
				pc = destPC
			}

		case 0x58 : // PC
			stack = push(stack, big.NewInt(int64(pc-1))) 

		case 0x59: // MSIZE
			stack = push(stack, big.NewInt(int64(len(memory))))

		case 0x20: // SHA3
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var offset, size *big.Int
			stack, offset = pop(stack)
			stack, size = pop(stack)
			off := int(offset.Uint64())
			sz := int(size.Uint64())
			
			var data []byte
			if sz > 0 {
				memory = expandMemory(memory, off, sz)
				data = memory[off : off+sz]
			}
			hash := sha3.NewLegacyKeccak256()
			hash.Write(data)
			hashBytes := hash.Sum(nil)
			stack = push(stack, new(big.Int).SetBytes(hashBytes))

		case 0x5a : // GAS
			maxUint256 := new(big.Int).Sub(uint256, big.NewInt(1))
			stack = push(stack, maxUint256)
         
		case 0x52 : // MSTORE
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var offset *big.Int
			var value *big.Int
			stack, offset = pop(stack)
			stack, value = pop(stack)
			off := int(offset.Uint64())
			memory = expandMemory(memory, off, 32)
			valueBytes := value.Bytes()
			padded := make([]byte, 32)
			copy(padded[32-len(valueBytes):], valueBytes)
			copy(memory[off:off+32], padded)

		case 0x51 : // MLOAD
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var offset *big.Int
			stack, offset = pop(stack)
			off := int(offset.Uint64())
			memory = expandMemory(memory, off, 32)
			value := new(big.Int).SetBytes(memory[off : off+32])
			stack = push(stack, value)

		case 0x53 : // MSTORE8
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var offset *big.Int
			var value *big.Int
			stack, offset = pop(stack)
			stack, value = pop(stack)
			off := int(offset.Uint64())
			memory = expandMemory(memory, off, 1)
			valueBytes := value.Bytes()
			if len(valueBytes) > 0 {
				memory[off] = valueBytes[len(valueBytes)-1]
			} else {
				memory[off] = 0
			}

		case 0x30: // ADDRESS
			if tx.To != nil {
				stack = push(stack, new(big.Int).Set(tx.To))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x33: // CALLER
			if tx.From != nil {
				stack = push(stack, new(big.Int).Set(tx.From))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x32: // ORIGIN
			if tx.Origin != nil {
				stack = push(stack, new(big.Int).Set(tx.Origin))
			} else {
				stack = push(stack, big.NewInt(0))
			}
			
		case 0x3A: // GASPRICE
			if tx.GasPrice != nil {
				stack = push(stack, new(big.Int).Set(tx.GasPrice))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x34: // CALLVALUE
			if tx.Value != nil {
				stack = push(stack, new(big.Int).Set(tx.Value))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x36: // CALLDATASIZE
			stack = push(stack, big.NewInt(int64(len(tx.Data))))

		case 0x35: // CALLDATALOAD
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var offset *big.Int
			stack, offset = pop(stack)
			off := int(offset.Uint64())
			res := make([]byte, 32)
			for i := 0; i < 32; i++ {
				if off+i < len(tx.Data) {
					res[i] = tx.Data[off+i]
				} else {
					res[i] = 0
				}
			}
			stack = push(stack, new(big.Int).SetBytes(res))

		case 0x37: // CALLDATACOPY
			if len(stack) < 3 {
				return stack, false, logs, nil
			}
			var destOffset, dataOffset, size *big.Int
			stack, destOffset = pop(stack)
			stack, dataOffset = pop(stack)
			stack, size = pop(stack)
			
			destOff := int(destOffset.Uint64())
			dataOff := int(dataOffset.Uint64())
			sz := int(size.Uint64())
			
			if sz > 0 {
				memory = expandMemory(memory, destOff, sz)
				for i := 0; i < sz; i++ {
					if dataOff+i < len(tx.Data) {
						memory[destOff+i] = tx.Data[dataOff+i]
					} else {
						memory[destOff+i] = 0
					}
				}
			}

		case 0x48 : // BASE FEE
			if block.Basefee != nil {
				stack = push(stack, new(big.Int).Set(block.Basefee))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x41: // COINBASE
			if block.Coinbase != nil {
				stack = push(stack, new(big.Int).Set(block.Coinbase))
			} else {
				stack = push(stack, big.NewInt(0))
			}
			
		case 0x42: // TIMESTAMP
			if block.Timestamp != nil {
				stack = push(stack, new(big.Int).Set(block.Timestamp))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x43: // NUMBER
			if block.Number != nil {
				stack = push(stack, new(big.Int).Set(block.Number))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x44: // DIFFICULTY
			if block.Difficulty != nil {
				stack = push(stack, new(big.Int).Set(block.Difficulty))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x45: // GASLIMIT
			if block.GasLimit != nil {
				stack = push(stack, new(big.Int).Set(block.GasLimit))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x46: // CHAINID
			if block.ChainId != nil {
				stack = push(stack, new(big.Int).Set(block.ChainId))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x40: // BLOCKHASH
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			stack, _ = pop(stack)
			stack = push(stack, big.NewInt(0))

		case 0x31: // BALANCE
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var addrBig *big.Int
			stack, addrBig = pop(stack)
			addrStr := fmt.Sprintf("0x%040x", addrBig)
			account, exists := state[addrStr]
			if exists && account.Balance != nil {
				stack = push(stack, new(big.Int).Set(account.Balance))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x38 : //CODESIZE
		     stack = push (stack , big.NewInt(int64(len(code))))

		case 0x39 : //CODECOPY
			if len(stack) < 3 {
				return stack, false, logs, nil
			}
			var destOffset, codeOffset, size *big.Int
			stack, destOffset = pop(stack)
			stack, codeOffset = pop(stack)
			stack, size = pop(stack)
			destOff := int(destOffset.Uint64())
			codeOff := int(codeOffset.Uint64())
			sz := int(size.Uint64())
			if sz > 0 {
				memory = expandMemory(memory, destOff, sz)
				for i := 0; i < sz; i++ {
					if codeOff+i < len(code) {
						memory[destOff+i] = code[codeOff+i]
					} else {
						memory[destOff+i] = 0
					}
				}
			}

		case 0x3B: // EXTCODESIZE
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var addrBig *big.Int
			stack, addrBig = pop(stack)
			addrStr := fmt.Sprintf("0x%040x", addrBig)
			account, exists := state[addrStr]
			if exists && account.Code != nil {
				stack = push(stack, big.NewInt(int64(len(account.Code))))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x3C: // EXTCODECOPY
			if len(stack) < 4 {
				return stack, false, logs, nil
			}
			var addrBig, destOffset, codeOffset, size *big.Int
			stack, addrBig = pop(stack)
			stack, destOffset = pop(stack)
			stack, codeOffset = pop(stack)
			stack, size = pop(stack)
			addrStr := fmt.Sprintf("0x%040x", addrBig)
			account, exists := state[addrStr]
			var extCode []byte
			if exists && account.Code != nil {
				extCode = account.Code
			}
			destOff := int(destOffset.Uint64())
			codeOff := int(codeOffset.Uint64())
			sz := int(size.Uint64())
			if sz > 0 {
				memory = expandMemory(memory, destOff, sz)
				for i := 0; i < sz; i++ {
					if codeOff+i < len(extCode) {
						memory[destOff+i] = extCode[codeOff+i]
					} else {
						memory[destOff+i] = 0
					}
				}
			}

		case 0x3d: // RETURNDATASIZE
			stack = push(stack, big.NewInt(int64(len(lastReturnData))))

		case 0x3e: // RETURNDATACOPY
			if len(stack) < 3 {
				return stack, false, logs, nil
			}
			var destOffset, offset, size *big.Int
			stack, destOffset = pop(stack)
			stack, offset = pop(stack)
			stack, size = pop(stack)
			destOff := int(destOffset.Uint64())
			off := int(offset.Uint64())
			sz := int(size.Uint64())
			if off+sz > len(lastReturnData) {
				return stack, false, logs, nil // Revert if out of bounds
			}
			if sz > 0 {
				memory = expandMemory(memory, destOff, sz)
				copy(memory[destOff:destOff+sz], lastReturnData[off:off+sz])
			}

		case 0x3f: // EXTCODEHASH
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var addrBig *big.Int
			stack, addrBig = pop(stack)
			addrStr := fmt.Sprintf("0x%040x", addrBig)
			account, exists := state[addrStr]
			if exists && len(account.Code) > 0 {
				hash := sha3.NewLegacyKeccak256()
				hash.Write(account.Code)
				stack = push(stack, new(big.Int).SetBytes(hash.Sum(nil)))
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0x47: // SELFBALANCE
			if tx.To != nil {
				addrStr := fmt.Sprintf("0x%040x", tx.To)
				account, exists := state[addrStr]
				if exists && account.Balance != nil {
					stack = push(stack, new(big.Int).Set(account.Balance))
				} else {
					stack = push(stack, big.NewInt(0))
				}
			} else {
				stack = push(stack, big.NewInt(0))
			}
	    
		case 0x54: // SLOAD
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var key *big.Int
			stack, key = pop(stack)
			var addrStr string
			if tx.To != nil {
				addrStr = fmt.Sprintf("0x%040x", tx.To)
			} else {
				addrStr = fmt.Sprintf("0x%040x", 0)
			}
			account := state[addrStr]
			keyStr := fmt.Sprintf("0x%064x", key)
			if account.Storage != nil && account.Storage[keyStr] != nil {
				stack = push(stack, new(big.Int).Set(account.Storage[keyStr]))
			} else {
				stack = push(stack, big.NewInt(0))
			}
	    
		case 0x55 : // SSTORE
			if tx.IsStatic {
				return stack, false, logs, nil
			}
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var key , value *big.Int
			stack , key = pop(stack)
			stack , value = pop(stack)
			var addrStr string
			if tx.To != nil {
				addrStr = fmt.Sprintf("0x%040x", tx.To)
			} else {
				addrStr = fmt.Sprintf("0x%040x", big.NewInt(0))
			}
			account := state[addrStr]
			if account.Storage == nil {
				account.Storage = make(map[string]*big.Int)
			}
			keyStr := fmt.Sprintf("0x%064x", key)
			account.Storage[keyStr] = value 
			state[addrStr] = account

		case 0xa0, 0xa1, 0xa2, 0xa3, 0xa4: // LOG0 to LOG4
			numTopics := int(op - 0xa0)
			if len(stack) < 2+numTopics {
				return stack, false, logs, nil
			}
			var offset, size *big.Int
			stack, offset = pop(stack)
			stack, size = pop(stack)
			var topics []*big.Int
			for j := 0; j < numTopics; j++ {
				var topic *big.Int
				stack, topic = pop(stack)
				topics = append(topics, topic)
			}
			off := int(offset.Uint64())
			sz := int(size.Uint64())
			if sz > 0 {
				memory = expandMemory(memory, off, sz)
			}
			data := make([]byte, sz)
			if sz > 0 {
				copy(data, memory[off:off+sz])
			}
			var logAddr *big.Int
			if tx.To != nil {
				logAddr = new(big.Int).Set(tx.To)
			} else {
				logAddr = big.NewInt(0)
			}
			logs = append(logs, Log{
				Address: logAddr,
				Data:    data,
				Topics:  topics,
			})
		case 0xf3: // RETURN
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var offset, size *big.Int
			stack, offset = pop(stack)
			stack, size = pop(stack)
			off := int(offset.Uint64())
			sz := int(size.Uint64())
			if sz > 0 {
				memory = expandMemory(memory, off, sz)
			}
			returnData := make([]byte, sz)
			if sz > 0 {
				copy(returnData, memory[off:off+sz])
			}
			return stack, true, logs, returnData

		case 0xfd: // REVERT
			if len(stack) < 2 {
				return stack, false, logs, nil
			}
			var offset, size *big.Int
			stack, offset = pop(stack)
			stack, size = pop(stack)
			off := int(offset.Uint64())
			sz := int(size.Uint64())
			if sz > 0 {
				memory = expandMemory(memory, off, sz)
			}
			returnData := make([]byte, sz)
			if sz > 0 {
				copy(returnData, memory[off:off+sz])
			}
			return stack, false, logs, returnData

		case 0xf1, 0xf4, 0xfa: // CALL, DELEGATECALL, STATICCALL
			isStaticCall := op == 0xfa
			isDelegateCall := op == 0xf4
			numArgs := 7
			if isStaticCall || isDelegateCall {
				numArgs = 6
			}
			if len(stack) < numArgs {
				return stack, false, logs, nil
			}
			var gas, addr, value, argsOffset, argsSize, retOffset, retSize *big.Int
			stack, gas = pop(stack)
			_ = gas
			stack, addr = pop(stack)
			if !isStaticCall && !isDelegateCall {
				stack, value = pop(stack)
			} else if isDelegateCall {
				if tx.Value != nil {
					value = tx.Value
				} else {
					value = big.NewInt(0)
				}
			} else {
				value = big.NewInt(0)
			}
			stack, argsOffset = pop(stack)
			stack, argsSize = pop(stack)
			stack, retOffset = pop(stack)
			stack, retSize = pop(stack)
			argsOff := int(argsOffset.Uint64())
			argsSz := int(argsSize.Uint64())
			retOff := int(retOffset.Uint64())
			retSz := int(retSize.Uint64())
			var callData []byte
			if argsSz > 0 {
				memory = expandMemory(memory, argsOff, argsSz)
				callData = make([]byte, argsSz)
				copy(callData, memory[argsOff:argsOff+argsSz])
			}
			addrStr := fmt.Sprintf("0x%040x", addr)
			targetAccount, exists := state[addrStr]
			var callCode []byte
			if exists {
				callCode = targetAccount.Code
			}
			innerTx := Tx{
				To:       addr,
				From:     tx.To,
				Origin:   tx.Origin,
				GasPrice: tx.GasPrice,
				Value:    value,
				Data:     callData,
				IsStatic: tx.IsStatic || isStaticCall,
			}
			if isDelegateCall {
				innerTx.To = tx.To
				innerTx.From = tx.From
				innerTx.Value = tx.Value
			}
			if innerTx.To == nil {
				innerTx.To = big.NewInt(0)
			}
			if innerTx.From == nil {
				innerTx.From = big.NewInt(0)
			}
			_, innerSuccess, innerLogs, innerRet := Evm(callCode, innerTx, block, state)
			lastReturnData = innerRet
			if innerSuccess {
				stack = push(stack, big.NewInt(1))
				logs = append(logs, innerLogs...)
			} else {
				stack = push(stack, big.NewInt(0))
			}
			if retSz > 0 {
				memory = expandMemory(memory, retOff, retSz)
				copyLen := retSz
				if len(lastReturnData) < copyLen {
					copyLen = len(lastReturnData)
				}
				copy(memory[retOff:retOff+copyLen], lastReturnData[:copyLen])
			}

		case 0xf0: // CREATE
			if tx.IsStatic {
				return stack, false, logs, nil
			}
			if len(stack) < 3 {
				return stack, false, logs, nil
			}
			var value, offset, size *big.Int
			stack, value = pop(stack)
			stack, offset = pop(stack)
			stack, size = pop(stack)
			
			off := int(offset.Uint64())
			sz := int(size.Uint64())
			var initCode []byte
			if sz > 0 {
				memory = expandMemory(memory, off, sz)
				initCode = make([]byte, sz)
				copy(initCode, memory[off:off+sz])
			}
			newAddr := big.NewInt(int64(len(state) + 123456))
			newAddrStr := fmt.Sprintf("0x%040x", newAddr)
			innerTx := Tx{
				To:       newAddr,
				From:     tx.To,
				Origin:   tx.Origin,
				GasPrice: tx.GasPrice,
				Value:    value,
				Data:     nil,
			}
			if innerTx.From == nil {
				innerTx.From = big.NewInt(0)
			}
			_, innerSuccess, innerLogs, deployedCode := Evm(initCode, innerTx, block, state)
			lastReturnData = deployedCode
			if innerSuccess {
				state[newAddrStr] = Account{
					Balance: value,
					Code:    deployedCode,
					Storage: make(map[string]*big.Int),
				}
				stack = push(stack, newAddr)
				logs = append(logs, innerLogs...)
			} else {
				stack = push(stack, big.NewInt(0))
			}

		case 0xff: // SELFDESTRUCT
			if tx.IsStatic {
				return stack, false, logs, nil
			}
			if len(stack) < 1 {
				return stack, false, logs, nil
			}
			var target *big.Int
			stack, target = pop(stack)
			targetStr := fmt.Sprintf("0x%040x", target)
			targetAccount := state[targetStr]
			var myAddrStr string
			if tx.To != nil {
				myAddrStr = fmt.Sprintf("0x%040x", tx.To)
			} else {
				myAddrStr = fmt.Sprintf("0x%040x", 0)
			}
			myAccount := state[myAddrStr]
			if targetAccount.Balance == nil {
				targetAccount.Balance = big.NewInt(0)
			}
			if myAccount.Balance != nil {
				targetAccount.Balance = new(big.Int).Add(targetAccount.Balance, myAccount.Balance)
			}
			state[targetStr] = targetAccount
			delete(state, myAddrStr)
			return stack, true, logs, nil

		default:
			return stack, false, logs, nil	
		}
		_ = op // delete this; it's only here to make the compiler think you're already using `op`
	}
	return stack, true, logs, nil
}


