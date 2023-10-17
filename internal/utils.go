package internal

import "go.mongodb.org/mongo-driver/bson/primitive"

func ContainsObjectID(s []primitive.ObjectID, str primitive.ObjectID) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
