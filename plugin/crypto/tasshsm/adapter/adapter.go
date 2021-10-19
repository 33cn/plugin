package adapter

//#cgo CFLAGS: -I../include
//#cgo LDFLAGS: -L../linux64 -lTassSDF4PCIeSM -ltass_pcie_api -lTassSDF4PCIeSM
//#cgo LDFLAGS: -ldl -lpthread -lc
//#include <stdlib.h>
//#include <stdio.h>
//#include <SDF4PCIeSM.h>
//#include <TassAPI4PCIeSM.h>
//unsigned char SM2ID[16] = {0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38};
//unsigned char sm3Hash[32] = { 0 };
//unsigned char RS[64] = { 0 };
//void* g_hDev = NULL, *g_hSess = NULL;
//const unsigned char*skCipherByKek = NULL;
//unsigned char*RSPtr = RS;
//unsigned int signatureLen = 64;
//unsigned int *signatureLenPtr = &signatureLen;
//int setupHSM()
//{
//    auto rt = SDF_OpenDevice(&g_hDev);
//    if (rt) {
//        printf("SDF_OpenDevice failed %#08x\n", rt);
//        return rt;
//    }
//    rt = SDF_OpenSession(g_hDev, &g_hSess);
//    if (rt) {
//        printf("SDF_OpenSession failed %#08x\n", rt);
//        SDF_CloseDevice(g_hDev);
//        return rt;
//    }
//}
//int pass_voidHandle(void *handle) {
//    (void)handle;
//}
//void closeHSM() {
//    SDF_CloseSession(g_hSess);
//    SDF_CloseDevice(g_hDev);
//}
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/33cn/chain33/common/crypto"
)

var (
//g_hDev      *C.int
//g_hSess     *C.void
//g_passwd    []byte
//g_passwdLen int

//SM2ID   = []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}
)

const (
	SDF_Success = C.int(0)
	SM2IDSize   = C.uint(16)
)

//OpenHSMSession:打开TASS HSM PCIe设备并建立session
func OpenHSMSession() error {
	C.pass_voidHandle(C.g_hDev)
	if rt := C.setupHSM(); int(rt) != 0 {
		return errors.New(fmt.Sprintf("Failde to setup HSM with error code:%#08x", int(rt)))
	}
	return nil
}

//CloseHSMSession:关闭设备
func CloseHSMSession() error {
	C.closeHSM()
	return nil
}

//获取私钥访问权限
func GetPrivateKeyAccessRight(passwordStr string, keyIndex int) error {
	passwd := (*C.uchar)(unsafe.Pointer(C.CString(passwordStr)))
	defer C.free(unsafe.Pointer(passwd))
	passwdLen := len(passwordStr)
	rt := C.TassGetPrivateKeyAccessRight(C.g_hSess, C.uint(keyIndex), C.TA_ALG_ECC_SECP_256K1, passwd, C.uint(passwdLen))
	if SDF_Success != rt {
		return errors.New(fmt.Sprintf("GetPrivateKeyAccessRight failed %#08x", int(rt)))
	}
	return nil
}

//释放私钥访问权限
func ReleaeAccessRight(keyIndex int) error {
	rt := C.SDF_ReleasePrivateKeyAccessRight(C.g_hSess, C.uint(keyIndex))
	if SDF_Success != rt {
		return errors.New(fmt.Sprintf("ReleaeAccessRight failed %#08x", int(rt)))
	}
	return nil
}

//获取签名
func SignSecp256k1(msg []byte, keyIndex int) (signatureR, signatureS []byte, err error) {
	hash := crypto.Sha256(msg)
	hash2sign := (*C.uchar)(C.CBytes(hash))
	defer C.free(unsafe.Pointer(hash2sign))

	rt := C.TassECCPrivateKeySign_Eth(C.g_hSess, C.TA_ALG_ECC_SECP_256K1, C.uint(keyIndex), C.uint(256), C.skCipherByKek, C.uint(0), hash2sign, C.uint(32), C.RSPtr, C.signatureLenPtr)
	if SDF_Success != rt {
		return nil, nil, errors.New(fmt.Sprintf("TassECCPrivateKeySign failed %#08x", int(rt)))
	}

	r := C.GoBytes(unsafe.Pointer(&C.RS[0]), C.int(32))
	s := C.GoBytes(unsafe.Pointer(&C.RS[32]), C.int(32))
	return r, s, nil
}

