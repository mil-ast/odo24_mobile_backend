package register_service

import (
	"encoding/binary"
	"errors"
	"log"
	"math/rand"
	"net/mail"
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/api/utils"
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
	rnd *rand.Rand
}

func NewRegisterService(passwordSalt string) *RegisterService {
	src := rand.NewSource(time.Now().UnixNano())
	return &RegisterService{
		rnd: rand.New(src),
	}
}

func (srv *RegisterService) SendEmailCodeConfirmation(email *mail.Address) error {
	existsCode, err := services.GetEmailCodeConfirmation(email)
	if err != nil {
		return err
	}
	if existsCode != nil {
		log.Printf("SendEmailCodeConfirmation code is exists")
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

	salt, err := utils.GenerateSalt()
	if err != nil {
		return err
	}

	newPassword, err := utils.GetPasswordHash([]byte(password), salt)
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

	var userID int64
	err = pg.QueryRow(`INSERT INTO profiles.users (login,password_hash,oauth,last_login_dt,salt) VALUES($1,$2,$3,now()::timestamp without time zone,$4) RETURNING user_id`, email.Address, newPassword, false, salt).Scan(&userID)
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

	salt, err := utils.GenerateSalt()
	if err != nil {
		return err
	}

	newPassword, err := utils.GetPasswordHash([]byte(password), salt)
	if err != nil {
		return err
	}

	pg := db.Conn()
	_, err = pg.Exec("UPDATE profiles.users SET password_hash=$1,salt=$2 WHERE login=$3;", newPassword, salt, email.Address)
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
