package api

import (
	"context"
	"io"
	sync "sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DummyServerStream implements the non‑message pieces of grpc.ServerStream.
type DummyServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

func (d DummyServerStream) Context() context.Context        { return d.Ctx }
func (d DummyServerStream) SendHeader(md metadata.MD) error { return nil }
func (d DummyServerStream) SetHeader(md metadata.MD) error  { return nil }
func (d DummyServerStream) SetTrailer(md metadata.MD)       {}

// ServerStreamAdapter lets us capture Send() calls into a channel.
type ServerStreamAdapter[T any] struct {
	DummyServerStream
	Msgs chan<- *T
}

func (s *ServerStreamAdapter[T]) Send(m *T) error {
	select {
	case <-s.Context().Done():
		return s.Context().Err()
	case s.Msgs <- m:
		return nil
	}
}

// ProxyClientStream adapts that channel back into a grpc.ServerStreamingClient.
type ProxyClientStream[T any] struct {
	Msgs    <-chan *T
	ErrOnce sync.Once
	Err     error
	Ctx     context.Context
}

func (p *ProxyClientStream[T]) Recv() (*T, error) {
	m, ok := <-p.Msgs
	if ok {
		return m, nil
	}
	// channel closed → return stored error (if any), else EOF
	if p.Err != nil {
		return nil, p.Err
	}
	return nil, io.EOF
}

// implement ClientStream for grpc.ServerStreamingClient:
func (p *ProxyClientStream[T]) Header() (metadata.MD, error) { return nil, nil }
func (p *ProxyClientStream[T]) Trailer() metadata.MD         { return nil }
func (p *ProxyClientStream[T]) CloseSend() error             { return nil }
func (p *ProxyClientStream[T]) Context() context.Context     { return p.Ctx }

// SendMsg is a no-op for server‑streaming RPCs.
func (p *ProxyClientStream[T]) SendMsg(m any) error {
	return nil
}

// RecvMsg must populate the passed-in message.
// We call our Recv() to get the next *e2e.Basic, then copy it into m.
func (p *ProxyClientStream[T]) RecvMsg(m any) error {
	msg, err := p.Recv()
	if err != nil {
		return err
	}
	// m is always *e2e.Basic here
	out, ok := m.(*T)
	if !ok {
		return errors.Errorf("unexpected message type %T", m)
	}
	*out = *msg
	return nil
}

// ClientStreamAdapter implements grpc.ServerStreamingClient[T] (and its
// underlying grpc.ClientStream) by reading from a channel.
type ClientStreamAdapter[T any] struct {
	Msgs    <-chan *T
	ErrOnce sync.Once
	Err     error
	Ctx     context.Context
}

// Recv gives you the next value from the channel, or EOF/error when done.
func (c *ClientStreamAdapter[T]) Recv() (*T, error) {
	var zero *T
	m, ok := <-c.Msgs
	if ok {
		return m, nil
	}
	// channel closed → return stored error or EOF
	if c.Err != nil {
		return zero, c.Err
	}
	return zero, io.EOF
}

// Header is a no‑op here.
func (c *ClientStreamAdapter[T]) Header() (metadata.MD, error) {
	return nil, nil
}

// Trailer is a no‑op here.
func (c *ClientStreamAdapter[T]) Trailer() metadata.MD {
	return nil
}

// CloseSend is a no‑op for server‑streaming RPCs.
func (c *ClientStreamAdapter[T]) CloseSend() error {
	return nil
}

// Context returns the caller’s context.
func (c *ClientStreamAdapter[T]) Context() context.Context {
	return c.Ctx
}

// SendMsg is a no‑op (unconditionally nil) for server‑streaming.
func (c *ClientStreamAdapter[T]) SendMsg(_ any) error {
	return nil
}

// RecvMsg must fill in the supplied pointer with the next message.
func (c *ClientStreamAdapter[T]) RecvMsg(m any) error {
	msg, err := c.Recv()
	if err != nil {
		return err
	}
	// m should be *T
	out, ok := m.(*T)
	if !ok {
		return errors.Errorf("ClientStreamAdapter: wrong type %T, wanted *%T", m, msg)
	}
	*out = *msg
	return nil
}

// Compile‑time check that we satisfy the right gRPC interfaces:
var (
	_ grpc.ServerStream               = (*ServerStreamAdapter[int])(nil)
	_ grpc.ServerStreamingServer[int] = (*ServerStreamAdapter[int])(nil)
	_ grpc.ClientStream               = (*ClientStreamAdapter[int])(nil)
	_ grpc.ServerStreamingClient[int] = (*ClientStreamAdapter[int])(nil)
)
