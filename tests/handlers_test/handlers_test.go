package handlers_test

import (
	"bytes"
	"encoding/json"
	"github.com/shiryaevgit/myProject/database"
	"github.com/shiryaevgit/myProject/pkg/handlers"
	"github.com/shiryaevgit/myProject/pkg/models"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Conn() *database.UserRepository {
	args := m.Called()
	return args.Get(0).(*database.UserRepository)
}

func TestCreateUserHandler(t *testing.T) {
	// Создаем мок UserRepository
	mockRepo := new(MockUserRepository)

	// Создаем обработчик, передавая мок UserRepository
	handler := handlers.NewHandlerServ(mockRepo.Conn())

	// Создаем тестовый запрос
	user := models.User{Login: "test", FullName: "Test User"}
	jsonBody, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/users", nil)
	req.Body = ioutil.NopCloser(bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Ожидаемый вызов метода Conn в моке
	mockRepo.On("Conn").Return(&database.UserRepository{})

	// Выполняем запрос
	recorder := httptest.NewRecorder()
	handler.CreateUser(recorder, req)

	// Проверяем, что метод Conn был вызван
	mockRepo.AssertExpectations(t)

	// Проверяем код ответа сервера
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}
func TestGetUserByIdHandler(t *testing.T) {
	// Создаем мок UserRepository
	mockRepo := new(MockUserRepository)

	// Создаем обработчик, передавая мок UserRepository
	handler := handlers.NewHandlerServ(mockRepo.Conn())

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/users/1", nil)

	// Ожидаемый вызов метода Conn в моке
	mockRepo.On("Conn").Return(&database.UserRepository{})

	// Ожидаемый вызов метода QueryRow с возвращаемыми значениями
	mockRepo.On("Conn").Return(&database.UserRepository{}).Once()
	mockRepo.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mock.AnythingOfType("*models.User"), nil)

	// Выполняем запрос
	recorder := httptest.NewRecorder()
	handler.GetUserById(recorder, req)

	// Проверяем, что методы Conn и QueryRow были вызваны
	mockRepo.AssertExpectations(t)

	// Проверяем код ответа сервера
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}
