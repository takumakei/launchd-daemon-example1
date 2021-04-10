package launch

import (
	"net"
	"os"

	"github.com/hashicorp/go-multierror"
)

func SocketFiles(name string) ([]*os.File, error) {
	fds, err := activateSocket(name)
	if err != nil {
		return nil, err
	}
	files := make([]*os.File, len(fds))
	for i, fd := range fds {
		files[i] = os.NewFile(uintptr(fd), "")
	}
	return files, nil
}

func SocketListeners(name string) ([]net.Listener, error) {
	files, err := SocketFiles(name)
	if err != nil {
		return nil, err
	}
	result := make([]net.Listener, len(files))
	errs := make([]error, len(files))
	for i, file := range files {
		result[i], errs[i] = net.FileListener(file)
	}
	return result, multierror.Append(&multierror.Error{}, errs...).ErrorOrNil()
}
