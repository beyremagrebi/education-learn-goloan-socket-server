package types

type UserConnection struct {
	UserId     string `json:"userId"`
	FullName   string `json:"fullName"`
	FacilityId string `json:"FacilityId"`
}

type User struct {
	UserID   string `json:"userId"`
	SocketID string `json:"socketId"`
	Status   string `json:"status"`
}
