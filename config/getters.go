package config

import "github.com/google/uuid"

func GetAgent() uuid.UUID {
	u, err := uuid.FromBytes(ID[:16])
	if err != nil {
		return uuid.Nil
	}
	return u
}
