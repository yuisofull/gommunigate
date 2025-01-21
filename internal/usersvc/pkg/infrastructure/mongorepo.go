package infrastructure

import (
	"context"
	uuid "github.com/google/uuid"
	"github.com/yuisofull/gommunigate/internal/usersvc/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

type mongoRepository struct {
	client     *mongo.Client
	db         string
	collection string
}

func NewMongoRepository(client *mongo.Client, db, collection string) *mongoRepository {
	return &mongoRepository{
		client:     client,
		db:         db,
		collection: collection,
	}
}

func (m *mongoRepository) CreateUser(ctx context.Context, u model.User) error {
	collection := m.client.Database(m.db).Collection(m.collection)
	id := oidFromUUID(*u.UUID)
	_, err := collection.InsertOne(ctx, createUserQuery{
		UUID:           id,
		Email:          u.Email,
		PhoneNumber:    u.PhoneNumber,
		UserName:       u.UserName,
		ProfilePicture: u.ProfilePicture,
		Bio:            u.Bio,
		AuthProvider:   u.AuthProvider,
	})
	return err
}

func (m *mongoRepository) GetUser(ctx context.Context, uid string) (model.User, error) {
	collection := m.client.Database(m.db).Collection(m.collection)
	var resp getUserResponse
	q := getUserQuery{UUID: oidFromUUID(uid)}
	err := collection.FindOne(ctx, q).Decode(&resp)
	if err != nil {
		return model.User{}, err
	}

	id := uuidFromOID(resp.UUID)
	return model.User{
		UUID:           &id,
		Email:          resp.Email,
		PhoneNumber:    resp.PhoneNumber,
		UserName:       resp.UserName,
		ProfilePicture: resp.ProfilePicture,
		Bio:            resp.Bio,
	}, nil
}

func (m *mongoRepository) UpdateUser(ctx context.Context, u model.User) error {
	collection := m.client.Database(m.db).Collection(m.collection)
	id := oidFromUUID(*u.UUID)
	filter := updateUserQuery{UUID: id}
	query := bson.D{{"$set", updateUserQuery{
		Email:          u.Email,
		PhoneNumber:    u.PhoneNumber,
		UserName:       u.UserName,
		ProfilePicture: u.ProfilePicture,
		Bio:            u.Bio,
	}}}
	_, err := collection.UpdateOne(ctx, filter, query)
	return err
}

func (m *mongoRepository) DeleteUser(ctx context.Context, uid string) error {
	collection := m.client.Database(m.db).Collection(m.collection)
	_, err := collection.DeleteOne(ctx, deleteUserQuery{UUID: oidFromUUID(uid)})
	return err
}

func (m *mongoRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

type createUserQuery struct {
	UUID           []byte  `bson:"_id,omitempty"`
	Email          *string `bson:"email,omitempty"`
	PhoneNumber    *string `bson:"phoneNumber,omitempty"`
	UserName       *string `bson:"userName,omitempty"`
	ProfilePicture *string `bson:"profilePicture,omitempty"`
	Bio            *string `bson:"bio,omitempty"`
	AuthProvider   *string `bson:"authProvider,omitempty"`
}

type getUserQuery struct {
	UUID []byte `bson:"_id"`
}

type getUserResponse struct {
	UUID           []byte  `bson:"_id,omitempty"`
	Email          *string `bson:"email,omitempty"`
	PhoneNumber    *string `bson:"phoneNumber,omitempty"`
	UserName       *string `bson:"userName,omitempty"`
	ProfilePicture *string `bson:"profilePicture,omitempty"`
	Bio            *string `bson:"bio,omitempty"`
	AuthProvider   *string `bson:"authProvider,omitempty"`
}

type updateUserQuery struct {
	UUID           []byte  `bson:"_id,omitempty"`
	Email          *string `bson:"email,omitempty"`
	PhoneNumber    *string `bson:"phoneNumber,omitempty"`
	UserName       *string `bson:"userName,omitempty"`
	ProfilePicture *string `bson:"profilePicture,omitempty"`
	Bio            *string `bson:"bio,omitempty"`
}

type deleteUserQuery struct {
	UUID []byte `bson:"_id"`
}

func oidFromUUID(uuid2 string) []byte {
	if uuid2 == "" {
		id := uuid.New()
		return id[:]
	}
	Uuid, err := uuid.Parse(uuid2)
	if err != nil {
		return []byte(uuid2)
	}
	return Uuid[:]
}

func uuidFromOID(oid []byte) string {
	if len(oid) == 0 {
		return uuid.New().String()
	}
	id, err := uuid.FromBytes(oid)
	if err != nil {
		return string(oid)
	}
	return id.String()
}
