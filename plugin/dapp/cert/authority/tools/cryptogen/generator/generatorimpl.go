// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package generator

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"path/filepath"

	"fmt"

	"github.com/33cn/plugin/plugin/dapp/cert/authority/tools/cryptogen/factory/csp"
	"github.com/33cn/plugin/plugin/dapp/cert/authority/tools/cryptogen/generator/utils"
	ut "github.com/33cn/plugin/plugin/dapp/cert/authority/utils"
	ty "github.com/33cn/plugin/plugin/dapp/cert/types"
	"github.com/tjfoc/gmsm/sm2"
)

// EcdsaCA ecdsa CA结构
type EcdsaCA struct {
	Name       string
	Signer     crypto.Signer
	SignCert   *x509.Certificate
	CertConfig *CertConfig
}

// SM2CA SM2 CA结构
type SM2CA struct {
	Name       string
	Signer     crypto.Signer
	SignCert   *sm2.Certificate
	Sm2Key     csp.Key
	CertConfig *CertConfig
}

// NewCA 根据类型生成CA生成器
func NewCA(baseDir string, cacfg *CertConfig, signType int) (CAGenerator, error) {
	if signType == ty.AuthECDSA {
		return newEcdsaCA(baseDir, cacfg)
	} else if signType == ty.AuthSM2 {
		return newSM2CA(baseDir, cacfg)
	} else {
		return nil, fmt.Errorf("Invalid sign type")
	}
}

func newEcdsaCA(baseDir string, certConfig *CertConfig) (*EcdsaCA, error) {
	err := os.MkdirAll(baseDir, 0750)
	if err != nil {
		return nil, err
	}

	var ca *EcdsaCA
	priv, signer, err := utils.GeneratePrivateKey(baseDir, csp.ECDSAP256KeyGen)
	if err != nil {
		return nil, err
	}

	ecPubKey, err := utils.GetECPublicKey(priv)
	if err != nil {
		return nil, err
	}

	template := x509Template(certConfig.CA.Expire)
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageDigitalSignature |
		x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}

	subject := pkix.Name{
		Country:  []string{certConfig.CA.Country},
		Locality: []string{certConfig.CA.Locality},
		Province: []string{certConfig.CA.Province},
	}
	subject.CommonName = certConfig.CA.CommonName

	template.Subject = subject
	template.SubjectKeyId = priv.SKI()
	template.PublicKey = ecPubKey

	x509Cert, err := genCertificateECDSA(baseDir, certConfig.Name, &template, &template, signer)
	if err != nil {
		return nil, err
	}
	ca = &EcdsaCA{
		Name:       certConfig.Name,
		Signer:     signer,
		SignCert:   x509Cert,
		CertConfig: certConfig,
	}

	return ca, nil
}

// SignCertificate 证书签名
func (ca *EcdsaCA) SignCertificate(baseDir, fileName string, sans []string, pub interface{}, isCA bool) (*x509.Certificate, error) {
	template := x509Template(ca.CertConfig.CA.Expire)
	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageDigitalSignature |
			x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
		template.SubjectKeyId = ut.SKI(pub.(*ecdsa.PublicKey).Curve, pub.(*ecdsa.PublicKey).X, pub.(*ecdsa.PublicKey).Y)
	} else {
		template.KeyUsage = x509.KeyUsageDigitalSignature
		template.ExtKeyUsage = []x509.ExtKeyUsage{}
	}

	subject := pkix.Name{
		Country:  []string{ca.CertConfig.CA.Country},
		Locality: []string{ca.CertConfig.CA.Locality},
		Province: []string{ca.CertConfig.CA.Province},
	}
	subject.CommonName = ca.CertConfig.CA.CommonName

	template.Subject = subject
	template.DNSNames = sans
	template.PublicKey = pub

	cert, err := genCertificateECDSA(baseDir, fileName, &template, ca.SignCert, ca.Signer)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// GenerateLocalOrg 生成组织证书
