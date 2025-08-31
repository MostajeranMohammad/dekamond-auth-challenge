package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/repositories/mockrepositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUsersUsecase_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepositories.NewMockUserRepository(ctrl)
	service := NewUsersService(mockRepo)

	tests := []struct {
		name       string
		userID     uint32
		setupMock  func()
		wantUser   entities.User
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:   "successful user retrieval",
			userID: 123,
			setupMock: func() {
				expectedUser := entities.User{
					Id:        123,
					Phone:     "+1234567890",
					CreatedAt: time.Now(),
				}
				mockRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(expectedUser, nil)
			},
			wantUser: entities.User{
				Id:        123,
				Phone:     "+1234567890",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func() {
				mockRepo.EXPECT().GetUserById(gomock.Any(), uint32(999)).Return(entities.User{}, errors.New("user not found"))
			},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "user not found",
		},
		{
			name:   "database error",
			userID: 456,
			setupMock: func() {
				mockRepo.EXPECT().GetUserById(gomock.Any(), uint32(456)).Return(entities.User{}, errors.New("database connection error"))
			},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			user, err := service.GetUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				assert.Equal(t, entities.User{}, user)
			} else {
				assert.NoError(t, err)
				// For time comparison, check if they're close (within 1 second)
				assert.Equal(t, tt.wantUser.Id, user.Id)
				assert.Equal(t, tt.wantUser.Phone, user.Phone)
				assert.WithinDuration(t, tt.wantUser.CreatedAt, user.CreatedAt, time.Second)
			}
		})
	}
}

func TestUsersUsecase_GetAllUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepositories.NewMockUserRepository(ctrl)
	service := NewUsersService(mockRepo)

	now := time.Now()
	sampleUsers := []entities.User{
		{Id: 1, Phone: "+1234567890", CreatedAt: now},
		{Id: 2, Phone: "+0987654321", CreatedAt: now.Add(-time.Hour)},
	}

	tests := []struct {
		name          string
		page          uint32
		limit         uint32
		phoneSearch   *string
		creationFrom  *time.Time
		creationTo    *time.Time
		setupMock     func()
		wantUsers     []entities.User
		wantErr       bool
		wantErrMsg    string
		expectedSkip  uint32
		expectedLimit uint32
	}{
		{
			name:  "successful retrieval with default pagination",
			page:  0, // Should default to 1
			limit: 0, // Should default to 10
			setupMock: func() {
				mockRepo.EXPECT().GetAllUsers(
					gomock.Any(),
					uint32(0),  // skip = (1-1) * 10 = 0
					uint32(10), // default limit
					(*string)(nil),
					(*time.Time)(nil),
					(*time.Time)(nil),
				).Return(sampleUsers, nil)
			},
			wantUsers:     sampleUsers,
			wantErr:       false,
			expectedSkip:  0,
			expectedLimit: 10,
		},
		{
			name:  "successful retrieval with custom pagination",
			page:  2,
			limit: 5,
			setupMock: func() {
				mockRepo.EXPECT().GetAllUsers(
					gomock.Any(),
					uint32(5), // skip = (2-1) * 5 = 5
					uint32(5), // limit
					(*string)(nil),
					(*time.Time)(nil),
					(*time.Time)(nil),
				).Return(sampleUsers, nil)
			},
			wantUsers:     sampleUsers,
			wantErr:       false,
			expectedSkip:  5,
			expectedLimit: 5,
		},
		{
			name:        "with phone search",
			page:        1,
			limit:       10,
			phoneSearch: stringPtr("+1234"),
			setupMock: func() {
				mockRepo.EXPECT().GetAllUsers(
					gomock.Any(),
					uint32(0),
					uint32(10),
					stringPtr("+1234"),
					(*time.Time)(nil),
					(*time.Time)(nil),
				).Return([]entities.User{sampleUsers[0]}, nil)
			},
			wantUsers: []entities.User{sampleUsers[0]},
			wantErr:   false,
		},
		{
			name:         "with date range",
			page:         1,
			limit:        10,
			creationFrom: timePtr(now.Add(-2 * time.Hour)),
			creationTo:   timePtr(now),
			setupMock: func() {
				mockRepo.EXPECT().GetAllUsers(
					gomock.Any(),
					uint32(0),
					uint32(10),
					(*string)(nil),
					timePtr(now.Add(-2*time.Hour)),
					timePtr(now),
				).Return(sampleUsers, nil)
			},
			wantUsers: sampleUsers,
			wantErr:   false,
		},
		{
			name:  "repository error",
			page:  1,
			limit: 10,
			setupMock: func() {
				mockRepo.EXPECT().GetAllUsers(
					gomock.Any(),
					uint32(0),
					uint32(10),
					(*string)(nil),
					(*time.Time)(nil),
					(*time.Time)(nil),
				).Return(nil, errors.New("database error"))
			},
			wantUsers:  nil,
			wantErr:    true,
			wantErrMsg: "database error",
		},
		{
			name:  "empty result",
			page:  1,
			limit: 10,
			setupMock: func() {
				mockRepo.EXPECT().GetAllUsers(
					gomock.Any(),
					uint32(0),
					uint32(10),
					(*string)(nil),
					(*time.Time)(nil),
					(*time.Time)(nil),
				).Return([]entities.User{}, nil)
			},
			wantUsers: []entities.User{},
			wantErr:   false,
		},
		{
			name:  "page 3 with limit 7",
			page:  3,
			limit: 7,
			setupMock: func() {
				mockRepo.EXPECT().GetAllUsers(
					gomock.Any(),
					uint32(14), // skip = (3-1) * 7 = 14
					uint32(7),
					(*string)(nil),
					(*time.Time)(nil),
					(*time.Time)(nil),
				).Return([]entities.User{}, nil)
			},
			wantUsers: []entities.User{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			users, err := service.GetAllUsers(
				context.Background(),
				tt.page,
				tt.limit,
				tt.phoneSearch,
				tt.creationFrom,
				tt.creationTo,
			)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUsers, users)
			}
		})
	}
}

