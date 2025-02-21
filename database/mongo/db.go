package mongo

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DB_TIMEOUT = 10 * time.Second
)

// MongoDB represents a connection to a mongo database
type MongoDB struct {
	Client     *mongo.Client
	clientOpts *options.ClientOptions
	dbname     string
	uri        string
}

// NewDatabase creates a database to load and save documents in a MongoDB database.
func NewDatabase() *MongoDB {
	log.Trace("--> mongo.NewDatabase")
	defer log.Trace("<-- mongo.NewDatabase")

	uri := os.Getenv("MONGODB_URI")
	dbname := os.Getenv("MONGODB_DATABASE")

	m := &MongoDB{
		uri:    uri,
		dbname: dbname,
	}

	// Wait for MongoDB to become active before proceeding
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	m.clientOpts = options.Client().ApplyURI(m.uri)
	m.Client, err = mongo.Connect(ctx, m.clientOpts)
	if err != nil {
		log.WithField("error", err).Fatal("unable to connect to the MongoDB database")
		return nil
	}

	// Check the connection
	err = m.Client.Ping(ctx, nil)
	if err != nil {
		log.WithField("error", err).Fatal("unable to ping the MongoDB database")
		err = nil
	}

	return m
}

// ListDocuments returns the ID of each document in a collection in the database.
func (m *MongoDB) ListDocuments(collectionName string, filter interface{}) ([]string, error) {
	log.Trace("--> mongoDB.ListDocuments")
	defer log.Trace("<-- mongoDB.ListDocuments")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	if m.clientOpts == nil {
		var err error
		m.clientOpts = options.Client().ApplyURI(m.uri)
		m.Client, err = mongo.Connect(ctx, m.clientOpts)
		if err != nil {
			log.Error("Unable to connect to the MongoDB database, error:", err)
			return nil, err
		}
	}

	db := m.Client.Database(m.dbname)
	collection := db.Collection(collectionName)
	if collection == nil {
		log.WithField("collection", collectionName).Error("Failed to create the collection")
		return nil, ErrCollectionNotAccessable
	}

	opts := options.Find().SetProjection(bson.M{"_id": 1})
	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("Failed to read the collection")
		return nil, ErrCollectionNotAccessable
	}

	type result struct {
		ID string `bson:"_id"`
	}
	var results []result
	err = cur.All(ctx, &results)
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("error getting IDs for the collection")
		return nil, ErrCollectionNotAccessable
	}

	idList := make([]string, 0, len(results))
	for _, r := range results {
		idList = append(idList, r.ID)
	}

	return idList, nil
}

// ReadAll reads all documents from the database that match the filter
func (m *MongoDB) ReadAll(collectionName string, filter interface{}, data interface{}, sortBy interface{}, limit int64) error {
	log.Trace("--> mongo.ReadAll")
	defer log.Trace("<-- mongoDB.ReadAll")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	db := m.Client.Database(m.dbname)
	collection := db.Collection(collectionName)
	if collection == nil {
		log.WithFields(log.Fields{"collection": collectionName}).Error("Failed to create the collection")
		return ErrCollectionNotAccessable
	}
	log.WithField("collection", collection.Name()).Debug("collection")

	// Limit the number of documents to return
	findOptions := options.Find()
	findOptions.SetLimit(limit)

	curr, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": err}).Error("unable to find the document")
		return err
	}
	err = curr.All(ctx, data)
	if err != nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": err}).Error("unable to decode the documents")
		return ErrInvalidDocument
	}
	return nil
}

// Read loads a document identified by documentID from the collection into data.
func (m *MongoDB) Read(collectionName string, filter interface{}, data interface{}) error {
	log.Trace("--> mongoDB.Read")
	defer log.Trace("<-- mongoDB.Read")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	db := m.Client.Database(m.dbname)
	collection := db.Collection(collectionName)
	if collection == nil {
		log.WithFields(log.Fields{"collection": collectionName}).Error("Failed to create the collection")
		return ErrCollectionNotAccessable
	}
	log.WithField("collection", collection.Name()).Debug("collection")

	res := collection.FindOne(ctx, filter)
	if res.Err() != nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": res.Err()}).Error("unable to find the document")
		return res.Err()
	}
	if res == nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter}).Error("unable to find the document")
		return ErrDocumentNotFound
	}
	err := res.Decode(data)
	if err != nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": err}).Error("unable to decode the document")
		return ErrInvalidDocument
	}
	return nil
}

// Write stores data into a documeent within the specified collection.
func (m *MongoDB) Write(collectionName string, filter interface{}, data interface{}) error {
	log.Trace("--> mongoDB.Write")
	defer log.Trace("<-- mongoDB.Write")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := m.Client.Database(m.dbname)
	if db == nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName}).Error("unable to create or access the database")
		return ErrDbInaccessable
	}

	collection := db.Collection(collectionName)
	if collection == nil {
		if err := db.CreateCollection(ctx, collectionName); err != nil {
			log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("unable to create the collection")
			return err
		}
		collection = db.Collection(collectionName)
	}

	update := bson.D{{Key: "$set", Value: data}}
	_, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "error": err, "data": data}).Error("unable to insert or update the document the collection")
		return err
	}
	log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "data": data}).Info("updated document in the collection")

	return nil
}

// Write stores data into multiple documeents within the specified collection.
func (m *MongoDB) WriteAll(collectionName string, filter interface{}, data interface{}) error {
	log.Trace("--> mongoDB.Write")
	defer log.Trace("<-- mongoDB.Write")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := m.Client.Database(m.dbname)
	if db == nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName}).Error("unable to create or access the database")
		return ErrDbInaccessable
	}

	collection := db.Collection(collectionName)
	if collection == nil {
		if err := db.CreateCollection(ctx, collectionName); err != nil {
			log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("unable to create the collection")
			return err
		}
		collection = db.Collection(collectionName)
	}

	// Matching 2 documents, modifying none. Why?
	update := bson.D{{Key: "$set", Value: data}}
	_, err := collection.UpdateMany(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "error": err, "data": data}).Error("unable to insert or update the document the collection")
		return err
	}
	log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "data": data}).Info("updated document in the collection")

	return nil
}

// Find returns the entries that match the provided query
func (m *MongoDB) Find() error {
	log.Trace("--> mongoDB.Find")
	defer log.Trace("<-- mongo")

	// TODO: implement

	return nil
}

// Close closes the mongo database client connection
func (m *MongoDB) Close() error {
	log.Trace("--> mongoDB.Close")
	defer log.Trace("<-- mongoDB.Close")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()
	if err := m.Client.Disconnect(ctx); err != nil {
		log.WithField("error", err).Error("unable to close the mongo database client")
		return err
	}
	return nil
}

// String returns the name of the database
func (db *MongoDB) String() string {
	return "mongo"
}
