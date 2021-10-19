/**
* Copyright (C) 2010-2023 TASS
* @file TassSDF4PCIECrypotoCard.h
* @brief 根据GM/T 0018 《密码设备应用接口规范》声明相关函数
* @detail 本文件为实现SDF接口提供了相应的模板，使用者可根据该模板完成
*         特定设备的SDF接口开发
* @author Lgl
* @version 1.0.0
* @date 2021/02/19
* Change History :
* <Date>     | <Version>  | <Author>       | <Description>
*---------------------------------------------------------------------
* 2021/02/19 | 1.0.0      | Lgl         | Create file
*---------------------------------------------------------------------
* 2021/09/19 | 2.0.0      | Lgl         | Optimized interface function
*---------------------------------------------------------------------
*/
#pragma once
#ifdef __cplusplus
extern "C" {
#endif
	/**
	*分组密码算法标识
	*/

#define SGD_SM1_ECB		0X00000101		//SM1算法ECB加密模式
#define SGD_SM1_CBC		0X00000102		//SM1算法CBC加密模式
#define SGD_SM1_CFB		0X00000104		//SM1算法CFB加密模式
#define SGD_SM1_OFB		0X00000108		//SM1算法OFB加密模式
#define SGD_SM1_MAC		0X00000110		//SM1算法MAC运算

#define SGD_SSF33_ECB	0X00000201		//SSF33算法ECB加密模式
#define SGD_SSF33_CBC	0X00000202		//SSF33算法CBC加密模式
#define SGD_SSF33_CFB	0X00000204		//SSF33算法CFB加密模式
#define SGD_SSF33_OFB	0X00000208		//SSF33算法OFB加密模式
#define SGD_SSF33_MAC	0X00000210		//SSF33算法MAC运算

#define SGD_SM4_ECB		0X00000401		//SM4算法ECB加密模式
#define SGD_SM4_CBC		0X00000402		//SM4算法CBC加密模式
#define SGD_SM4_CFB		0X00000404		//SM4算法CFB加密模式
#define SGD_SM4_OFB		0X00000408		//SM4算法OFB加密模式
#define SGD_SM4_MAC		0X00000410		//SM4算法MAC运算

#define SGD_ZUC_EEA3	0X00000801		//ZUC祖冲之机密性算法128-EEA3算法
#define SGD_ZUC_EIA3	0X00000802		//ZUC祖冲之机密性算法128-EIA3算法

	/*TASS扩展*/
#define SGD_DES_ECB     0x80000101    // DES算法ECB加密模式
#define SGD_DES_CBC     0x80000102    // DES算法CBC加密模式
#define SGD_DES_CFB     0x80000104    // DES算法CFB加密模式
#define SGD_DES_OFB     0x80000108    // DES算法OFB加密模式
#define SGD_DES_MAC		0X80000110	  // DES算法MAC运算

#define SGD_AES_ECB     0x80000201    // AES算法ECB加密模式
#define SGD_AES_CBC     0x80000202    // AES算法CBC加密模式
#define SGD_AES_CFB     0x80000204    // AES算法CFB加密模式
#define SGD_AES_OFB     0x80000208    // AES算法OFB加密模式
#define SGD_AES_MAC     0x80000210    // AES算法MAC运算


	/**
	*非对称密码算法标识
	*/
#define SGD_RSA			0X00010000		//RSA算法
#define SGD_SM2			0X00020100		//SM2椭圆曲线密码算法
#define SGD_SM2_1		0X00020200		//SM2椭圆曲线签名算法
#define SGD_SM2_2		0X00020400		//SM2椭圆曲线密钥交换协议
#define SGD_SM2_3		0X00020800		//SM2椭圆曲线加密算法

	/**
	*密码杂凑算法标识
	*/
#define SGD_SM3			0X00000001		//SM3杂凑算法
#define SGD_SHA256		0X00000004		//SHA_256杂凑算法

	typedef struct DeviceInfo_st {
		unsigned char IssuerName[40];//设备生产厂商名称
		unsigned char DeviceName[16];//设备型号
		unsigned char DeviceSerial[16];//设备编号，包含日期（8字符）、批次号（3字符）、流水号（5字符）
		unsigned int DeviceVersion;//密码设备内部软件版本号
		unsigned int StandardVersion;//密码设备支持的接口规范版本号
		unsigned int AsymAlgAbility[2];///<非对称算法能力，前4字节表示支持的算法，非对称算法标识按位异或，后4字节表示算法的最大模长，表示方法为支持模长按位异或的结果
		unsigned int SymAlgAbility;//对称算法能力，对称算法标识按位异或
		unsigned int HashAlgAbility;//杂凑算法能力，杂凑算法标识按位异或
		unsigned int BufferSize;//支持的最大文件存储空间（单位字节）
	}DEVICEINFO;

#define RSAref_MAX_BITS		2048//4096
#define RSAref_MAX_LEN		((RSAref_MAX_BITS + 7) / 8)
#define RSAref_MAX_PBITS	((RSAref_MAX_BITS + 1) / 2)
#define RSAref_MAX_PLEN		((RSAref_MAX_PBITS + 7) / 8)

	typedef struct RSArefPublicKey_st {
		unsigned int bits;
		unsigned char m[RSAref_MAX_LEN];
		unsigned char e[RSAref_MAX_LEN];
	}RSArefPublicKey;

	typedef struct RSArefPrivateKey_st {
		unsigned int bits;
		unsigned char m[RSAref_MAX_LEN];
		unsigned char e[RSAref_MAX_LEN];
		unsigned char d[RSAref_MAX_LEN];
		unsigned char prime[2][RSAref_MAX_PLEN];
		unsigned char pexp[2][RSAref_MAX_PLEN];
		unsigned char coef[RSAref_MAX_PLEN];
	}RSArefPrivateKey;

#define ECCref_MAX_BITS		512
#define ECCref_MAX_LEN		((ECCref_MAX_BITS + 7) / 8)

	typedef struct ECCrefPublicKey_st {
		unsigned int bits;
		unsigned char x[ECCref_MAX_LEN];
		unsigned char y[ECCref_MAX_LEN];
	}ECCrefPublicKey;

	typedef struct ECCrefPrivateKey_st {
		unsigned int bits;
		unsigned char K[ECCref_MAX_LEN];
	}ECCrefPrivateKey;

	typedef struct ECCCipher_st {
		unsigned char x[ECCref_MAX_LEN];
		unsigned char y[ECCref_MAX_LEN];
		unsigned char M[32];
		unsigned int L;
		unsigned char C[1024];
	}ECCCipher;

	typedef struct ECCSignature_st {
		unsigned char r[ECCref_MAX_LEN];
		unsigned char s[ECCref_MAX_LEN];
	}ECCSignature;

	/**
	*@brief	GMT 0018-2012 未定义ECCCIPHERBLOB/ECCPUBLICKEYBLOB，推测使用如下定义
	*/
	typedef ECCCipher ECCCIPHERBLOB;
	typedef ECCrefPublicKey ECCPUBLICKEYBLOB;

	typedef struct SDF_ENVELOPEDKEYBLOB {
		unsigned long ulAsymmAlgID;
		unsigned long ulSymmAlgID;
		ECCCIPHERBLOB ECCCipherBlob;
		ECCPUBLICKEYBLOB PubKey;
		unsigned char cbEncryptedPriKey[ECCref_MAX_LEN];
	}ENVELOPEDKEYBLOB, * PENVELOPEDKEYBLOB;

	/**
	*以下设备管理类函数
	*/
	/**
	* @brief 打开密码设备
	* @param	phDeviceHandle	[OUT]	返回设备句柄
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	phDeviceHandle由函数初始化并填写内容
	*/
	int SDF_OpenDevice(void** phDeviceHandle);

	/**
	* @brief 关闭密码设备，并释放相关资源
	* @param	hDeviceHandle	[IN]	已打开的设备句柄
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_CloseDevice(void* hDeviceHandle);

	/**
	* @brief	创建与密码设备的会话
	* @param	hDeviceHandle	[IN]	已打开的设备句柄
	* @param	phSessionHandle	[OUT]	返回与密码设备建立的新会话句柄
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_OpenSession(void* hDeviceHandle, void** phSessionHandle);

	/**
	* @brief	关闭与密码设备已建立的会话，并释放相关资源
	* @param	hSessionHandle	[IN]	与密码设备已建立的会话句柄
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_CloseSession(void* hSessionHandle);

	/**
	* @brief	获取密码设备能力描述
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pstDeviceInfo	[OUT]	设备能力描述信息
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_GetDeviceInfo(
		void* hSessionHandle,
		DEVICEINFO* pstDeviceInfo);

	/**
	* @brief	产生随机数
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiLength		[IN]	欲获取的随机数长度
	* @param	pucRandom		[OUT]	缓冲区指针，用于存放获取的随机数
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_GenerateRandom(
		void* hSessionHandle,
		unsigned int uiLength,
		unsigned char* pucRandom);

	/**
	* @brief	获取密码设备内部存储的指定索引私钥的使用权
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储私钥的索引值
	*									TASS补充：正数获取SM2密钥权限，负数获取RSA密钥权限
	* @param	pucPassword		[IN]	使用私钥权限的识别码，默认为a1234567
	* @param	uiPwdLength		[IN]	私钥访问控制码长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	本标准涉及密码设备存储的密钥对索引值的起始索引值为1，最大为n，
	*		密码设备的实际存储容量决定n值
	*/
	int SDF_GetPrivateKeyAccessRight(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		unsigned char* pucPassword,
		unsigned int uiPwdLength);

	/**
	* @brief	释放密码设备存储的指定索引私钥的使用授权
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储私钥的索引值
	*									TASS补充：正数释放SM2密钥权限，负数释放RSA密钥权限
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ReleasePrivateKeyAccessRight(
		void* hSessionHandle,
		unsigned int uiKeyIndex);

	/**
	*以下密钥管理类函数
	*/

	/**
	* @brief	导出密码设备内部存储的指定索引位置的签名公钥
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储的RSA密钥对索引值
	* @param	pucPublicKey	[OUT]	RSA公钥结构
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ExportSignPublicKey_RSA(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		RSArefPublicKey* pucPublicKey);

	/**
	* @brief	导出密码设备内部存储的指定索引位置的加密公钥
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储的RSA密钥对索引值
	* @param	pucPublicKey	[OUT]	RSA公钥结构
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ExportEncPublicKey_RSA(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		RSArefPublicKey* pucPublicKey);

	/**
	* @brief	请求密码设备产生指定模长的RSA密钥对（明文）
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	指定密钥模长
	* @param	pucPublicKey	[OUT]	RSA公钥结构
	*									TASS补充：若需要保存到加密机内部加密密钥对
	*											  可设置pucPublicKey->bits为索引值，此时不输出pucPrivateKey
	* @param	pucPrivateKey	[OUT]	RSA私钥结构（pucPublicKey->bits输出为私钥实际长度，可从私钥结构体的成员m、e、d依次拷贝pucPublicKey->bits长度的数据，即为真实的私钥密文数据）
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_GenerateKeyPair_RSA(
		void* hSessionHandle,
		unsigned int uiKeyBits,
		RSArefPublicKey* pucPublicKey,
		RSArefPrivateKey* pucPrivateKey);

	/**
	* @brief	生成会话密钥并用指定索引的内部加密公钥加密输出，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiIPKIndex		[IN]	密码设备内部存储公钥的索引值
	* @param	uiKeyBits		[IN]	指定产生的会话密钥长度，支持2048bits（256字节）
	* @param	pucKey			[OUT]	缓冲区指针，用于存放返回的密钥密文
	* @param	puiKeyLength	[IN/OUT]	返回的密钥密文长度
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	公钥加密数据是填充方式按照PKCS#1 v1.5的要求进行
	*/
	int SDF_GenerateKeyWithIPK_RSA(
		void* hSessionHandle,
		unsigned int uiIPKIndex,
		unsigned int uiKeyBits,
		unsigned char* pucKey,
		unsigned int* puiKeyLength,
		void** phKeyHandle);

	/**
	* @brief	生成会话密钥并用外部公钥加密输出，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyBits		[IN]	指定产生的会话密钥长度，仅支持128（16字节）
	* @param	pucPublicKey	[IN]	输入的外部RSA公钥结构
	* @param	pucKey			[OUT]	缓冲区指针，用于存放返回的密钥密文
	* @param	puiKeyLength	[OUT]	返回的密钥密文长度
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	公钥加密数据是填充方式按照PKCS#1 v1.5的要求进行
	*/
	int SDF_GenerateKeyWithEPK_RSA(
		void* hSessionHandle,
		unsigned int uiKeyBits,
		RSArefPublicKey* pucPublicKey,
		unsigned char* pucKey,
		unsigned int* puiKeyLength,
		void** phKeyHandle);

	/**
	* @brief	导入会话密钥并用内部私钥解密，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiISKIndex		[IN]	密码设备内部存储加密私钥的索引值，对应于加密时的公钥
	* @param	pucKey			[IN]	缓冲区指针，用于存放输入的密钥密文
	* @param	puiKeyLength	[IN]	输入密钥密文长度
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	填充方式与公钥加密时相同
	*/
	int SDF_ImportKeyWithISK_RSA(
		void* hSessionHandle,
		unsigned int uiISKIndex,
		unsigned char* pucKey,
		unsigned int uiKeyLength,
		void** phKeyHandle);

	/**
	* @brief	将由内部公钥加密的会话密钥转换为由外部指定的公钥加密，可用于数字信封转换
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储的内部RSA密钥对索引值
	* @param	pucPublicKey	[IN]	外部RSA公钥结构
	* @param	pucDEInput		[IN]	缓冲区指针，用于存放输入的会话密钥密文
	* @param	uiDELength		[IN]	输入的会话密钥密文长度
	* @param	pucDEOuput		[OUT]	缓冲区指针，用于存放输出的会话密钥密文
	* @param	puiDELength		[OUT]	输出的会话密钥密文长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	填充方式与公钥加密时相同
	*/
	int SDF_ExchangeDigitEnvelopeBaseOnRSA(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		RSArefPublicKey* pucPublicKey,
		unsigned char* pucDEInput,
		unsigned int uiDELength,
		unsigned char* pucDEOuput,
		unsigned int* puiDELength);

	/**
	* @brief	导出密码设备内部存储的指定索引位置的签名公钥
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储的ECC密钥对索引值
	* @param	pucPublicKey	[OUT]	ECC公钥结构
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ExportSignPublicKey_ECC(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		ECCrefPublicKey* pucPublicKey);

	/**
	* @brief	导出密码设备内部存储的指定索引位置的加密公钥
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储的ECC密钥对索引值
	* @param	pucPublicKey	[OUT]	ECC公钥结构
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ExportEncPublicKey_ECC(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		ECCrefPublicKey* pucPublicKey);

	/**
	* @brief	请求密码设备产生指定和模长的ECC密钥对
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiAlgID			[IN]	指定算法标识
	* @param	uiKeyBits		[IN]	指定密钥模长，只支持256bit（32字节）
	* @param	pucPublicKey	[OUT]	ECC公钥结构
	*									TASS补充：若需要保存到加密机内部加密密钥对
	*											  可设置pucPublicKey->bits为索引值，此时不输出pucPrivateKey
	* @param	pucPrivateKey	[OUT]	RSA私钥结构
	*									TASS补充：若需要保存到加密机内部签名密钥对，
	*											  可设置pucPrivateKey->bits为索引值，此时不输出pucPrivateKey
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_GenerateKeyPair_ECC(
		void* hSessionHandle,
		unsigned int uiAlgID,
		unsigned int uiKeyBits,
		ECCrefPublicKey* pucPublicKey,
		ECCrefPrivateKey* pucPrivateKey);

	/**
	* @brief	生成会话密钥并用指定索引的内部ECC加密公钥加密输出，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiIPKIndex		[IN]	密码设备内部存储公钥的索引值
	* @param	uiKeyBits		[IN]	指定产生的会话密钥长度，支持128bits（16字节）
	* @param	pucKey			[OUT]	缓冲区指针，用于存放返回的密钥密文
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	公钥加密数据是填充方式按照PKCS#1 v1.5的要求进行
	*/
	int SDF_GenerateKeyWithIPK_ECC(
		void* hSessionHandle,
		unsigned int uiIPKIndex,
		unsigned int uiKeyBits,
		ECCCipher* pucKey,
		void** phKeyHandle);

	/**
	* @brief	生成会话密钥并用外部ECC公钥加密输出，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyBits		[IN]	指定产生的会话密钥长度，支持128bits（16字节）
	* @param	uiAlgID			[IN]	外部ECC公钥的算法标识
	* @param	pucPublicKey	[IN]	输入的外部ECC公钥结构
	* @param	pucKey			[OUT]	缓冲区指针，用于存放返回的密钥密文
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	公钥加密数据是填充方式按照PKCS#1 v1.5的要求进行
	*/
	int SDF_GenerateKeyWithEPK_ECC(
		void* hSessionHandle,
		unsigned int uiKeyBits,
		unsigned int uiAlgID,
		ECCrefPublicKey* pucPublicKey,
		ECCCipher* pucKey,
		void** phKeyHandle);

	/**
	* @brief	导入会话密钥并用内部ECC加密私钥解密，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiISKIndex		[IN]	密码设备内部存储加密私钥的索引值，对应于加密时的公钥
	* @param	pucKey			[IN]	缓冲区指针，用于存放输入的密钥密文
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	填充方式与公钥加密时相同
	*/
	int SDF_ImportKeyWithISK_ECC(
		void* hSessionHandle,
		unsigned int uiISKIndex,
		ECCCipher* pucKey,
		void** phKeyHandle);

	/**
	* @brief	使用ECC密钥协商算法，为计算会话密钥而产生协商参数，同时返回指定索引位置的ECC公钥、临时ECC密钥对的公钥及协商句柄
	* @param	hSessionHandle			[IN]	与设备建立的会话句柄
	* @param	uiISKIndex				[IN]	密码设备内部存储加密私钥的索引值，该私钥用于参与密钥协商
	* @param	uiKeyBits				[IN]	要求协商的密钥长度，支持128bits（16字节）
	* @param	pucSponsorID			[IN]	参与密钥协商的发起方ID值
	* @param	uiSponsorIDLength		[IN]	发起方ID长度
	* @param	pucSponsorPublicKey		[OUT]	返回的发起方ECC公钥结构
	* @param	pucSponsorTmpPublicKey	[OUT]	返回的发起方临时ECC公钥结构
	* @param	phAgreementHandle		[OUT]	返回的协商句柄，用于计算协商密钥
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	为协商会话密钥，协商的发起方应首先调用本函数
	*/
	int SDF_GenerateAgreementDataWithECC(
		void* hSessionHandle,
		unsigned int uiISKIndex,
		unsigned int uiKeyBits,
		unsigned char* pucSponsorID,
		unsigned int uiSponsorIDLength,
		ECCrefPublicKey* pucSponsorPublicKey,
		ECCrefPublicKey* pucSponsorTmpPublicKey,
		void** phAgreementHandle);

	/**
	* @brief	使用ECC密钥协商算法，使用自身协商句柄和响应方的协商参数计算会话密钥，同时返回会话密钥句柄
	* @param	hSessionHandle			[IN]	与设备建立的会话句柄
	* @param	pucResponseID			[IN]	外部输入的响应方ID值
	* @param	uiResponseIDLength		[IN]	外部输入的响应方ID长度
	* @param	pucResponsePublicKey	[IN]	外部输入的响应方ECC公钥结构
	* @param	pucResponseTmpPublicKey	[IN]	外部输入的响应方临时ECC公钥结构
	* @param	hAgreementHandle		[IN]	协商句柄，用于计算协商密钥
	* @param	phKeyHandle				[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	协商的发起方获得响应方的协商参数后调用本函数，计算会话密钥，使用SM2算法计算会话密钥的过程见GM/T 0009
	*/
	int SDF_GenerateKeyWithECC(
		void* hSessionHandle,
		unsigned char* pucResponseID,
		unsigned int uiResponseIDLength,
		ECCrefPublicKey* pucResponsePublicKey,
		ECCrefPublicKey* pucResponseTmpPublicKey,
		void* hAgreementHandle,
		void** phKeyHandle);

	/**
	* @brief	使用ECC密钥协商算法，产生协商参数并计算会话密钥，同时返回产生的协商参数和密钥句柄
	* @param	hSessionHandle			[IN]	与设备建立的会话句柄
	* @param	uiISKIndex				[IN]	密码设备内部存储加密私钥的索引值，该私钥用于参与密钥协商
	* @param	uiKeyBits				[IN]	要求协商的密钥长度，支持128bits（16字节）
	* @param	pucResponseID			[IN]	响应方ID值
	* @param	uiResponseIDLength		[IN]	响应方ID长度
	* @param	pucSponsorID			[IN]	发起方ID值
	* @param	uiSponsorIDLength		[IN]	发起方ID长度
	* @param	pucSponsorPublicKey		[IN]	外部输入的发起方ECC公钥结构
	* @param	pucSponsorTmpPublicKey	[IN]	外部输入的发起方临时ECC公钥结构
	* @param	pucResponsePublicKey	[OUT]	返回的响应方ECC公钥结构
	* @param	pucResponseTmpPublicKey	[OUT]	返回的响应方临时ECC公钥结构
	* @param	phKeyHandle				[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	本函数由响应方调用。使用SM2算法计算会话密钥的过程见GM/T 0009
	*/
	int SDF_GenerateAgreementDataAndKeyWithECC(
		void* hSessionHandle,
		unsigned int uiISKIndex,
		unsigned int uiKeyBits,
		unsigned char* pucResponseID,
		unsigned int uiResponseIDLength,
		unsigned char* pucSponsorID,
		unsigned int uiSponsorIDLength,
		ECCrefPublicKey* pucSponsorPublicKey,
		ECCrefPublicKey* pucSponsorTmpPublicKey,
		ECCrefPublicKey* pucResponsePublicKey,
		ECCrefPublicKey* pucResponseTmpPublicKey,
		void** phKeyHandle);

	/**
	* @brief	将由内部公钥加密的会话密钥转换为由外部指定的公钥加密，可用于数字信封转换
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备存储的内部ECC密钥对索引值
	* @param	uiAlgID			[IN]	外部ECC公钥的算法标识
	* @param	pucPublicKey	[IN]	外部ECC公钥结构
	* @param	pucEncDataIn	[IN]	缓冲区指针，用于存放输入的会话密钥密文
	* @param	pucEncDataOut	[OUT]	缓冲区指针，用于存放输出的会话密钥密文
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	填充方式与公钥加密时相同
	*/
	int SDF_ExchangeDigitEnvelopeBaseOnECC(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		unsigned int uiAlgID,
		ECCrefPublicKey* pucPublicKey,
		ECCCipher* pucEncDataIn,
		ECCCipher* pucEncDataOut);

	/**
	* @brief	生成会话密钥并用密钥加密密钥加密输出，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyBits		[IN]	指定产生的会话密钥长度，支持128bits（16字节）
	* @param	uiAlgID			[IN]	算法标识，指定对称加密算法,仅支持SGD_SM1_ECB
	* @param	uiKEKIndex		[IN]	密码设备内部存储的加密密钥的索引值
	* @param	pucKey			[OUT]	缓冲区指针，用于存放返回的密钥密文
	* @param	puiKeyLength	[OUT]	返回的密钥密文长度
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	*									TASS补充：若需要保存到加密机内部，可使用如下方式
	*											  如保存到1号索引密钥（仅支持SM4算法），则int idx = 1; void* pIdx = &idx; 传入 &pIdx 即可
	*                                             uiKeyBits=128，导入SM4算法
	*											  此时phKeyHandle不能调用SDF_DestroyKey释放
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	加密模式使用ECB模式
	*/
	int SDF_GenerateKeyWithKEK(
		void* hSessionHandle,
		unsigned int uiKeyBits,
		unsigned int uiAlgID,
		unsigned int uiKEKIndex,
		unsigned char* pucKey,
		unsigned int* puiKeyLength,
		void** phKeyHandle);

	/**
	* @brief	导入会话密钥并用密钥加密密钥解密，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiAlgID			[IN]	算法标识，指定对称加密算法
	* @param	uiKEKIndex		[IN]	密码设备内部存储的加密密钥的索引值
	* @param	pucKey			[IN]	缓冲区指针，用于存放输入的密钥密文
	* @param	puiKeyLength	[IN]	返回的密钥密文长度
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	*									TASS补充：若需要保存到加密机内部（仅支持SM4算法），可使用如下方式
	*											  如保存到1号索引密钥，则int idx = 1; void* pIdx = &idx; 传入 &pIdx 即可
	*                                             uiKeyBits=128，导入SM4算法
	*											  此时phKeyHandle不能调用SDF_DestroyKey释放
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	加密模式使用ECB模式
	*/
	int SDF_ImportKeyWithKEK(
		void* hSessionHandle,
		unsigned int uiAlgID,
		unsigned int uiKEKIndex,
		unsigned char* pucKey,
		unsigned int uiKeyLength,
		void** phKeyHandle);

	/**
	* @brief	导入明文会话密钥，同时返回密钥句柄
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pucKey			[IN]	缓冲区指针，用于存放输入的密钥明文
	* @param	puiKeyLength	[IN]	输入的密钥明文长度
	* @param	phKeyHandle		[OUT]	返回的密钥句柄，传入前需先赋值为NULL
	*									TASS补充：若需要保存到加密机内部（仅支持SM4算法），可使用如下方式
	*											  如保存到1号索引密钥，则int idx = 1; void* pIdx = &idx; 传入 &pIdx 即可
	*                                             uiKeyBits=128，导入SM4算法
	*											  此时phKeyHandle不能调用SDF_DestroyKey释放
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ImportKey(
		void* hSessionHandle,
		unsigned char* pucKey,
		unsigned int uiKeyLength,
		void** phKeyHandle);

	/**
	* @brief	销毁会话密钥，并释放为密钥句柄分配的内存等资源
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	hKeyHandle		[IN]	输入的密钥句柄，若输入指向索引的指针则删除该内部对称密钥
	*									TASS补充：若需要删除加密机内部密钥，可使用如下方式
	*											  如删除1号索引密钥，则int idx = 1; 传入 &idx 即可
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	在对称算法运算完成后，应调用本函数销毁会话密钥
	*/
	int SDF_DestroyKey(
		void* hSessionHandle,
		void* hKeyHandle);

	/**
	*以下非对称算法运算类函数
	*/

	/**
	* @brief	指定使用外部公钥对数据进行运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pucPublicKey	[IN]	外部RSA公钥结构
	* @param	pucDataInput	[IN]	缓冲区指针，用于存放输入的数据
	* @param	uiInputLength	[IN]	输入的数据长度
	* @param	pucDataOutput	[OUT]	缓冲区指针，用于存放输出的数据
	* @param	puiOutputLength	[OUT]	输出的数据长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	数据格式由应用层封装
	*/
	int SDF_ExternalPublicKeyOperation_RSA(
		void* hSessionHandle,
		RSArefPublicKey* pucPublicKey,
		unsigned char* pucDataInput,
		unsigned int uiInputLength,
		unsigned char* pucDataOutput,
		unsigned int* puiOutputLength);

	/**
	* @brief	使用内部指定索引的公钥对数据进行运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备内部存储公钥的索引值
	* @param	pucDataInput	[IN]	缓冲区指针，用于存放输入的数据
	* @param	uiInputLength	[IN]	输入的数据长度
	* @param	pucDataOutput	[OUT]	缓冲区指针，用于存放输出的数据
	* @param	puiOutputLength	[OUT]	输出的数据长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	索引范围仅限于内部签名密钥对，数据格式由应用层封装
	*/
	int SDF_InternalPublicKeyOperation_RSA(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		unsigned char* pucDataInput,
		unsigned int uiInputLength,
		unsigned char* pucDataOutput,
		unsigned int* puiOutputLength);

	/**
	* @brief	使用内部指定索引的私钥对数据进行运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiKeyIndex		[IN]	密码设备内部存储私钥的索引值
	* @param	pucDataInput	[IN]	缓冲区指针，用于存放输入的数据
	* @param	uiInputLength	[IN]	输入的数据长度
	* @param	pucDataOutput	[OUT]	缓冲区指针，用于存放输出的数据
	* @param	puiOutputLength	[OUT]	输出的数据长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	索引范围仅限于内部签名密钥对，数据格式由应用层封装
	*/
	int SDF_InternalPrivateKeyOperation_RSA(
		void* hSessionHandle,
		unsigned int uiKeyIndex,
		unsigned char* pucDataInput,
		unsigned int uiInputLength,
		unsigned char* pucDataOutput,
		unsigned int* puiOutputLength);

	/**
		* @brief	使用外部 ECC 私钥对数据进行签名运算
		* @param	hSessionHandle	[IN]	与设备建立的会话句柄
		* @param	uiAlgID			[IN]	算法标识，指定使用的ECC算法
		* @param	pucPrivateKey	[IN]	外部 ECC 私钥结构
		* @param	pucData			[IN]	缓冲区指针，用于存放外部输入的数据
		* @param	uiDataLength	[IN]	输入的数据长度
		* @param	pucSignature	[OUT]	缓冲区指针，用于存放输入的签名值数据
		* @return
		*   @retval	0		成功
		*   @retval	非0		失败，返回错误代码
		* @note	输入数据为待签数据的杂凑值。当使用 SM2 算法时，该输入数据为待签数据经过SM2 签名预处理的结果，预处理过程见 GM/T AAAA。
		*/
	int SDF_ExternalSign_ECC(
		void* hSessionHandle,
		unsigned int uiAlgID,
		ECCrefPrivateKey* pucPrivateKey,
		unsigned char* pucData,
		unsigned int uiDataLength,
		ECCSignature* pucSignature);

	/**
	* @brief	使用外部ECC公钥对ECC签名值进行验证运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiAlgID			[IN]	算法标识，指定使用的ECC算法
	* @param	pucPublicKey	[IN]	外部ECC公钥结构
	* @param	pucData			[IN]	缓冲区指针，用于存放输入的数据
	* @param	uiDataLength	[IN]	输入的数据长度
	* @param	pucSignature	[IN]	缓冲区指针，用于存放输入的签名值数据
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	输入数据为待签数据的杂凑值。当使用SM2算法时，该输入数据位待签数据经过SM2签名预处理的结果，预处理过程见GM/T0009
	*/
	int SDF_ExternalVerify_ECC(
		void* hSessionHandle,
		unsigned int uiAlgID,
		ECCrefPublicKey* pucPublicKey,
		unsigned char* pucData,
		unsigned int uiDataLength,
		ECCSignature* pucSignature);

	/**
	* @brief	使用内部ECC公钥对ECC签名值进行验证运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiISKIndex		[IN]	密码设备内部存储的ECC签名私钥的索引值
	* @param	pucData			[IN]	缓冲区指针，用于存放外部输入的数据
	* @param	uiDataLength	[IN]	输入的数据长度
	* @param	pucSignature	[OUT]	缓冲区指针，用于存放输出的签名值数据
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	输入数据为待签数据的杂凑值。当使用SM2算法时，该输入数据位待签数据经过SM2签名预处理的结果，预处理过程见GM/T0009
	*/
	int SDF_InternalSign_ECC(
		void* hSessionHandle,
		unsigned int uiISKIndex,
		unsigned char* pucData,
		unsigned int uiDataLength,
		ECCSignature* pucSignature);

	/**
	* @brief	使用内部ECC公钥对ECC签名值进行验证运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiIPKIndex		[IN]	密码设备内部存储的ECC签名公钥的索引值
	* @param	pucData			[IN]	缓冲区指针，用于存放外部输入的数据
	* @param	uiDataLength	[IN]	输入的数据长度
	* @param	pucSignature	[IN]	缓冲区指针，用于存放输入的签名值数据
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	输入数据为待签数据的杂凑值。当使用SM2算法时，该输入数据位待签数据经过SM2签名预处理的结果，预处理过程见GM/T0009
	*/
	int SDF_InternalVerify_ECC(
		void* hSessionHandle,
		unsigned int uiIPKIndex,
		unsigned char* pucData,
		unsigned int uiDataLength,
		ECCSignature* pucSignature);

	/**
	* @brief	使用外部ECC公钥对数据进行加密运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiAlgID			[IN]	算法标识，指定使用的ECC算法
	* @param	pucPublicKey	[IN]	外部ECC公钥结构
	* @param	pucData			[IN]	缓冲区指针，用于存放输入的数据
	* @param	uiDataLength	[IN]	输入的数据长度
	* @param	pucEncData		[OUT]	缓冲区指针，用于存放输出的数据密文
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ExternalEncrypt_ECC(
		void* hSessionHandle,
		unsigned int uiAlgID,
		ECCrefPublicKey* pucPublicKey,
		unsigned char* pucData,
		unsigned int uiDataLength,
		ECCCipher* pucEncData);

	/**
	* @brief	使用外部 ECC 私钥进行解密运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiAlgID			[IN]	算法标识，指定使用的ECC算法
	* @param	pucPrivateKey	[IN]	外部ECC私钥结构
	* @param	pucEncData		[IN]	缓冲区指针，用于存放输入的数据密文
	* @param	pucData			[OUT]	缓冲区指针，用于存放输出的数据明文
	* @param	puiDataLength	[OUT]	缓冲区指针，输出的数据明文长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ExternalDecrypt_ECC(
		void* hSessionHandle,
		unsigned int uiAlgID,
		ECCrefPrivateKey* pucPrivateKey,
		ECCCipher* pucEncData,
		unsigned char* pucData,
		unsigned int* puiDataLength);

	/**
	*以下函数并未包含在GM/T 0018中，但从功能完整性考虑，定义为扩展实现
	*/
	/**
	* @brief	使用内部ECC私钥对数据进行解密运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiISKIndex		[IN]	密码设备内部存储的ECC解密私钥的索引值
	* @param	pucEncData		[IN]	缓冲区指针，用于存放输入的数据密文
	* @param	pucData			[OUT]	缓冲区指针，用于存放输出的数据
	* @param	uiDataLength	[OUT]	输出的数据长度
	*
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_InternalDecrypt_ECC(
		void* hSessionHandle,
		unsigned int uiISKIndex,
		ECCCipher* pucEncData,
		unsigned char* pucData,
		unsigned int* puiDataLength);

	/**
	* @brief	使用内部ECC公钥对数据进行加密运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiIPKIndex		[IN]	密码设备内部存储的ECC加密公钥的索引值
	* @param	pucData			[IN]	缓冲区指针，用于存放输入的数据
	* @param	uiDataLength	[IN]	输入的数据长度
	* @param	pucEncData		[OUT]	缓冲区指针，用于存放输出的数据密文
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_InternalEncrypt_ECC(
		void* hSessionHandle,
		unsigned int uiIPKIndex,
		unsigned char* pucData,
		unsigned int uiDataLength,
		ECCCipher* pucEncData);

	/**
	*对称算法运算类函数
	*/
	/**
	* @brief	使用指定的密钥句柄和IV对数据进行对称加密运算
	* @param	hSessionHandle		[IN]		与设备建立的会话句柄
	* @param	hKeyHandle			[IN]		指定的密钥句柄
	*											TASS补充：若需要使用加密机内部索引的对称密钥，可使用如下方式
	*													  如使用1号索引密钥，则int idx = 1; 传入 &idx 即可
	* @param	uiAlgID				[IN]		算法标识，指定的对称加密算法
	* @param	pucIV				[IN/OUT]	缓冲区指针，用于存放输入和返回的IV数据
	* @param	pucData				[IN]		缓冲区指针，用于存放输入的数据明文
	* @param	uiDataLength		[IN]		输入的数据明文长度
	* @param	pucEncData			[OUT]		缓冲区指针，用于存放输出的数据密文
	* @param	puiEncDataLength	[OUT]		输出的数据密文长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	此函数不对数据进行填充处理，输入的数据必须是指定算法分组长度的整数倍
	*/
	int SDF_Encrypt(
		void* hSessionHandle,
		void* hKeyHandle,
		unsigned int uiAlgID,
		unsigned char* pucIV,
		unsigned char* pucData,
		unsigned int uiDataLength,
		unsigned char* pucEncData,
		unsigned int* puiEncDataLength);

	/**
	* @brief	使用指定的密钥句柄和IV对数据进行对称解密运算
	* @param	hSessionHandle		[IN]		与设备建立的会话句柄
	* @param	hKeyHandle			[IN]		指定的密钥句柄
	*											TASS补充：若需要使用加密机内部索引的对称密钥，可使用如下方式
	*													  如使用1号索引密钥，则int idx = 1; 传入 &idx 即可
	* @param	uiAlgID				[IN]		算法标识，指定的对称加密算法
	* @param	pucIV				[IN/OUT]	缓冲区指针，用于存放输入和返回的IV数据
	* @param	pucEncData			[IN]		缓冲区指针，用于存放输入的数据密文
	* @param	puiEncDataLength	[IN]		输入的数据密文长度
	* @param	pucData				[OUT]		缓冲区指针，用于存放输出的数据明文
	* @param	puiDataLength		[OUT]		输出的数据明文长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	此函数不对数据进行填充处理，输入的数据必须是指定算法分组长度的整数倍
	*/
	int SDF_Decrypt(
		void* hSessionHandle,
		void* hKeyHandle,
		unsigned int uiAlgID,
		unsigned char* pucIV,
		unsigned char* pucEncData,
		unsigned int uiEncDataLength,
		unsigned char* pucData,
		unsigned int* puiDataLength);

	/**
	* @brief	使用指定的密钥句柄和IV对数据进行对称加密运算
	* @param	hSessionHandle	[IN]		与设备建立的会话句柄
	* @param	hKeyHandle			[IN]		指定的密钥句柄
	*											TASS补充：若需要使用加密机内部索引的密钥，可使用如下方式
	*													  如使用1号索引密钥，则int idx = 1; 传入 &idx 即可
	* @param	uiAlgID			[IN]		算法标识，指定的对称加密算法
	* @param	pucIV			[IN/OUT]	缓冲区指针，用于存放输入和返回的IV数据
	* @param	pucData			[IN]		缓冲区指针，用于存放输入的数据明文
	* @param	uiDataLength	[IN]		输入的数据明文长度
	* @param	pucMAC			[OUT]		缓冲区指针，用于存放输出的MAC值
	* @param	puiMACLength	[OUT]		输出的MAC值长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	此函数不对数据进行填充处理，输入的数据必须是指定算法分组长度的整数倍
	*/
	int SDF_CalculateMAC(
		void* hSessionHandle,
		void* hKeyHandle,
		unsigned int uiAlgID,
		unsigned char* pucIV,
		unsigned char* pucData,
		unsigned int uiDataLength,
		unsigned char* pucMAC,
		unsigned int* puiMACLength);

	/**
	*杂凑运算类函数
	*/

	/**
	* @brief	三步式数据杂凑运算第一步
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	uiAlgID			[IN]	指定杂凑算法标识
	* @param	pucPublicKey	[IN]	签名者公钥。当uiAlgID位SGD_SM3时有效
	* @param	pucID			[IN]	签名者的ID值。当uiAlgID位SGD_SM3时有效
	* @param	uiIDLength		[IN]	签名者ID长度。当uiAlgID位SGD_SM3时有效
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note	uiIDLength非零且uiAlgID为SGD_SM3时，函数执行SM2的预处理1操作。计算过程见GM/T 0009
	*/
	int SDF_HashInit(
		void* hSessionHandle,
		unsigned int uiAlgID,
		ECCrefPublicKey* pucPublicKey,
		unsigned char* pucID,
		unsigned int uiIDLength);

	/**
	* @brief	三步式数据杂凑运算第二步，对输入的明文进行杂凑运算
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pucData			[IN]	缓冲区指针，用于存放输入的数据明文
	* @param	uiDataLength	[IN]	输入的数据明文长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_HashUpdate(
		void* hSessionHandle,
		unsigned char* pucData,
		unsigned int uiDataLength);

	/**
	* @brief	三步式数据杂凑运算第三步，杂凑运算结束返回杂凑数据并清除中间数据
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pucHash			[OUT]	缓冲区指针，用于存放输出的杂凑数据
	* @param	puiHashLength	[OUT]	返回的杂凑数据长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_HashFinal(
		void* hSessionHandle,
		unsigned char* pucHash,
		unsigned int* puiHashLength);

	/**
	*用户文件操作类函数
	*/

	/**
	* @brief	在密码设备内部创建用于存储用户数据的文件
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pucFileName		[IN]	缓冲区指针，用于存放输入的文件名，最大长度128字节。
	* @param	uiNameLen		[IN]	文件名长度
	* @param	uiFileSize		[IN]	文件所占存储空间的长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_CreateFile(
		void* hSessionHandle,
		unsigned char* pucFileName,
		unsigned int uiNameLen,
		unsigned int uiFileSize);

	/**
	* @brief	读取在密码设备内部存储用户数据的文件的内容
	* @param	hSessionHandle	[IN]		与设备建立的会话句柄
	* @param	pucFileName		[IN]		缓冲区指针，用于存放输入的文件名，最大长度128字节
	* @param	uiNameLen		[IN]		文件名长度
	* @param	uiOffset		[IN]		指定读取文件时的偏移值
	* @param	puiFileLength	[IN/OUT]	入参时指定读取文件内容的长度；出参时返回实际读取文件内容的长度
	* @param	pucBuffer		[OUT]		缓冲区指针，用于存放读取的文件数据
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_ReadFile(
		void* hSessionHandle,
		unsigned char* pucFileName,
		unsigned int uiNameLen,
		unsigned int uiOffset,
		unsigned int* puiFileLength,
		unsigned char* pucBuffer);

	/**
	* @brief	向密码设备内部存储用户数据的文件中写入内容
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pucFileName		[IN]	缓冲区指针，用于存放输入的文件名，最大长度128字节
	* @param	uiNameLen		[IN]	文件名长度
	* @param	uiOffset		[IN]	指定写入文件时的偏移值
	* @param	uiFileLength	[IN]	指定写入文件内容的长度
	* @param	pucBuffer		[IN]	缓冲区指针，用于存放输入的写文件数据
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_WriteFile(
		void* hSessionHandle,
		unsigned char* pucFileName,
		unsigned int uiNameLen,
		unsigned int uiOffset,
		unsigned int uiFileLength,
		unsigned char* pucBuffer);

	/**
	* @brief	在密码设备内部创建用于存储用户数据的文件
	* @param	hSessionHandle	[IN]	与设备建立的会话句柄
	* @param	pucFileName		[IN]	缓冲区指针，用于存放输入的文件名，最大长度128字节
	* @param	uiNameLen		[IN]	文件名长度
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	*/
	int SDF_DeleteFile(
		void* hSessionHandle,
		unsigned char* pucFileName,
		unsigned int uiNameLen);

	
	/*
	*	以下接口仅作为扩展功能使用
	*/
	/**
	* @brief 通过指定id打开设备，并获取设备句柄
	*
	* @param id				[in]	要打开的设备ID
	* @param phDevice		[out]	返回设备句柄
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	int TassGetDeviceHandleByID(unsigned char id[16], void** phDevice);

	/**
	* @brief	设置配置文件路径
	* @param	cfgPath		[IN]	配置文件路径，不包含配置文件名字
	*
	* @return
	*   @retval	0		成功
	*   @retval	非0		失败，返回错误代码
	* @note cfgPath只传路径，不用传配置文件的名字
	*/
	int TassSetCfgPath(const char* cfgPath);

		/**
		*函数返回代码定义
		*/
#define SDR_OK					0X0							//操作成功
#define SDR_BASE				0X01000000					//错误码基础值
#define SDR_UNKNOWERR			SDR_BASE + 0X00000001		//未知错误
#define SDR_NOTSUPPORT			SDR_BASE + 0X00000002		//不支持的接口调用
#define SDR_COMMFAIL			SDR_BASE + 0X00000003		//与设备通讯失败
#define SDR_HARDFAIL			SDR_BASE + 0X00000004		//运算模块无响应
#define SDR_OPENDEVICE			SDR_BASE + 0X00000005		//打开设备失败
#define SDR_OPENSESSION			SDR_BASE + 0X00000006		//创建会话失败
#define SDR_PARDENY				SDR_BASE + 0X00000007		//无私钥使用权限
#define SDR_KEYNOTEXIST			SDR_BASE + 0X00000008		//不存在的密钥调用
#define SDR_ALGNOTSUPPORT		SDR_BASE + 0X00000009		//不支持的算法调用
#define SDR_ALGMODNOTSUPPORT	SDR_BASE + 0X0000000A		//不支持的算法模式调用
#define SDR_PKOPERR				SDR_BASE + 0X0000000B		//公钥运算失败
#define SDR_SKOPERR				SDR_BASE + 0X0000000C		//私钥运算失败
#define SDR_SIGNERR				SDR_BASE + 0X0000000D		//签名运算失败
#define SDR_VERIFYERR			SDR_BASE + 0X0000000E		//验证签名失败
#define SDR_SYMOPERR			SDR_BASE + 0X0000000F		//对称算法运算失败
#define SDR_STEPERR				SDR_BASE + 0X00000010		//多步运算步骤错误
#define SDR_FILESIZEERR			SDR_BASE + 0X00000011		//文件长度超出限制
#define SDR_FILENOEXIST			SDR_BASE + 0X00000012		//指定的文件不存在
#define SDR_FILEOFSERR			SDR_BASE + 0X00000013		//文件起始位置错误
#define SDR_KEYTYPEERR			SDR_BASE + 0X00000014		//密钥类型错误
#define SDR_KEYERR				SDR_BASE + 0X00000015		//密钥错误
#define SDR_ENCDATAERR			SDR_BASE + 0X00000016		//ECC加密数据错误
#define SDR_RANDERR				SDR_BASE + 0X00000017		//随机数产生失败
#define SDR_PRKRERR				SDR_BASE + 0X00000018		//私钥使用权限获取失败
#define SDR_MACERR				SDR_BASE + 0X00000019		//MAC运算失败
#define SDR_FILEEXISTS			SDR_BASE + 0X0000001A		//指定文件已存在
#define SDR_FILEWERR			SDR_BASE + 0X0000001B		//文件写入失败
#define SDR_NOBUFFER			SDR_BASE + 0X0000001C		//存储空间不足
#define SDR_INARGERR			SDR_BASE + 0X0000001D		//输入参数错误
#define SDR_OUTARGERR			SDR_BASE + 0X0000001E		//输出参数错误

#define TASSR_OK				0X0							//操作成功
#define TASSR_BASE				0X02000000					//天安错误码基础值
#define TASSR_UNKNOWERR			TASSR_BASE + 0X00000001		//未知错误
#define TASSR_BUFFTOOSMALL		TASSR_BASE + 0X00000002		//缓冲区不足

#ifdef __cplusplus
}
#endif
