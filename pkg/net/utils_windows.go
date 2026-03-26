//go:build windows
// +build windows

package net

func GetMaxOpenFiles() int {
	return 1024 * 1024 * 2
}
