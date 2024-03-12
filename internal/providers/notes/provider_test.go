package notes

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/notes/mocks"
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

func TestProvider_NoteCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		giveNote         models.Note
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
			},
		},
		{
			name: "crypto error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveNote: models.Note{
				Title: "title",
				Text:  "text",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
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
				store.EXPECT().NoteCreate(
					mock.Anything,
					mock.AnythingOfType("models.Note"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveNote: models.Note{
				Title: "title",
				Text:  "text",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
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
				store.EXPECT().NoteCreate(
					mock.Anything,
					mock.AnythingOfType("models.Note"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveNote: models.Note{
				Title: "title",
				Text:  "text",
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

			np := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := np.NoteCreate(ctx, tc.giveNote)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_NoteUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		giveNote         models.Note
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
			},
		},
		{
			name: "crypto error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveNote: models.Note{
				ID:    1,
				Title: "title",
				Text:  "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
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
				store.EXPECT().NoteUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Note"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveNote: models.Note{
				ID:    1,
				Title: "title",
				Text:  "note",
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
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
				store.EXPECT().NoteUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Note"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveNote: models.Note{
				ID:    1,
				Title: "title",
				Text:  "note",
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

			np := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := np.NoteUpdate(ctx, tc.giveNote)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_NoteDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		giveNoteID       int
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().NoteDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveNoteID:       1,
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().NoteDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveNoteID:       1,
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

			np := &Provider{
				log:   log,
				store: store,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := np.NoteDelete(ctx, tc.giveNoteID)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_Notes(t *testing.T) {
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
		wantNotes        []models.Note
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Notes(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Note{}, errors.New("some error"))
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errNoteProvider)
			},
		},
		{
			name: "crypto error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Notes(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Note{
						{
							ID:    1,
							Title: "title",
							Text:  "text",
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
				require.ErrorIs(t, err, errNoteProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Notes(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Note{
						{
							ID:    1,
							Title: "title",
							Text:  "text",
						},
					}, nil)
			},
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("dectext", nil)
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
			wantNotes: []models.Note{
				{
					ID:    1,
					Title: "title",
					Text:  "dectext",
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

			np := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			gotNotes, err := np.Notes(ctx)
			tc.wantErrFunc(t, err)
			require.Equal(t, tc.wantNotes, gotNotes)
		})
	}
}
