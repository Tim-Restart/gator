package main

import (
	"fmt"
	"gator/internal/config"
	"log"
	"github.com/google/uuid"
	"gator/internal/database"

	"context"
	"database/sql"
	"errors"
)

func getUserId(user string) (uuid.UUID, error) {
	var id uuid.UUID
	ctx := context.Background()
	userName, err := s.db.GetUser(ctx, user)
	if errors.Is(err, sql.ErrNoRows) {
		log.Print("User doesn't exist")
	} else if err != nil {
		fmt.Print("Error checking existing user")
		return nil, err
	} else {
		id = userName.ID
	}
	return id, nil
}