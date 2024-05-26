//go:build (amd64 && !generic) || (arm64 && !generic)
// +build amd64,!generic arm64,!generic

//nolint:unparam // 忽视本文件所有golangci-linter检查
package bn256

import "golang.org/x/sys/cpu"

// This file contains forward declarations for the architecture-specific
// assembly implementations of these functions, provided that they exist.

//nolint:varcheck
var hasBMI2 = cpu.X86.HasBMI2

// go:noescape
func gfpNeg(c, a *gfP)

//go:noescape
func gfpAdd(c, a, b *gfP)

//go:noescape
func gfpSub(c, a, b *gfP)

//go:noescape
func gfpMul(c, a, b *gfP)