func TestUsersUsecase_PaginationCalculation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepositories.NewMockUserRepository(ctrl)
	service := NewUsersService(mockRepo)

	tests := []struct {
		name          string
		page          uint32
		limit         uint32
		expectedSkip  uint32
		expectedLimit uint32
	}{
		{
			name:          "page 0 defaults to 1",
			page:          0,
			limit:         5,
			expectedSkip:  0, // (1-1) * 5 = 0
			expectedLimit: 5,
		},
		{
			name:          "limit 0 defaults to 10",
			page:          1,
			limit:         0,
			expectedSkip:  0, // (1-1) * 10 = 0
			expectedLimit: 10,
		},
		{
			name:          "both 0 use defaults",
			page:          0,
			limit:         0,
			expectedSkip:  0, // (1-1) * 10 = 0
			expectedLimit: 10,
		},
		{
			name:          "page 1",
			page:          1,
			limit:         20,
			expectedSkip:  0, // (1-1) * 20 = 0
			expectedLimit: 20,
		},
		{
			name:          "page 2",
			page:          2,
			limit:         20,
			expectedSkip:  20, // (2-1) * 20 = 20
			expectedLimit: 20,
		},
		{
			name:          "page 5, limit 3",
			page:          5,
			limit:         3,
			expectedSkip:  12, // (5-1) * 3 = 12
			expectedLimit: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().GetAllUsers(
				gomock.Any(),
				tt.expectedSkip,
				tt.expectedLimit,
				(*string)(nil),
				(*time.Time)(nil),
				(*time.Time)(nil),
			).Return([]entities.User{}, nil)

			_, err := service.GetAllUsers(
				context.Background(),
				tt.page,
				tt.limit,
				nil,
				nil,
				nil,
			)

			assert.NoError(t, err)
		})
	}
}

func TestUsersUsecase_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepositories.NewMockUserRepository(ctrl)
	service := NewUsersService(mockRepo)

	now := time.Now()
	testUser := entities.User{
		Id:        123,
		Phone:     "+1234567890",
		CreatedAt: now,
	}

	// Test getting a specific user
	mockRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(testUser, nil)

	user, err := service.GetUser(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, testUser.Id, user.Id)
	assert.Equal(t, testUser.Phone, user.Phone)

	// Test getting all users
	allUsers := []entities.User{testUser}
	mockRepo.EXPECT().GetAllUsers(
		gomock.Any(),
		uint32(0),
		uint32(10),
		(*string)(nil),
		(*time.Time)(nil),
		(*time.Time)(nil),
	).Return(allUsers, nil)

	users, err := service.GetAllUsers(context.Background(), 1, 10, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, testUser.Id, users[0].Id)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
