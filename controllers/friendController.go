package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	helper "github.com/Mutay1/chat-backend/helpers"
	"github.com/Mutay1/chat-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Friend struct {
	FriendID string `json:"friendID"`
	State    bool   `json:"state"`
}

func findFriend(userID string, friendID string) (f models.Friendship, err error) {
	var friendship models.Friendship
	senderID, _ := primitive.ObjectIDFromHex(userID)
	recipientID, _ := primitive.ObjectIDFromHex(friendID)
	filter := bson.D{
		{"$or",
			bson.A{
				bson.M{
					"$and": []interface{}{
						bson.M{"requester._id": senderID},
						bson.M{"recipient._id": recipientID},
					},
				},
				bson.M{
					"$and": []interface{}{
						bson.M{"requester._id": recipientID},
						bson.M{"recipient._id": senderID},
					},
				},
			},
		},
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	err = friendshipCollection.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		fmt.Println(err)
		return friendship, err
	}
	defer cancel()
	return friendship, nil
}

func updateFriendship(friendship models.Friendship) error {
	data, err := helper.ToDoc(friendship)
	if err != nil {
		return err
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	_, err = friendshipCollection.UpdateOne(ctx, bson.M{"_id": friendship.ID}, bson.D{
		{"$set", data},
	})
	if err != nil {
		return err
	}

	defer cancel()
	return nil
}

func GetFriends() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.GetString("uid"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid UserID"})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		filter := bson.D{
			{"$and",
				bson.A{
					bson.D{{"accepted", bson.D{{"$eq", true}}}},
					bson.M{
						"$or": []interface{}{
							bson.M{"requester._id": id},
							bson.M{"recipient._id": id},
						},
					},
				},
			},
		}
		cursor, err := friendshipCollection.Find(ctx, filter)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var requestsLoaded []models.Friendship
		var friendsLoaded []models.Friend
		ID := c.GetString(("uid"))
		if err = cursor.All(ctx, &requestsLoaded); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred"})
		}
		for _, user := range requestsLoaded {
			if user.Requester.ID.Hex() != ID {

				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				friendsLoaded = append(friendsLoaded, user.Requester)
			} else {
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				friendsLoaded = append(friendsLoaded, user.Recipient)
			}
		}
		defer cancel()
		c.JSON(http.StatusOK, friendsLoaded)
	}
}

func Archive() gin.HandlerFunc {
	return func(c *gin.Context) {
		var friend Friend
		if err := c.BindJSON(&friend); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(friend.State)
		friendship, err := findFriend(c.GetString("uid"), friend.FriendID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if friendship.Recipient.ID.Hex() == friend.FriendID {
			friendship.Recipient.Archived = friend.State
			fmt.Println("UPDATING")
		} else {
			friendship.Requester.Archived = friend.State
			fmt.Println("UPDATING")
		}
		err = updateFriendship(friendship)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": true,
		})
	}
}

func Favorite() gin.HandlerFunc {
	return func(c *gin.Context) {
		var friend Friend
		if err := c.BindJSON(&friend); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		friendship, err := findFriend(c.GetString("uid"), friend.FriendID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if friendship.Recipient.ID.Hex() == friend.FriendID {
			friendship.Recipient.Favorite = friend.State
		} else {
			friendship.Requester.Favorite = friend.State
		}
		err = updateFriendship(friendship)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": true,
		})
	}
}

func Block() gin.HandlerFunc {
	return func(c *gin.Context) {
		var friend Friend
		if err := c.BindJSON(&friend); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		friendship, err := findFriend(c.GetString("uid"), friend.FriendID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if friendship.Recipient.ID.Hex() == friend.FriendID {
			friendship.Recipient.Blocked = friend.State
		} else {
			friendship.Requester.Blocked = friend.State
		}
		err = updateFriendship(friendship)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": true,
		})
	}
}
