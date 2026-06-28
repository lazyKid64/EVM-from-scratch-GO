package evm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type testCase struct {
	Name  string     `json:"name"`
	Hint  string     `json:"hint"`
	Tx    *txJSON    `json:"tx"`
	Block *blockJSON `json:"block"`
	State map[string]*accountJSON `json:"state"`
	Code  code       `json:"code"`
	Want  want       `json:"expect"`
}

type accountJSON struct {
	Balance *hexBigInt            `json:"balance"`
	Code    *code                 `json:"code"`
	Storage map[string]*hexBigInt `json:"storage"`
}

type blockJSON struct {
	Basefee    *hexBigInt `json:"basefee"`
	Coinbase   *hexBigInt `json:"coinbase"`
	Timestamp  *hexBigInt `json:"timestamp"`
	Number     *hexBigInt `json:"number"`
	Difficulty *hexBigInt `json:"difficulty"`
	GasLimit   *hexBigInt `json:"gaslimit"`
	ChainId    *hexBigInt `json:"chainid"`
}

type txJSON struct {
	To       *hexBigInt `json:"to"`
	From     *hexBigInt `json:"from"`
	Origin   *hexBigInt `json:"origin"`
	GasPrice *hexBigInt `json:"gasprice"`
	Value    *hexBigInt `json:"value"`
	Data     string     `json:"data"`
}

type code struct {
	Bin string `json:"bin"`
	Asm string `json:"asm"`
}

type want struct {
	Stack   []hexBigInt `json:"stack"`
	Logs    []logJSON   `json:"logs"`
	Success bool        `json:"success"`
	Return  string      `json:"return"`
}

type logJSON struct {
	Address *hexBigInt   `json:"address"`
	Data    string       `json:"data"`
	Topics  []*hexBigInt `json:"topics"`
}

type stringLog struct {
	Address string
	Data    string
	Topics  []string
}

// A hexBigInt is a *big.Int that can be read from a JSON hex string.
type hexBigInt struct {
	*big.Int
}

// UnmarshalJSON unmarshals the buffer into i.Int; it expects the input to be
// string-quoted.
func (i *hexBigInt) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if i.Int == nil {
		i.Int = new(big.Int)
	}
	return i.Int.UnmarshalJSON([]byte(s))
}

// StackInts returns the underlying *big.Int values of w.Stack, unwrapping them
// from within the JSON-unmarshalling helper.
func (w *want) StackInts() []*big.Int {
	b := make([]*big.Int, len(w.Stack))
	for i, s := range w.Stack {
		b[i] = s.Int
	}
	return b
}

// toHexStrings converts an array of *big.Ints into hex-formatted strings to match
// the format of numbers used in evm.json
func toHexStrings(ints []*big.Int) []string {
	b := make([]string, len(ints))
	for i, s := range ints {
		b[i] = "0x" + s.Text(16)
	}
	return b
}

