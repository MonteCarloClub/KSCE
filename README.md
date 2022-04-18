# KSCE
Smart contract Engine



4.18

智能合约通过交易来执行，主要是交易的执行的测试过程 我理解应该是通过Process来执行交易，依次通过上面交易流程的方法，最后的**TransitionDb() **方法则是主要负责执行交易，影响DB状态。通过判断contractCreation来决定交易是创建合约交易还是执行合约交易，再调用方法来create或call(call会调用run)。

但是Process(block *types.Block, statedb *state.StateDB, cfg vm.Config)与stateDB， Block有关，跟其他部分联系比较紧密，主要就是要考虑跟其他部分对接。