func SignSM2Internal(msg []byte, keyIndex int) (signatureR, signatureS []byte, err error) {
	var signPubKey C.ECCrefPublicKey
	rt := C.SDF_ExportSignPublicKey_ECC(C.g_hSess, C.uint(keyIndex), &signPubKey)
	if SDF_Success != rt {
		return nil, nil, errors.New(fmt.Sprintf("SDF_ExportSignPublicKey_ECC failed:%08x", int(rt)))
	}
	//fmt.Println("SDF_ExportSignPublicKey_ECC success for chain33 verification\n")
	//fmt.Printf("chain33 PubKey.bits: %d\n", int(signPubKey.bits))
	//fmt.Printf("chain33 PubKey.x: %s\n", common.ToHex(C.GoBytes(unsafe.Pointer(&signPubKey.x[0]), C.int(64))))
	//fmt.Printf("chain33 PubKey.y: %s\n", common.ToHex(C.GoBytes(unsafe.Pointer(&signPubKey.y[0]), C.int(64))))

	rt = C.SDF_HashInit(C.g_hSess, C.SGD_SM3, &signPubKey, &C.SM2ID[0], SM2IDSize)
	if SDF_Success != rt {
		return nil, nil, errors.New(fmt.Sprintf("SDF_HashInit failed %#08x", int(rt)))
	}
	//fmt.Println("SDF_HashInit success\n")

	msg2C := (*C.uchar)(C.CBytes(msg))
	defer C.free(unsafe.Pointer(msg2C))
	rt = C.SDF_HashUpdate(C.g_hSess, msg2C, C.uint(len(msg)))
	if SDF_Success != rt {
		return nil, nil, errors.New(fmt.Sprintf("SDF_HashUpdate failed %#08x", int(rt)))
	}
	//fmt.Println("SDF_HashUpdate success\n")

	hashlen := C.uint(32)
	rt = C.SDF_HashFinal(C.g_hSess, &C.sm3Hash[0], &hashlen)
	if SDF_Success != rt {
		return nil, nil, errors.New(fmt.Sprintf("SDF_HashFinal failed %#08x", int(rt)))
	}
	//fmt.Printf("SDF_HashFinal success! sm3Hash = %s\n", common.ToHex(C.GoBytes(unsafe.Pointer(&C.sm3Hash[0]), C.int(32))))
	//0x726106a793a45af8ee1b9b6781b87391e4b49304a5f12f6dff82c7edb5f58390

	var sign C.ECCSignature
	rt = C.SDF_InternalSign_ECC(C.g_hSess, C.uint(keyIndex), &C.sm3Hash[0], hashlen, &sign)
	if SDF_Success != rt {
		return nil, nil, errors.New(fmt.Sprintf("SDF_InternalSign_ECC failed %#08x", int(rt)))
	}
	//fmt.Println("SDF_InternalSign_ECC success! ")

	r := C.GoBytes(unsafe.Pointer(&sign.r[0]), C.int(64))
	s := C.GoBytes(unsafe.Pointer(&sign.s[0]), C.int(64))
	return r, s, nil
}

func MakeDERsignature(rb, sb []byte) []byte {
	if rb[0] > 0x7F {
		rb = append([]byte{0}, rb...)
	}

	if sb[0] > 0x7F {
		sb = append([]byte{0}, sb...)
	}
	// total length of returned signature is 1 byte for each magic and
	// length (6 total), plus lengths of r and s
	length := 6 + len(rb) + len(sb)
	b := make([]byte, length)

	b[0] = 0x30
	b[1] = byte(length - 2)
	b[2] = 0x02
	b[3] = byte(len(rb))
	offset := copy(b[4:], rb) + 4
	b[offset] = 0x02
	b[offset+1] = byte(len(sb))
	copy(b[offset+2:], sb)
	return b
}
