package register_service

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
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
	existsCode, _ := services.GetEmailCodeConfirmation(email)
	if existsCode != nil {
		return nil
	}

	code := srv.generateConfirmationCode()

	err := services.AddEmailCodeConfirmation(email, code)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["code"] = code

	err = sendmail.SendEmail(email.Address, sendmail.TypeConfirmEmail, data)
	return err
}

func (srv *RegisterService) PasswordRecoverySendEmailCodeConfirmation(email *mail.Address) error {
	existsCode, _ := services.GetEmailCodeConfirmation(email)
	if existsCode != nil {
		return ErrCodeHasAlreadyBeenSent
	}

	code := srv.generateConfirmationCode()

	err := services.AddEmailCodeConfirmation(email, code)
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

	confirmCode := binary.LittleEndian.Uint16(item.Value)

	if code != confirmCode {
		return errors.New("the code does not match")
	}

	hasherNewPassword := sha1.New()
	_, err = hasherNewPassword.Write([]byte(password))
	if err != nil {
		return err
	}
	sumNewPasswd := hasherNewPassword.Sum(srv.passwordSalt)

	pg := db.Conn()

	var user struct {
		UserID uint64
		Login  string
	}
	err = pg.QueryRow("select * from profiles.register_by_email($1,$2);", email.Address, sumNewPasswd).Scan(&user.UserID, &user.Login)
	if err != nil {
		if err.Error() == "pq: login is exists" {
			return ErrLoginAlreadyExists
		}
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

	confirmCode := binary.LittleEndian.Uint16(item.Value)

	if code != confirmCode {
		return errors.New("the code does not match")
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
