#pragma once

#include "TassType4PCIeSM.h"
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

	/**
	* @brief 搜索设备
	*
	* @param id		[out]		设备句柄集合
	* @param idLen	[in|out]	输入时，标识id缓冲区大小
	*							输出时，标识输出id长度，通过 *idLen / TA_DEVICE_ID_SIZE 获取设备数量
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassScanDevice(unsigned char* id, unsigned int* idLen);

	/**
	* @brief 打开设备
	*
	* @param id				[in]	要打开的设备ID
	* @param phDevice		[out]	返回设备句柄
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassOpenDevice(unsigned char id[TA_DEVICE_ID_SIZE], void** phDevice);

	/**
	* @brief 关闭设备
	*
	* @param pDevice		[in]	已打开的设备句柄
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassCloseDevice(void* hDevice);

	/**
	* @brief 打开会话
	*
	* @param pDevice	[in]	已打开的设备句柄
	* @param hSess		[out]	打开的会话句柄
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassOpenSession(void* hDevice, void** phSess);

	/**
	* @brief 关闭会话
	*
	* @param hSess		[in]	已打开的会话句柄
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassCloseSession(void* hSess);

	/**
	* @brief 设置超时时间
	*
	* @param timout	 [in]	超时时间，毫秒
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassSetTimeout(void* hDevice, int timout);

	/**
	* @brief 获取超时时间
	*
	* @param timout	 [out]	超时时间，毫秒
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassGetTimeout(void* hDevice, int* timout);

	/**
	* @brief 获取错误码描述信息
	*
	* @param err		[in]	错误码
	*
	* @return 错误码描述信息
	*
	*/
	const char* TassGetErrorDesc(int err);

	/*
	* 设备管理类指令
	*/

	/**
	* @brief 获取密码卡信息
	*
	* @param hSess		[in]		会话句柄
	* @param devInfo	[out]		设备信息
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassGetDeviceInfo(void* hSess, TassDevInfo* devInfo);

	/**
	* @brief 设备自检
	*
	* @param hSess		[in]	会话句柄
	* @param res		[in]	自检结果
	*							1B：SM4 IP结果
	*							1B：SM2 IP结果
	*							1B：密管芯片结果
	*							1B：WNG8芯片结果
	*
	* @return
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassDeviceSelfCheck(void* hSess,
		unsigned char res[4]);

	/**
	* @brief 恢复出厂设置
	*
	* @param hSess		[in]	会话句柄
	* @param cb			[in]	管理平台私钥签名回调函数
	*
	* @return
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassRestoreFactory(void* hSess,
		const TassSignCb cb);

	/**
	* @brief 设置设备基础信息
	*
	* @param hSess				[in]	会话句柄
	* @param devSn				[in]	设备序列号
	* @param selfCheckCycle		[in]	自检周期
	* @param cb					[in]	管理平台私钥签名回调函数
	*
	* @return
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassSetDeviceBaseInfo(void* hSess,
		const unsigned char devSn[4],
		unsigned int selfCheckCycle,
		const TassSignCb cb);

	/*
	* 设备密钥管理类指令
	*/

	/**
	* @brief 生成设备密钥
	*
	* @param hSess			[in]	会话句柄
	* @param type			[in]	生成的设备密钥类型
	*									TA_DEV_SIGN/TA_DEV_ENC/TA_DEV_KEK有效
	* @param keyInfo		[in]	密钥信息
	* @param bootAuth		[in]	是否开机认证，0--否，非0--是
	*									type=TA_DEV_KEK时有效
	* @param kekcv			[out]	KEK校验值，为NULL时不输出
	*									type=TA_DEV_KEK时有效
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassGenDeviceKey(void* hSess,
		TassDevKeyType type,
		const unsigned char keyInfo[4],
		TassBool bootAuth,
		unsigned char kekcv[16]);

	/**
	* @brief 获取设备公钥
	*
	* @param hSess		[in]	会话句柄
	* @param type		[in]	获取的设备密钥类型
	*								TA_DEV_KEY_PLATFORM/TA_DEV_KEY_SIGN/TA_DEV_KEY_ENC有效
	* @param keyInfo	[out]	密钥信息，为NULL时不输出
	* @param pk			[out]	公钥
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassGetDevicePublicKey(void* hSess,
		TassDevKeyType type,
		unsigned char keyInfo[4],
		unsigned char pk[64]);

	/**
	* @brief 设置平台公钥
	*
	* @param hSess		[in]	会话句柄
	* @param keyInfo	[in]	密钥信息
	* @param pk			[in]	公钥
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassSetPlatformPublicKey(void* hSess,
		const unsigned char keyInfo[4],
		const unsigned char pk[64]);

	/**
	* @brief 导入管理员公钥
	*
	* @param hSess		[in]	会话句柄
	* @param keyInfo	[in]	密钥信息
	* @param pk			[in]	导入的公钥
	* @param cb			[in]	（管理平台私钥）签名回调
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassImportAdminPublicKey(void* hSess,
		const unsigned char keyInfo[4],
		const unsigned char pk[64],
		const TassSignCb cb);

	/**
	* @brief 增加管理员公钥
	*
	* @param hSess			[in]	会话句柄
	* @param keyInfo		[in]	密钥信息
	* @param pk				[in]	增加的公钥
	* @param adminPk		[in]	（已经存在的）管理员公钥
	* @param cb				[in]	（adminPk对应私钥）签名回调
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassAddAdminPublicKey(void* hSess,
		const unsigned char keyInfo[4],
		const unsigned char pk[64],
		const unsigned char adminPk[64],
		const TassSignCb cb);

	/**
	* @brief 删除管理员公钥
	*
	* @param hSess			[in]	会话句柄
	* @param keyInfo		[in]	密钥信息
	* @param pk				[in]	删除的公钥
	* @param adminPk		[in]	（已经存在的）管理员公钥，可以与pubKey相同
	* @param cb				[in]	（adminPk对应私钥）签名回调
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassDeleteAdminPublicKey(void* hSess,
		const unsigned char keyInfo[4],
		const unsigned char pk[64],
		const unsigned char adminPk[64],
		const TassSignCb cb);

	/**
	* @brief 导出设备加密密钥对
	*
	* @param hSess							[in]	会话句柄
	* @param protectPk						[in]	保护公钥，一般为另一个密码卡的签名公钥
	* @param adminPk						[in]	管理员公钥
	* @param adminCb						[in]	（adminPk对应私钥）签名回调
	* @param platformCb						[in]	（平台公钥对应私钥）签名回调
	* @param keyInfo						[out]	密钥信息，为NULL时不输出
	* @param devEncPk						[out]	设备加密公钥，为NULL时不输出
	* @param devEncSkEnvelopByProtectPk		[out]	设备加密私钥信封，通过protectPk加密
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassExportDeviceEncKeyPair(void* hSess,
		const unsigned char protectPk[64],
		const unsigned char adminPk[64], const TassSignCb adminCb,
		const TassSignCb platformCb,
		unsigned char keyInfo[4],
		unsigned char devEncPk[64],
		unsigned char devEncSkEnvelopByProtectPk[144]);

	/**
	* @brief 导出设备KEK
	*
	* @param hSess					[in]	会话句柄
	* @param protectPk				[in]	保护公钥，一般为另一个密码卡的加密公钥
	* @param adminPk				[in]	管理员公钥
	* @param adminCb				[in]	（adminPk对应私钥）签名回调
	* @param platformCb				[in]	（平台公钥对应私钥）签名回调
	* @param keyInfo				[out]	密钥信息，为NULL时不输出
	* @param kekCipherByProtectPk	[out]	设备KEK密文，通过protectPk加密
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassExportDeviceKEK(void* hSess,
		const unsigned char protectPk[64],
		const unsigned char adminPk[64], const TassSignCb adminCb,
		const TassSignCb platformCb,
		unsigned char keyInfo[4],
		unsigned char kekCipherByProtectPk[112]);

	/**
	* @brief 导入设备加密密钥对
	*
	* @param hSess					[in]	会话句柄
	* @param keyInfo				[in]	密钥信息
	* @param pk						[in]	公钥
	* @param skEnvelopByDevSignPk	[in]	设备签名公钥加密的私钥私钥信封
	* @param cb						[in]	（平台公钥对应私钥）签名回调

	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassImportDeviceEncKeyPair(void* hSess,
		const unsigned char keyInfo[4],
		const unsigned char pk[64],
		const unsigned char skEnvelopByDevSignPk[144],
		const TassSignCb cb);

	/**
	* @brief 导入设备KEK
	*
	* @param hSess					[in]	会话句柄
	* @param bootAuth				[in]	是否开机认证，0--否，非0--是
	* @param keyInfo				[in]	密钥信息
	* @param kekCipherByDevEncPk	[in]	设备加密密钥对加密的KEK密文
	* @param cb						[in]	（平台公钥对应私钥）签名回调
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassImportDeviceKEK(void* hSess,
		TassBool bootAuth,
		const unsigned char keyInfo[4],
		const unsigned char kekCipherByDevEncPk[112],
		const TassSignCb cb);

	/**
	* @brief 开机认证
	*
	* @param hSess             [in]	 会话句柄
	* @param cb                [in]  （平台公钥对应私钥）签名回调

	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassBootAuth(void* hSess, const TassSignCb cb);

	/**
	* @brief 开机认证
	*
	* @param hSess             [in]	 会话句柄
	* @param state             [in]	 要设置的工作状态
	*									TA_DEV_STATE_WORK/TA_DEV_STATE_MNG有效
	* @param cb                [in]  （平台公钥对应私钥）签名回调
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassSetDeviceState(void* hSess, TassDevState state, const TassSignCb cb);

	/*
	* 应用密钥管理类指令
	*/

	/*
	3.3.1
	后续废弃，仅用于测试。
	仅在管理状态下可用
	*/
	int TassAuthSM2PublicKey(void* hSess,
		const unsigned char pk[64],
		unsigned char authCode[4]);

	/*
	3.3.1
	* @brief 扩展认证保护密钥
	*
	* @param hSess             [in]	 会话句柄
	* @param alg			   [in]	 保护密钥类型, 0–SM4，2-SM2，4–RSA
	* @param protectKey		   [in]	 保护密钥，当保护密钥类型是 0 时，保护密钥是16字节的 SM4 密钥；
	*										   当保护密钥类型是 2 时，保护密钥是 64字节 SM2 公钥；
	*										   当保护密钥类型是 4 时，保护密钥是RSA2048公钥，此时是4字节的公钥长度+公钥内容
	* @param protectKeyLen	   [in]	  保护密钥长度
	* @param cb                [in]  （平台公钥对应私钥）签名回调
	* @param authCode		   [out]  认证码
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassAuthPortectKey(void* hSess,
		TassAlg alg,
		const unsigned char* protectKey, unsigned int protectKeyLen,
		const TassSignCb cb,
		unsigned char authCode[4]);

	/*
	3.3.3
	同时适用于19150
	*/
	int TassGenSM2KeyPair(void* hSess,
		unsigned char skCipherByKek[32],
		unsigned char pk[64]);

	/*
	3.3.4
	同时适用于19150
	密钥信息校验值，为NULL时不输出
	*/
	int TassGenSM4Key(void* hSess,
		unsigned char keyCipherByKek[16],
		unsigned char kcv[16]);

	/*
	3.3.7
	同时适用于19150
	密钥信息校验值，为NULL时不输出
	*/
	int TassGen_ExportSM4KeyBySM4Key(void* hSess,
		const unsigned char protectKeyCipherByKek[16],
		unsigned char keyCipherByProtectKey[16],
		unsigned char keyCipherByKek[16],
		unsigned char kcv[16]);

	/*
	3.3.8
	同时适用于19150
	*/
	int TassExportSM2PrivateKeyBySM2PublicKey(void* hSess,
		const unsigned char protectPk[64],
		const unsigned char authCode[4],
		const unsigned char skCipherByKek[32],
		unsigned char skEnvelopByProtectPk[144]);
	/*
	3.3.9
	同时适用于19150
	*/
	int TassImportSM2PrivateKeyBySM2PrivateKey(void* hSess,
		const unsigned char protectSkCipherByKek[32],
		const unsigned char skEnvelopByProtectPk[144],
		unsigned char skCipherByKek[32]);

	/*
	3.3.12
	同时适用于19150
	密钥信息校验值，为NULL时不输出
	*/
	int TassExportSM4KeyBySM2PublicKey(void* hSess,
		const unsigned char protectPk[64],
		const unsigned char authCode[4],
		const unsigned char keyCipherByKek[16],
		unsigned char keyCipherByProtectPk[112],
		unsigned char kcv[16]);

	/*
	3.3.13
	同时适用于19150
	*/
	int TassImportSM4KeyBySM2PrivateKey(void* hSess,
		const unsigned char protectSkCipherByKek[32],
		const unsigned char keyCipherByProtectPk[112],
		const unsigned char kcv[16],
		unsigned char keyCipherByKek[16]);

	/*
	3.3.16
	同时适用于19150
	*/
	int TassImportSM4KeyBySM4Key(void* hSess,
		const unsigned char protectKeyCipherByKek[16],
		const unsigned char keyCipherByProtectKey[16],
		const unsigned char kcv[16],
		unsigned char keyCipherByKek[16]);

	/*
	3.3.17
	密钥信息校验值，为NULL时不输出
	仅导出SM2和SM4密钥
	*/
	int TassExportKeyBySM4Key(void* hSess,
		const unsigned char protectKeyCipherByKek[16],
		const unsigned char authCode[4],
		TassAlg alg,
		const unsigned char* keyCipherByKek, unsigned int keyCipherByKekLen,
		unsigned char* keyCipherByProtectKey, unsigned int* keyCipherByProtectKeyLen,
		unsigned char kcv[16]);

	/*
	3.3.4
	*/
	int TassGenSymmKey(void* hSess,
		TassAlg keyAlg,
		unsigned int keyBits,		//密钥类型为 9 时，支持模长是128、256、384、512bit,其他类型仅支持 128(0x00000080)
		unsigned char* keyCipherByKek, unsigned int* keyCipherByKekLen,
		unsigned char kcv[16]);

	/*
	3.3.5
	* @brief 生成非对称密钥
	*
	* @param hSess						[in]	会话句柄
	* @param keyAlg						[in]	密钥类型, 2-SM2, 3–SECP_256R1, 4–RSA，8-SECP_256K1
	* @param keyBits					[in]	模长, 密钥类型是2或3 时都只支持256，密钥类型是4时只支持 2048
	* @param rsaE						[in]	公钥指数
	* @param skCipherByKek				[out]	私钥密文
	* @param skCipherByKekLen			[out]	私钥密文长度
	* @param pk							[out]	公钥
	* @param pkLen						[out]	公钥长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassGenAsymKeyPair(void* hSess,
		TassAlg keyAlg,
		unsigned int keyBits,
		TassRSA_E rsaE,
		unsigned char* skCipherByKek, unsigned int* skCipherByKekLen,
		unsigned char* pk, unsigned int* pkLen);

	/*
	3.3.9
	* @brief 将非对称私钥转加密（本地保护密钥 KEK 加密转为SM2加密）
	*
	* @param hSess						[in]	会话句柄
	* @param protectPk					[in]	经认证的SM2公钥
	* @param authCode					[in]	认证码
	* @param keyAlg						[in]	密钥类型，2-SM2、3–ECC（SECP 256r1）、4–RSA、8–ECC（SECP 256k1）
	* @param keyBits					[in]	模长
	* @param skCipherByKek				[in]	私钥密文
	* @param skCipherByKekLen			[in]	私钥密文长度
	* @param symmKeyCipher				[out]	随机对称密钥密文
	* @param symmKeyCipherLen			[out]	随机对称密钥密文长度
	* @param skEnvelopByProtectPk		[out]	业务密钥私钥密文
	* @param skEnvelopByProtectPkLen	[out]	业务密钥私钥密文长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassExportAsymPrivateKeyBySM2PublicKey(void* hSess,
		const unsigned char protectPk[64],
		const unsigned char authCode[4],
		TassAlg keyAlg,
		unsigned int keyBits,
		const unsigned char* skCipherByKek, unsigned int skCipherByKekLen,
		unsigned char* symmKeyCipher, unsigned int *symmKeyCipherLen,
		unsigned char* skEnvelopByProtectPk, unsigned int* skEnvelopByProtectPkLen);



	/*
	3.3.11
	* @brief 将非对称私钥转加密（SM2加密转为本地保护密钥KEK加密）
	*
	* @param hSess						[in]	会话句柄
	* @param protectSkCipherByKek		[in]	解密私钥密文
	* @param keyAlg						[in]	密钥类型, 2-SM2, 3–ECC 256r1, 4–RSA，8-ECC 256k1
	* @param keyBits					[in]	模长, 密钥类型是2或3 时都只支持256，密钥类型是4时只支持 2048
	* @param symmKeyCipher				[in]	随机对称密钥密文
	* @param symmKeyCipherLen			[in]	随机对称密钥密文长度
	* @param skEnvelopByProtectPk		[in]	业务密钥私钥密文，（数字信封）
	* @param skEnvelopByProtectPkLen	[in]	业务密钥私钥密文长度
	* @param skCipherByKek				[out]	KEK加密的业务密钥私钥密文
	* @param skCipherByKekLen			[out]	KEK加密的业务密钥私钥密文长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassImportAsymPrivateKeyBySM2PrivateKey(void* hSess,
		const unsigned char protectSkCipherByKek[32],
		TassAlg keyAlg,
		unsigned int keyBits,
		const unsigned char* symmKeyCipher, unsigned int symmKeyCipherLen,
		const unsigned char* skEnvelopByProtectPk, unsigned int skEnvelopByProtectPkLen,
		unsigned char* skCipherByKek, unsigned int* skCipherByKekLen);

	/*
	3.3.14
	*/
	int TassExportSymmKey(void* hSess,
		TassAlg protectKeyAlg,
		unsigned int protectKeyBits,
		const unsigned char* protectKey, unsigned int protectKeyLen,
		const unsigned char authCode[4],
		TassAlg keyAlg,
		unsigned int keyBits,
		const unsigned char* keyCipherByKek, unsigned int keyCipherByKekLen,
		unsigned char* keyCipherByProtectKey, unsigned int* keyCipherByProtectKeyLen,
		unsigned char kcv[16]);

	/*
	3.3.15
	* @brief 将对称密钥转加密（外部密钥加密转为本地保护密钥 KEK 加密）
	*
	* @param hSess						[in]	会话句柄
	* @param protectKeyAlg				[in]	外部保护密钥类型，0-SM4，2-SM2, 4–RSA
	* @param protectKeyBits				[in]	外部保护密钥模长
	* @param protectKey					[in]	外部保护密钥
	* @param protectKeyLen				[in]	外部保护密钥长度
	* @param keyAlg						[in]	对称密钥类型，0-SM4，1-SM1，5--AES，6-DES，7-SM7
	* @param keyBits					[in]	对称密钥模长，当前仅支持 128
	* @param keyCipherByProtectKey		[in]	业务密钥密文（即被公钥加密的对称密钥密文）
	* @param keyCipherByProtectKeyLen	[in]	业务密钥密文长度
	* @param kcv						[in]	密钥校验值
	* @param keyCipherByKek				[out]	外部密钥密文（即公钥对应的私钥或者 SM4 密钥密文）
	* @param keyCipherByKekLen			[out]	外部密钥密文长度
	*
	* @retval	成功返回0，失败返回非0
	*/
	int TassImportSymmKey(void* hSess,
		TassAlg protectKeyAlg,
		unsigned int protectKeyBits,
		const unsigned char* protectKey, unsigned int protectKeyLen,
		TassAlg keyAlg,
		unsigned int keyBits,
		const unsigned char* keyCipherByProtectKey, unsigned int keyCipherByProtectKeyLen,
		const unsigned char kcv[16],
		unsigned char* keyCipherByKek, unsigned int* keyCipherByKekLen);

	/*
	* 密码运算类指令
	*/

	int TassGenRandom(void* hSess,
		unsigned int randomLen,
		unsigned char* random);

	/*
	3.4.2 + 4.2.1
	同时适用于19150
	*/
	int TassSM2PrivateKeySign(void* hSess,
		unsigned int index,
		const unsigned char skCipherByKek[32],
		const unsigned char hash[32],
		unsigned char sig[64]);

	/*
	3.4.3 + 4.2.2
	同时适用于19150
	*/
	int TassSM2PublicKeyVerify(void* hSess,
		unsigned int index,
		const unsigned char pk[64],
		const unsigned char hash[32],
		const unsigned char sig[64]);

	/*
	3.4.4 + 4.2.3
	同时适用于19150
	*/
	int TassSM2PublicKeyEncrypt(void* hSess,
		unsigned int index,
		const unsigned char pk[64],
		const unsigned char* plain, unsigned int plainLen,
		unsigned char* cipher, unsigned int* cipherLen);

	/*
	3.4.5 + 4.2.4
	同时适用于19150
	*/
	int TassSM2PrivateKeyDecrypt(void* hSess,
		unsigned int index,
		const unsigned char skCipherByKek[32],
		const unsigned char* cipher, unsigned int cipherLen,
		unsigned char* plain, unsigned int* plainLen);

	/*
	3.4.6 + 4.2.5
	同时适用于19150
	*/
	int TassSM2KeyExchange(void* hSess,
		TassBool sponsor,
		unsigned int selfIndex,
		const unsigned char selfSkCipherByKek[32],
		const unsigned char selfPk[64],
		const unsigned char selfTmpSkCipherByKek[32],
		const unsigned char selfTmpPk[64],
		const unsigned char* selfId, unsigned int selfIdLen,
		const unsigned char peerPk[64],
		const unsigned char peerTmpPk[64],
		const unsigned char* peerId, unsigned int peerIdLen,
		unsigned int keyBytes,
		TassBool genPlainKey,
		unsigned char* key);

	/*
	* @brief ECC私钥签名
	*
	* @param hSess						[in]	会话句柄
	* @param alg						[in]	ECC密钥类型
	* @param index						[in]	索引
	* @param keyBits					[in]	模长，ECC_256R1和ECC_256K1均为256，索引为0时有效
	* @param skCipherByKek				[in]	签名私钥，索引为0时有效
	* @param skCipherByKekLen			[in]	签名私钥长度，索引为0时有效
	* @param hash						[in]	哈希值
	* @param hashLen					[in]	哈希值长度
	* @param sig						[out]	签名值
	* @param sigLen						[in/out]	签名值长度
	*
	* @retval	成功返回0，失败返回非0
	*/
	int TassECCPrivateKeySign(void* hSess,
		TassAlg alg,
		unsigned int index,
		unsigned int keyBits,
		const unsigned char* skCipherByKek, unsigned int skCipherByKekLen,
		const unsigned char* hash, unsigned int hashLen,
		unsigned char* sig, unsigned int* sigLen);

	/*
	* @brief ECC私钥签名（为复杂美项目特殊需求定制）
	*
	* @param hSess						[in]	会话句柄
	* @param alg						[in]	ECC密钥类型
	* @param index						[in]	索引
	* @param keyBits					[in]	模长，ECC_256R1和ECC_256K1均为256，索引为0时有效
	* @param skCipherByKek				[in]	签名私钥，索引为0时有效
	* @param skCipherByKekLen			[in]	签名私钥长度，索引为0时有效
	* @param hash						[in]	哈希值
	* @param hashLen					[in]	哈希值长度
	* @param sig						[out]	签名值
	* @param sigLen						[in/out]	签名值长度
	*
	* @retval	成功返回0，失败返回非0
	*/
	int TassECCPrivateKeySign_Eth(void* hSess,
		TassAlg alg,
		unsigned int index,
		unsigned int keyBits,
		const unsigned char* skCipherByKek, unsigned int skCipherByKekLen,
		const unsigned char* hash, unsigned int hashLen,
		unsigned char* sig, unsigned int* sigLen);
	/*
	* @brief ECC公钥验签
	*
	* @param hSess						[in]	会话句柄
	* @param alg						[in]	ECC密钥类型
	* @param index						[in]	索引
	* @param keyBits					[in]	模长，ECC_256R1和ECC_256K1均为256，索引为0时有效
	* @param pk							[in]	验签公钥，索引为0时有效
	* @param pkLen						[in]	验签公钥长度，索引为0时有效
	* @param hash						[in]	哈希值
	* @param hashLen					[in]	哈希值长度
	* @param sig						[out]	签名值
	* @param sigLen						[in/out]签名值长度
	*
	* @retval	成功返回0，失败返回非0
	*/
	int TassECCPublicKeyVerify(void* hSess,
		TassAlg alg,
		unsigned int index,
		unsigned int keyBits,
		const unsigned char* pk, unsigned int pkLen,
		const unsigned char* hash, unsigned int hashLen,
		const unsigned char* sig, unsigned int sigLen);

	/*
	3.4.9
	*/
	int TassECCKeyAgreement(void* hSess,
		TassAlg alg,
		unsigned int keyBits,
		const unsigned char* selfSkCipherByKek, unsigned int selfSkCipherByKekLen,
		const unsigned char* peerPk, unsigned int peerPkLen,
		TassBool genPlainKey,
		unsigned char* key, unsigned int* keyLen);

	/*
	3.4.10 + 4.2.9
	*/
	int TassRSAPrivateKeyOperate(void* hSess,
		unsigned int index,
		unsigned int keyBits,
		const unsigned char* skCipherByKek, unsigned int skCipherByKekLen,
		const unsigned char* inData, unsigned int inDataLen,
		unsigned char* outData);

	/*
	3.4.11 + 4.2.10
	*/
	int TassRSAPublicKeyOperate(void* hSess,
		unsigned int index,
		unsigned int keyBits,
		const unsigned char* pk, unsigned int pkLen,
		const unsigned char* inData, unsigned int inDataLen,
		unsigned char* outData);

	/*
	3.4.12 + 4.2.6
	同时适用于19150
	*/
	int TassSM4KeyOperate(void* hSess,
		TassSymmOp op,
		unsigned int index,
		const unsigned char keyCipherByKek[16],
		unsigned char* iv,
		const unsigned char* inData, unsigned int dataLen,
		unsigned char* outData);

	/*
	3.4.13 + 4.2.11
	*/
	int TassSymmKeyOperate(void* hSess,
		TassAlg alg,
		unsigned int keyBits,
		TassSymmOp op,
		unsigned int index,
		const unsigned char* keyCipherByKek, unsigned int keyCipherByKekLen,
		unsigned char* iv,
		const unsigned char* inData, unsigned int inDataLen,
		unsigned char* outData, unsigned int* outDataLen);

	int TassSM3Single(void* hSess,
		const unsigned char pk[64],
		const unsigned char* id, unsigned int idLen,
		const unsigned char* data, unsigned int dataLen,
		unsigned char hash[32]);

	int TassSM3Init(void* hSess,
		unsigned int uiAlgID,
		const unsigned char pk[64],
		const unsigned char* id, unsigned int idLen,
		unsigned char ctx[112]);

	int TassSM3Update(void* hSess,
		const unsigned char* data, unsigned int dataLen,
		unsigned char ctx[112]);

	int TassSM3Final(void* hSess,
		const unsigned char ctx[112],
		unsigned char hash[32]);

	/*
	* PKI扩展命令
	*/
	/*
	* 密钥管理命令
	*/

	/*
	4.1.1
	同时适用19150
	*/
	int TassGetIndexInfo(void* hSess,
		TassAlg alg,
		unsigned char* info, unsigned int* infoLen);

	/*
	4.1.2
	同时适用19150
	*/
	int TassSetLabel(void* hSess,
		TassAlg alg,
		unsigned int index,
		const unsigned char* label, unsigned int labelLen);

	/*
	4.1.3
	同时适用19150
	*/
	int TassGetLabel(void* hSess,
		TassAlg alg,
		unsigned int index,
		unsigned char* label, unsigned int* labelLen);

	/*
	4.1.4
	同时适用19150
	*/
	int TassGetIndex(void* hSess,
		TassAlg alg,
		const unsigned char* label, unsigned int labelLen,
		unsigned int* index);

	/*
	4.1.5
	同时适用19150
	*/
	/*
	* @brief 依据索引设置密钥属性
	*
	* @param hSess				[in]	会话句柄
	* @param alg				[in]	密钥类型
	* @param index				[in]	密钥索引
	* @param sk					[in]	公私钥标识，0–私钥，1–公钥
	* @param attr				[in]	密钥属性
	* @param attrLen			[in]	密钥属性长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassSetAttr(void* hSess,
		TassAlg alg,
		unsigned int index,
		TassBool sk,
		const unsigned char* attr, unsigned int attrLen);

	/*
	4.1.6
	同时适用19150
	*/
	/*
	* @brief 依据索引获取密钥属性
	*
	* @param hSess				[in]	会话句柄
	* @param alg				[in]	密钥类型
	* @param index				[in]	密钥索引
	* @param sk					[in]	公私钥标识，0–私钥，1–公钥
	* @param attr				[out]	密钥属性
	* @param attrLen			[out]	密钥属性长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassGetAttr(void* hSess,
		TassAlg alg,
		unsigned int index,
		TassBool sk,
		unsigned char* attr, unsigned int* attrLen);

	/*
	4.1.10
	同时适用19150
	*/
	int TassStoreSM2KeyPair(void* hSess,
		unsigned int index,
		const unsigned char* label, unsigned int labelLen,
		TassAsymKeyUsage usage,
		const unsigned char skCipherByKek[32],
		const unsigned char pk[64]);

	/**
	4.1.11
	* @brief 导入非对称密钥到密码卡
	*
	* @param hSess              [in]	会话句柄
	* @param alg				[in]	密钥类型
	* @param index				[in]	密钥索引，0-64，当密钥类型是 RSA 是为 0-4
	* @param label				[in]	密钥标签
	* @param labelLen			[in]	密钥标签长度
	* @param usage				[in]	0-签名密钥，1-加密密钥，2-密钥协商密钥
	* @param skCipherByKek		[in]	私钥密文
	* @param skCipherByKekLen	[in]	私钥密文长度
	* @param pk					[in]	公钥
	* @param pkLen				[in]	公钥长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassStoreAsymKeyPair(void* hSess,
		TassAlg alg,
		unsigned int index,
		const unsigned char* label, unsigned int labelLen,
		TassAsymKeyUsage usage,
		const unsigned char* skCipherByKek, unsigned int skCipherByKekLen,
		const unsigned char* pk, unsigned int pkLen);

	/*
	4.1.12
	同时适用19150
	*/
	int TassStoreSM4Key(void* hSess,
		unsigned int index,
		const unsigned char* label, unsigned int labelLen,
		const unsigned char keyCipherByKek[16],
		const unsigned char kcv[16]);

	/*
	4.1.13
	*/
	int TassStoreSymmKey(void* hSess,
		TassAlg alg,
		unsigned int index,
		const unsigned char* label, unsigned int labelLen,
		const unsigned char* keyCipherByKek, unsigned int keyCipherByKekLen,
		const unsigned char kcv[16]);

	/*
	4.1.14
	同时适用19150
	*/
	int TassDestroyKey(void* hSess,
		TassAlg alg,
		TassAsymKeyUsage usage,
		unsigned int index);

	/*
	4.1.8
	同时适用19150
	*/
	int TassGetSM2PublicKey(void* hSess,
		unsigned int index,
		TassAsymKeyUsage usage,
		unsigned char pk[64]);

	int TassGetAsymPublicKey(void* hSess,
		TassAlg alg,
		unsigned int index,
		TassAsymKeyUsage usage,
		unsigned char* pk, unsigned int* pkLen);

	/*
	4.1.14
	同时适用19150
	*/
	/*
	* @brief 依据索引设置密钥属性
	*
	* @param hSess					[in]	会话句柄
	* @param alg					[in]	密钥类型
	* @param index					[in]	密钥索引
	* @param usage					[in]	密钥用途，0–签名密钥，1–加密密钥, 2-密钥协商密钥
	* @param sk_keyCipherByKek		[out]	私钥
	* @param sk_keyCipherByKekLen	[in/out]私钥长度
	* @param sk_keyCipherByKek		[out]	公钥
	* @param sk_keyCipherByKekLen	[in/out]公钥长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassGetKey(void* hSess,
		TassAlg alg,
		unsigned int index,
		TassAsymKeyUsage usage,
		unsigned char* sk_keyCipherByKek, unsigned int* sk_keyCipherByKekLen,
		unsigned char* pk_kcv, unsigned int* pk_kcvLen);

	/*
	4.1.16 + 4.1.17
	2K FLASH 操作同时适用19150
	*/
	int TassFlashOperate(void* hSess,
		TassFlashFlag flag,
		TassFlashOp op,
		unsigned int offset,
		unsigned int dataLen,
		unsigned char* data);

	/**
	4.2.10
	* @brief ECC 密钥协商
	*
	* @param hSess              [in]	会话句柄
	* @param alg				[in]	密钥类型, 3–ECC_SECP_256R1, 8-ECC_SECP_256K1
	* @param index				[in]	密钥索引，1-64
	* @param pk					[in]	对方公钥
	* @param pkLen				[in]	对方公钥长度
	* @param agreementData		[out]	协商结果
	* @param agreementDataLen	[out]	协商结果长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassGenerateAgreementDataWithECC(void* hSess, 
		TassAlg alg, 
		unsigned int index, 
		unsigned char* pk, unsigned int pkLen,
		unsigned char* agreementData, unsigned int* agreementDataLen);

	/*
	* HMAC计算
	*/

	/**
	3.4.20、4.2.14
	* @brief HMAC单包计算
	*
	* @param hSess              [in]	会话句柄
	* @param index				[in]	密钥索引
	* @param key				[in]	密钥密文，索引为0时有效
	* @param keyLen				[in]	密钥密文长度，16/32/48/64，最大 64 字节，索引为0时有效
	* @param data				[in]	数据
	* @param dataLen			[in]	数据长度，最大不超过2000字节
	* @param hmac				[out]	HMAC 结果
	* @param hmacLen			[out]	HMAC 结果长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassHmacSingle(void* hSess,
		unsigned int index,
		unsigned char* key, unsigned int keyLen,
		unsigned char* data, unsigned int dataLen,
		unsigned char* hmac, unsigned int* hmacLen);

	/**
	3.4.21、4.2.15
	* @brief HMAC多包计算init
	*
	* @param hSess              [in]	会话句柄
	* @param index				[in]	密钥索引，仅支持HMAC密钥
	* @param key				[in]	密钥密文，索引为0时有效，仅支持HMAC密钥
	* @param keyLen				[in]	密钥密文长度，16/32/48/64，最大 64 字节，索引为0时有效，仅支持HMAC密钥
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassHmacInit(void* hSess,
		unsigned int index,
		unsigned char* key, unsigned int keyLen);

	/**
	3.4.22
	* @brief HMAC多包计算update
	*
	* @param hSess              [in]	会话句柄
	* @param data				[in]	数据
	* @param dataLen			[in]	数据长度，长度只能是64 字节的倍数
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassHmacUpdate(void* hSess, unsigned char* data, unsigned int dataLen);

	/**
	3.4.23、4.2.16
	* @brief HMAC多包计算final
	*
	* @param hSess              [in]	会话句柄
	* @param hmac				[out]	HMAC 结果
	* @param hmacLen			[out]	HMAC 结果长度
	*
	* @retval	成功返回0，失败返回非0
	*
	*/
	int TassHmacFinal(void* hSess, unsigned char* hmac, unsigned int* hmacLen);

	/**
	* @brief	获取密码设备内部存储的指定索引私钥的使用权
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储私钥的索引值
	* @param	uiAlgID			[IN]	密钥类型
	* @param	pucPassword		[IN]	使用私钥权限的识别码，默认为a1234567
	* @param	uiPwdLength		[IN]	私钥访问控制码长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	本标准涉及密码设备存储的密钥对索引值的起始索引值为1，最大为n，
	*		密码设备的实际存储容量决定n值
	*/
	int TassGetPrivateKeyAccessRight(void* hSessionHandle,
		unsigned int uiKeyIndex,
		TassAlg uiAlgID,
		unsigned char* pucPassword,
		unsigned int uiPwdLength);

#ifdef __cplusplus
}
#endif
