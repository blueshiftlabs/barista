// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package base provides some helpers to make constructing bar modules easier.
package base

import (
	"sync"
	"sync/atomic"

	l "github.com/soumya92/barista/logging"
	"github.com/soumya92/barista/notifier"
)

// Value provides atomic value storage with update notifications.
type Value struct {
	value atomic.Value

	subM sync.RWMutex
	subs []func()
}

// Subscribe creates a new ticker for value updates.
func (v *Value) Subscribe() <-chan struct{} {
	notifyFn, ch := notifier.New()
	v.subM.Lock()
	defer v.subM.Unlock()
	l.Attachf(v, ch, "$%d", len(v.subs))
	v.subs = append(v.subs, notifyFn)
	return ch
}

// Get returns the currently stored value.
func (v *Value) Get() interface{} {
	return v.value.Load()
}

// Set updates the stored values and notifies any subscribers.
func (v *Value) Set(value interface{}) {
	v.value.Store(value)
	l.Fine("%s: Store %#v", l.ID(v), value)
	v.subM.RLock()
	defer v.subM.RUnlock()
	for _, notifyFn := range v.subs {
		notifyFn()
	}
}

type valueOrErr struct {
	value interface{}
	err   error
}

// ErrorValue adds an error to Value, allowing storage of either
// a value (interface{}) or an error.
type ErrorValue struct {
	v       Value // of valueOrErr
	logInit sync.Once
}

func (e *ErrorValue) initLogging() {
	e.logInit.Do(func() { l.Attach(e, &e.v, "") })
}

// Subscribe creates a new ticker for value/error updates.
func (e *ErrorValue) Subscribe() <-chan struct{} {
	e.initLogging()
	return e.v.Subscribe()
}

// Get returns the currently stored value or error.
func (e *ErrorValue) Get() (interface{}, error) {
	e.initLogging()
	if v, ok := e.v.Get().(valueOrErr); ok {
		return v.value, v.err
	}
	// Uninitialised.
	return nil, nil
}

// Set updates the stored value and clears any error.
func (e *ErrorValue) Set(value interface{}) {
	e.initLogging()
	e.v.Set(valueOrErr{value: value})
}

// Error replaces the stored value and returns true if non-nil,
// and simply returns false if nil.
func (e *ErrorValue) Error(err error) bool {
	if err == nil {
		return false
	}
	e.initLogging()
	e.v.Set(valueOrErr{err: err})
	return true
}