func TestEVM(t *testing.T) {
	var tests []testCase
	t.Run("setup", func(t *testing.T) {
		const testSrc = "../evm.json"
		f, err := os.Open(testSrc)
		if err != nil {
			fatalAndBugReport(t, "os.Open(%q) error %v", testSrc, err)
		}
		defer f.Close()

		if err := json.NewDecoder(f).Decode(&tests); err != nil {
			fatalAndBugReport(t, "json.NewDecoder(%q).Decode(%T) error %v", testSrc, &tests, err)
		}
	})
	if t.Failed() {
		return
	}

	for i, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			bin, err := hex.DecodeString(tt.Code.Bin)
			if err != nil {
				fatalAndBugReport(t, "hex.DecodeString(%q) error %v", tt.Code.Bin, err)
			}

			var tx Tx
			if tt.Tx != nil {
				if tt.Tx.To != nil && tt.Tx.To.Int != nil { tx.To = tt.Tx.To.Int }
				if tt.Tx.From != nil && tt.Tx.From.Int != nil { tx.From = tt.Tx.From.Int }
				if tt.Tx.Origin != nil && tt.Tx.Origin.Int != nil { tx.Origin = tt.Tx.Origin.Int }
				if tt.Tx.GasPrice != nil && tt.Tx.GasPrice.Int != nil { tx.GasPrice = tt.Tx.GasPrice.Int }
				if tt.Tx.Value != nil && tt.Tx.Value.Int != nil { tx.Value = tt.Tx.Value.Int }
				if tt.Tx.Data != "" { tx.Data, _ = hex.DecodeString(tt.Tx.Data) }
			}

			var block Block
			if tt.Block != nil {
				if tt.Block.Basefee != nil && tt.Block.Basefee.Int != nil { block.Basefee = tt.Block.Basefee.Int }
				if tt.Block.Coinbase != nil && tt.Block.Coinbase.Int != nil { block.Coinbase = tt.Block.Coinbase.Int }
				if tt.Block.Timestamp != nil && tt.Block.Timestamp.Int != nil { block.Timestamp = tt.Block.Timestamp.Int }
				if tt.Block.Number != nil && tt.Block.Number.Int != nil { block.Number = tt.Block.Number.Int }
				if tt.Block.Difficulty != nil && tt.Block.Difficulty.Int != nil { block.Difficulty = tt.Block.Difficulty.Int }
				if tt.Block.GasLimit != nil && tt.Block.GasLimit.Int != nil { block.GasLimit = tt.Block.GasLimit.Int }
				if tt.Block.ChainId != nil && tt.Block.ChainId.Int != nil { block.ChainId = tt.Block.ChainId.Int }
			}

			state := make(State)
			if tt.State != nil {
				for addr, acc := range tt.State {
					var account Account
					if acc != nil {
						if acc.Balance != nil && acc.Balance.Int != nil {
							account.Balance = acc.Balance.Int
						}
						if acc.Code != nil && acc.Code.Bin != "" {
							account.Code, _ = hex.DecodeString(acc.Code.Bin)
						}
						if acc.Storage != nil {
							account.Storage = make(map[string]*big.Int)
							for k, v := range acc.Storage {
								keyBig := new(big.Int)
								keyBig.SetString(k[2:], 16)
								keyStr := fmt.Sprintf("0x%064x", keyBig)
								if v != nil && v.Int != nil {
									account.Storage[keyStr] = v.Int
								}
							}
						}
					}
					state[addr] = account
				}
			}

			got, gotSuccess, gotLogs, gotReturn := Evm(bin, tx, block, state)
			if gotSuccess != tt.Want.Success {
				t.Errorf("Evm(…) got success = %t; want %t", gotSuccess, tt.Want.Success)
			}
			if diff := cmp.Diff(toHexStrings(tt.Want.StackInts()), toHexStrings(got), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Evm(…) stack mismatch; diff (-want +got)\n%s", diff)
			}

			// Format expected logs
			var wantStrLogs []stringLog
			if tt.Want.Logs != nil {
				for _, wl := range tt.Want.Logs {
					var addr string
					if wl.Address != nil && wl.Address.Int != nil {
						addr = fmt.Sprintf("0x%040x", wl.Address.Int)
					}
					var topics []string
					for _, tpc := range wl.Topics {
						if tpc != nil && tpc.Int != nil {
							topics = append(topics, fmt.Sprintf("0x%064x", tpc.Int))
						}
					}
					wantStrLogs = append(wantStrLogs, stringLog{
						Address: addr,
						Data:    wl.Data,
						Topics:  topics,
					})
				}
			}

			// Format actual logs
			var gotStrLogs []stringLog
			for _, gl := range gotLogs {
				var addr string
				if gl.Address != nil {
					addr = fmt.Sprintf("0x%040x", gl.Address)
				}
				var topics []string
				for _, tpc := range gl.Topics {
					if tpc != nil {
						topics = append(topics, fmt.Sprintf("0x%064x", tpc))
					}
				}
				gotStrLogs = append(gotStrLogs, stringLog{
					Address: addr,
					Data:    hex.EncodeToString(gl.Data),
					Topics:  topics,
				})
			}

			if diff := cmp.Diff(wantStrLogs, gotStrLogs, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Evm(…) logs mismatch; diff (-want +got)\n%s", diff)
			}

			gotReturnHex := hex.EncodeToString(gotReturn)
			if gotReturnHex != tt.Want.Return {
				t.Errorf("Evm(…) return data mismatch; got %s, want %s", gotReturnHex, tt.Want.Return)
			}

			if t.Failed() {
				t.Logf("✕  %v", tt.Name)
				t.Logf("EVM Instructions:\n%v", tt.Code.Asm)
				if tt.Hint != "" {
					t.Log("#####")
					t.Logf("##### HINT: %s", tt.Hint)
					t.Log("#####")
				}
			}
		})
		if t.Failed() {
			t.Fatalf("Progress: %d/%d", i, len(tests))
		} else {
			t.Logf("✓  %v", tt.Name)
		}
	}
}

// fatalAndBugReport calls t.Errorf(format, a...) and then t.Fatal() with a
// message requesting that the student files a bug report. It's intended use is
// as a replacement for t.Fatal() when the error is in the test setup, not in
// the student's implementation.
func fatalAndBugReport(t *testing.T, format string, a ...interface{}) {
	t.Helper()
	t.Errorf(format, a...)
	t.Fatal("The error wasn't in your code. Please file a bug report at https://github.com/w1nt3r-eth/evm-from-scratch/issues/new")
}
