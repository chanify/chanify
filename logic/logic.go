package logic

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
	"github.com/chanify/chanify/pb"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

var (
	ApiEndpoint            = "https://api.chanify.net"
	MockPusher  APNSPusher = nil

	randReader = rand.Read
)

type Options struct {
	Name     string
	Version  string
	Endpoint string
	DataPath string
	DBUrl    string
	Secret   string
}

type Logic struct {
	srvless  bool
	db       model.DB
	secKey   *crypto.SecretKey
	Name     string
	NodeID   string
	Version  string
	Endpoint string
	Features []string

	apnsPClient *apns2.Client
	apnsDClient *apns2.Client
}

type APNSPusher interface {
	Push(n *apns2.Notification) (*apns2.Response, error)
}

const authKey = "MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQgQ6vCLkUeDj223nfPfKGrjG+Coc53EbKHmO6Oa9YcHiGgCgYIKoZIzj0DAQehRANCAAQNwg3W2eOqNlX0nl9kGbfmMxwSZoO4RmqKoKJnH/vGkU8csJuN5Dg4JiI6ni5PEx+A1rb19DuDm4AzwBVvl8Jt"

func (opts *Options) fixDataPath() {
	if len(opts.DBUrl) <= 0 && len(opts.Secret) <= 0 && len(opts.DataPath) > 0 {
		s, err := os.Stat(opts.DataPath)
		if err == nil && s.IsDir() {
			opts.DBUrl = "sqlite://" + filepath.Join(opts.DataPath, "chanify.db")
		}
	}
}

func NewLogic(opts *Options) (*Logic, error) {
	opts.fixDataPath()
	l := &Logic{
		srvless:  false,
		Name:     opts.Name,
		Version:  opts.Version,
		Endpoint: opts.Endpoint,
		Features: []string{"msg.text"},
	}
	if len(opts.DBUrl) <= 0 {
		if len(opts.Secret) > 0 {
			opts.DBUrl = "nosql://?secret=" + url.QueryEscape(opts.Secret)
			l.srvless = true
		} else {
			return nil, errors.New("In serverless mode, secret is required")
		}
	}
	if err := l.loadDB(opts.DBUrl); err != nil {
		return nil, err
	}
	l.NodeID = l.secKey.ToID(0x01)
	if !l.srvless {
		l.Features = append([]string{"store.device"}, l.Features...)
		key, _ := base64.RawStdEncoding.DecodeString(authKey)
		akey, _ := x509.ParsePKCS8PrivateKey(key)
		tk := &token.Token{
			AuthKey: akey.(*ecdsa.PrivateKey),
			KeyID:   "CPBF9RLA6G",
			TeamID:  "P4XS4AVCLW",
		}
		l.apnsPClient = apns2.NewTokenClient(tk).Production()
		l.apnsDClient = apns2.NewTokenClient(tk).Development()
	}
	log.Printf("Node server name: %s, version: %s, serverless: %v, node-id: %s\n", l.Name, l.Version, l.srvless, l.NodeID)
	return l, nil
}

func (l *Logic) Close() {
	if l.db != nil {
		l.db.Close()
		l.db = nil
	}
}

func (l *Logic) GetUser(uid string) (*model.User, error) {
	return l.db.GetUser(uid)
}

func (l *Logic) GetUserKey(uid string) ([]byte, error) {
	u, err := l.db.GetUser(uid)
	if err != nil {
		return nil, err
	}
	return u.SecretKey, nil
}

func (l *Logic) UpsertUser(uid string, key string, serverless bool) (*model.User, error) {
	pk, err := model.CalcUserKey(uid, key)
	if err != nil {
		return nil, err
	}
	u, err := l.db.GetUser(uid)
	if err != nil {
		u, err = l.createUser(uid, pk, serverless)
		if err != nil {
			return nil, err
		}
	} else {
		if u.IsServerless() != serverless {
			u.SetServerless(serverless)
			if err := l.db.UpsertUser(u); err != nil {
				return nil, err
			}
		}
	}
	if pk != nil && len(u.PublicKey) <= 0 {
		u.PublicKey = pk.MarshalPublicKey()
	}
	return u, nil
}

func (l *Logic) BindDevice(uid string, uuid string, key string) error {
	pk, err := model.CalcDeviceKey(uuid, key)
	if err != nil {
		return err
	}
	return l.db.BindDevice(uid, uuid, pk.MarshalPublicKey())
}

func (l *Logic) UnbindDevice(uid string, uuid string) error {
	return l.db.UnbindDevice(uid, uuid)
}

func (l *Logic) UpdatePushToken(uid string, uuid string, token string, sandbox bool) error {
	tk, err := model.DecodePushToken(token)
	if err != nil {
		return err
	}
	return l.db.UpdatePushToken(uid, uuid, tk, sandbox)
}

func (l *Logic) GetDevices(uid string) ([]*model.Device, error) {
	return l.db.GetDevices(uid)
}

func (l *Logic) CreateMessage(tk *model.Token) *pb.Message {
	return &pb.Message{
		From:    tk.GetNodeID(),
		Channel: tk.GetChannel(),
	}
}

func (l *Logic) GetAPNS(sandbox bool) APNSPusher {
	if MockPusher != nil {
		return MockPusher
	}
	if sandbox {
		return l.apnsDClient
	}
	return l.apnsPClient
}

func (l *Logic) SendAPNS(uid string, data []byte, devices []*model.Device) error {
	notification := &apns2.Notification{}
	notification.Topic = "net.chanify.ios"
	notification.Payload = payload.NewPayload().MutableContent().AlertLocKey("NewMsg").
		Custom("uid", uid).
		Custom("src", l.NodeID).
		Custom("msg", base64.RawURLEncoding.EncodeToString(data))
	for _, dev := range devices {
		notification.DeviceToken = hex.EncodeToString(dev.Token)
		l.GetAPNS(dev.Sandbox).Push(notification) // nolint: errcheck
	}
	return nil
}

func (l *Logic) loadDB(dburl string) error {
	var err error
	l.db, err = model.InitDB(dburl)
	if err != nil {
		log.Println("Open database failed:", err)
		return err
	}
	var secret []byte
	if err := l.db.GetOption("secret", &secret); err == nil {
		l.secKey, _ = crypto.LoadSecretKey(secret)
	}
	return l.fixSecretKey()
}

func (l *Logic) fixSecretKey() error {
	if l.secKey == nil {
		l.secKey = crypto.GenerateSecretKey(nil)
		if err := l.db.SetOption("secret", l.secKey.MarshalSecretKey()); err != nil {
			log.Println("Save secret key failed")
			return err
		}
		log.Println("Generate new secret key")
	}
	return nil
}

func (l *Logic) createUser(uid string, pk *crypto.PublicKey, serverless bool) (*model.User, error) {
	u := &model.User{
		Uid:       uid,
		PublicKey: pk.MarshalPublicKey(),
		SecretKey: make([]byte, 64),
	}
	if _, err := randReader(u.SecretKey); err != nil {
		return nil, err
	}
	u.SetServerless(serverless)
	if err := l.db.UpsertUser(u); err != nil {
		return nil, err
	}
	return u, nil
}
