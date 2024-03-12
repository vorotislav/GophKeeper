package cards

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/cards/mocks"
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

func TestProvider_CardCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		giveCard         models.Card
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "crypto error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveCard: models.Card{
				Name:     "name",
				Number:   "title",
				CVC:      "cvc",
				ExpMonth: 1,
				ExpYear:  1,
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "store error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encnumber", nil)
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("enccvc", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().CardCreate(
					mock.Anything,
					mock.AnythingOfType("models.Card"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveCard: models.Card{
				Name:     "name",
				Number:   "title",
				CVC:      "cvc",
				ExpMonth: 1,
				ExpYear:  1,
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "success",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encnumber", nil)
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("enccvc", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().CardCreate(
					mock.Anything,
					mock.AnythingOfType("models.Card"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveCard: models.Card{
				Name:     "name",
				Number:   "title",
				CVC:      "cvc",
				ExpMonth: 1,
				ExpYear:  1,
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

			cp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := cp.CardCreate(ctx, tc.giveCard)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_CardUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		prepareCrypto    func(crp *mocks.Crypto)
		giveCard         models.Card
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "crypto error number",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveCard: models.Card{
				ID:       1,
				Name:     "name",
				Number:   "number",
				CVC:      "cvc",
				ExpMonth: 1,
				ExpYear:  1,
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "crypto error cvc",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("enc number", nil)
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveCard: models.Card{
				ID:       1,
				Name:     "name",
				Number:   "number",
				CVC:      "cvc",
				ExpMonth: 1,
				ExpYear:  1,
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "store error",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encnumber", nil)
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("enccvc", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().CardUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Card"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveCard: models.Card{
				ID:       1,
				Name:     "name",
				Number:   "number",
				CVC:      "cvc",
				ExpMonth: 1,
				ExpYear:  1,
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "success",
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("encnumber", nil)
				crp.EXPECT().EncryptString(mock.AnythingOfType("string")).
					Once().
					Return("enccvc", nil)
			},
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().CardUpdate(
					mock.Anything,
					mock.AnythingOfType("models.Card"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveCard: models.Card{
				ID:       1,
				Name:     "name",
				Number:   "number",
				CVC:      "cvc",
				ExpMonth: 1,
				ExpYear:  1,
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

			cp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := cp.CardUpdate(ctx, tc.giveCard)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_CardDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name             string
		prepareStore     func(store *mocks.Storage)
		giveCardID       int
		giveTokenPayload token.Payload
		wantErrFunc      require.ErrorAssertionFunc
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().CardDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(errors.New("some error"))
			},
			giveCardID:       1,
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().CardDelete(
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int")).
					Once().
					Return(nil)
			},
			giveCardID:       1,
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

			cp := &Provider{
				log:   log,
				store: store,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			err := cp.CardDelete(ctx, tc.giveCardID)
			tc.wantErrFunc(t, err)
		})
	}
}

func TestProvider_Cards(t *testing.T) {
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
		wantCards        []models.Card
	}{
		{
			name: "token error",
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "store error",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Cards(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Card{}, errors.New("some error"))
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "crypto error 1",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Cards(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Card{
						{
							ID:       1,
							Name:     "name",
							Number:   "number",
							CVC:      "cvc",
							ExpMonth: 1,
							ExpYear:  1,
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
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "crypto error 2",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Cards(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Card{
						{
							ID:       1,
							Name:     "name",
							Number:   "number",
							CVC:      "cvc",
							ExpMonth: 1,
							ExpYear:  1,
						},
					}, nil)
			},
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("dec number", nil)
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("some error"))
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, errCardProvider)
			},
		},
		{
			name: "success",
			prepareStore: func(store *mocks.Storage) {
				store.EXPECT().Cards(
					mock.Anything,
					mock.AnythingOfType("int")).
					Once().
					Return([]models.Card{
						{
							ID:       1,
							Name:     "name",
							Number:   "number",
							CVC:      "cvc",
							ExpMonth: 1,
							ExpYear:  1,
						},
					}, nil)
			},
			prepareCrypto: func(crp *mocks.Crypto) {
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("dec number", nil)
				crp.EXPECT().DecryptString(mock.AnythingOfType("string")).
					Once().
					Return("dec cvc", nil)
			},
			giveTokenPayload: token.Payload{ID: 1},
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
			wantCards: []models.Card{
				{
					ID:       1,
					Name:     "name",
					Number:   "dec number",
					CVC:      "dec cvc",
					ExpMonth: 1,
					ExpYear:  1,
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

			cp := &Provider{
				log:    log,
				store:  store,
				crypto: crp,
			}

			ctx := context.Background()
			if tc.giveTokenPayload.ID != 0 {
				ctx = token.ToContext(ctx, tc.giveTokenPayload)
			}

			gotCards, err := cp.Cards(ctx)
			tc.wantErrFunc(t, err)
			require.Equal(t, tc.wantCards, gotCards)
		})
	}
}
