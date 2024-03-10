package users

import (
	"context"
	"encoding/json"
	"errors"

	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"

	"go.uber.org/zap"
	"net/http"
)

type UserProvider interface {
	UserCreate(ctx context.Context, um models.UserMachine) (models.Session, error)
	UserLogin(ctx context.Context, um models.UserMachine) (models.Session, error)
	//UserLogout(ctx context.Context, um models.UserMachine) error
}

type Handler struct {
	log          *zap.Logger
	userProvider UserProvider
}

func NewHandler(log *zap.Logger, up UserProvider) *Handler {
	return &Handler{
		log:          log.Named("user handler"),
		userProvider: up,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in input

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed to register user decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed to register user decode", err.Error()))

		return
	}

	res, err := h.userProvider.UserCreate(r.Context(), models.UserMachine{
		User: models.User{
			Login:    in.User.Login,
			Password: in.User.Password,
		},
		Machine: models.Machine{
			IPAddress:  in.Machine.IPAddress,
			MACAddress: in.Machine.MACAddress,
			PublicKey:  in.Machine.PublicKey,
		},
	})

	if err != nil {
		h.log.Error("failed to user register", zap.Error(err))

		responder.JSON(w, httpErr.NewInternalError("failed to register user decode", err.Error()))

		return
	}

	s := session{
		ID:           res.ID,
		Token:        res.AccessToken,
		RefreshToken: res.RefreshToken,
	}

	responder.JSON(w, &output{
		session:    s,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success user register")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in input

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed to login user decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed to login user decode", err.Error()))

		return
	}

	res, err := h.userProvider.UserLogin(r.Context(), models.UserMachine{
		User: models.User{
			Login:    in.User.Login,
			Password: in.User.Password,
		},
		Machine: models.Machine{
			IPAddress:  in.Machine.IPAddress,
			MACAddress: in.Machine.MACAddress,
			PublicKey:  in.Machine.PublicKey,
		},
	})

	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Error("failed to user login", zap.Error(err))

			responder.JSON(w, httpErr.NewNotFoundError("user not found"))

			return
		}

		h.log.Error("failed to user register", zap.Error(err))

		responder.JSON(w, httpErr.NewInternalError("failed to login user", err.Error()))

		return
	}

	s := session{
		ID:           res.ID,
		Token:        res.AccessToken,
		RefreshToken: res.RefreshToken,
	}

	responder.JSON(w, &output{
		session:    s,
		statusCode: http.StatusOK,
	})

	h.log.Info("success user login")
}

type input struct {
	User struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"user"`
	Machine struct {
		IPAddress  string `json:"ip_address"`
		MACAddress string `json:"mac_address"`
		PublicKey  string `json:"public_key"`
	} `json:"machine"`
}

type session struct {
	ID           int64  `json:"id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type output struct {
	session    session
	statusCode int
}

// ToJSON converts output structure into JSON representation.
func (o *output) ToJSON() ([]byte, error) { return json.Marshal(o.session) }

// StatusCode allows to customize output HTTP status code (when responder.JSON is used).
func (o *output) StatusCode() int { return o.statusCode }
