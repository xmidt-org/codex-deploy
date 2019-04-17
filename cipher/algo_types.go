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

package cipher

type AlgorithmType string

const (
	None          AlgorithmType = "none"
	Box           AlgorithmType = "box"
	RSASymmetric  AlgorithmType = "rsa-sym"
	RSAAsymmetric AlgorithmType = "rsa-asy"
)

func ParseAlogrithmType(algo string) AlgorithmType {
	if algo == string(Box) {
		return Box
	} else if algo == string(RSASymmetric) {
		return RSASymmetric
	} else if algo == string(RSAAsymmetric) {
		return RSAAsymmetric
	}
	return None
}
