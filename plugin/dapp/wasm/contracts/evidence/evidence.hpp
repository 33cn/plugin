#ifdef __cplusplus  //而这一部分就是告诉编译器，如果定义了__cplusplus(即如果是cpp文件，
extern "C" {        //因为cpp文件默认定义了该宏),则采用C语言方式进行编译
#endif

/**合约开发
新建 cpp 和 hpp 文件，并导入 common.h 头文件，其中 common.h 中声明了 chain33 中的回调函数，是合约调用 chain33 系统方法的接口。
合约中的导出方法的所有参数都只能是数字类型，且必须有一个数字类型的返回值，其中非负值表示执行成功，负值表示执行失败。 */
//1.调用上链数据操作接口
int AddStateTx();
int ModStateTx();
int DelStateTx();
//2.调用本地数据操作接口
int AddLocalTx();
int ModLocalTx();
int DelLocalTx();

#ifdef __cplusplus
}
#endif
