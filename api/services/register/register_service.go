package register_service

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"log"
	"math/rand"
	"net/mail"
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/db"
	"odo24_mobile_backend/sendmail"
	"time"
)

var (
	ErrLoginAlreadyExists     = errors.New("errLoginAlreadyExists")
	ErrCodeHasAlreadyBeenSent = errors.New("code has already been sent")
	ErrCodeDoesNotMatch       = errors.New("code does not match")
)

type RegisterService struct {
	rnd          *rand.Rand
	passwordSalt []byte
}

func NewRegisterService(passwordSalt string) *RegisterService {
	src := rand.NewSource(time.Now().UnixNano())
	return &RegisterService{
		rnd:          rand.New(src),
		passwordSalt: []byte(passwordSalt),
	}
}

func (srv *RegisterService) SendEmailCodeConfirmation(email *mail.Address) error {
	existsCode, err := services.GetEmailCodeConfirmation(email)
	if err != nil {
		return err
	}
	if existsCode != nil {
		return nil
	}

	code := srv.generateConfirmationCode()

	err = services.AddEmailCodeConfirmation(email, code)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["code"] = code

	err = sendmail.SendEmail(email.Address, sendmail.TypeConfirmEmail, data)
	return err
}

func (srv *RegisterService) PasswordRecoverySendEmailCodeConfirmation(email *mail.Address) error {
	existsCode, err := services.GetEmailCodeConfirmation(email)
	if err != nil {
		return err
	}
	if existsCode != nil {
		return ErrCodeHasAlreadyBeenSent
	}

	code := srv.generateConfirmationCode()

	err = services.AddEmailCodeConfirmation(email, code)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["code"] = code

	err = sendmail.SendEmail(email.Address, sendmail.TypeRepairConfirmCode, data)
	return err
}

func (srv *RegisterService) RegisterByEmail(email *mail.Address, code uint16, password string) error {
	item, err := services.GetEmailCodeConfirmation(email)
	if err != nil {
		return err
	}

	if item == nil {
		log.Println("RegisterByEmail GetEmailCodeConfirmation code is empty")
		return ErrCodeDoesNotMatch
	}

	confirmCode := binary.LittleEndian.Uint16(item.Value)

	if code != confirmCode {
		return ErrCodeDoesNotMatch
	}

	hasherNewPassword := sha1.New()
	_, err = hasherNewPassword.Write([]byte(password))
	if err != nil {
		return err
	}

	pg := db.Conn()

	var emailIsExists bool
	err = pg.QueryRow("SELECT EXISTS(SELECT 1 from profiles.users u WHERE u.login=$1)", email.Address).Scan(&emailIsExists)
	if err != nil {
		return err
	}
	if emailIsExists {
		return ErrLoginAlreadyExists
	}

	sumNewPasswd := hasherNewPassword.Sum(srv.passwordSalt)
	var userID int64
	err = pg.QueryRow(`INSERT INTO profiles.users (login,password_hash,oauth,last_login_dt) VALUES($1,$2,$3,now()::timestamp without time zone) RETURNING user_id`, email.Address, sumNewPasswd, false).Scan(&userID)
	if err != nil {
		return err
	}

	services.DeleteEmailCodeConfirmation(email)
	return nil
}

func (srv *RegisterService) PasswordRecovery(email *mail.Address, code uint16, password string) error {
	item, err := services.GetEmailCodeConfirmation(email)
	if err != nil {
		return err
	}

	if item == nil {
		log.Println("PasswordRecovery GetEmailCodeConfirmation code is empty")
		return ErrCodeDoesNotMatch
	}

	confirmCode := binary.LittleEndian.Uint16(item.Value)

	if code != confirmCode {
		return ErrCodeDoesNotMatch
	}

	hasherNewPassword := sha1.New()
	_, err = hasherNewPassword.Write([]byte(password))
	if err != nil {
		return err
	}
	sumNewPasswd := hasherNewPassword.Sum(srv.passwordSalt)

	pg := db.Conn()
	_, err = pg.Exec("UPDATE profiles.users SET password_hash=$1 WHERE login=$2;", sumNewPasswd, email.Address)
	if err != nil {
		return err
	}

	services.DeleteEmailCodeConfirmation(email)
	return nil
}

func (ctrl *RegisterService) generateConfirmationCode() uint16 {
	var value = ctrl.rnd.Uint32()
	value %= (9999 - 1000)
	value += 1000
	return uint16(value)
}
