package toolspkg

import (
	"github.com/google/uuid"
)

func GenerateUUID() string {
	uid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return uid.String()
}
