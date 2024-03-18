// Package session предоставляет хранилище текущей сессии для клиентов.
package session

import "GophKeeper/internal/models"

// Storage описывает структуру для хранения текущий сессии.
type Storage struct {
	session models.Session
}

// GetSession возвращает сохраненную сессию.
func (s *Storage) GetSession() models.Session {
	return s.session
}

// SaveSession сохраняет переданную сессию.
func (s *Storage) SaveSession(ses models.Session) {
	s.session = ses
}
