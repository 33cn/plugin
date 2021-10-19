#pragma once
#ifdef __cplusplus
extern "C" {
#endif
#define TA_DEVICE_ID_SIZE		16

	typedef enum {
		TA_DEV_STATE_INIT = 0x0,		//初始状态，无密钥
		TA_DEV_STATE_WORK = 0x1,		//工作状态
		TA_DEV_STATE_MNG = 0x2,			//管理状态
	}TassDevState;

	typedef enum {
		TA_BOOL_FALSE = 0,	//私钥
		TA_BOOL_TRUE = !TA_BOOL_FALSE,	//公钥
	}TassBool;

	typedef enum {
		TA_DEV_KEY_PLATFORM = 0,	//管理平台密钥（公钥）
		TA_DEV_KEY_SIGN = 1,		//设备签名密钥
		TA_DEV_KEY_ENC = 2,			//设备加密密钥
		TA_DEV_KEY_KEK = 3,			//设备KEK
	}TassDevKeyType;

	typedef enum {
		TA_ALG_SM4 = 0,
		TA_ALG_SM1 = 1,
		TA_ALG_AES = 5,
		TA_ALG_DES = 6,
		TA_ALG_SM7 = 7,
		TA_ALG_SM2 = 2,
		TA_ALG_ECC_SECP_256R1 = 3,
		TA_ALG_ECC_SECP_256K1 = 8,
		TA_ALG_RSA = 4,
		TA_ALG_HMAC = 9,
	}TassAlg;

	typedef enum {
		TA_RSA_E_3 = 3,
		TA_RSA_E_65537 = 65537,
	}TassRSA_E;

	typedef enum {
		TA_SYMM_ECB_ENC = 0x00000000,
		TA_SYMM_ECB_DEC = 0x00000001,
		TA_SYMM_CBC_ENC = 0x00000100,
		TA_SYMM_CBC_DEC = 0x10000101,
		TA_SYMM_CFB_ENC = 0x00000200,
		TA_SYMM_CFB_DEC = 0x10000201,
		TA_SYMM_OFB_ENC = 0x00000300,
		TA_SYMM_OFB_DEC = 0x10000301,
		TA_SYMM_MAC = 0x00000400,
	}TassSymmOp;

	typedef enum {
		TA_SM2_KEY_EXCHANGE_SPONSOR = 0,
		TA_SM2_KEY_EXCHANGE_RESPONSE = 1,
	}TassSM2KeyExchangeRole;

	typedef enum {
		TA_ASYM_SIGN = 0,
		TA_ASYM_ENC = 1,
		TA_ASYM_KEY_EX = 2,
	}TassAsymKeyUsage;

	typedef enum {
		TA_ASYM_ENCRYPT = 0,
		TA_ASYM_DECRYPT = 1,
	}TassSymmKeyUsage;

	typedef enum {
		TA_FLASH_2K = 0,
		TA_FLASH_32K = 1,
	}TassFlashFlag;

	typedef enum {
		TA_FLASH_GET_SIZE = 0,
		TA_FLASH_READ = 1,
		TA_FLASH_WRITE = 2,
		TA_FLASH_ERASE = 3,
	}TassFlashOp;

	typedef struct {
		unsigned char sigHead[16];				//签名数据头
		unsigned char hwVer[4];					//硬件版本
		unsigned char fpgaVer[4];				//FPGA版本
		unsigned char keyMngChipVer[4];			//密管芯片版本
		unsigned char devId[16];				//设备唯一标识，不可指定
		unsigned char devSn[4];					//设备序列号，可指定
		unsigned char fpgaHwCv[32];				//FPGA固件校验值
		unsigned char devSignKeyPairInfo[4];	//设备签名密钥对信息
		unsigned char platformPkInfo[4];		//管理平台公钥信息
		unsigned char devEncKeyPairInfo[4];		//设备加密密钥对信息
		unsigned char devKEKInfo[4];			//设备本地保护密钥信息
		unsigned char kekCv[16];				//设备本地保护密钥校验值
		unsigned char adminPkAblity[4];			//管理员公钥存储能力，大端格式
		unsigned char curAdminPkNum[4];			//当前管理员公钥个数，大端格式
		unsigned char adminPkInfo[5][4];		//管理员公钥信息5 * 4Bytes= 20Bytes
		unsigned char devState[4];				//设备状态，大端格式
		unsigned char temp[8];					//温度
		unsigned char valtage[8];				//电压
		unsigned char current[8];				//电流
		unsigned char selfCheckCycle[4];		//自检周期，大端格式
		unsigned char lastSelfCheckInfo[8];		//上次自检信息
		unsigned char channelNum[4];			//通道总数，大端格式
		unsigned char channelAlg[256];			//通道算法
		unsigned char sig[64];					//设备签名公钥对上述数据的签名
	}TassDevInfo;

	/**
	* @brief 签名回调函数，用于需要私钥签名的接口调用
	*
	* @param data		[in]	待签名数据
	* @param dataLen	[in]	data长度
	* @param sig		[out]	签名值
	*
	* @return 成功返回0，失败返回非0
	*
	*/
	typedef int (*TassSignCb)(const unsigned char* data, int len, unsigned char sig[64]);
#ifdef __cplusplus
}
#endif
