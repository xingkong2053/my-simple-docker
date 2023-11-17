package util

type CleanFn = func() error

var CleanFnNil = func() error { return nil }
