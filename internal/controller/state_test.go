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
	"testing"

	tessaiov1 "github.com/TessaIO/kube-state-machine/api/v1"
	"k8s.io/utils/ptr"
)

func TestNewStateMachineManager(t *testing.T) {
	tc := []struct {
		name                  string
		currentState          string
		states                []tessaiov1.State
		expectedManagerStates map[string]string
	}{
		{
			name:                  "only init state",
			currentState:          "init",
			states:                []tessaiov1.State{},
			expectedManagerStates: map[string]string{},
		},
		{
			name:                  "only one state",
			currentState:          "init",
			states:                []tessaiov1.State{{Name: "state1", Next: nil}},
			expectedManagerStates: map[string]string{"init": "state1"},
		},
		{
			name:                  "two consequent states",
			currentState:          "init",
			states:                []tessaiov1.State{{Name: "state1", Next: ptr.To("state2")}, {Name: "state2", Next: nil}},
			expectedManagerStates: map[string]string{"init": "state1", "state1": "state2"},
		},
		{
			name:                  "two states with circular reference",
			currentState:          "init",
			states:                []tessaiov1.State{{Name: "state1", Next: ptr.To("state2")}, {Name: "state2", Next: ptr.To("state1")}},
			expectedManagerStates: map[string]string{"init": "state1", "state1": "state2", "state2": "state1"},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateMachineManager(tt.currentState, tt.states)
			if sm == nil {
				t.Fatal("NewStateMachineManager() returned nil")
			}
			if !equalMaps(sm.states, tt.expectedManagerStates) {
				t.Fatalf("NewStateMachineManager() returned %v, expected %v", sm.states, tt.expectedManagerStates)
			}
		})
	}
}

func TestStateMachineManager_Add(t *testing.T) {
	sm := &StateMachineManager{
		states: make(map[string]string),
	}

	sm.Add("init", "state1")
	if !equalMaps(sm.states, map[string]string{"init": "state1"}) {
		t.Fatalf("Add() returned %v, expected %v", sm.states, map[string]string{"init": "state1"})
	}

	sm.Add("state1", "state2")
	if !equalMaps(sm.states, map[string]string{"init": "state1", "state1": "state2"}) {
		t.Fatalf("Add() returned %v, expected %v", sm.states, map[string]string{"init": "state1", "state1": "state2"})
	}
}

func TestStateMachineManager_Next(t *testing.T) {
	sm := &StateMachineManager{
		states:       map[string]string{"init": "state1", "state1": "state2"},
		currentState: ptr.To("init"),
	}

	next := sm.Next()
	if next == nil || *next != "state1" {
		t.Fatalf("Next() returned %v, expected %v", next, "state1")
	}

	next = sm.Next()
	if next == nil || *next != "state2" {
		t.Fatalf("Next() returned %v, expected %v", next, "state2")
	}

	next = sm.Next()
	if next != nil {
		t.Fatalf("Next() returned %v, expected %v", next, nil)
	}
}

func equalMaps(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}

	return true
}
