// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build crt.strace

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"strings"

	fcntl2 "modernc.org/ccir/libc/fcntl"
)

func cmdString(cmd int32) string {
	switch cmd {
	case fcntl2.XF_DUPFD:
		return "F_DUPFD"
	case fcntl2.XF_GETFD:
		return "F_GETFD"
	case fcntl2.XF_GETFL:
		return "F_GETFL"
	case fcntl2.XF_GETLK:
		return "F_GETLK"
	case fcntl2.XF_GETOWN:
		return "F_GETOWN"
	case fcntl2.XF_SETFD:
		return "F_SETFD"
	case fcntl2.XF_SETFL:
		return "F_SETFL"
	case fcntl2.XF_SETLK:
		return "F_SETLK"
	case fcntl2.XF_SETLKW:
		return "F_SETLKW"
	case fcntl2.XF_SETOWN:
		return "F_SETOWN"
	default:
		return fmt.Sprintf("%#x", cmd)
	}
}

func modeString(flag int32) string {
	if flag == 0 {
		return "0"
	}

	var a []string
	for _, v := range []struct {
		int32
		string
	}{
		{fcntl2.XO_APPEND, "O_APPEND"},
		{fcntl2.XO_CREAT, "O_CREAT"},
		{fcntl2.XO_DSYNC, "O_DSYNC"},
		{fcntl2.XO_EXCL, "O_EXCL"},
		{fcntl2.XO_NOCTTY, "O_NOCTTY"},
		{fcntl2.XO_NONBLOCK, "O_NONBLOCK"},
		{fcntl2.XO_RDONLY, "O_RDONLY"},
		{fcntl2.XO_RDWR, "O_RDWR"},
		{fcntl2.XO_WRONLY, "O_WRONLY"},
	} {
		if flag&v.int32 != 0 {
			a = append(a, v.string)
		}
	}
	return strings.Join(a, "|")
}
