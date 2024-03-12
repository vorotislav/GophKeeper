package users

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/users/mocks"
	"GophKeeper/internal/repository"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestNewUsersProvider(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	repo := repository.TestRepository(t)
	require.NotNil(t, repo)

	auth := mocks.NewAuthorizer(t)
	require.NotNil(t, auth)

	p := NewUsersProvider(log, repo, auth)
	require.NotNil(t, p)
}

func TestUsers_UserCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareAuth     func(auth *mocks.Authorizer)
		prepareStore    func(store *mocks.Storage)
		giveUserMachine models.UserMachine
		wantErrFunc     require.ErrorAssertionFunc
		wantSession     models.Session
	}{
		{
			name: "user create store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserCreate(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{}, errors.New("some error"))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errUserProvider)
			},
		},
		{
			name: "generate tokens error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserCreate(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{
						ID:       1,
						Login:    "user",
						Password: "login",
					}, nil)
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().GenerateToken(mock.AnythingOfType("token.Payload")).
					Once().
					Return("", errors.New("some error"))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errUserProvider)
			},
		},
		{
			name: "sessions create error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserCreate(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{
						ID:       1,
						Login:    "user",
						Password: "login",
					}, nil)
				store.EXPECT().SessionCreate(mock.Anything, mock.AnythingOfType("models.Session")).
					Once().
					Return(models.Session{}, errors.New("some error"))
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().GenerateToken(mock.AnythingOfType("token.Payload")).
					Once().
					Return("some token", nil)
				auth.EXPECT().GetRefreshTokenDurationLifetime().
					Once().
					Return(time.Duration(0))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errUserProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserCreate(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{
						ID:       1,
						Login:    "user",
						Password: "login",
					}, nil)
				store.EXPECT().SessionCreate(mock.Anything, mock.AnythingOfType("models.Session")).
					Once().
					Return(models.Session{
						ID:           1,
						UserID:       1,
						AccessToken:  "some token",
						RefreshToken: "refresh token",
						IPAddress:    "127.0.0.1",
						CreatedAt:    time.Time{},
						UpdatedAt:    time.Time{},
					}, nil)
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().GenerateToken(mock.AnythingOfType("token.Payload")).
					Once().
					Return("some token", nil)
				auth.EXPECT().GetRefreshTokenDurationLifetime().
					Once().
					Return(time.Duration(0))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
			wantSession: models.Session{
				ID:           1,
				UserID:       1,
				AccessToken:  "some token",
				RefreshToken: "refresh token",
				IPAddress:    "127.0.0.1",
				CreatedAt:    time.Time{},
				UpdatedAt:    time.Time{},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth := mocks.NewAuthorizer(t)
			if tc.prepareAuth != nil {
				tc.prepareAuth(auth)
			}

			store := mocks.NewStorage(t)
			if tc.prepareStore != nil {
				tc.prepareStore(store)
			}

			up := &Users{
				log:   log,
				store: store,
				auth:  auth,
			}

			gotSession, err := up.UserCreate(context.Background(), tc.giveUserMachine)
			tc.wantErrFunc(t, err)
			assert.Equal(t, tc.wantSession, gotSession)
		})
	}
}

func TestUsers_UserLogin(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareAuth     func(auth *mocks.Authorizer)
		prepareStore    func(store *mocks.Storage)
		giveUserMachine models.UserMachine
		wantErrFunc     require.ErrorAssertionFunc
		wantSession     models.Session
	}{
		{
			name: "login error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{}, errors.New("some error"))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errUserProvider)
			},
		},
		{
			name: "token generate error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{}, nil)
				store.EXPECT().CheckSessionFromClient(mock.Anything, mock.AnythingOfType("string")).
					Once().
					Return([]int64{}, nil)
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().GenerateToken(mock.AnythingOfType("token.Payload")).
					Once().
					Return("", errors.New("some error"))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errUserProvider)
			},
		},
		{
			name: "sessions create error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{}, nil)
				store.EXPECT().CheckSessionFromClient(mock.Anything, mock.AnythingOfType("string")).
					Once().
					Return([]int64{}, nil)
				store.EXPECT().SessionCreate(mock.Anything, mock.AnythingOfType("models.Session")).
					Once().
					Return(models.Session{}, errors.New("some error"))
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().GenerateToken(mock.AnythingOfType("token.Payload")).
					Once().
					Return("some token", nil)
				auth.EXPECT().GetRefreshTokenDurationLifetime().
					Once().
					Return(time.Duration(0))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errUserProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{}, nil)
				store.EXPECT().CheckSessionFromClient(mock.Anything, mock.AnythingOfType("string")).
					Once().
					Return([]int64{}, nil)
				store.EXPECT().SessionCreate(mock.Anything, mock.AnythingOfType("models.Session")).
					Once().
					Return(models.Session{
						ID:           1,
						UserID:       1,
						AccessToken:  "some token",
						RefreshToken: "refresh token",
						IPAddress:    "127.0.0.1",
						CreatedAt:    time.Time{},
						UpdatedAt:    time.Time{},
					}, nil)
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().GenerateToken(mock.AnythingOfType("token.Payload")).
					Once().
					Return("some token", nil)
				auth.EXPECT().GetRefreshTokenDurationLifetime().
					Once().
					Return(time.Duration(0))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
			wantSession: models.Session{
				ID:           1,
				UserID:       1,
				AccessToken:  "some token",
				RefreshToken: "refresh token",
				IPAddress:    "127.0.0.1",
				CreatedAt:    time.Time{},
				UpdatedAt:    time.Time{},
			},
		},
		{
			name: "success with remove old session",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.User")).
					Once().
					Return(models.User{}, nil)
				store.EXPECT().CheckSessionFromClient(mock.Anything, mock.AnythingOfType("string")).
					Once().
					Return([]int64{2}, nil)
				store.EXPECT().SessionCreate(mock.Anything, mock.AnythingOfType("models.Session")).
					Once().
					Return(models.Session{
						ID:           1,
						UserID:       1,
						AccessToken:  "some token",
						RefreshToken: "refresh token",
						IPAddress:    "127.0.0.1",
						CreatedAt:    time.Time{},
						UpdatedAt:    time.Time{},
					}, nil)
				store.EXPECT().RemoveSession(mock.Anything, mock.AnythingOfType("int64")).
					Once().
					Return(nil)
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().GenerateToken(mock.AnythingOfType("token.Payload")).
					Once().
					Return("some token", nil)
				auth.EXPECT().GetRefreshTokenDurationLifetime().
					Once().
					Return(time.Duration(0))
			},
			giveUserMachine: models.UserMachine{
				User: models.User{
					Login:    "login",
					Password: "pass",
				},
				Machine: models.Machine{
					IPAddress: "0.0.0.0",
				},
			},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
			wantSession: models.Session{
				ID:           1,
				UserID:       1,
				AccessToken:  "some token",
				RefreshToken: "refresh token",
				IPAddress:    "127.0.0.1",
				CreatedAt:    time.Time{},
				UpdatedAt:    time.Time{},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth := mocks.NewAuthorizer(t)
			if tc.prepareAuth != nil {
				tc.prepareAuth(auth)
			}

			store := mocks.NewStorage(t)
			if tc.prepareStore != nil {
				tc.prepareStore(store)
			}

			up := &Users{
				log:   log,
				store: store,
				auth:  auth,
			}

			gotSession, err := up.UserLogin(context.Background(), tc.giveUserMachine)
			tc.wantErrFunc(t, err)
			assert.Equal(t, tc.wantSession, gotSession)
		})
	}
}
