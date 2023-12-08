package main

// #include <stdlib.h>
import "C"

import (
	"encoding/json"
	"errors"
	"net"
	"unsafe"
)

//export metadata
func metadata() uint64 {
	metadata := map[string]any{
		"type": "test",
		"status": map[string]any{
			"stability": map[string]any{
				"development": []string{"logs"},
			},
		},
	}
	return toLeakedJSON(metadata)
}

//export defaultConfig
func defaultConfig() uint64 {
	config := map[string]string{}
	return toLeakedJSON(config)
}

var server *Server

//export start
func start() uint64 {
	if server != nil {
		return toLeakedJSON(map[string]any{"error": "already started"})
	}

	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return toLeakedJSON(map[string]any{"error": err.Error()})
	}
	server = &Server{
		listener: listener,
	}
	go server.Serve()

	return 0
}

type Server struct {
	listener net.Listener
}

func (s *Server) Serve() {
	for {
		conn, err := s.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			// TODO: Error callback
			break
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	// TODO: Publish callback.
	conn.Close()
}

func (s *Server) Close() error {
	return s.listener.Close()
}

//export stop
func stop() uint64 {
	if server == nil {
		return toLeakedJSON(map[string]any{"error": "not started"})
	}

	server.Close()
	server = nil

	return 0
}

// toLeakedJSON returns an uint64 with the pointer and the size of
// a string with the encoded JSON.
// The pointer is not automatically managed by TinyGo hence it must be freed by the host.
func toLeakedJSON(o any) uint64 {
	d, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	ptr, size := stringToLeakedPtr(string(d))
	return (uint64(ptr) << uint64(32)) | uint64(size)
}

// stringToPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
// The returned pointer aliases the string hence the string must be kept alive
// until ptr is no longer needed.
func stringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}

// stringToLeakedPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
// The pointer is not automatically managed by TinyGo hence it must be freed by the host.
func stringToLeakedPtr(s string) (uint32, uint32) {
	size := C.ulong(len(s))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), s)
	return uint32(uintptr(ptr)), uint32(size)
}

func main() {}