func (ca *EcdsaCA) GenerateLocalOrg(baseDir, fileName string, orgCfg *CertConfig) (CAGenerator, error) {
	err := createFolderStructure(baseDir, true)
	if err != nil {
		return nil, err
	}

	keystore := filepath.Join(baseDir, "keystore")
	priv, signer, err := utils.GeneratePrivateKey(keystore, csp.ECDSAP256KeyGen)
	if err != nil {
		return nil, err
	}

	ecPubKey, err := utils.GetECPublicKey(priv)
	if err != nil {
		return nil, err
	}

	cert, err := ca.SignCertificate(filepath.Join(baseDir, "signcerts"), fileName, []string{}, ecPubKey, true)
	if err != nil || cert == nil {
		return nil, err
	}

	err = x509Export(filepath.Join(baseDir, "cacerts", x509Filename(fileName)), ca.SignCert.Raw)
	if err != nil {
		return nil, err
	}

	err = x509Export(filepath.Join(baseDir, "intermediatecerts", x509Filename(orgCfg.Name)), cert.Raw)
	if err != nil {
		return nil, err
	}

	orgCA := &EcdsaCA{
		Name:       ca.Name,
		Signer:     signer,
		SignCert:   cert,
		CertConfig: orgCfg,
	}
	return orgCA, nil
}

// GenerateLocalUser 生成本地用户
func (ca *EcdsaCA) GenerateLocalUser(baseDir, fileName string) error {
	err := createFolderStructure(baseDir, false)
	if err != nil {
		return err
	}

	keystore := filepath.Join(baseDir, "keystore")
	priv, _, err := utils.GeneratePrivateKey(keystore, csp.ECDSAP256KeyGen)
	if err != nil {
		return err
	}

	ecPubKey, err := utils.GetECPublicKey(priv)
	if err != nil {
		return err
	}

	cert, err := ca.SignCertificate(filepath.Join(baseDir, "signcerts"), fileName, []string{}, ecPubKey, false)
	if err != nil || cert == nil {
		return err
	}

	err = x509Export(filepath.Join(baseDir, "cacerts", x509Filename(ca.Name)), ca.SignCert.Raw)
	return err
}

func x509Template(expire int) x509.Certificate {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	expiry := time.Duration(expire) * 24 * time.Hour
	notBefore := time.Now().Add(-5 * time.Minute).UTC()

	x509 := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(expiry).UTC(),
		BasicConstraintsValid: true,
	}
	return x509
}

func genCertificateECDSA(baseDir, fileName string, template, parent *x509.Certificate, priv interface{}) (*x509.Certificate, error) {
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, template.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	fullName := filepath.Join(baseDir, fileName+"-cert.pem")
	certFile, err := os.Create(fullName)
	if err != nil {
		return nil, err
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return nil, err
	}

	x509Cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}
	return x509Cert, nil
}

func newSM2CA(baseDir string, certConfig *CertConfig) (*SM2CA, error) {
	var ca *SM2CA
	priv, signer, err := utils.GeneratePrivateKey(baseDir, csp.SM2P256KygGen)
	if err != nil {
		return nil, err
	}

	smPubKey, err := utils.GetSM2PublicKey(priv)
	if err != nil {
		return nil, err
	}

	template := x509Template(certConfig.CA.Expire)
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageDigitalSignature |
		x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}

	subject := pkix.Name{
		Country:  []string{certConfig.CA.Country},
		Locality: []string{certConfig.CA.Locality},
		Province: []string{certConfig.CA.Province},
	}
	subject.CommonName = certConfig.CA.CommonName

	template.Subject = subject
	template.SubjectKeyId = priv.SKI()

	sm2cert := utils.ParseX509CertificateToSm2(&template)
	sm2cert.PublicKey = smPubKey
	x509Cert, err := genCertificateGMSM2(baseDir, certConfig.Name, sm2cert, sm2cert, signer)
	if err != nil {
		return nil, err
	}

	ca = &SM2CA{
		Name:       certConfig.Name,
		Signer:     signer,
		SignCert:   x509Cert,
		Sm2Key:     priv,
		CertConfig: certConfig,
	}
	return ca, nil
}

