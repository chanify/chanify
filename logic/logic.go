package logic

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io/ioutil"
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

// variable define
var (
	APIEndpoint            = "https://api.chanify.net"
	MockPusher  APNSPusher = nil

	randReader = rand.Read

	ErrNoSupportMethod = errors.New("no support method")
	ErrNotFound        = errors.New("not found")
	ErrInvalidContent  = errors.New("invalid content")
	ErrSystemLimited   = errors.New("system limited")
)

// Options for init logic
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

// Logic instance
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

	infoData    []byte
	infoSign    string
	whitelist   map[string]bool
	filepath    string
	apnsPClient *apns2.Client
	apnsDClient *apns2.Client
}

// APNSPusher is the interface of APNS2
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

// NewLogic with options
func NewLogic(opts *Options) (*Logic, error) {
	opts.fixOptions()
	l := &Logic{
		srvless:      false,
		registerable: opts.Registerable,
		Name:         opts.Name,
		Version:      opts.Version,
		Endpoint:     opts.Endpoint,
		Features:     []string{"platform.watchos", "msg.text", "msg.link"},
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
			return nil, errors.New("in serverless mode, secret is required")
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
			l.Features = append(l.Features, "msg.image", "msg.file")
			fixPath(filepath.Join(l.filepath, "images")) // nolint: errcheck
			fixPath(filepath.Join(l.filepath, "files"))  // nolint: errcheck
			log.Println("Files path:", l.filepath)
		}
	}
	l.InitInfo()
	log.Printf("Node server name: %s, version: %s, serverless: %v, node-id: %s\n", l.Name, l.Version, l.srvless, l.NodeID)
	return l, nil
}

// Close and cleanup logic instance
func (l *Logic) Close() {
	if l.db != nil {
		l.db.Close()
		l.db = nil
	}
}

// CanFileStore return file stroage is available
func (l *Logic) CanFileStore() bool {
	return len(l.filepath) > 0
}

// GetUser find user info with user id
func (l *Logic) GetUser(uid string) (*model.User, error) {
	return l.db.GetUser(uid)
}

// GetUserKey find user key with user id
func (l *Logic) GetUserKey(uid string) ([]byte, error) {
	u, err := l.db.GetUser(uid)
	if err != nil {
		return nil, err
	}
	return u.SecretKey, nil
}

// UpsertUser insert or update user info
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

// BindDevice to user
func (l *Logic) BindDevice(uid string, uuid string, key string, devType int) error {
	pk, err := model.CalcDeviceKey(uuid, key)
	if err != nil {
		return err
	}
	return l.db.BindDevice(uid, uuid, pk.MarshalPublicKey(), devType)
}

// UnbindDevice from user
func (l *Logic) UnbindDevice(uid string, uuid string) error {
	return l.db.UnbindDevice(uid, uuid)
}

// UpdatePushToken for APNS
func (l *Logic) UpdatePushToken(uid string, uuid string, token string, sandbox bool) error {
	tk, err := model.DecodePushToken(token)
	if err != nil {
		return err
	}
	return l.db.UpdatePushToken(uid, uuid, tk, sandbox)
}

// GetDeviceKey return device key with device uuid
func (l *Logic) GetDeviceKey(uuid string) ([]byte, error) {
	return l.db.GetDeviceKey(uuid)
}

// GetDevices return all devices with user id
func (l *Logic) GetDevices(uid string) ([]*model.Device, error) {
	return l.db.GetDevices(uid)
}

// Decrypt data with node secret key
func (l *Logic) Decrypt(data []byte) ([]byte, error) {
	return l.secKey.Decrypt(data)
}

// VerifyToken chekc sender token
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

// LoadFile read with file type & data
func (l *Logic) LoadFile(tname string, name string) ([]byte, error) {
	if len(l.filepath) <= 0 {
		return nil, ErrNoSupportMethod
	}
	fh, err := hex.DecodeString(name)
	if err != nil || len(fh) <= 0 {
		return nil, ErrNotFound
	}
	path := filepath.Join(l.filepath, tname, hex.EncodeToString(fh))
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// SaveFile save with file type & data
func (l *Logic) SaveFile(tname string, data []byte) (string, error) {
	if len(l.filepath) <= 0 {
		return "", ErrNoSupportMethod
	}
	if len(data) <= 0 {
		return "", ErrInvalidContent
	}
	key := sha1.Sum(data)
	name := hex.EncodeToString(key[:])
	path := filepath.Join(l.filepath, tname, name)
	if err := saveFile(path, data); err != nil {
		return "", err
	}
	return filepath.Join("/files/"+tname, name), nil
}

// SendAPNS send message to APNS
func (l *Logic) SendAPNS(uid string, data []byte, devices []*model.Device, priority int) int {
	notification := &apns2.Notification{
		Topic:      "net.chanify.ios",
		Expiration: time.Now().Add(24 * time.Hour),
		Payload: payload.NewPayload().MutableContent().AlertLocKey("NewMsg").
			Custom("uid", uid).
			Custom("src", l.NodeID).
			Custom("msg", crypto.Base64Encode.EncodeToString(data)),
	}
	if priority == 5 { // only 10 or 5
		notification.Priority = priority
	}
	n := len(devices)
	for _, dev := range devices {
		notification.DeviceToken = hex.EncodeToString(dev.Token)
		res, err := l.getAPNS(dev.Sandbox).Push(notification)
		if err != nil {
			log.Println("Send apns failed:", res.StatusCode, res.Reason)
			n--
		}
	}
	return n
}

func (l *Logic) getAPNS(sandbox bool) APNSPusher {
	if MockPusher != nil {
		return MockPusher
	}
	if sandbox {
		return l.apnsDClient
	}
	return l.apnsPClient
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
		UID:       uid,
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
