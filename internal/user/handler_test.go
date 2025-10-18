package user

import (
"bytes"
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"testing"
"github.com/gin-gonic/gin"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"

"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
)

// MockService is a mock implementation of the user service
type MockService struct {
mock.Mock
}

func (m *MockService) RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
args := m.Called(ctx, req)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error) {
args := m.Called(ctx, req)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) GetUserByID(ctx context.Context, id uint) (*User, error) {
args := m.Called(ctx, id)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error) {
args := m.Called(ctx, id, req)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) DeleteUser(ctx context.Context, id uint) error {
args := m.Called(ctx, id)
return args.Error(0)
}

// MockAuthService is a mock implementation of the auth service
type MockAuthService struct {
mock.Mock
}

func (m *MockAuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
args := m.Called(tokenString)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockAuthService) GenerateToken(userID uint, email string, name string) (string, error) {
args := m.Called(userID, email, name)
return args.String(0), args.Error(1)
}

func TestHandler_Register(t *testing.T) {
tests := []struct {
name           string
requestBody    interface{}
setupMocks     func(*MockService, *MockAuthService)
expectedStatus int
checkResponse  func(*testing.T, *httptest.ResponseRecorder)
}{
{
name: "successful registration",
requestBody: RegisterRequest{
Name:     "John Doe",
Email:    "john@example.com",
Password: "password123",
},
setupMocks: func(ms *MockService, mas *MockAuthService) {
user := &User{
ID:    1,
Name:  "John Doe",
Email: "john@example.com",
}
ms.On("RegisterUser", mock.Anything, mock.AnythingOfType("user.RegisterRequest")).Return(user, nil)
mas.On("GenerateToken", uint(1), "john@example.com", "John Doe").Return("mock-token", nil)
},
expectedStatus: http.StatusOK,
checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
var response map[string]interface{}
err := json.Unmarshal(w.Body.Bytes(), &response)
assert.NoError(t, err)
assert.Contains(t, response, "token")
assert.Contains(t, response, "user")
},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
mockService := &MockService{}
mockAuthService := &MockAuthService{}
tt.setupMocks(mockService, mockAuthService)

handler := NewHandler(mockService, mockAuthService)

w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)

// Set up request
var reqBody []byte
if str, ok := tt.requestBody.(string); ok {
reqBody = []byte(str)
} else {
reqBody, _ = json.Marshal(tt.requestBody)
}

c.Request, _ = http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
c.Request.Header.Set("Content-Type", "application/json")

handler.Register(c)

assert.Equal(t, tt.expectedStatus, w.Code)
tt.checkResponse(t, w)

mockService.AssertExpectations(t)
mockAuthService.AssertExpectations(t)
})
}
}

func TestHandler_GetUser(t *testing.T) {
tests := []struct {
name           string
userID         string
setupMocks     func(*MockService, *MockAuthService)
setupContext   func(*gin.Context)
expectedStatus int
checkResponse  func(*testing.T, *httptest.ResponseRecorder)
}{
{
name:   "successful get user",
userID: "1",
setupMocks: func(ms *MockService, mas *MockAuthService) {
user := &User{
ID:    1,
Name:  "John Doe",
Email: "john@example.com",
}
ms.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
},
setupContext: func(c *gin.Context) {
claims := &auth.Claims{UserID: 1}
c.Set(auth.KeyUser, claims)
},
expectedStatus: http.StatusOK,
checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
var response map[string]interface{}
err := json.Unmarshal(w.Body.Bytes(), &response)
assert.NoError(t, err)
assert.Equal(t, float64(1), response["id"])
assert.Equal(t, "John Doe", response["name"])
assert.Equal(t, "john@example.com", response["email"])
},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
mockService := &MockService{}
mockAuthService := &MockAuthService{}
tt.setupMocks(mockService, mockAuthService)

handler := NewHandler(mockService, mockAuthService)

w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)

// Create a proper HTTP request
req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
c.Request = req
c.Params = gin.Params{{Key: "id", Value: tt.userID}}

tt.setupContext(c)

handler.GetUser(c)

assert.Equal(t, tt.expectedStatus, w.Code)
tt.checkResponse(t, w)

mockService.AssertExpectations(t)
mockAuthService.AssertExpectations(t)
})
}
}
