package executor

import (
	"fmt"
	"strconv"

	"github.com/perlin-network/life/exec"
)

// Resolver defines imports for WebAssembly modules ran in Life.
type Resolver struct{}

// ResolveFunc defines a set of import functions that may be called within a WebAssembly module.
func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	switch module {
	case "env":
		switch field {
		case "setStateDB":
			return func(vm *exec.VirtualMachine) int64 {
				keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				keyLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				key := vm.Memory[keyPtr : keyPtr+keyLen]
				valuePtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				valueLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				value := make([]byte, valueLen)
				copy(value, vm.Memory[valuePtr:valuePtr+valueLen])
				setStateDB(key, value)
				return 0
			}

		case "getStateDBSize":
			return func(vm *exec.VirtualMachine) int64 {
				keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				keyLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				key := vm.Memory[keyPtr : keyPtr+keyLen]
				return int64(getStateDBSize(key))
			}

		case "getStateDB":
			return func(vm *exec.VirtualMachine) int64 {
				keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				keyLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				key := vm.Memory[keyPtr : keyPtr+keyLen]
				valuePtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				valueLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				value, err := getStateDB(key)
				if err != nil {
					for i := 0; i < valueLen; i++ {
						vm.Memory[valuePtr+i] = 0
					}
					return 0
				}
				if valueLen != len(value) {
					return 0
				}
				copy(vm.Memory[valuePtr:valuePtr+valueLen], value)
				return int64(valueLen)
			}

		case "setLocalDB":
			return func(vm *exec.VirtualMachine) int64 {
				keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				keyLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				key := vm.Memory[keyPtr : keyPtr+keyLen]
				valuePtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				valueLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				value := make([]byte, valueLen)
				copy(value, vm.Memory[valuePtr:valuePtr+valueLen])
				setLocalDB(key, value)
				return 0
			}

		case "getLocalDBSize":
			return func(vm *exec.VirtualMachine) int64 {
				keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				keyLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				key := vm.Memory[keyPtr : keyPtr+keyLen]
				return int64(getLocalDBSize(key))
			}

		case "getLocalDB":
			return func(vm *exec.VirtualMachine) int64 {
				keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				keyLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				key := vm.Memory[keyPtr : keyPtr+keyLen]
				valuePtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				valueLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				value, err := getLocalDB(key)
				if err != nil {
					copy(vm.Memory[valuePtr:valuePtr+valueLen], make([]byte, valueLen))
				}
				if valueLen != len(value) {
					return 0
				}
				for i, c := range value {
					vm.Memory[valuePtr+i] = c
				}
				return int64(valueLen)
			}

		case "getBalance":
			return func(vm *exec.VirtualMachine) int64 {
				addrPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				addrLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				addr := string(vm.Memory[addrPtr : addrPtr+addrLen])
				execPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				execLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				exec := string(vm.Memory[execPtr : execPtr+execLen])
				balance, _, err := getBalance(addr, exec)
				if err != nil {
					return -1
				}
				return balance
			}

		case "getFrozen":
			return func(vm *exec.VirtualMachine) int64 {
				addrPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				addrLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				addr := string(vm.Memory[addrPtr : addrPtr+addrLen])
				execPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				execLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				exec := string(vm.Memory[execPtr : execPtr+execLen])
				_, frozen, err := getBalance(addr, exec)
				if err != nil {
					return -1
				}
				return frozen
			}

		case "transfer":
			return func(vm *exec.VirtualMachine) int64 {
				fromPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				fromLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				fromAddr := string(vm.Memory[fromPtr : fromPtr+fromLen])
				toPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				toLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				toAddr := string(vm.Memory[toPtr : toPtr+toLen])
				amount := vm.GetCurrentFrame().Locals[4]
				err := transfer(fromAddr, toAddr, amount)
				if err != nil {
					return -1
				}
				return 0
			}

		case "transferToExec":
			return func(vm *exec.VirtualMachine) int64 {
				fromPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				fromLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				fromAddr := string(vm.Memory[fromPtr : fromPtr+fromLen])
				toPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				toLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				toAddr := string(vm.Memory[toPtr : toPtr+toLen])
				amount := vm.GetCurrentFrame().Locals[4]
				err := transferToExec(fromAddr, toAddr, amount)
				if err != nil {
					return -1
				}
				return 0
			}

		case "transferWithdraw":
			return func(vm *exec.VirtualMachine) int64 {
				fromPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				fromLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				fromAddr := string(vm.Memory[fromPtr : fromPtr+fromLen])
				toPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				toLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				toAddr := string(vm.Memory[toPtr : toPtr+toLen])
				amount := vm.GetCurrentFrame().Locals[4]
				err := transferWithdraw(fromAddr, toAddr, amount)
				if err != nil {
					return -1
				}
				return 0
			}

		case "execAddress":
			return func(vm *exec.VirtualMachine) int64 {
				namePtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				nameLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				name := string(vm.Memory[namePtr : namePtr+nameLen])
				addr := []byte(execAddress(name))
				addrPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				addrLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				copy(vm.Memory[addrPtr:addrPtr+addrLen], addr)
				return 0
			}

		case "execFrozen":
			return func(vm *exec.VirtualMachine) int64 {
				addrPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				addrLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				addr := string(vm.Memory[addrPtr : addrPtr+addrLen])
				amount := vm.GetCurrentFrame().Locals[2]
				err := execFrozen(addr, amount)
				if err != nil {
					return -1
				}
				return 0
			}

		case "execActive":
			return func(vm *exec.VirtualMachine) int64 {
				addrPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				addrLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				addr := string(vm.Memory[addrPtr : addrPtr+addrLen])
				amount := vm.GetCurrentFrame().Locals[2]
				err := execActive(addr, amount)
				if err != nil {
					return -1
				}
				return 0
			}

		case "execTransfer":
			return func(vm *exec.VirtualMachine) int64 {
				fromPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				fromLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				fromAddr := string(vm.Memory[fromPtr : fromPtr+fromLen])
				toPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				toLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				toAddr := string(vm.Memory[toPtr : toPtr+toLen])
				amount := vm.GetCurrentFrame().Locals[4]
				err := execTransfer(fromAddr, toAddr, amount)
				if err != nil {
					return -1
				}
				return 0
			}

		case "execTransferFrozen":
			return func(vm *exec.VirtualMachine) int64 {
				fromPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				fromLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				fromAddr := string(vm.Memory[fromPtr : fromPtr+fromLen])
				toPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				toLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				toAddr := string(vm.Memory[toPtr : toPtr+toLen])
				amount := vm.GetCurrentFrame().Locals[4]
				err := execTransferFrozen(fromAddr, toAddr, amount)
				if err != nil {
					return -1
				}
				return 0
			}

		case "getFrom":
			return func(vm *exec.VirtualMachine) int64 {
				fromAddr := []byte(getFrom())
				fromPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				fromLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				copy(vm.Memory[fromPtr:fromPtr+fromLen], fromAddr)
				return 0
			}

		case "getHeight":
			return func(vm *exec.VirtualMachine) int64 { return getHeight() }

		case "getRandom":
			return func(vm *exec.VirtualMachine) int64 { return getRandom() }

		case "printlog":
			return func(vm *exec.VirtualMachine) int64 {
				logPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				logLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				logInfo := string(vm.Memory[logPtr : logPtr+logLen])
				printlog(logInfo)
				return 0
			}

		case "printint":
			return func(vm *exec.VirtualMachine) int64 {
				n := vm.GetCurrentFrame().Locals[0]
				printlog(strconv.FormatInt(n, 10))
				return 0
			}

		case "sha256":
			return func(vm *exec.VirtualMachine) int64 {
				dataPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				dataLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				data := vm.Memory[dataPtr : dataPtr+dataLen]
				sumPtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				sumLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				copy(vm.Memory[sumPtr:sumPtr+sumLen], sha256(data))
				return 0
			}
		case "getENVSize":
			return func(vm *exec.VirtualMachine) int64 {
				n := vm.GetCurrentFrame().Locals[0]
				return int64(getENVSize(int(n)))
			}

		case "getENV":
			return func(vm *exec.VirtualMachine) int64 {
				n := vm.GetCurrentFrame().Locals[0]
				valuePtr := int(uint32(vm.GetCurrentFrame().Locals[1]))
				valueLen := int(uint32(vm.GetCurrentFrame().Locals[2]))
				value := getENV(int(n))
				copy(vm.Memory[valuePtr:valuePtr+valueLen], value)
				return int64(len(value))
			}
		case "totalENV":
			return func(vm *exec.VirtualMachine) int64 {
				return int64(totalENV())
			}

		default:
			log.Error("ResolveFunc", "unknown field", field)
		}

	default:
		log.Error("ResolveFunc", "unknown module", module)
	}
	return nil
}

// ResolveGlobal defines a set of global variables for use within a WebAssembly module.
func (r *Resolver) ResolveGlobal(module, field string) int64 {
	fmt.Printf("Resolve global: %s %s\n", module, field)
	switch module {
	case "env":
		switch field {
		case "__life_magic":
			return 424
		default:
			log.Error("ResolveGlobal", "unknown field", field)
		}
	default:
		log.Error("ResolveGlobal", "unknown module", module)
	}

	return 0
}
