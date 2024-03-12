package passwords

import (
	"context"
	"errors"
	"testing"

	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/passwords/mocks"
	"GophKeeper/internal/repository"
	"GophKeeper/internal/token"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewProvider(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	repo := repository.TestRepository(t)
	require.NotNil(t, repo)

	crp := mocks.NewCrypto(t)

	p := NewProvider(log, repo, crp)
	require.NotNil(t, p)
}

func TestProvider_PasswordCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		givePass         models.Password
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "crypto error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			givePass: models.Password{
				Title:    "title",
				Login:    "login",
				Password: "pass",
				URL:      "url",
				Note:     "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "store error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encpass", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().PasswordCreate(
					mock.Anything,
					mock.AnythingOfType("models.Password"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			givePass: models.Password{
				Title:    "title",
				Login:    "login",
				Password: "pass",
				URL:      "url",
				Note:     "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "success",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encpass", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().PasswordCreate(
					mock.Anything,
					mock.AnythingOfType("models.Password"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			givePass: models.Password{
				Title:    "title",
				Login:    "login",
				Password: "pass",
				URL:      "url",
				Note:     "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := mocks.NewStorage(t)
			if tc.prepareStore != nil {
				tc.prepareStore(store)
			}

			crp := mocks.NewCrypto(t)
			if tc.prepareCrypto != nil {
				tc.prepareCrypto(crp)
			}

			pp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := pp.PasswordCreate(ctx, tc.givePass)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_PasswordUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		givePass         models.Password
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "crypto error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			givePass: models.Password{
				Title:    "title",
				Login:    "login",
				Password: "pass",
				URL:      "url",
				Note:     "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "store error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encpass", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().PasswordUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Password"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			givePass: models.Password{
				Title:    "title",
				Login:    "login",
				Password: "pass",
				URL:      "url",
				Note:     "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "success",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encpass", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().PasswordUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Password"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			givePass: models.Password{
				Title:    "title",
				Login:    "login",
				Password: "pass",
				URL:      "url",
				Note:     "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := mocks.NewStorage(t)
			if tc.prepareStore != nil {
				tc.prepareStore(store)
			}

			crp := mocks.NewCrypto(t)
			if tc.prepareCrypto != nil {
				tc.prepareCrypto(crp)
			}

			pp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := pp.PasswordUpdate(ctx, tc.givePass)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_PasswordDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		givePassID       int
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().PasswordDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			givePassID:       1,
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().PasswordDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			givePassID:       1,
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := mocks.NewStorage(t)
			if tc.prepareStore != nil {
				tc.prepareStore(store)
			}

			pp := &Provider{
				log:   log,
				store: store,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := pp.PasswordDelete(ctx, tc.givePassID)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_Passwords(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
		wantPasswords    []models.Password
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Passwords(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Password{}, errors.New("some error"))
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "crypto error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Passwords(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Password{
						{
							ID:       1,
							Title:    "title",
							Login:    "login",
							Password: "pass",
							URL:      "url",
							Note:     "note",
						},
					}, nil)
			},
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errPasswordProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Passwords(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Password{
						{
							ID:       1,
							Title:    "title",
							Login:    "login",
							Password: "pass",
							URL:      "url",
							Note:     "note",
						},
					}, nil)
			},
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("decpass", nil)
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
			wantPasswords: []models.Password{
				{
					ID:       1,
					Title:    "title",
					Login:    "login",
					Password: "decpass",
					URL:      "url",
					Note:     "note",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := mocks.NewStorage(t)
			if tc.prepareStore != nil {
				tc.prepareStore(store)
			}

			crp := mocks.NewCrypto(t)
			if tc.prepareCrypto != nil {
				tc.prepareCrypto(crp)
			}

			pp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			gotPasswords, err := pp.Passwords(ctx)
			tc.wantErrFunc(t, err)
			require.Equal(t, tc.wantPasswords, gotPasswords)
		})
	}
}
