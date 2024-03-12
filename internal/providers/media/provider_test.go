package media

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/media/mocks"
	"GophKeeper/internal/repository"
	"GophKeeper/internal/token"
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
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

func TestProvider_MediaCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		giveMedia        models.Media
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
			},
		},
		{
			name: "crypto error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveMedia: models.Media{
				Title:     "title",
				Body:      []byte(`some body`),
				MediaType: "type",
				Note:      "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
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
				store.EXPECT().MediaCreate(
					mock.Anything,
					mock.AnythingOfType("models.Media"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveMedia: models.Media{
				Title:     "title",
				Body:      []byte(`some body`),
				MediaType: "type",
				Note:      "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
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
				store.EXPECT().MediaCreate(
					mock.Anything,
					mock.AnythingOfType("models.Media"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveMedia: models.Media{
				Title:     "title",
				Body:      []byte(`some body`),
				MediaType: "type",
				Note:      "note",
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

			mp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := mp.MediaCreate(ctx, tc.giveMedia)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_MediaUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		giveMedia        models.Media
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
			},
		},
		{
			name: "crypto error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveMedia: models.Media{
				ID:        1,
				Title:     "title",
				Body:      []byte(`some body`),
				MediaType: "type",
				Note:      "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
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
				store.EXPECT().MediaUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Media"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveMedia: models.Media{
				ID:        1,
				Title:     "title",
				Body:      []byte(`some body`),
				MediaType: "type",
				Note:      "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
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
				store.EXPECT().MediaUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Media"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveMedia: models.Media{
				ID:        1,
				Title:     "title",
				Body:      []byte(`some body`),
				MediaType: "type",
				Note:      "note",
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

			mp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := mp.MediaUpdate(ctx, tc.giveMedia)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_MediaDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		giveMediaID      int
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().MediaDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveMediaID:      1,
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().MediaDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveMediaID:      1,
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

			mp := &Provider{
				log:   log,
				store: store,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := mp.MediaDelete(ctx, tc.giveMediaID)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_Medias(t *testing.T) {
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
		wantMedia        []models.Media
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Medias(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Media{}, errors.New("some error"))
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errMediaProvider)
			},
		},
		{
			name: "crypto error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Medias(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Media{
						{
							ID:        1,
							Title:     "title",
							Body:      []byte(`some body`),
							MediaType: "type",
							Note:      "note",
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
				require.ErrorIs(t, err, errMediaProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Medias(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Media{
						{
							ID:        1,
							Title:     "title",
							Body:      []byte(`some body`),
							MediaType: "type",
							Note:      "note",
						},
					}, nil)
			},
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("dec media body", nil)
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
			wantMedia: []models.Media{
				{
					ID:        1,
					Title:     "title",
					Body:      []byte(`dec media body`),
					MediaType: "type",
					Note:      "note",
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

			mp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			gotMedias, err := mp.Medias(ctx)
			tc.wantErrFunc(t, err)
			require.Equal(t, tc.wantMedia, gotMedias)
		})
	}
}
