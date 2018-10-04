package storage

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"sync"

	"bytes"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	"github.com/polyverse/play-with-polyverse/pwd/types"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

var (
	s3ObjectKey = "play-with-polyverse.json"
)

type s3Storage struct {
	rw       sync.Mutex
	db       *DB
	s3Bucket string
	s3svc    *s3.S3
}

func (store *s3Storage) SessionGet(id string) (*types.Session, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	s, found := store.db.Sessions[id]
	if !found {
		return nil, NotFoundError
	}

	return s, nil
}

func (store *s3Storage) SessionGetAll() ([]*types.Session, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	sessions := make([]*types.Session, len(store.db.Sessions))
	i := 0
	for _, s := range store.db.Sessions {
		sessions[i] = s
		i++
	}

	return sessions, nil
}

func (store *s3Storage) SessionPut(session *types.Session) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	store.db.Sessions[session.Id] = session

	return store.save()
}

func (store *s3Storage) SessionDelete(id string) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	_, found := store.db.Sessions[id]
	if !found {
		return nil
	}
	for _, i := range store.db.WindowsInstancesBySessionId[id] {
		delete(store.db.WindowsInstances, i)
	}
	store.db.WindowsInstancesBySessionId[id] = []string{}
	for _, i := range store.db.InstancesBySessionId[id] {
		delete(store.db.Instances, i)
	}
	store.db.InstancesBySessionId[id] = []string{}
	for _, i := range store.db.ClientsBySessionId[id] {
		delete(store.db.Clients, i)
	}
	store.db.ClientsBySessionId[id] = []string{}
	delete(store.db.Sessions, id)

	return store.save()
}

func (store *s3Storage) SessionCount() (int, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	return len(store.db.Sessions), nil
}

func (store *s3Storage) InstanceGet(name string) (*types.Instance, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	i := store.db.Instances[name]
	if i == nil {
		return nil, NotFoundError
	}
	return i, nil
}

func (store *s3Storage) InstancePut(instance *types.Instance) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	_, found := store.db.Sessions[string(instance.SessionId)]
	if !found {
		return NotFoundError
	}

	store.db.Instances[instance.Name] = instance
	found = false
	for _, i := range store.db.InstancesBySessionId[string(instance.SessionId)] {
		if i == instance.Name {
			found = true
			break
		}
	}
	if !found {
		store.db.InstancesBySessionId[string(instance.SessionId)] = append(store.db.InstancesBySessionId[string(instance.SessionId)], instance.Name)
	}

	return store.save()
}

func (store *s3Storage) InstanceDelete(name string) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	instance, found := store.db.Instances[name]
	if !found {
		return nil
	}

	instances := store.db.InstancesBySessionId[string(instance.SessionId)]
	for n, i := range instances {
		if i == name {
			instances = append(instances[:n], instances[n+1:]...)
			break
		}
	}
	store.db.InstancesBySessionId[string(instance.SessionId)] = instances
	delete(store.db.Instances, name)

	return store.save()
}

func (store *s3Storage) InstanceCount() (int, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	return len(store.db.Instances), nil
}

func (store *s3Storage) InstanceFindBySessionId(sessionId string) ([]*types.Instance, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	instanceIds := store.db.InstancesBySessionId[sessionId]
	instances := make([]*types.Instance, len(instanceIds))
	for i, id := range instanceIds {
		instances[i] = store.db.Instances[id]
	}

	return instances, nil
}

func (store *s3Storage) WindowsInstanceGetAll() ([]*types.WindowsInstance, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	instances := []*types.WindowsInstance{}

	for _, s := range store.db.WindowsInstances {
		instances = append(instances, s)
	}

	return instances, nil
}

func (store *s3Storage) WindowsInstancePut(instance *types.WindowsInstance) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	_, found := store.db.Sessions[string(instance.SessionId)]
	if !found {
		return NotFoundError
	}
	store.db.WindowsInstances[instance.Id] = instance
	found = false
	for _, i := range store.db.WindowsInstancesBySessionId[string(instance.SessionId)] {
		if i == instance.Id {
			found = true
			break
		}
	}
	if !found {
		store.db.WindowsInstancesBySessionId[string(instance.SessionId)] = append(store.db.WindowsInstancesBySessionId[string(instance.SessionId)], instance.Id)
	}

	return store.save()
}

func (store *s3Storage) WindowsInstanceDelete(id string) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	instance, found := store.db.WindowsInstances[id]
	if !found {
		return nil
	}

	instances := store.db.WindowsInstancesBySessionId[string(instance.SessionId)]
	for n, i := range instances {
		if i == id {
			instances = append(instances[:n], instances[n+1:]...)
			break
		}
	}
	store.db.WindowsInstancesBySessionId[string(instance.SessionId)] = instances
	delete(store.db.WindowsInstances, id)

	return store.save()
}

