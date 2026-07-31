package application

import (
	"context"
	"fmt"
	"io"
	"sync"

	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	run "boyler/internal/runtime"
	"boyler/pkg/logger"
)

const chanMaxBuffer int = 30

type AttachStream interface {
	Receive() (*AttachInboundEvent, error)
	Send(*AttachOutboundEvent) error
}

type Attacher struct {
	runtime run.Runtime
	store   *storage.ContainerRepository
}

func NewAttacher(d Deps) *Attacher {
	return &Attacher{
		runtime: d.Runtime,
		store:   d.Store,
	}
}

func (a *Attacher) Execute(ctx context.Context, stream AttachStream) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := logger.FromContext(ctx)
	log.Debug("Start grpc-streaming")

	initRequest, err := stream.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive grpc-init message: %w", err)
	}
	if initRequest.Init == nil || initRequest.Init.ContainerID == "" {
		return fmt.Errorf("no container ID in init request")
	}

	containerID := initRequest.Init.ContainerID
	containerCore, err := a.store.Get(ctx, containerID)
	if err != nil {
		return fmt.Errorf("no container: %w", err)
	}

	slave, err := a.runtime.ExecPTY(ctx, int64(containerCore.PID))
	if err != nil {
		return fmt.Errorf("failed to create PTY terminal with [%d]: %w", containerCore.PID, err)
	}
	defer slave.Close()

	go observer(ctx, slave)

	var wg sync.WaitGroup
	userCmdBufChan := make(chan []byte, chanMaxBuffer)
	ptyOutChan := make(chan []byte, chanMaxBuffer)

	wg.Add(4)
	go grpcStreamReceiver(ctx, cancel, stream, userCmdBufChan, &wg)
	go ptyWriter(ctx, cancel, slave, userCmdBufChan, &wg)
	go ptyReader(ctx, cancel, slave, ptyOutChan, &wg)
	go grpcStreamSender(ctx, cancel, stream, ptyOutChan, &wg)

	wg.Wait()
	return nil
}


func observer(ctx context.Context, slave io.ReadWriteCloser) {
	<-ctx.Done()
	slave.Close()
}


func grpcStreamReceiver(ctx context.Context, cancel context.CancelFunc, stream AttachStream, ch chan<- []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)
	for {
		data, err := stream.Receive()
		if err != nil {
			cancel()
			return
		}
		
		select {
		case <-ctx.Done():
			return
		case ch <- data.Stdin:
		}
	}
}


func ptyWriter(ctx context.Context, cancel context.CancelFunc, slave io.ReadWriteCloser, ch <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			if _, err := slave.Write(data); err != nil {
				cancel()
				return
			}
		}
	}
}


func ptyReader(ctx context.Context, cancel context.CancelFunc, slave io.ReadWriteCloser, ch chan<- []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		ptyBuffer := make([]byte, 1024)
		n, err := slave.Read(ptyBuffer)
		if err != nil {
			cancel()
			return
		}
		buf := make([]byte, n)
		copy(buf, ptyBuffer[:n])

		select {
		case <-ctx.Done():
			return
		case ch <- buf:
		}
	}
}

func grpcStreamSender(ctx context.Context, cancel context.CancelFunc, stream AttachStream, ch <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	log := logger.FromContext(ctx)
	
	for {
		select {
		case <-ctx.Done():
			return
		case out, ok := <-ch:
			if !ok {
				return
			}
			if err := stream.Send(&AttachOutboundEvent{Stdout: out}); err != nil {
				log.Error("Failed to send output to client", "err", err)
				cancel()
				return
			}
		}
	}
}