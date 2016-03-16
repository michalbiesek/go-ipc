// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os"
	"unsafe"
)

// Destroyer is an object which can be permanently removed
type Destroyer interface {
	Destroy() error
}

// Blocker is an object, whose operations can be blockable or not
type Blocker interface {
	SetBlocking(bool) error
}

// Buffered is an interface for objects with a capacity for storing other objects
type Buffered interface {
	Cap() (int, error)
}

func accessModeToOsMode(mode int) (osMode int, err error) {
	if mode&O_READ_ONLY != 0 {
		if mode&(O_WRITE_ONLY|O_READWRITE) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return osMode | os.O_RDONLY, nil
	}
	if mode&O_WRITE_ONLY != 0 {
		if mode&O_READWRITE != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return osMode | os.O_WRONLY, nil
	}
	if mode&O_READWRITE != 0 {
		return osMode | os.O_RDWR, nil
	}
	return 0, fmt.Errorf("no access mode flags")
}

func createModeToOsMode(mode int) (int, error) {
	if mode&O_OPEN_OR_CREATE != 0 {
		if mode&(O_CREATE_ONLY|O_OPEN_ONLY) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_TRUNC, nil
	}
	if mode&O_CREATE_ONLY != 0 {
		if mode&O_OPEN_ONLY != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_EXCL, nil
	}
	if mode&O_OPEN_ONLY != 0 {
		return 0, nil
	}
	return 0, fmt.Errorf("no create mode flags")
}

func openModeToOsMode(mode int) (int, error) {
	var err error
	var createMode, accessMode int
	if createMode, err = createModeToOsMode(mode); err != nil {
		return 0, err
	}
	if accessMode, err = accessModeToOsMode(mode); err != nil {
		return 0, err
	}
	return createMode | accessMode, nil
}

func openOrCreateFile(opener func(int) error, mode int) (bool, error) {
	switch {
	case mode&(O_OPEN_ONLY|O_CREATE_ONLY) != 0:
		osMode, err := openModeToOsMode(mode)
		if err != nil {
			return false, err
		}
		if err = opener(osMode); err == nil {
			return (mode & O_CREATE_ONLY) != 0, nil
		}
		return false, err
	case mode&O_OPEN_OR_CREATE != 0:
		const attempts = 16
		amode, err := accessModeToOsMode(mode)
		if err == nil {
			for attempt := 0; attempt < attempts; attempt++ {
				if err = opener(amode | os.O_CREATE | os.O_EXCL); !os.IsExist(err) {
					return true, err
				}
				if err = opener(amode); !os.IsNotExist(err) {
					return false, err
				}
			}
		}
		return false, err
	default:
		return false, fmt.Errorf("unknown open mode")
	}
}

// from syscall package:
// use is a no-op, but the compiler cannot see that it is.
// Calling use(p) ensures that p is kept live until that point.
//go:noescape
func use(p unsafe.Pointer)
