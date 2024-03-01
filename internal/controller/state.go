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

type StateMachineManager struct {
	currentState *string
	states       map[string]string
}

func NewStateMachineManager(currentState string, states []kubetessaiov1.State) *StateMachineManager {
	s := &StateMachineManager{
		currentState: &currentState,
		states:       make(map[string]string),
	}

	if len(states) > 0 {
		s.Add(InitState, states[0].Name)
		for _, state := range states {
			if state.Next != nil {
				s.Add(state.Name, *state.Next)
			}
		}
	}

	return s
}

func (sm *StateMachineManager) Next(currentState string) *string {
	if s, ok := sm.states[currentState]; ok {
		return &s
	}

	return nil
}

func (sm *StateMachineManager) Add(currentState, nextState string) {
	sm.states[currentState] = nextState
}
