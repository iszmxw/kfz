package admin

import "goapi/app/response"

type Session struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func (s Session) UserDTO() response.AdminUserDTO {
	return response.AdminUserDTO{
		ID:       s.UserID,
		Username: s.Username,
		Name:     s.Name,
		Status:   s.Status,
		Roles:    s.Roles,
	}
}
