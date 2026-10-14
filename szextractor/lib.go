package szextractor

import (
	"context"
	"sync"

	"github.com/itchio/headway/state"
	"github.com/itchio/sevenzip-go/sz"
)

var _libMutex sync.Mutex
var _lib *sz.Lib
var _libErr error
var _loadOnce sync.Once

func GetLib(consumer *state.Consumer) (*sz.Lib, error) {
	_libMutex.Lock()
	defer _libMutex.Unlock()

	_loadOnce.Do(func() {
		// A process-wide once-init has no owning caller, so no caller's
		// cancellation may leave the library unloaded for everyone else.
		_libErr = EnsureDeps(context.Background(), consumer)
		if _libErr != nil {
			return
		}

		_lib, _libErr = sz.NewLib()
	})

	if _libErr != nil {
		return nil, _libErr
	}
	return _lib, nil
}
