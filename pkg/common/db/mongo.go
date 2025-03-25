package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func MongoInit(mongoUrl string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.TODO())
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	db := client.Database("mangacage")

	if err = db.CreateCollection(context.TODO(), "titles_on_moderation_covers"); err != nil {
		log.Println(err)
	}

	collection := db.Collection("titles_on_moderation_covers")

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"title_on_moderation_id": 1},
		Options: options.Index().SetUnique(true),
	}

	if _, err = collection.Indexes().CreateOne(context.TODO(), indexModel); err != nil {
		log.Println(err)
	}

	indexModel = mongo.IndexModel{
		Keys:    bson.M{"title_id": 1},
		Options: options.Index().SetUnique(true),
	}

	if _, err = collection.Indexes().CreateOne(context.TODO(), indexModel); err != nil {
		log.Println(err)
	}

	if err = db.CreateCollection(context.TODO(), "titles_covers"); err != nil {
		log.Println(err)
	}

	collection = db.Collection("titles_covers")

	if _, err := collection.Indexes().CreateOne(context.TODO(), indexModel); err != nil {
		log.Println(err)
	}

	if err = db.CreateCollection(context.TODO(), "chapters_on_moderation_pages"); err != nil {
		log.Println(err)
	}

	collection = db.Collection("chapters_on_moderation_pages")

	indexModel = mongo.IndexModel{
		Keys:    bson.M{"chapter_id": 1},
		Options: options.Index().SetUnique(true),
	}

	if _, err = collection.Indexes().CreateOne(context.TODO(), indexModel); err != nil {
		log.Println(err)
	}

	if err = db.CreateCollection(context.TODO(), "chapters_pages"); err != nil {
		log.Println(err)
	}

	collection = db.Collection("chapters_pages")

	if _, err = collection.Indexes().CreateOne(context.TODO(), indexModel); err != nil {
		log.Println(err)
	}

	return client, nil
}
