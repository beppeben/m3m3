package web

import (
	"errors"
	"fmt"
	"github.com/beppeben/m3m3/utils"
	"time"
)

type User struct {
	Name  string
	Pass  string
	Email string
}

func (h WebserviceHandler) ILogin(nameOrMail, pass string) (string, error, string) {
	if nameOrMail == "" || pass == "" {
		return "", errors.New("Error: empty fields"), "ERROR_FORMAT"
	}
	var name, storedPass string
	if utils.ValidateEmail(nameOrMail) {
		//we are given an email address
		user, err := h.repo.GetUserByEmail(nameOrMail)
		if err != nil {
			return "", err, "ERROR_EMAIL_NOT_EXISTING"
		}
		name = user.Name
		storedPass = user.Pass
	} else {
		//we are given a name
		user, err := h.repo.GetUserByName(nameOrMail)
		if err != nil {
			return "", err, "ERROR_NAME_NOT_EXISTING"
		}
		name = nameOrMail
		storedPass = user.Pass
	}
	if utils.Hash(pass) != storedPass {
		err := fmt.Errorf("Error: wrong password for %s", name)
		return "", err, "ERROR_WRONG_PASS"
	}
	return name, nil, ""
}

func (h WebserviceHandler) ICreateSession(name string) (string, error, string) {
	token := utils.RandString(32)
	expire := time.Now().AddDate(0, 2, 0)
	err := h.repo.InsertAccessToken(token, name, expire)
	if err != nil {
		return "", err, "ERROR_DB"
	}
	go func() {
		time.Sleep(time.Hour * 24 * 30 * 2)
		h.repo.DeleteAccessToken(token)
	}()
	return token, nil, ""
}

func (h WebserviceHandler) ILogout(token string) (error, string) {
	err := h.repo.DeleteAccessToken(token)
	if err != nil {
		return err, "DB_ERROR"
	}
	return nil, ""
}

func (h WebserviceHandler) IConfirmEmail(token string) (string, error, string) {
	user, err := h.repo.InsertUserFromTempToken(token)
	if err != nil {
		return "", err, "DB_ERROR"
	}
	return user.Name, nil, ""
}

func (h WebserviceHandler) IRegister(u *User) (string, error, string) {
	var err error
	if u.Name == "" || u.Pass == "" || u.Email == "" {
		return "", errors.New("Error: Empty fields"), "ERROR_FORMAT"
	}
	if !utils.ValidateEmail(u.Email) {
		err = fmt.Errorf("Error: email % is invalid", u.Email)
		return "", err, "ERROR_EMAIL_INVALID"
	}
	_, err = h.repo.GetUserByName(u.Name)
	if err == nil {
		err = fmt.Errorf("Error: username %s used already", u.Name)
		return "", err, "ERROR_NAME_USED"
	}
	_, err = h.repo.GetUserByEmail(u.Email)
	if err == nil {
		err = fmt.Errorf("Error: email %s used already", u.Email)
		return "", err, "ERROR_EMAIL_USED"
	}

	u.Pass = utils.Hash(u.Pass)
	token := utils.RandString(32)
	expire := time.Now().AddDate(0, 0, 1)
	err = h.repo.InsertTempToken(u, token, expire)

	if err != nil {
		return "", err, "ERROR_DB"
	}

	go func() {
		time.Sleep(time.Hour * 24)
		h.repo.DeleteTempToken(token)
	}()

	return token, nil, ""
}
