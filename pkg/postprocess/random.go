// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package postprocess

import (
	_ "embed"
	"sync"
)

//go:embed assets/pregen_rand
var randomAsset []byte

var cacheLock sync.Mutex
var randomCache []int
var currentIndex int

func init() {
	randomCache = parseRandomBytesToInts(randomAsset)
	currentIndex = 0
}

func parseRandomBytesToInts(data []byte) []int {
	var numbers []int
	for _, b := range data {
		numbers = append(numbers, int(b))
	}
	return numbers
}

func getRandomBoolWithWrap() bool {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	number := randomCache[currentIndex]
	currentIndex = (currentIndex + 1) % len(randomCache)
	isEven := (number % 2) == 0

	return isEven
}
