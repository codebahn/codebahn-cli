package config

import "os"

// lockPath returns the path to the sibling lock file used to serialize
// concurrent reads and writes of the config file across processes.
func lockPath() string {
	return ConfigPath() + ".lock"
}

// LockConfig acquires an exclusive, cross-process lock guarding the config
// file. It blocks until the lock is available. Callers must release it with
// UnlockConfig.
func LockConfig() (*os.File, error) {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(lockPath(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

// UnlockConfig releases a lock acquired by LockConfig and closes the
// underlying file handle.
func UnlockConfig(f *os.File) error {
	err := unlockFile(f)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}