// SignCertificate 证书签名
func (ca *SM2CA) SignCertificate(baseDir, fileName string, sans []string, pub interface{}, isCA bool) (*x509.Certificate, error) {
	template := x509Template(ca.CertConfig.CA.Expire)
	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageDigitalSignature |
			x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
		template.SubjectKeyId = ut.SKI(pub.(*sm2.PublicKey).Curve, pub.(*sm2.PublicKey).X, pub.(*sm2.PublicKey).Y)
	} else {
		template.KeyUsage = x509.KeyUsageDigitalSignature
		template.ExtKeyUsage = []x509.ExtKeyUsage{}
	}

	subject := pkix.Name{
		Country:  []string{ca.CertConfig.CA.Country},
		Locality: []string{ca.CertConfig.CA.Locality},
		Province: []string{ca.CertConfig.CA.Province},
	}
	subject.CommonName = ca.CertConfig.CA.CommonName

	template.Subject = subject
	template.DNSNames = sans
	template.PublicKey = pub

	sm2Tpl := utils.ParseX509CertificateToSm2(&template)
	cert, err := genCertificateGMSM2(baseDir, fileName, sm2Tpl, ca.SignCert, ca.Signer)
	if err != nil {
		return nil, err
	}

	return utils.ParseSm2CertificateToX509(cert), nil
}

// GenerateLocalOrg 生成组织证书
func (ca *SM2CA) GenerateLocalOrg(baseDir, fileName string, orgCfg *CertConfig) (CAGenerator, error) {
	err := createFolderStructure(baseDir, true)
	if err != nil {
		return nil, err
	}

	keystore := filepath.Join(baseDir, "keystore")
	priv, signer, err := utils.GeneratePrivateKey(keystore, csp.SM2P256KygGen)
	if err != nil {
		return nil, err
	}

	sm2PubKey, err := utils.GetSM2PublicKey(priv)
	if err != nil {
		return nil, err
	}

	cert, err := ca.SignCertificate(filepath.Join(baseDir, "signcerts"), fileName, []string{}, sm2PubKey, true)
	if err != nil || cert == nil {
		return nil, err
	}

	err = x509Export(filepath.Join(baseDir, "cacerts", x509Filename(ca.Name)), ca.SignCert.Raw)
	if err != nil {
		return nil, err
	}

	err = x509Export(filepath.Join(baseDir, "intermediatecerts", x509Filename(orgCfg.Name)), cert.Raw)
	if err != nil {
		return nil, err
	}

	orgCA := &SM2CA{
		Name:       orgCfg.Name,
		Signer:     signer,
		SignCert:   utils.ParseX509CertificateToSm2(cert),
		CertConfig: orgCfg,
	}
	return orgCA, nil
}

// GenerateLocalUser 生成本地用户
func (ca *SM2CA) GenerateLocalUser(baseDir, fileName string) error {
	err := createFolderStructure(baseDir, false)
	if err != nil {
		return err
	}

	keystore := filepath.Join(baseDir, "keystore")
	priv, _, err := utils.GeneratePrivateKey(keystore, csp.SM2P256KygGen)
	if err != nil {
		return err
	}

	sm2PubKey, err := utils.GetSM2PublicKey(priv)
	if err != nil {
		return err
	}

	cert, err := ca.SignCertificate(filepath.Join(baseDir, "signcerts"), fileName, []string{}, sm2PubKey, false)
	if err != nil || cert == nil {
		return err
	}

	err = x509Export(filepath.Join(baseDir, "cacerts", x509Filename(ca.Name)), ca.SignCert.Raw)
	return err
}

func genCertificateGMSM2(baseDir, fileName string, template, parent *sm2.Certificate, key crypto.Signer) (*sm2.Certificate, error) {
	certBytes, err := utils.CreateCertificateToMem(template, parent, key)
	if err != nil {
		return nil, err
	}

	fullName := filepath.Join(baseDir, fileName+"-cert.pem")

	err = utils.CreateCertificateToPem(fullName, template, parent, key)
	if err != nil {
		return nil, err
	}

	x509Cert, err := sm2.ReadCertificateFromMem(certBytes)
	if err != nil {
		return nil, err
	}
	return x509Cert, nil
}

func createFolderStructure(rootDir string, isOrg bool) error {
	var folders = []string{
		filepath.Join(rootDir, "cacerts"),
		filepath.Join(rootDir, "keystore"),
		filepath.Join(rootDir, "signcerts"),
	}
	if isOrg {
		folders = append(folders, filepath.Join(rootDir, "intermediatecerts"))
	}

	for _, folder := range folders {
		err := os.MkdirAll(folder, 0750)
		if err != nil {
			return err
		}
	}

	return nil
}

func x509Filename(name string) string {
	return name + "-cert.pem"
}

func x509Export(path string, cert []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
}
