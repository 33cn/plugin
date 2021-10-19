#pragma once

#include "TassType4PCIeSM.h"

#ifdef __cplusplus
extern "C" {
#endif

#define TASS_DEV_ADMIN_NAME			"admin"		//设备管理员名称
#define TASS_DEV_ID_STR_SIZE		32			//设备ID长度
#define TASS_NAME_SIZE_MAX			8			//管理员名称最大长度
#define TASS_UKEY_SN_SIZE			5			//UKey序列号长度
#define TASS_KEY_LABEL_SIZE_MAX		128			//密钥标签最大长度
#define TASS_KEY_ATTR_SIZE_MAX		1024		//密钥属性最大长度
#define TASS_SK_SIZE_MAX			4096		//私钥最大长度
#define TASS_PK_SIZE_MAX			1024		//公钥最大长度
#define TASS_KEY_SIZE_MAX			64			//对称密钥最大长度
#define TASS_KCV_SIZE				16			//密钥校验值长度
#define TASS_DEV_ADMIN_CNT_MAX		1			//设备管理员数量
#define TASS_KEY_ADMIN_CNT_MAX		4			//密钥管理员最大数量

#define TASS_ECC_INDEX_MAX		64
#define TASS_SYMM_INDEX_MAX		64

	typedef enum {
		TA_DEV_ADMIN = 0,	//设备管理员
		TA_KEY_ADMIN = 1,	//密钥管理员
	}TassAdminType;

	typedef struct {
		TassAdminType type;
		char name[TASS_NAME_SIZE_MAX + 1];	//名称
		char ukeySn[TASS_UKEY_SN_SIZE + 1];	//UKey序列号，制作UKey后有效
	}TassAdminInfo;

	typedef struct {
		TassAlg alg;										//算法
		int index;											//索引
		char label[TASS_KEY_LABEL_SIZE_MAX + 1];			//标签
		unsigned char sk_key[TASS_SK_SIZE_MAX];				//私钥/对称密钥密文
		unsigned int sk_keyLen;								//私钥/对称密钥密文长度
		unsigned char sk_keyAttr[TASS_KEY_ATTR_SIZE_MAX];	//私钥/对称密钥属性
		unsigned int sk_keyAttrLen;							//私钥/对称密钥属性长度
		unsigned char pk_kcv[TASS_PK_SIZE_MAX];				//公钥/对称密钥校验值
		unsigned int pk_kcvLen;								//公钥/对称密钥校验值长度
		unsigned char pkAttr[TASS_KEY_ATTR_SIZE_MAX];		//公钥属性，非对称时有效
		unsigned int pkAttrLen;								//公钥属性长度
	}TassKeyInfo;

	/**
	 * @brief 搜索设备
	 *
	 * @param	id		[in]		设备ID缓冲区，传NULL时通过len返回需要的缓冲区大小
	 * @param	idLen	[in|out]	输入时标识id缓冲区大小
	 *								输出时标识id实际长度
	 *								多个设备ID间以‘\0’分隔，最后以两个'\0'结尾
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlScanDevice(char* id, unsigned int* idLen);

	/**
	 * @brief 打开设备
	 *
	 * @param	id			[in]	要打开的设备ID，通过TassScan获取
	 * @param	phDevice	[out]	设备句柄
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlOpenDevice(const char id[TASS_DEV_ID_STR_SIZE + 1], void** phDevice);

	/**
	 * @brief 关闭设备
	 *
	 * @param	hDevice		[in]	已打开的设备句柄
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlCloseDevice(void* hDevice);

	/**
	* @brief 管理员管理功能
	*
	* @func TassListAdmin
	* @func TassLogin
	* @func TassLogout
	* @func TassMakeAdminUKey
	* @func TassUpdatePWD
	* @func TassAddAdmin
	* @func TassRemoveAdmin
	*
	*/

	/**
	 * @brief 获取用户列表
	 * @param	hDevice		[in]		已打开的设备句柄
	 * @param	info		[in]		管理员信息，传NULL时通过len返回需要的缓冲区大小
	 * @param	infoLen		[in|out]	输入时标识info缓冲区大小
	 *									输出时标识info实际长度
	 *									用户数量可通过计算“len / sizeof(TassAdminInfo)”获取
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 打开设备后必须首先执行登录接口，才能执行其他操作
	 */
	int TassCtlListAdmin(void* hDevice,
		TassAdminInfo* info, unsigned int* infoLen);

	/**
	 * @brief 申请口令哈希加密公钥
	 *
	 * @param	hDevice		[in]	已打开的设备句柄
	 * @param	pk			[out]	用于加密口令哈希的公钥
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 打开设备后必须首先执行登录接口，才能执行其他操作
	 */
	int TassCtlRequestPwdHashEncPublicKey(void* hDevice, unsigned char pk[64]);

	/**
	 * @brief 通过哈希加密公钥加密口令hash值得到哈希值密文
	 *
	 * @param	hDevice			[in]	已打开的设备句柄
	 * @param	pk				[in]	用于加密口令哈希的公钥
	 * @param	pwd				[in]	用户登录口令
	 * @param	pwdHashCipher	[out]	管理员口令的哈希值密文
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlEncryptPwdHash(void* hDevice, const unsigned char pk[64], const char* pwd, unsigned char pwdHashCipher[128]);

	/**
	 * @brief 登录
	 *
	 * @param	hDevice			[in]	已打开的设备句柄
	 * @param	name			[in]	管理员名称，设备管理员固定为“admin”
	 * @param	pwdHashCipher	[in]	管理员口令的哈希值密文
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 打开设备后必须首先执行登录接口，才能执行其他操作
	 */
	int TassCtlLogin(void* hDevice,
		const char* name,
		const unsigned char pwdHashCipher[128]);

	/**
	 * @brief 登出
	 *
	 * @param	hDevice		[in]	已打开的设备句柄
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 */
	int TassCtlLogout(void* hDevice);

	/**
	 * @brief 修改口令
	 *
	 * @param	hDevice				[in]	已打开的设备句柄
	 * @param	name				[in]	已登录的用户名
	 * @param	oldPwdHashCipher	[in]	旧管理员口令的哈希值密文
	 * @param	newPwdHashCipher	[in]	新管理员口令的哈希值密文
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlUpdatePwdHash(void* hDevice,
		const char* name,
		const unsigned char oldPwdHashCipher[128],
		const unsigned char newPwdHashCipher[128]);

	/**
	 * @brief 绑定UKey（暂未启用）
	 *
	 * @param	hDevice			[in]	已打开的设备句柄
	 * @param	name			[in]	已登录的用户名称
	 * @param	pwdHashCipher	[in]	管理员口令的哈希值密文
	 * @param	ukeyId			[in]	UKey序号
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 制作UKey后，登录时必须插入该UKey
	 */
	int TassCtlBindUKey(void* hDevice,
		const char* name,
		const char pwdHashCipher[128],
		const char* ukeyId);

	/**
	 * @brief 添加密钥管理员（暂未启用）
	 *
	 * @param	hDevice				[in]	已打开的设备句柄
	 * @param	devPwdHashCipher	[in]	设备管理员口令的哈希值密文
	 * @param	name				[in]	增加的密钥管理员名称，不能与已存在管理员名称相同
	 * @param	pwdHashCipher		[in]	增加的密钥管理员口令的哈希值密文
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录设备管理员
	 *
	 */
	int TassCtlAddKeyAdmin(void* hDevice,
		const unsigned char devPwdHashCipher[128],
		const char* name,
		const unsigned char pwdHashCipher[128]);

	/**
	 * @brief 删除密钥管理员（暂未启用）
	 *
	 * @param	hDevice				[in]	已打开的设备句柄
	 * @param	devPwdHashCipher	[in]	设备管理员口令的哈希值密文
	 * @param	name				[in]	删除的密钥管理员名称
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录设备管理员
	 *
	 */
	int TassCtlRemoveKeyAdmin(void* hDevice,
		const unsigned char devPwdHashCipher[128],
		const char* name);

	/**
	* @brief 设备管理功能
	*
	* @func TassCtlDeviceFormat
	* @func TassCtlDeviceInit
	* @func TassCtlDeviceInitRestoreFactory
	* @func TassCtlBootAuth
	* @func TassGetInfo
	* @func TassSetInfo
	* @func TassSelfCheck
	* @func TassRestoreFactory
	* @func TassInitialize
	* @func TassBootAuth
	* @func TassExportDevEncKeyPair
	* @func TassExportDevKEK
	* @func TassImportDevKEK
	*
	*/

	/**
	 * @brief 设备格式化
	 *			清除设备中的全部数据，包括设备密钥FLASH数据等
	 *
	 * @param	hDevice			[in]	已打开的设备句柄
	  * @param	pwdHashCipher	[in]	设备管理员口令的哈希值密文
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note
	 *	1. 需要设备管理员登录
	 *	2. 执行成功后会清除登录状态，需要使用默认设备管理员口令登录
	 */
	int TassCtlDeviceFormat(void* hDevice,
		const unsigned char pwdHashCipher[128]);

	/**
	 * @brief 设备初始化
	 *			使用随机密钥初始化设备
	 *
	 * @param	hDevice				[in]	已打开的设备句柄
	 * @param	newPwdHashCipher	[in]	新的设备管理员口令密文
	 * @param	bootAuth			[in]	是否开机认证
	  * @param	devSn				[in]	设备序列号，可以从设备表面标签或包装盒外部标签获取
	  * @param	selfCheckCycle		[in]	设备自检周期，暂时不启用
	  * @param	kekCv				[out]	设备本地保护密钥校验值，为NULL时不输出
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note
	 *	1. 设备必须为格式化状态（TassCtlDeviceFormat成功）
	 *	2. 需要设备管理员登录
	 *	3. 执行成功后会清除登录状态，需要使用新的设备管理员口令登录
	 */
	int TassCtlDeviceInit(void* hDevice,
		const unsigned char newPwdHashCipher[128],
		TassBool bootAuth,
		const unsigned char devSn[4], unsigned int selfCheckCycle,
		unsigned char kekCv[16]);

	/**
	 * @brief 设备初始化恢复出厂
	 *			使用出厂默认密钥初始化设备
	 *
	 * @param	hDevice			[in]	已打开的设备句柄
	  * @param	devSn			[in]	设备序列号，可以从设备表面标签或包装盒外部标签获取
	  * @param	selfCheckCycle	[in]	设备自检周期，暂时不启用
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note
	 *	1. 设备必须为格式化状态（TassCtlDeviceFormat成功）
	 *	2. 需要设备管理员登录
	 */
	int TassCtlDeviceInitRestoreFactory(void* hDevice,
		const unsigned char devSn[4],
		unsigned int selfCheckCycle);

	/**
	 * @brief 获取密码卡信息（暂未启用）
	 *
	 * @param	hDevice		[in]		已打开的设备句柄
	 * @param	info		[in]		信息类型，详见TassInfoType说明
	 * @param	buf			[out]		具体信息，传NULL时通过len返回需要的缓冲区大小
	 * @param	len			[in|out]	输入时标识buf缓冲区大小
	 *									输出时标识buf实际长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlGetInfo(void* hDevice,
		TassDevInfo info,
		void* buf, unsigned int* len);

	/**
	 * @brief 设置密码卡信息（暂未启用）
	 *
	 * @param	hDevice		[in]	已打开的设备句柄
	 * @param	info		[in]	信息类型，仅TA_DEV_SN和TA_SELF_CHECK_CYCLE有效
	 * @param	buf			[in]	具体信息
	 * @param	len			[in]	信息长度
	 * @param	sig			[in]	管理员私钥对函数名和除hDevice及sig之外所有入参的签名
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录设备管理员
	 *
	 */
	int TassCtlSetInfo(void* hDevice,
		TassDevInfo info,
		void* buf, unsigned int len,
		const unsigned char sig[64]);

	/**
	 * @brief 自检（暂未启用）
	 *			使用随机密钥替换密码卡默认密钥
	 *
	 * @param	hDevice		[in]	已打开的设备句柄
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlSelfCheck(void* hDevice);

	/**
	 * @brief 恢复出厂设置
	 *			恢复密码卡默认密钥和状态
	 *
	 * @param	hDevice			[in]	已打开的设备句柄
	 * @param	pwdHashCipher	[in]	设备管理员口令的哈希值密文
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录设备管理员
	 *
	 */
	int TassCtlRestoreFactory(void* hDevice,
		const unsigned char pwdHashCipher[128]);

	/**
	* @brief 生成设备加密密钥对（暂未启用）
	*
	*
	* @param	hDevice		[in]	已打开的设备句柄
	*
	* @return
	*   @retval 0		成功
	*   @retval other	失败
	*
	* @note 须先登录设备管理员
	*
	*/
	int TassCtlGenDevEncKeyPair(void* hDevice);

	/**
	* @brief 生成设备本地保护密钥（暂未启用）
	*
	*
	* @param	hDevice		[in]	已打开的设备句柄
	* @param	bootAuth	[in]	是否开机认证
	*
	* @return
	*   @retval 0		成功
	*   @retval other	失败
	*
	* @note 须先登录设备管理员
	*
	*/
	int TassCtlGenDevKEK(void* hDevice,
		TassBool bootAuth);

	/**
	* @brief 导入设备加密密钥对（暂未启用）
	*
	*
	* @param	hDevice					[in]	已打开的设备句柄
	* @param	pwdHashCipher			[in]	设备管理员口令的哈希值密文
	* @param	pk						[in]	设备加密密钥对公钥
	* @param	skEnvelopByDevSignPk	[in]	设备签名密钥对封装的设备加密密钥对私钥信封
	*
	* @return
	*   @retval 0		成功
	*   @retval other	失败
	*
	* @note 须先登录设备管理员
	*
	*/
	int TassCtlImportDevEncKeyPair(void* hDevice,
		const unsigned char pwdHashCipher[128],
		const unsigned char pk[64],
		const unsigned char skEnvelopByDevSignPk[144]);

	/**
	* @brief 导入设备本地保护密钥（暂未启用）
	*
	*
	* @param	hDevice					[in]	已打开的设备句柄
	* @param	pwdHashCipher			[in]	设备管理员口令的哈希值密文
	* @param	bootAuth				[in]	是否开机认证
	* @param	kekCipherByDevEncPk		[in]	设备加密密钥对加密的设备本地保护密钥密文
	*
	* @return
	*   @retval 0		成功
	*   @retval other	失败
	*
	* @note 须先登录设备管理员
	*
	*/
	int TassCtlImportDevKEK(void* hDevice,
		const unsigned char pwdHashCipher[128],
		TassBool bootAuth,
		const unsigned char kekCipherByDevEncPk[112]);

	/**
	 * @brief 开机认证
	 *
	 * @param	hDevice			[in]	已打开的设备句柄
	 * @param	pwdHashCipher	[in]	设备管理员口令的哈希值密文
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录设备管理员
	 */
	int TassCtlBootAuth(void* hDevice,
		const unsigned char pwdHashCipher[128]);

	/**
	 * @brief 导出设备加密密钥对（暂未启用）
	 *
	 * @param	hDevice				[in]	已打开的设备句柄
	 * @param	keyPwdHashCipher	[in]	密钥管理员口令的哈希值密文
	 * @param	devPwdHashCipher	[in]	设备管理员口令的哈希值密文
	 * @param	pk					[in]	加密公钥，通常为另一个设备的签名公钥
	 * @param	encPk				[out]	设备加密密钥对公钥
	 * @param	encSkEnvelopByPk	[out]	加密公钥封装的设备加密密钥对私钥信封
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录密钥管理员，同时需要设备管理员口令进行授权
	 */
	int TassCtlExportDevEncKeyPair(void* hDevice,
		const unsigned char keyPwdHashCipher[128],
		const unsigned char devPwdHashCipher[128],
		const unsigned char pk[64],
		unsigned char encPk[64],
		unsigned char encSkEnvelopByPk[144]);

	/**
	 * @brief 导出设备本地保护密钥（暂未启用）
	 *
	 * @param	hDevice				[in]	已打开的设备句柄
	 * @param	keyPwdHashCipher	[in]	密钥管理员口令的哈希值密文
	 * @param	devPwdHashCipher	[in]	设备管理员口令的哈希值密文
	 * @param	pk					[in]	加密公钥，通常为另一个设备的加密公钥
	 * @param	encKek				[out]	加密公钥加密的设备本地保护密钥密文
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录密钥管理员，同时需要设备管理员口令进行授权
	 */
	int TassCtlExportDevKEK(void* hDevice,
		const unsigned char keyPwdHashCipher[128],
		const unsigned char devPwdHashCipher[128],
		const unsigned char pk[64],
		unsigned char encKek[112]);

	/**
	* @brief 用户密钥管理
	*
	* @func TassListKey
	* @func TassGenerateKey
	* @func TassDestroyKey
	* @func TassImportPlainKey
	* @func TassBackup
	* @func TassRecover
	*
	*/

	/**
	 * @brief 获取密钥列表
	 *
	 * @param	hDevice		[in]		已打开的设备句柄
	 * @param	alg			[in]		密钥列表的算法
	 * @param	info		[out]		密钥信息，传NULL时通过len返回需要的缓冲区大小
	 * @param	len			[in|out]	输入时标识info缓冲区大小
	 *									输出时标识info实际长度
	 *									密钥数量可通过计算“len / sizeof(TassKeyInfo)”获取
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlListKey(void* hDevice,
		TassAlg alg,
		TassKeyInfo* info, int* len);

	/**
	 * @brief 生成密钥
	 *
	 * @param	hDevice		[in]		已打开的设备句柄
	 * @param	alg			[in]		密钥类型
	 * @param	keyBits		[in]		模长,
	 *									密钥类型为 2-SM2、3-ECC_256R1、8-ECC_256K1时，模长只支持256
	 *									密钥类型为 4-RSA时，模长只支持2048
	 *									密钥类型为 9-HMAC时，支持模长是128、256、384、512bit,
	 *									类型为0-SM4、1-SM1、5-AES、6-DES、7-SM7时，仅模长仅支持128
	 * @param	index		[in]		密钥索引，为0时根据标签存储密钥，为-1时不存储
	 * @param	label		[in]		密钥标签
	 * @param	pwd			[in]		私钥口令，当前未启用
	 *									type为非对称密钥时有效
	 * @param	sk_key		[out]		私钥密文或对称密钥密文，传NULL时通过sk_keyLen返回需要的缓冲区大小
	 * @param	sk_keyLen	[in|out]	输入时标识sk_key缓冲区大小
	 *									输出时标识sk_key实际长度
	 * @param	pk_kcv		[out]		公钥，传NULL时通过pk_kcvLen返回需要的缓冲区大小
	 * @param	pk_kcvLen	[in|out]	输入时标识pk_kcv缓冲区大小
	 *									输出时标识pk_kcv实际长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlGenerateKey(void* hDevice,
		TassAlg alg,
		unsigned int keyBits,
		unsigned int index,
		const char* label,
		const char* pwd,
		unsigned char* sk_key, unsigned int* sk_keyLen,
		unsigned char* pk_kcv, unsigned int* pk_kcvLen);

	/**
	 * @brief 导入明文密钥
	 *
	 * @param	hDevice		[in]		已打开的设备句柄
	 * @param	alg			[in]		导入的密钥算法，暂时仅支持TA_ALG_SM2和TA_ALG_SM4
	 * @param	index		[in]		导入的密钥索引，为0时根据标签存储密钥，为-1时不存储
	 * @param	label		[in]		导入的密钥标签
	 * @param	pwd			[in]		私钥口令，当前未启用
	 *									type为非对称密钥时有效
	 * @param	sk_key		[in]		私钥或对称密钥明文
	 * @param	sk_keyLen	[in]		sk_key长度
	 * @param	pk_kcv		[in]		公钥
	 *									type为非对称密钥时有效
	 * @param	pk_kcvLen	[in]		公钥长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlImportPlainKey(void* hDevice,
		TassAlg alg,
		unsigned int index,
		const char* label,
		const char* pwd,
		const unsigned char* sk_key, unsigned int sk_keyLen,
		const unsigned char* pk_kcv, unsigned int pk_kcvLen);

	/**
	 * @brief 通过sm2密钥保护导出非对称密钥
	 *
	 * @param	hDevice						[in] 已打开的设备句柄
	 * @param	sm2Pk						[in] SM2保护公钥
	 * @param	exportedKeyAlg				[in] 待导出的密钥算法，支持SM2/ECC_SECP_256R1/RSA/ECC_SECP_256K1
	 * @param	exportedKeyIndex			[in] 待导出的密钥索引
	 * @param	devPwdHashCipher			[in] 设备管理员口令的哈希值密文
	 * @param	symmKeyCipher				[out]随机对称密钥密文
	 * @param	symmKeyCipherkLen			[out]随机对称密钥密文长度
	 * @param	exportedPk					[out] 待导出的公钥
	 * @param	exportedPkLen				[out] 待导出的公钥长度
	 * @param	exportedKeyCipherByPk		[out] SM2公钥加密的待导出私钥密钥密文
	 * @param	exportedKeyCipherByPkLen	[out] SM2公钥加密的待导出私钥密钥密文长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlExportKey(void* hDevice, 
		const unsigned char sm2Pk[64], 
		TassAlg exportedKeyAlg,
		unsigned int exportedKeyIndex,
		const unsigned char devPwdHashCipher[128],
		unsigned char* symmKeyCipher,
		unsigned int* symmKeyCipherLen,
		unsigned char* exportedPk,
		unsigned int* exportedPkLen,
		unsigned char* exportedKeyCipherByPk,
		unsigned int* exportedKeyCipherByPkLen);

	/**
	* @brief 通过sm2密钥保护导入非对称密钥
	*
	* @param	hDevice						[in] 已打开的设备句柄
	* @param	sm2Sk						[in] SM2私钥
	* @param	importedKeyAlg				[in] 待导入的密钥算法,支持SM2/ECC_SECP_256R1/RSA/ECC_SECP_256K1
	* @param	importedKeyIndex			[in] 待导入的密钥索引
	* @param	importedKeyLabel			[in] 待导入的密钥标签
	* @param	importedKeyLabelLen			[in] 待导入的密钥标签长度
	* @param	symmKeyCipher				[in] 随机对称密钥密文
	* @param	symmKeyCipherkLen			[in] 随机对称密钥密文长度
	* @param	importedPk					[in] 待导入的公钥
	* @param	importedPkLen				[in] 待导入的公钥长度
	* @param	importedKeyCipherByPk		[in] SM2公钥加密的待导入私钥密钥密文
	* @param	importedKeyCipherByPkLen	[in] SM2公钥加密的待导入私钥密钥密文长度
	*
	* @return
	*   @retval 0		成功
	*   @retval other	失败
	*
	*/
	int TassCtlImportKey(void* hDevice,
		const unsigned char sm2Sk[32],
		TassAlg importedKeyAlg,
		unsigned int importedKeyIndex,
		const unsigned char* importedKeyLabel,
		unsigned int importedKeyLabelLen,
		unsigned char* symmKeyCipher,
		unsigned int symmKeyCipherLen,
		const unsigned char* importedPk,
		unsigned int importedPkLen,
		const unsigned char* importedKeyCipherByPk,
		unsigned int importedKeyCipherByPkLen);

	/**
	 * @brief 删除密钥
	 *
	 * @param	hDevice		[in]		已打开的设备句柄
	 * @param	alg			[in]		密钥算法
	 * @param	index		[in]		密钥索引
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlDestroyKey(void* hDevice,
		TassAlg alg,
		unsigned int index);

	/**
	 * @brief 修改私钥权限口令
	 *
	 * @param	hDevice				[in]		已打开的设备句柄
	 * @param	alg					[in]		密钥算法，只支持非对称密钥
	 * @param	index				[in]		密钥索引
	 * @param	privateKeyPwd		[in]		私钥口令
	 * @param	privateKeyPwdLen	[in]		私钥口令长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 */
	int TassCtlChangePrivateKeyAccessRight(void* hDevice,
		TassAlg alg, unsigned int index,
		const unsigned char* privateKeyPwd, 
		unsigned int privateKeyPwdLen);
	
	/**
	 * @brief 备份，包括设备信息、密钥和FLASH索引信息
	 *
	 * @param	hDevice				[in]		已打开的设备句柄
	 * @param	keyPwdHashCipher	[in]		密钥管理员口令的哈希值密文
	 * @param	devPwdHashCipher	[in]		设备管理员口令的哈希值密文
	 * @param	info				[out]		备份信息，传NULL时通过infoLen返回需要的缓冲区大小
	 * @param	infoLen				[in|out]	输入时标识info缓冲区大小
	 *											输出时标识info实际长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录密钥管理员，同时需要设备管理员口令进行授权
	 *
	 */
	int TassCtlBackup(void* hDevice,
		const unsigned char keyPwdHashCipher[128],
		const unsigned char devPwdHashCipher[128],
		unsigned char* info, unsigned int* infoLen);

	/**
	 * @brief 恢复设备信息
	 *
	 * @param	hDevice			[in]		已打开的设备句柄
	 * @param	pwdHashCipher	[in]		设备管理员口令的哈希值密文
	 * @param	info			[in]		备份的设备信息
	 * @param	infoLen			[in]		info长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录设备管理员
	 *
	 */
	int TassCtlRecover(void* hDevice,
		const unsigned char pwdHashCipher[128],
		const unsigned char* info, unsigned int infoLen);

	/**
	 * @brief 校验备份数据信息的有效性
	 *
	 * @param	hDevice			[in]		已打开的设备句柄
	 * @param	pwdHashCipher	[in]		设备管理员口令的哈希值密文
	 * @param	info			[in]		备份的设备信息
	 * @param	infoLen			[in]		info长度
	 *
	 * @return
	 *   @retval 0		成功
	 *   @retval other	失败
	 *
	 * @note 须先登录设备管理员
	 *
	 */
	int TassCtlCheckBackupInfoValid(void* hDevice,
		const unsigned char pwdHashCipher[128],
		const unsigned char* info, unsigned int infoLen);

#ifdef __cplusplus
}
#endif
