/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package nmxutil

import (
	"sync"
)

type SRWaiter struct {
	c     chan error
	token interface{}
}

type SingleResource struct {
	acquired  bool
	waitQueue []SRWaiter
	mtx       sync.Mutex
}

func NewSingleResource() SingleResource {
	return SingleResource{}
}

// Appends an entry to the wait queue and returns a channel for the caller to
// block on.  The caller must either read from the channel or call
// StopWaiting().
func (s *SingleResource) Acquire(token interface{}) <-chan error {
	s.mtx.Lock()

	if !s.acquired {
		s.acquired = true
		s.mtx.Unlock()

		// Indicate immediate acquisition.
		ch := make(chan error)
		close(ch)
		return ch
	}

	// XXX: Verify no duplicates.

	w := SRWaiter{
		c:     make(chan error),
		token: token,
	}
	s.waitQueue = append(s.waitQueue, w)

	s.mtx.Unlock()

	return w.c
}

// @return                      true if a pending waiter acquired the resource;
//                              false if the resource is now free.
func (s *SingleResource) Release() bool {
	initiate := func() *SRWaiter {
		s.mtx.Lock()
		defer s.mtx.Unlock()

		if !s.acquired {
			panic("SingleResource release without acquire")
			return nil
		}

		if len(s.waitQueue) == 0 {
			s.acquired = false
			return nil
		}

		w := s.waitQueue[0]
		s.waitQueue = s.waitQueue[1:]

		return &w
	}

	w := initiate()
	if w == nil {
		return false
	}

	w.c <- nil
	return true
}

func (s *SingleResource) StopWaiting(token interface{}, err error) {
	getw := func() *SRWaiter {
		s.mtx.Lock()
		defer s.mtx.Unlock()

		for i, w := range s.waitQueue {
			if w.token == token {
				s.waitQueue = append(s.waitQueue[:i], s.waitQueue[i+1:]...)
				return &w
			}
		}

		return nil
	}

	if w := getw(); w != nil {
		w.c <- err
	}
}

func (s *SingleResource) Abort(err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, w := range s.waitQueue {
		w.c <- err
	}
	s.waitQueue = nil
}

func (s *SingleResource) Acquired() bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.acquired
}