func (store *s3Storage) ClientGet(id string) (*types.Client, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	i := store.db.Clients[id]
	if i == nil {
		return nil, NotFoundError
	}
	return i, nil
}
func (store *s3Storage) ClientPut(client *types.Client) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	_, found := store.db.Sessions[string(client.SessionId)]
	if !found {
		return NotFoundError
	}

	store.db.Clients[client.Id] = client
	found = false
	for _, i := range store.db.ClientsBySessionId[string(client.SessionId)] {
		if i == client.Id {
			found = true
			break
		}
	}
	if !found {
		store.db.ClientsBySessionId[string(client.SessionId)] = append(store.db.ClientsBySessionId[string(client.SessionId)], client.Id)
	}

	return store.save()
}
func (store *s3Storage) ClientDelete(id string) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	client, found := store.db.Clients[id]
	if !found {
		return nil
	}

	clients := store.db.ClientsBySessionId[string(client.SessionId)]
	for n, i := range clients {
		if i == client.Id {
			clients = append(clients[:n], clients[n+1:]...)
			break
		}
	}
	store.db.ClientsBySessionId[string(client.SessionId)] = clients
	delete(store.db.Clients, id)

	return store.save()
}
func (store *s3Storage) ClientCount() (int, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	return len(store.db.Clients), nil
}
func (store *s3Storage) ClientFindBySessionId(sessionId string) ([]*types.Client, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	clientIds := store.db.ClientsBySessionId[sessionId]
	clients := make([]*types.Client, len(clientIds))
	for i, id := range clientIds {
		clients[i] = store.db.Clients[id]
	}

	return clients, nil
}

func (store *s3Storage) LoginRequestPut(loginRequest *types.LoginRequest) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	store.db.LoginRequests[loginRequest.Id] = loginRequest
	return nil
}
func (store *s3Storage) LoginRequestGet(id string) (*types.LoginRequest, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	if lr, found := store.db.LoginRequests[id]; !found {
		return nil, NotFoundError
	} else {
		return lr, nil
	}
}
func (store *s3Storage) LoginRequestDelete(id string) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	delete(store.db.LoginRequests, id)
	return nil
}

func (store *s3Storage) UserFindByProvider(providerName, providerUserId string) (*types.User, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	if userId, found := store.db.UsersByProvider[fmt.Sprintf("%s_%s", providerName, providerUserId)]; !found {
		return nil, NotFoundError
	} else {
		if user, found := store.db.Users[userId]; !found {
			return nil, NotFoundError
		} else {
			return user, nil
		}
	}
}

func (store *s3Storage) UserPut(user *types.User) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	store.db.UsersByProvider[fmt.Sprintf("%s_%s", user.Provider, user.ProviderUserId)] = user.Id
	store.db.Users[user.Id] = user

	return store.save()
}
func (store *s3Storage) UserGet(id string) (*types.User, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	if user, found := store.db.Users[id]; !found {
		return nil, NotFoundError
	} else {
		return user, nil
	}
}

func (store *s3Storage) PlaygroundPut(playground *types.Playground) error {
	store.rw.Lock()
	defer store.rw.Unlock()

	store.db.Playgrounds[playground.Id] = playground

	return store.save()
}
func (store *s3Storage) PlaygroundGet(id string) (*types.Playground, error) {
	store.rw.Lock()
	defer store.rw.Unlock()
	if playground, found := store.db.Playgrounds[id]; !found {
		return nil, NotFoundError
	} else {
		return playground, nil
	}
	return nil, NotFoundError
}

func (store *s3Storage) PlaygroundGetAll() ([]*types.Playground, error) {
	store.rw.Lock()
	defer store.rw.Unlock()

	playgrounds := make([]*types.Playground, len(store.db.Playgrounds))
	i := 0
	for _, p := range store.db.Playgrounds {
		playgrounds[i] = p
		i++
	}

	return playgrounds, nil
}

func (store *s3Storage) load() error {

	getReq := store.s3svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &store.s3Bucket,
		Key:    &s3ObjectKey,
	})

	getObjOut, err := getReq.Send()
	if err != nil {
		return errors.Wrapf(err, "Unable to get DB state from S3")
	}

	if err == nil {
		jsonblob, err := ioutil.ReadAll(getObjOut.Body)
		if err != nil {
			return errors.Wrapf(err, "Unable to read Get Object body from S3")
		}

		err = json.Unmarshal(jsonblob, &store.db)
		if err != nil {
			return errors.Wrapf(err, "Unable to unmarshall JSON obtained from S3 into DB")
		}
	} else {
		log.Infof("Generating a new blank DB due to first invocation")
		store.db = &DB{
			Sessions:                    map[string]*types.Session{},
			Instances:                   map[string]*types.Instance{},
			Clients:                     map[string]*types.Client{},
			WindowsInstances:            map[string]*types.WindowsInstance{},
			LoginRequests:               map[string]*types.LoginRequest{},
			Users:                       map[string]*types.User{},
			Playgrounds:                 map[string]*types.Playground{},
			WindowsInstancesBySessionId: map[string][]string{},
			InstancesBySessionId:        map[string][]string{},
			ClientsBySessionId:          map[string][]string{},
			UsersByProvider:             map[string]string{},
		}
	}
	return nil
}

func (store *s3Storage) save() error {

	jsonblob, err := json.Marshal(store.db)
	if err != nil {
		return errors.Wrapf(err, "Unable to serialize DB into JSON for storage.")
	}

	putReq := store.s3svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: &store.s3Bucket,
		Key:    &s3ObjectKey,
		Body:   bytes.NewReader(jsonblob),
	})

	putObjOut, err := putReq.Send()
	if err != nil {
		return errors.Wrapf(err, "Unable to push DB state to S3")
	}

	log.Infof("Put DB to S3 succeeded: %s", putObjOut.String())
	return nil
}

func DynamoDbStorage(config aws.Config, s3Bucket string) (StorageApi, error) {
	s := &s3Storage{
		s3svc:    s3.New(config),
		s3Bucket: s3Bucket,
	}

	err := s.load()
	if err != nil {
		return nil, err
	}

	return s, nil
}
