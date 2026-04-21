package flowctrl

import (
	"fmt"
	"sync"
)

type ctrlWrapper struct {
	c        *Controller
	refCount int
}

func newCtrlWrapper(rate int) *ctrlWrapper {
	return &ctrlWrapper{
		c:        NewController(rate),
		refCount: 0,
	}
}

type KeyFlowCtrl struct {
	mutex   sync.RWMutex
	current map[string]*ctrlWrapper // uid -> controller
}

func NewKeyFlowCtrl() *KeyFlowCtrl {
	return &KeyFlowCtrl{current: make(map[string]*ctrlWrapper)}
}

func (k *KeyFlowCtrl) Acquire(key string, rate int) *Controller {
	k.mutex.Lock()
	ctrl, ok := k.current[key]
	if !ok {
		ctrl = newCtrlWrapper(rate)
		k.current[key] = ctrl
	}
	ctrl.refCount++
	k.mutex.Unlock()
	return ctrl.c
}

func (k *KeyFlowCtrl) Release(key string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	ctrl, ok := k.current[key]
	if !ok {
		return fmt.Errorf("key not in map. Possible reason: Release without Acquire.")
	}
	ctrl.refCount--
	if ctrl.refCount < 0 {
		return fmt.Errorf("internal error: refs < 0")
	}
	if ctrl.refCount == 0 {
		ctrl.c.Close() // avoid goroutine leak
		delete(k.current, key)
	}
	return nil
}
