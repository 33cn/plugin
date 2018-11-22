// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"errors"

	ty "github.com/33cn/plugin/plugin/dapp/cert/types"
)

// GetLocalValidator 根据类型获取校验器
func GetLocalValidator(authConfig *AuthConfig, signType int) (Validator, error) {
	var lclValidator Validator
	var err error

	if signType == ty.AuthECDSA {
		lclValidator = NewEcdsaValidator()
	} else if signType == ty.AuthSM2 {
		lclValidator = NewGmValidator()
	} else {
		return nil, ty.ErrUnknowAuthSignType
	}

	err = lclValidator.Setup(authConfig)
	if err != nil {
		authLogger.Error("Failed to set up local validator config", "Error", err)
		return nil, errors.New("Failed to initialize local validator")
	}

	return lclValidator, nil
}
