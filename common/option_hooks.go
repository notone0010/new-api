package common

import "sync"

var (
	optionHooksMu sync.RWMutex
	optionHooks   []func(string, string)
)

func RegisterOptionUpdateHook(hook func(string, string)) {
	if hook == nil {
		return
	}
	optionHooksMu.Lock()
	optionHooks = append(optionHooks, hook)
	optionHooksMu.Unlock()
}

func NotifyOptionUpdated(key, value string) {
	optionHooksMu.RLock()
	hooks := append([]func(string, string){}, optionHooks...)
	optionHooksMu.RUnlock()
	for _, hook := range hooks {
		hook(key, value)
	}
}
