/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package db

const (
	// default event type
	EventDefault = iota

	// event type for online and offline events
	EventState
)

var (
	eventMarshal = map[int]string{
		EventDefault: "default",
		EventState:   "state",
	}
	eventUnmarshal = map[string]int{
		"default": EventDefault,
		"state":   EventState,
	}
)

func MarshalEvent(event int) string {
	if value, ok := eventMarshal[event]; ok {
		return value
	}
	return eventMarshal[EventDefault]
}

func UnmarshalEvent(event string) int {
	if value, ok := eventUnmarshal[event]; ok {
		return value
	}
	return EventDefault
}
