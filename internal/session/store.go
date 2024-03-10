package session

import "GophKeeper/internal/models"

type Storage struct {
	session models.Session
}

func (s *Storage) GetSession() models.Session {
	return s.session
}

func (s *Storage) SaveSession(ses models.Session) {
	s.session = ses
}
