// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl

import (
	"crypto"
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
	"github.com/33cn/plugin/plugin/dapp/cert/authority/tools/cryptogen/generator"
	"github.com/33cn/plugin/plugin/dapp/cert/authority/tools/cryptogen/generator/utils"
	ty "github.com/33cn/plugin/plugin/dapp/cert/types"
	"github.com/tjfoc/gmsm/sm2"
)

// EcdsaCA ecdsa CA结构
type EcdsaCA struct {
	Name     string
	Signer   crypto.Signer
	SignCert *x509.Certificate
}

// SM2CA SM2 CA结构
type SM2CA struct {
	Name     string
	Signer   crypto.Signer
	SignCert *sm2.Certificate
	Sm2Key   csp.Key
}

// NewCA 根据类型生成CA生成器
func NewCA(baseDir, name string, signType int) (generator.CAGenerator, error) {
	if signType == ty.AuthECDSA {
		return newEcdsaCA(baseDir, name)
	} else if signType == ty.AuthSM2 {
		return newSM2CA(baseDir, name)
	} else {
		return nil, fmt.Errorf("Invalid sign type")
	}
}

func newEcdsaCA(baseDir, name string) (*EcdsaCA, error) {
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

	template := x509Template()
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageDigitalSignature |
		x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}

	subject := subjectTemplate()
	subject.CommonName = name

	template.Subject = subject
	template.SubjectKeyId = priv.SKI()
	template.PublicKey = ecPubKey

	x509Cert, err := genCertificateECDSA(baseDir, name, &template, &template, signer)
	if err != nil {
		return nil, err
	}
	ca = &EcdsaCA{
		Name:     name,
		Signer:   signer,
		SignCert: x509Cert,
	}

	return ca, nil
}

// SignCertificate 证书签名
func (ca *EcdsaCA) SignCertificate(baseDir, name string, sans []string, pub interface{}) (*x509.Certificate, error) {
	template := x509Template()
	template.KeyUsage = x509.KeyUsageDigitalSignature
	template.ExtKeyUsage = []x509.ExtKeyUsage{}

	subject := subjectTemplate()
	subject.CommonName = name

	template.Subject = subject
	template.DNSNames = sans
	template.PublicKey = pub

	cert, err := genCertificateECDSA(baseDir, name, &template, ca.SignCert, ca.Signer)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// GenerateLocalUser 生成本地用户
func (ca *EcdsaCA) GenerateLocalUser(baseDir, name string) error {
	err := createFolderStructure(baseDir, true)
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

	cert, err := ca.SignCertificate(filepath.Join(baseDir, "signcerts"), name, []string{}, ecPubKey)
	if err != nil || cert == nil {
		return err
	}

	err = x509Export(filepath.Join(baseDir, "cacerts", x509Filename(ca.Name)), ca.SignCert.Raw)
	return err
}

func subjectTemplate() pkix.Name {
	return pkix.Name{
		Country:  []string{"US"},
		Locality: []string{"San Francisco"},
		Province: []string{"California"},
	}
}

func x509Template() x509.Certificate {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	expiry := 3650 * 24 * time.Hour
	notBefore := time.Now().Add(-5 * time.Minute).UTC()

	x509 := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(expiry).UTC(),
		BasicConstraintsValid: true,
	}
	return x509

}

func genCertificateECDSA(baseDir, name string, template, parent *x509.Certificate, priv interface{}) (*x509.Certificate, error) {
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, template.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Join(baseDir, name+"-cert.pem")
	certFile, err := os.Create(fileName)
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

func newSM2CA(baseDir, name string) (*SM2CA, error) {
	var ca *SM2CA
	priv, signer, err := utils.GeneratePrivateKey(baseDir, csp.SM2P256KygGen)
	if err != nil {
		return nil, err
	}

	smPubKey, err := utils.GetSM2PublicKey(priv)
	if err != nil {
		return nil, err
	}

	template := x509Template()
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageDigitalSignature |
		x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}

	subject := subjectTemplate()
	subject.CommonName = name

	template.Subject = subject
	template.SubjectKeyId = priv.SKI()

	sm2cert := utils.ParseX509CertificateToSm2(&template)
	sm2cert.PublicKey = smPubKey
	x509Cert, err := genCertificateGMSM2(baseDir, name, sm2cert, sm2cert, signer)
	if err == nil {
		ca = &SM2CA{
			Name:     name,
			Signer:   signer,
			SignCert: x509Cert,
			Sm2Key:   priv,
		}
	}

	return ca, nil
}

// SignCertificate 证书签名
func (ca *SM2CA) SignCertificate(baseDir, name string, sans []string, pub interface{}) (*x509.Certificate, error) {
	template := x509Template()
	template.KeyUsage = x509.KeyUsageDigitalSignature
	template.ExtKeyUsage = []x509.ExtKeyUsage{}

	subject := subjectTemplate()
	subject.CommonName = name

	template.Subject = subject
	template.DNSNames = sans
	template.PublicKey = pub

	sm2Tpl := utils.ParseX509CertificateToSm2(&template)
	cert, err := genCertificateGMSM2(baseDir, name, sm2Tpl, ca.SignCert, ca.Signer)
	if err != nil {
		return nil, err
	}

	return utils.ParseSm2CertificateToX509(cert), nil
}

// GenerateLocalUser 生成本地用户
func (ca *SM2CA) GenerateLocalUser(baseDir, name string) error {
	err := createFolderStructure(baseDir, true)
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

	cert, err := ca.SignCertificate(filepath.Join(baseDir, "signcerts"), name, []string{}, sm2PubKey)
	if err != nil || cert == nil {
		return err
	}

	err = x509Export(filepath.Join(baseDir, "cacerts", x509Filename(ca.Name)), ca.SignCert.Raw)
	return err
}

func genCertificateGMSM2(baseDir, name string, template, parent *sm2.Certificate, key crypto.Signer) (*sm2.Certificate, error) {
	certBytes, err := utils.CreateCertificateToMem(template, parent, key)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Join(baseDir, name+"-cert.pem")

	err = utils.CreateCertificateToPem(fileName, template, parent, key)
	if err != nil {
		return nil, err
	}

	x509Cert, err := sm2.ReadCertificateFromMem(certBytes)
	if err != nil {
		return nil, err
	}
	return x509Cert, nil
}

func createFolderStructure(rootDir string, local bool) error {
	var folders = []string{
		filepath.Join(rootDir, "cacerts"),
	}
	if local {
		folders = append(folders, filepath.Join(rootDir, "keystore"),
			filepath.Join(rootDir, "signcerts"))
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
