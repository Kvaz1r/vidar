// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package config_test

import (
	"io"
)

type mockOpener struct {
	OpenCalled chan bool
	OpenInput  struct {
		Path chan string
	}
	OpenOutput struct {
		F   chan io.ReadCloser
		Err chan error
	}
	CreateCalled chan bool
	CreateInput  struct {
		Path chan string
	}
	CreateOutput struct {
		F   chan io.WriteCloser
		Err chan error
	}
}

func newMockOpener() *mockOpener {
	m := &mockOpener{}
	m.OpenCalled = make(chan bool, 100)
	m.OpenInput.Path = make(chan string, 100)
	m.OpenOutput.F = make(chan io.ReadCloser, 100)
	m.OpenOutput.Err = make(chan error, 100)
	m.CreateCalled = make(chan bool, 100)
	m.CreateInput.Path = make(chan string, 100)
	m.CreateOutput.F = make(chan io.WriteCloser, 100)
	m.CreateOutput.Err = make(chan error, 100)
	return m
}
func (m *mockOpener) Open(path string) (f io.ReadCloser, err error) {
	m.OpenCalled <- true
	m.OpenInput.Path <- path
	return <-m.OpenOutput.F, <-m.OpenOutput.Err
}
func (m *mockOpener) Create(path string) (f io.WriteCloser, err error) {
	m.CreateCalled <- true
	m.CreateInput.Path <- path
	return <-m.CreateOutput.F, <-m.CreateOutput.Err
}

type mockReadCloser struct {
	ReadCalled chan bool
	ReadInput  struct {
		P chan []byte
	}
	ReadOutput struct {
		N   chan int
		Err chan error
	}
	CloseCalled chan bool
	CloseOutput struct {
		Ret0 chan error
	}
}

func newMockReadCloser() *mockReadCloser {
	m := &mockReadCloser{}
	m.ReadCalled = make(chan bool, 100)
	m.ReadInput.P = make(chan []byte, 100)
	m.ReadOutput.N = make(chan int, 100)
	m.ReadOutput.Err = make(chan error, 100)
	m.CloseCalled = make(chan bool, 100)
	m.CloseOutput.Ret0 = make(chan error, 100)
	return m
}
func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	m.ReadCalled <- true
	m.ReadInput.P <- p
	return <-m.ReadOutput.N, <-m.ReadOutput.Err
}
func (m *mockReadCloser) Close() error {
	m.CloseCalled <- true
	return <-m.CloseOutput.Ret0
}

type mockWriteCloser struct {
	WriteCalled chan bool
	WriteInput  struct {
		P chan []byte
	}
	WriteOutput struct {
		N   chan int
		Err chan error
	}
	CloseCalled chan bool
	CloseOutput struct {
		Ret0 chan error
	}
}

func newMockWriteCloser() *mockWriteCloser {
	m := &mockWriteCloser{}
	m.WriteCalled = make(chan bool, 100)
	m.WriteInput.P = make(chan []byte, 100)
	m.WriteOutput.N = make(chan int, 100)
	m.WriteOutput.Err = make(chan error, 100)
	m.CloseCalled = make(chan bool, 100)
	m.CloseOutput.Ret0 = make(chan error, 100)
	return m
}
func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	m.WriteCalled <- true
	m.WriteInput.P <- p
	return <-m.WriteOutput.N, <-m.WriteOutput.Err
}
func (m *mockWriteCloser) Close() error {
	m.CloseCalled <- true
	return <-m.CloseOutput.Ret0
}
