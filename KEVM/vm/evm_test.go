package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"fmt"
	//"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

func main(){
	var (
		env            = NewEVM(Context{}, nil, params.TestChainConfig, Config{})
		//stack          = newstack()
		//pc             = uint64(0)
		//evmInterpreter = NewEVMInterpreter(env, env.vmConfig)
	)
	var code []byte
	var gas uint64
	var value *big.Int
	var address common.Address
	caller := vm.AccountRef(address)
	//address := []byte("0x5B38Da6a701c568545dCfcB03FcB875f56beddC4")
	//caller = &AccountRef{address}
	ret, contractAddr, leftOverGas, err := env.Create(caller, code, gas, value)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println(ret)
	fmt.Println(contractAddr)
	fmt.Println(leftOverGas)
	var c *Contract
	_, err = run(env, c, ret, false)
}
