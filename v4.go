package tdclient

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tada-team/tdproto"
	"github.com/tada-team/tdproto/tdapi"
)

func (s Session) Ping() error {
	resp := new(struct {
		tdapi.Resp
		Result string `json:"result"`
	})
	return s.doGet("/api/v4/ping", resp)
}

func (s Session) Me(teamUid string) (tdproto.Contact, error) {
	resp := new(struct {
		tdapi.Resp
		Result tdproto.Team `json:"result"`
	})

	if !tdproto.ValidUid(teamUid) {
		return tdproto.Contact{}, errors.New("invalid team uid")
	}

	if err := s.doGet("/api/v4/teams/"+teamUid, resp); err != nil {
		return tdproto.Contact{}, err
	}

	if !resp.Ok {
		return tdproto.Contact{}, resp.Error
	}

	return resp.Result.Me, nil
}

func (s Session) Contacts(teamUid string) ([]tdproto.Contact, error) {
	resp := new(struct {
		tdapi.Resp
		Result []tdproto.Contact `json:"result"`
	})

	if !tdproto.ValidUid(teamUid) {
		return resp.Result, errors.New("invalid team uid")
	}

	if err := s.doGet("/api/v4/teams/"+teamUid+"/contacts/", resp); err != nil {
		return resp.Result, err
	}

	if !resp.Ok {
		return resp.Result, resp.Error
	}

	return resp.Result, nil
}

func (s Session) AddContact(teamUid string, phone string) (tdproto.Contact, error) {
	req := map[string]interface{}{
		"phone": phone,
	}

	resp := new(struct {
		tdapi.Resp
		Result tdproto.Contact `json:"result"`
	})

	if err := s.doPost(fmt.Sprintf("/api/v4/teams/%s/contacts", teamUid), req, resp); err != nil {
		return resp.Result, err
	}

	if !resp.Ok {
		return resp.Result, resp.Error
	}

	return resp.Result, nil
}

func (s Session) AuthBySmsSendCode(phone string) (tdapi.SmsCode, error) {
	req := map[string]interface{}{
		"phone": phone,
	}

	resp := new(struct {
		tdapi.Resp
		Result tdapi.SmsCode `json:"result"`
	})

	if err := s.doPost("/api/v4/auth/sms/send-code", req, resp); err != nil {
		return resp.Result, err
	}

	if !resp.Ok {
		return resp.Result, resp.Error
	}

	return resp.Result, nil
}

func (s Session) AuthBySmsGetToken(phone, code string) (tdapi.Auth, error) {
	req := map[string]interface{}{
		"phone": phone,
		"code":  code,
	}

	resp := new(struct {
		tdapi.Resp
		Result tdapi.Auth `json:"result"`
	})

	if err := s.doPost("/api/v4/auth/sms/get-token", req, resp); err != nil {
		return resp.Result, err
	}

	if !resp.Ok {
		return resp.Result, resp.Error
	}

	return resp.Result, nil
}

func (s Session) AuthByPasswordGetToken(login, password string) (tdapi.Auth, error) {
	req := map[string]interface{}{
		"login":    login,
		"password": password,
	}

	resp := new(struct {
		tdapi.Resp
		Result tdapi.Auth `json:"result"`
	})

	if err := s.doPost("/api/v4/auth/password/get-token", req, resp); err != nil {
		return resp.Result, err
	}

	if !resp.Ok {
		return resp.Result, resp.Error
	}

	return resp.Result, nil
}