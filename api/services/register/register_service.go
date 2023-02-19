package register_service

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net/mail"
	"odo24_mobile_backend/db"
	"odo24_mobile_backend/sendmail"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type RegisterService struct {
	rnd  *rand.Rand
	memc *memcache.Client
}

func NewRegisterService() *RegisterService {
	src := rand.NewSource(time.Now().UnixNano())
	return &RegisterService{
		rnd:  rand.New(src),
		memc: newMemcachedClient(),
	}
}

func (srv *RegisterService) SendEmailCodeConfirmation(email *mail.Address) error {
	code := srv.generateConfirmationCode()
	rawCode := make([]byte, 2)
	binary.LittleEndian.PutUint16(rawCode, code)

	cacheKey := strings.Replace(email.Address, "@", ".", -1)

	item := memcache.Item{
		Key:        cacheKey,
		Value:      []byte(rawCode),
		Expiration: 1200,
	}

	for i := 0; i < 2; i++ {
		err := srv.memc.Add(&item)
		if err != nil {
			if errors.Is(err, memcache.ErrNoServers) || errors.Is(err, memcache.ErrServerError) {
				srv.memc = newMemcachedClient()
				continue
			}
			return err
		}
		break
	}

	data := make(map[string]interface{})
	data["code"] = code

	err := sendmail.SendEmail(email.Address, sendmail.TypeConfirmEmail, data)
	return err
}

func (srv *RegisterService) RegisterByEmail(email *mail.Address, code uint16, password string) error {
	cacheKey := strings.Replace(email.Address, "@", ".", -1)

	item, err := srv.memc.Get(cacheKey)
	if err != nil {
		return err
	}

	confirmCode := binary.LittleEndian.Uint16(item.Value)

	if code != confirmCode {
		return errors.New("the code does not match")
	}

	pg := db.Conn()

	var user struct {
		UserID uint64
		Login  string
	}
	err = pg.QueryRow("select * from profiles.register_by_email($1, $2);", email.Address, password).Scan(&user.UserID, &user.Login)
	if err != nil {
		return err
	}

	srv.memc.Delete(cacheKey)
	return nil
}

func (ctrl *RegisterService) generateConfirmationCode() uint16 {
	var value = ctrl.rnd.Uint32()
	value %= (9999 - 1000)
	value += 1000
	return uint16(value)
}

func newMemcachedClient() *memcache.Client {
	return memcache.New("localhost:11211")
}
