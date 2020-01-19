/*
 * Copyright 2020 Torben Schinke
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
 */

package clio

import (
	"crypto/sha512"
	"hash"
)

// Options define how the clio database is opened.
type Options struct {
	Paths      []string // Paths contains all directories to search for databases
	HMACSecret []byte   // HMACSecret can be used to verify the database against tampering
}

// newHash returns the hasher instance to use globally. The size is always 32 byte.
// The current implementation uses sha512-256 which is likely to be 60% faster than sha256
// on 64bit. If sha3 gets compromised one day, we simply need to increase the version number but keep the size.
func (o Options) newHash() hash.Hash {
	return sha512.New512_256()
}
