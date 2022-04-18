                              在阅读go-ethereum的evm部分代码基础上进行实现

![image](https://user-images.githubusercontent.com/45748258/149899808-540eef7a-56a5-4726-b16c-65f9d2d43a96.png)

源码分析：



首先 evm 模块的核心对象是 `EVM`，它代表了一个以太坊虚拟机，用于创建或调用某个合约。每次处理一个交易对象时，都会创建一个 `EVM` 对象。

`EVM` 对象内部主要依赖三个对象：解释器 `Interpreter`、虚拟机相关配置对象 `vm.Config`、以太坊状态数据库 `StateDB`。

`StateDB` 主要的功能就是用来提供数据的永久存储和查询。

`Interpreter` 是一个接口，在代码中由 `EVMInterpreter` 实现具体功能。这是一个解释器对象，循环解释执行给定的合约指令，直接遇到退出指令。每执行一次指令前，都会做一些检查操作，确保 gas、栈空间等充足。但各指令真正的解释执行代码却不在这个对象中，而是记录在 `vm.Config` 的 `JumpTable` 字段中。

`vm.Config` 为虚拟机和解释器提供了配置信息，其中最重要的就是 `JumpTable`。`JumpTable` 是 `vm.Config` 的一个字段，它是一个由 256 个 `operation` 对象组成的数组。解释器每拿到一个准备执行的新指令时，就会从 `JumpTable` 中获取指令相关的信息，即 `operation` 对象。这个对象中包含了解释执行此条指令的函数、计算指令的 gas 消耗的函数等。

在代码中，根据以太坊的版本不同，`JumpTable` 可能指向四个不同的对象：`constantinopleInstructionSet`、`byzantiumInstructionSet`、`homesteadInstructionSet`、`frontierInstructionSet`。这四套指令集多数指令是相同的，只是随着版本的更新，新版本比旧版本支持更多的指令集。







interpreter.go （68）的NewEVMInterpreter函数 ：创建一个EVMInterpreter实例

interpreter.go （116）的Run函数： 循环解释执行合约代码的input data

这是执行的主要过程代码，接收三个参数，contract对象， input（因为是一个合约调用的交易，所以这里为nil）readonly 标志位，这个标志位只有在staticcall的时候是true的，其他的调用都是false;

1. 判断解释器的intPool为nil，则先创建一个intPool,然后将新创建的intPool放入poolOfIntPools

2. evm 调用堆栈+1

3. 解释器循环执行指定，当遇到STOP RETURN 或者合约自毁 或者 当执行过程中遇到错误的时候， 直到所有指令被执行结束，循环才会停止。

   3.1 循环首先获取当前pc 计数器， 然后从jumptable里面拿出要执行的opcode, opcode是以太坊虚拟机指令，一共不超过256个，正好一个byte大小能装下。下面展示了一个 sha3的opteration

   ![img](http://ww1.sinaimg.cn/large/c26c1fe3gy1g26y17xrepj20f804adfv.jpg)

 execute表示指令对应的执行方法
​ gasCost表示执行这个指令需要消耗的gas
​ validateStack计算是不是解析器栈溢出
​ memorySize用于计算operation的占用内存大小

3.2 根据不同的指令，指令的memorysize等，调用operation.gasCost()方法计算执行operation指令需要消耗的gas。

3.3 调用operation.execute(&pc, in.evm, contract, mem, stack)执行指令对应的方法。

3.4 operation.reverts值是true或者operation.halts值是true的指令，会跳出主循环，否则继续遍历下个op。

3.5 operation指令集里面有4个特殊的指令LOG0，LOG1，LOG2，LOG3，它们的指令执行方法makeLog()会产生日志数据，日志内容包括EVM解析栈内容，指令内存数据，区块信息，合约信息等。这些日志数据会写入到tx的Receipt的logs里面，并存入本地levleldb数据库。



evm.go 

**create**方法   创建合约对象  用于创建合约的交易

1.1 判断evm执行栈深度不能超过1024，

1.2 发送方持有的以太坊数量大于此次合约交易金额。

1.3 获取合约调用者账户nonce，然后将nonce+1存入stateDB

1.4 根据合约地址获取合约hash值

1.5 记录一个状态快照，用来失败回滚。

1.6 为这个合约地址创建一个合约账户，并为这个合约账户设置nonce值为1

1.5 产生以太坊资产转移，发送方地址账户金额减value值，合约账户的金额加value值。

1.6 根据发送方地址和合约地址，以及金额value 值和gas，合约代码和代码hash值，创建一个合约对象

1.7 设置合约对象的bytecode, bytecodehash和 合约地址

1.8 run方法来执行合约，内部调用evm的解析器来执行合约指令，如果是预编译好的合约，则预编译执行合约就行



**run**函数 

**run**方法来执行合约，内部调用evm的解析器来执行合约指令，如果是预编译好的合约，则执行预编译合约就行。



四个合约调用方法 分别为 Call, CallCode,DelegateCall,StaticCall  

**NewEVM**方法 主要是根据当前的区块号以及相关配置，设置EVM的解释器.



core/types/transaction.go`中定义了交易的数据结构



core/state_processor.go  

**Process()**方法对block中的每个交易`tx`调用**ApplyTransaction()**来执行交易，入参**state**存储了各个账户的信息，如账户余额、合约代码(仅对合约账户而言)，将其理解为一个内存中的数据库。其中每个账户以**state object**表示。



**ApplyTransaction()**方法完成以下功能

- 调用**AsMessage()**用tx生成`core.Message` 其实现就是将`tx`中的一些字段存入`Message`以及从`tx`的数字签名中反解出`tx`的**sender**，重点关注其中的**data** 字段：对普通转账交易，该字段为空，对创建一个新的合约，该字段为新的合约的**代码**，对执行一个已经在区块链上存在的合约，该参数为合约代码的**输入参数**
- 调用**NewEVMContext()**创建一个EVM运行上下文`vm.Context`。注意其中的**Coinbase**字段填入的矿工的地址，`Transfer`是具体的转账方法，其实就是操作**sender**和**recipient**的账户余额
- 调用**NewEVM()**创建一个虚拟机运行环境`EVM`，它主要作用是汇集之前的信息以及创建一个代码解释器(Interpreter)，这个解释器之后会用来解释并执行合约代码
- 接下来就是调用**ApplyMessage()**将以上的信息**作用**在当前以太坊状态上，使得状态机发生状态变换



**ApplyMessage() **方法通过给定的message计算新的DB状态，继而改变旧的DB状态并返回EVM执行的返回结果和gas使用情况。

```go
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) ([]byte, uint64, bool, error) {
   return NewStateTransition(evm, msg, gp).TransitionDb()
}
```

**TransitionDb() **方法则是主要负责执行交易，影响DB状态。

1. preCheck 函数主要进行执行交易前的检查，目前包含下面两个步骤

   1.1 检查msg 里面的nonce值与db里面存储的账户的nonce值是否一致。

   1.2 buyGas方法主要是判断交易账户是否可以支付足够的gas执行交易，如果可以支付，则设置stateTransaction 的gas值 和 initialGas 值。并且从交易执行账户扣除相应的gas值。

   2.Intrinsic函数计算固定的gas消耗，之后进行支付。

   3.通过判断contractCreation来决定交易是创建合约交易还是执行合约交易

```go
if contractCreation {
   ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
} else {
   // Increment the nonce for the next transaction
   st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
   ret, st.gas, vmerr = evm.Call(sender, st.to(), st.data, st.gas, st.value)
}
```

   

  contracts.go 包含了 预编译合约 如ecrecover、sha256hash等

   gas_table.go  返回了各种指令消耗的gas的函数

   instructions.go 包含了指令的操作 如opAdd、opSub等等

   interface.go包含了StateDB的接口以及CallContext的接口

- core/vm/intpool.go` 常量池
- `core/vm/jump_table.go` 指令跳转表
- `core/vm/logger.go` 状态日志
- `core/vm/logger_json.go` json形式日志
- `core/vm/memory.go` evm 可操作内存
- `core/vm/memory_table.go` evm内存操作表，衡量一些操作耗费内存大小
- `core/vm/opcodes.go` 定义操作码的名称和编号
- `core/vm/stacks.go` evm栈操作
- `core/vm/stack_table.go` evm栈验证函数



下一步：

参考上述代码，主要先实现交易的执行、合约的创建、合约的执行等核心功能

​									

​																																																	蔡栋梁  4.7

主要是交易的执行的测试过程 我理解应该是提供接口来执行交易，依次通过上面交易流程的方法，最后的**TransitionDb() **方法则是主要负责执行交易，影响DB状态。通过判断contractCreation来决定交易是创建合约交易还是执行合约交易，再调用方法来create或call(call会调用run)。

