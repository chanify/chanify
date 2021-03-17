package logic

import (
	"errors"
	"log"
	"net/url"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type Options struct {
	Name     string
	Version  string
	Endpoint string
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
}

func NewLogic(opts *Options) (*Logic, error) {
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
	if l.srvless {
		log.Println("Running in serverless mode")
	} else {
		l.Features = append([]string{"store.device"}, l.Features...)
		log.Println("Running in serverful mode")
	}
	log.Printf("Node server name: %s, version: %s, node-id: %s\n", l.Name, l.Version, l.NodeID)
	return l, nil
}

func (l *Logic) Close() {
	if l.db != nil {
		l.db.Close()
		l.db = nil
	}
}

func (l *Logic) GetUserKey(uid string) ([]byte, error) {
	u, err := l.db.GetUser(uid)
	if err != nil {
		return nil, err
	}
	return u.SecretKey, nil
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
