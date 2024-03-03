/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	kubetessaiov1 "github.com/TessaIO/kube-state-machine/api/v1"
)

const (
	// InitState is the initial state of the state machine
	InitState = "init"
)

// StateMachineManager is a simple state machine manager which controls the state transitions
type StateMachineManager struct {
	currentState *string
	states       map[string]string
}

// NewStateMachineManager creates a new state machine manager
func NewStateMachineManager(currentState string, states []kubetessaiov1.State) *StateMachineManager {
	s := &StateMachineManager{
		currentState: &currentState,
		states:       make(map[string]string),
	}

	if len(states) > 0 {
		s.Add(InitState, states[0].Name)
		for _, state := range states {
			if state.Next != nil {
				// this is a simple check to avoid infinite loops
				if s.Exist(state.Name) {
					break
				}
				s.Add(state.Name, *state.Next)
			}
		}
	}

	return s
}

// Next returns the next state and advances the current state of the state machine
func (sm *StateMachineManager) Next() *string {
	if s, ok := sm.states[*sm.currentState]; ok {
		sm.currentState = &s
		return &s
	}

	return nil
}

// Add adds a new state transition to the manager
func (sm *StateMachineManager) Add(currentState, nextState string) {
	sm.states[currentState] = nextState
}

// Add adds a new state transition to the manager
func (sm *StateMachineManager) Exist(state string) bool {
	_, ok := sm.states[state]
	return ok
}
