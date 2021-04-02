package logic

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

var (
	ApiEndpoint            = "https://api.chanify.net"
	MockPusher  APNSPusher = nil

	randReader = rand.Read

	ErrNoSupportMethod = errors.New("No support method")
	ErrInvalidContent  = errors.New("Invalid content")
	ErrSystemLimited   = errors.New("SystemLimited")
)

type Options struct {
	Name         string
	Version      string
	Endpoint     string
	DataPath     string
	FilePath     string
	DBUrl        string
	Secret       string
	Registerable bool
	RegUsers     []string
}

type Logic struct {
	srvless      bool
	registerable bool
	db           model.DB
	secKey       *crypto.SecretKey
	Name         string
	NodeID       string
	Version      string
	Endpoint     string
	Features     []string

	whitelist   map[string]bool
	filepath    string
	apnsPClient *apns2.Client
	apnsDClient *apns2.Client
}

type APNSPusher interface {
	Push(n *apns2.Notification) (*apns2.Response, error)
}

const authKey = "MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQgQ6vCLkUeDj223nfPfKGrjG+Coc53EbKHmO6Oa9YcHiGgCgYIKoZIzj0DAQehRANCAAQNwg3W2eOqNlX0nl9kGbfmMxwSZoO4RmqKoKJnH/vGkU8csJuN5Dg4JiI6ni5PEx+A1rb19DuDm4AzwBVvl8Jt"

func (opts *Options) fixOptions() {
	if len(opts.DataPath) > 0 {
		s, err := os.Stat(opts.DataPath)
		if err == nil && s.IsDir() {
			if len(opts.Secret) <= 0 && len(opts.DBUrl) <= 0 {
				opts.DBUrl = "sqlite://" + filepath.Join(opts.DataPath, "chanify.db")
			}
			if len(opts.FilePath) <= 0 {
				opts.FilePath = filepath.Join(opts.DataPath, "files")
			}
		}
	}
}

func NewLogic(opts *Options) (*Logic, error) {
	opts.fixOptions()
	l := &Logic{
		srvless:      false,
		registerable: opts.Registerable,
		Name:         opts.Name,
		Version:      opts.Version,
		Endpoint:     opts.Endpoint,
		Features:     []string{"msg.text"},
	}
	if l.registerable {
		log.Println("Register user enabled")
	} else {
		log.Println("Register user disabled")
		whitelist := map[string]bool{}
		for _, u := range opts.RegUsers {
			whitelist[u] = true
		}
		l.whitelist = whitelist
		log.Println("Find", len(whitelist), "user(s) in whitelist")
		l.Features = append([]string{"register.limit"}, l.Features...)
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
		l.filepath = opts.FilePath
		l.apnsPClient = apns2.NewTokenClient(tk).Production()
		l.apnsDClient = apns2.NewTokenClient(tk).Development()
		if len(l.filepath) > 0 {
			l.Features = append(l.Features, "msg.image")
			FixPath(filepath.Join(l.filepath, "images")) // nolint: errcheck
			log.Println("Files path:", l.filepath)
		}
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

func (l *Logic) CanFileStore() bool {
	return len(l.filepath) > 0
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

func (l *Logic) GetDeviceKey(uuid string) ([]byte, error) {
	return l.db.GetDeviceKey(uuid)
}

func (l *Logic) GetDevices(uid string) ([]*model.Device, error) {
	return l.db.GetDevices(uid)
}

func (l *Logic) VerifyToken(tk *model.Token) bool {
	if tk.IsExpires() {
		return false
	}
	key, err := l.GetUserKey(tk.GetUserID())
	if err != nil {
		return false
	}
	return tk.VerifySign(key)
}

func (l *Logic) LoadImageFile(name string) ([]byte, error) {
	if len(l.filepath) <= 0 {
		return nil, ErrNoSupportMethod
	}
	return LoadFile(filepath.Join(l.filepath, "images", name))
}

func (l *Logic) SaveImageFile(data []byte) (string, error) {
	if len(l.filepath) <= 0 {
		return "", ErrNoSupportMethod
	}
	if len(data) <= 0 {
		return "", ErrInvalidContent
	}
	key := sha1.Sum(data)
	name := hex.EncodeToString(key[:])
	path := filepath.Join(l.filepath, "images", name)
	if err := SaveFile(path, data); err != nil {
		return "", err
	}
	return filepath.Join("/files/images", name), nil
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

func (l *Logic) SendAPNS(uid string, msg *model.Message, data []byte, devices []*model.Device, priority int) int {
	notification := &apns2.Notification{
		Topic:      "net.chanify.ios",
		Expiration: time.Now().Add(24 * time.Hour),
		Payload: payload.NewPayload().MutableContent().AlertLocKey("NewMsg").
			Custom("uid", uid).
			Custom("src", l.NodeID).
			Custom("msg", base64.RawURLEncoding.EncodeToString(data)),
	}
	if priority == 5 { // only 10 or 5
		notification.Priority = priority
	}
	n := len(devices)
	for _, dev := range devices {
		notification.DeviceToken = hex.EncodeToString(dev.Token)
		res, err := l.GetAPNS(dev.Sandbox).Push(notification)
		if err != nil {
			log.Println("Send apns failed:", res.StatusCode, res.Reason)
			n -= 1
		}
	}
	return n
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
	if !l.canRegisterUser(uid) {
		return nil, ErrSystemLimited
	}
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

func (l *Logic) canRegisterUser(uid string) bool {
	if l.registerable {
		return true
	}
	_, ok := l.whitelist[uid]
	return ok
}
