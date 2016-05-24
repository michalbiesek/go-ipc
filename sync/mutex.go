// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"

	"github.com/pkg/errors"
)

// NewMutex creates a new interprocess mutex.
// It uses the default implementation on the current platform.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewMutex(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	if !checkMutexFlags(flag) {
		return nil, errors.Errorf("invalid open flags")
	}
	return newMutex(name, flag, perm)
}

// DestroyMutex permanently removes mutex with the given name.
func DestroyMutex(name string) error {
	return destroyMutex(name)
}
