package application

import (
	"context"
	"errors"
	"io"
	"sync"
)

// MockContainerService реализует интерфейс ContainerService для тестирования.
type MockContainerService struct {
	mu sync.Mutex

	// Хранилище состояния контейнеров для эмуляции логики
	Containers map[string]*ContainerState

	// Настраиваемые ошибки для тестирования негативных сценариев
	CreateErr   error
	StartErr    error
	StopErr     error
	DeleteErr   error
	RestartErr  error
	AttachErr   error

	// Кастомные обработчики (опционально для продвинутых тестов)
	AttachHandler func(ctx context.Context, cmd AttachContainerCommand) (*AttachSession, error)
}

type ContainerState struct {
	ID        string
	Name      string
	Image     string
	IsRunning bool
	PID       int64
}

// NewMockContainerService создает новый экземпляр заглушки
func NewMockContainerService() ContainerService {
	return &MockContainerService{
		Containers: make(map[string]*ContainerState),
	}
}

func (m *MockContainerService) CreateAndStart(ctx context.Context, cmd CreateContainerCommand) (*CreateContainerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.CreateErr != nil {
		return nil, m.CreateErr
	}

	id := cmd.ContainerName
	if id == "" {
		id = "mock-container-id-123"
	}

	m.Containers[id] = &ContainerState{
		ID:        id,
		Name:      cmd.ContainerName,
		Image:     cmd.ImageName,
		IsRunning: false,
	}

	return &CreateContainerResponse{
		ContainerContext: ContainerContext{ID: id},
		Status:           "created",
	}, nil
}

func (m *MockContainerService) Start(ctx context.Context, cmd StartContainerCommand) (*StartContainerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.StartErr != nil {
		return nil, m.StartErr
	}

	container, exists := m.Containers[cmd.ID]
	if !exists {
		return nil, errors.New("container not found")
	}

	container.IsRunning = true
	container.PID = 1337

	return &StartContainerResponse{
		PID:              container.PID,
		ContainerContext: ContainerContext{ID: cmd.ID},
	}, nil
}

func (m *MockContainerService) Stop(ctx context.Context, cmd StopContainerCommand) (*StopContainerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.StopErr != nil {
		return nil, m.StopErr
	}

	container, exists := m.Containers[cmd.ID]
	if !exists {
		return nil, errors.New("container not found")
	}

	container.IsRunning = false

	return &StopContainerResponse{
		ContainerContext: ContainerContext{ID: cmd.ID},
	}, nil
}

func (m *MockContainerService) Remove(ctx context.Context, cmd RemoveContainerCommand) (*RemoveContainerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.DeleteErr != nil {
		return nil, m.DeleteErr
	}

	if _, exists := m.Containers[cmd.ID]; !exists {
		return nil, errors.New("container not found")
	}

	delete(m.Containers, cmd.ID)

	return &RemoveContainerResponse{
		ContainerContext: ContainerContext{ID: cmd.ID},
	}, nil
}

func (m *MockContainerService) Restart(ctx context.Context, cmd RestartContainerCommand) (*RestartContainerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.RestartErr != nil {
		return nil, m.RestartErr
	}

	if _, exists := m.Containers[cmd.ID]; !exists {
		return nil, errors.New("container not found")
	}

	return &RestartContainerResponse{
		ContainerContext: ContainerContext{ID: cmd.ID},
	}, nil
}


func (m *MockContainerService) Attach(ctx context.Context, cmd AttachContainerCommand) (*AttachSession, error) {
	if m.AttachErr != nil {
		return nil, m.AttachErr
	}

	if m.AttachHandler != nil {
		return m.AttachHandler(ctx, cmd)
	}

	// Создаем пайпы для потоков ввода-вывода
	prStdout, pwStdout := io.Pipe()
	prStderr, pwStderr := io.Pipe()
	prStdin, pwStdin := io.Pipe() // prStdin - для чтения демоном, pwStdin - для записи клиентом

	// Имитация фоновой работы контейнера (отправка данных в stdout)
	go func() {
		defer pwStdout.Close()
		defer pwStderr.Close()
		_, _ = pwStdout.Write([]byte("Mock container started...\n"))
	}()

	// Имитация чтения из Stdin (если клиент отправляет данные в контейнер)
	go func() {
		defer prStdin.Close()
		buf := make([]byte, 1024)
		for {
			_, err := prStdin.Read(buf)
			if err != nil {
				break
			}
		}
	}()

	session := &AttachSession{
		Stdin:  pwStdin,  // Клиент пишет в Stdin, демон читает из prStdin
		Stdout: prStdout, // Демон пишет в pwStdout, клиент читает из prStdout
		Stderr: prStderr, // Демон пишет в pwStderr, клиент читает из prStderr
		Wait: func() error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	return session, nil
}
func (m *MockContainerService) Pause(ctx context.Context, cmd PauseContainerCommand) (*PauseContainerResponse, error){ return nil, nil }
func (m *MockContainerService) Unpause(ctx context.Context, cmd UnpauseContainerCommand) (*UnpauseContainerResponse, error){ return nil, nil }