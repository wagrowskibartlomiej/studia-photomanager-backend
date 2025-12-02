package main

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type DBUser struct {
	ID       int64
	Login    string
	Password string
}

type Photo struct {
	Filename string `json:"filename"`
	Public   bool   `json:"public"`
}

type PublicPhoto struct {
	User     string `json:"user"`
	Filename string `json:"filename"`
}

type UpdatePublicRequest struct {
	Filename string `json:"filename"`
	Public   int    `json:"public"` // 0 OR 1
}

type UserResponse struct {
	Login    string `json:"login"`
	IsBanned bool   `json:"isBanned"`
}

type ManageBanRequest struct {
	Login  string `json:"login"`
	Banned int    `json:"banned"` // 0 = unban, 1 = ban
}

