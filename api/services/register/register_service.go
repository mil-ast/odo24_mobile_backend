package register_service

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"log"
	"math/big"
	"net/mail"

	"github.com/mil-ast/odo24_mobile_backend/api/services"
	"github.com/mil-ast/odo24_mobile_backend/api/utils"
	"github.com/mil-ast/odo24_mobile_backend/db"
	"github.com/mil-ast/odo24_mobile_backend/sendmail"
)

var (
	ErrLoginAlreadyExists     = errors.New("errLoginAlreadyExists")
	ErrCodeHasAlreadyBeenSent = errors.New("code has already been sent")
	ErrCodeDoesNotMatch       = errors.New("code does not match")
	ErrCodeTooManyAttempts    = errors.New("code too many attempts")
)

type RegisterService struct{}

func NewRegisterService() *RegisterService {
	return &RegisterService{}
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

	code, err := srv.generateConfirmationCode()
	if err != nil {
		return err
	}

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

	code, err := srv.generateConfirmationCode()
	if err != nil {
		return err
	}

	err = services.AddEmailCodeConfirmation(email, code)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["code"] = code

	err = sendmail.SendEmail(email.Address, sendmail.TypeRepairConfirmCode, data)
	return err
}

func (srv *RegisterService) checkAttempts(email *mail.Address) error {
	attempts, err := services.GetConfirmCodeAttempts(email)
	if err != nil {
		return err
	}
	if attempts >= services.MaxConfirmCodeAttempts {
		services.DeleteEmailCodeConfirmation(email)
		return ErrCodeTooManyAttempts
	}
	return nil
}

func (srv *RegisterService) RegisterByEmail(email *mail.Address, code uint32, password string) error {
	item, err := services.GetEmailCodeConfirmation(email)
	if err != nil {
		return err
	}

	if item == nil {
		log.Println("RegisterByEmail GetEmailCodeConfirmation code is empty")
		return ErrCodeDoesNotMatch
	}

	if err := srv.checkAttempts(email); err != nil {
		return err
	}

	confirmCode := binary.LittleEndian.Uint32(item.Value)

	if code != confirmCode {
		attempts, err := services.IncrementConfirmCodeAttempts(email)
		if err != nil {
			return err
		}
		if attempts >= services.MaxConfirmCodeAttempts {
			services.DeleteEmailCodeConfirmation(email)
		}
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

	var userID uint64
	err = pg.QueryRow(`INSERT INTO profiles.users (login,password_hash,oauth,last_login_dt,salt) VALUES($1,$2,$3,now()::timestamp without time zone,$4) RETURNING user_id`, email.Address, newPassword, false, salt).Scan(&userID)
	if err != nil {
		return err
	}

	services.DeleteEmailCodeConfirmation(email)
	return nil
}

func (srv *RegisterService) PasswordRecovery(email *mail.Address, code uint32, password string) error {
	item, err := services.GetEmailCodeConfirmation(email)
	if err != nil {
		return err
	}

	if item == nil {
		log.Println("PasswordRecovery GetEmailCodeConfirmation code is empty")
		return ErrCodeDoesNotMatch
	}

	if err := srv.checkAttempts(email); err != nil {
		return err
	}

	confirmCode := binary.LittleEndian.Uint32(item.Value)

	if code != confirmCode {
		attempts, err := services.IncrementConfirmCodeAttempts(email)
		if err != nil {
			return err
		}
		if attempts >= services.MaxConfirmCodeAttempts {
			services.DeleteEmailCodeConfirmation(email)
		}
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

func (srv *RegisterService) generateConfirmationCode() (uint32, error) {
	// 6-значный код в диапазоне [100000, 999999]
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return 0, err
	}
	return uint32(n.Int64()) + 100000, nil
}