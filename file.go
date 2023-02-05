package goque

import (
	"os"
	"path/filepath"
	"syscall"
)

// goqueType defines the type of Goque data structure used.
type goqueType uint8

// The possible Goque types, used to determine compatibility when
// one stored type is trying to be opened by a different type.
const (
	goqueStack goqueType = iota
	goqueQueue
	goquePriorityQueue
	goquePrefixQueue
)

// checkGoqueType checks if the type of Goque data structure
// trying to be opened is compatible with the opener type.
//
// A file named 'GOQUE' within the data directory used by
// the structure stores the structure type, using the constants
// declared above.
//
// Stacks and Queues are 100% compatible with each other, while
// a PriorityQueue is incompatible with both.
//
// Returns true if types are compatible and false if incompatible.
func checkGoqueType(dataDir string, gt goqueType) (bool, error) {
	// Set the path to 'GOQUE' file.
	path := filepath.Join(dataDir, "GOQUE")

	// Read 'GOQUE' file for this directory.
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if os.IsNotExist(err) {
		// Write the metadata to a temporary file first.
		tempPath := filepath.Join(dataDir, "GOQUE.tmp")
		f, err = os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return false, err
		}

		// Create byte slice of goqueType.
		gtb := make([]byte, 1)
		gtb[0] = byte(gt)

		_, err = f.Write(gtb)
		if err != nil {
			f.Close()
			return false, err
		}

		err = f.Sync()
		if err != nil {
			f.Close()
			return false, err
		}
		err = f.Close()
		if err != nil {
			return false, err
		}
		// Atomically swap the file into place.
		err = os.Rename(tempPath, path)
		if err != nil {
			return false, err
		}
		err = syncDir(dataDir)
		if err != nil {
			return false, err
		}

		return true, nil
	}
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Get the saved type from the file.
	fb := make([]byte, 1)
	_, err = f.Read(fb)
	if err != nil {
		return false, err
	}

	// Convert the file byte to its goqueType.
	filegt := goqueType(fb[0])

	// Compare the types.
	if filegt == gt {
		return true, nil
	} else if filegt == goqueStack && gt == goqueQueue {
		return true, nil
	} else if filegt == goqueQueue && gt == goqueStack {
		return true, nil
	}

	return false, nil
}

// syncDir calls fsync(2) on a directory, which is necessary on Linux, and
// ignores EINVAL errors that may arise on other systems.
func syncDir(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Sync(); err != nil {
		if err == os.ErrInvalid {
			return nil
		}
		// Go < 1.8
		if syserr, ok := err.(*os.SyscallError); ok && syserr.Err == syscall.EINVAL {
			return nil
		}
		// Go >= 1.8
		if patherr, ok := err.(*os.PathError); ok && patherr.Err == syscall.EINVAL {
			return nil
		}
		return err
	}
	return nil
}
