package executor

// 将手动生成的local db 的代码和用table 生成的local db的代码分离出来
// 手动生成的local db, 将不生成任意资产标价的数据， 保留用coins 生成交易的数据， 来兼容为升级的app 应用
// 希望有全量数据的， 需要调用新的rpc