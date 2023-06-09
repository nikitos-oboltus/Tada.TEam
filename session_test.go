package tdclient

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/tada-team/kozma"
	"github.com/tada-team/tdproto"
	"github.com/tada-team/tdproto/tdapi"
)

func TestSession(t *testing.T) {
	testServer := mustEnv("TEST_SERVER")
	testAccountPhone := mustEnv("TEST_ACCOUNT_PHONE")
	testAccountCode := mustEnv("TEST_ACCOUNT_CODE")

	s, err := NewSession(testServer)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	t.Run("http ping", func(t *testing.T) {
		if err := s.Ping(); err != nil {
			t.Fatalf("%+v", err)
		}
	})

	t.Run("features smoke test", func(t *testing.T) {
		if _, err := s.Features(); err != nil {
			t.Fatalf("%+v", err)
		}
	})

	var team tdproto.Team

	t.Run("sms login", func(t *testing.T) {
		codeResp, err := s.AuthBySmsSendCode(testAccountPhone)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if codeResp.CodeLength != len(testAccountCode) {
			t.Fatalf("invalid code length: %+v", codeResp)
		}

		tokenResp, err := s.AuthBySmsGetToken(testAccountPhone, testAccountCode)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		//if len(tokenResp.Me.Teams) == 0 {
		//	t.Fatalf("invalid teams number: %d", len(tokenResp.Me.Teams))
		//}

		for _, v := range tokenResp.Me.Teams {
			if v.Me.CanAddToTeam {
				team = v
				break
			}
		}

		s.SetToken(tokenResp.Token)
	})

	if team.Uid == "" {
		team, err = s.createTeam(tdapiTeam{Name: "tdclient test"})
		if err != nil {
			t.Fatalf("%+v", err)
		}
		log.Println("new team created:", team.Uid)
	}

	var newContact tdproto.Contact
	t.Run("contacts list", func(t *testing.T) {
		anyPhone := "+79870000000"
		newContact, err = s.AddContact(team.Uid, anyPhone)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		contacts, err := s.Contacts(team.Uid)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		newContactFound := false
		for _, contact := range contacts {
			if contact.Jid == newContact.Jid {
				newContactFound = true
			}
		}
		if !newContactFound {
			t.Error("new contact not found:", newContact.Jid)
		}
	})

	t.Run("messages", func(t *testing.T) {
		message, err := s.SendPlaintextMessage(team.Uid, newContact.Jid, kozma.Say())
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if message.Chat != newContact.Jid {
			t.Error("invalid send message:", newContact.Jid)
		}

		t.Run("get messages", func(t *testing.T) {
			filter := new(tdapi.MessageFilter)
			filter.Lang = "ru"
			filter.Limit = 200
			messages, err := s.GetMessages(team.Uid, newContact.Jid, filter)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if len(messages) < 1 {
				t.Error("invalid get messages:", len(messages))
			}
		})
		t.Run("send-upload", func(t *testing.T) {
			file, err := ioutil.TempFile(".", "sample.txt")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(file.Name())
			fname := "sample.txt"
			newMsg, err := s.SendUploadMessage(team.Uid, newContact.Jid, fname, file)
			if err != nil {
				t.Fatal(err)
			}
			if len(newMsg.Links) < 1 {
				t.Error("expect to have upload in message")
			}
		})

		t.Run("delete messages", func(t *testing.T) {
			_, err := s.DeleteMessage(team.Uid, newContact.Jid, message.MessageId)
			if err != nil {
				t.Fatalf("%+v", err)
			}
		})

	})

	t.Run("me smoke test", func(t *testing.T) {
		me, err := s.Me(team.Uid)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if !me.CanAddToTeam {
			t.Fatal("cant add to team")
		}
	})

	t.Run("ws", func(t *testing.T) {
		ws, err := s.Ws(team.Uid)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		t.Run("ping", func(t *testing.T) {
			confirmId := ws.Ping()
			ev := new(tdproto.ServerConfirm)
			if err := ws.WaitFor(ev); err != nil {
				t.Fatalf("%+v", err)
			}
			if ev.Params.ConfirmId != confirmId {
				t.Error("confirmId mismatched: got:", ev.ConfirmId, "want:", confirmId)
			}
		})

		t.Run("create message", func(t *testing.T) {
			testText := kozma.Say()
			var messageUid string
			go ws.SendPlainMessage(newContact.Jid, testText)

			err = ws.ForeachMessage(func(messages chan tdproto.Message, errorChan chan error) {
				for {
					select {
					case m := <-messages:
						if m.Content.Text == testText {
							messageUid = m.MessageId
							errorChan <- nil
							return
						}
					case <-time.After(time.Second * 10):
						errorChan <- Timeout
						return
					}
				}
			})

			if err != nil {
				t.Fatal("failed to get the message ", err)
			}

			t.Run("delete message", func(t *testing.T) {
				go ws.DeleteMessage(messageUid)
				err = ws.ForeachMessage(func(messages chan tdproto.Message, errorChan chan error) {
					for {
						select {
						case m := <-messages:
							if m.Content.Text == testText {
								errorChan <- nil
								return
							}
						case <-time.After(time.Second * 10):
							errorChan <- Timeout
							return
						}
					}
				})

				if err != nil {
					t.Fatal("failed to get the message ", err)
				}
			})

			t.Run("close", func(t *testing.T) {
				if err := ws.Close(); err != nil {
					t.Fatalf("%+v", err)
				}
			})
		})
	})

	t.Run("create task", func(t *testing.T) {
		text := kozma.Say()
		task, err := s.CreateTask(team.Uid, tdapi.Task{
			Description: text,
			Tags:        []string{"autotest"},
			Assignee:    newContact.Jid,
			Deadline:    tdproto.IsoDatetime(time.Now().Add(time.Hour)),
			Public:      false,
			RemindAt:    tdproto.IsoDatetime(time.Now().Add(time.Minute)),
		})
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if task.Description != text {
			t.Error("task description mismatched: want:", text, "got:", task.Description)
		}
	})

	t.Run("chats", func(t *testing.T) {
		chats, err := s.GetChats(team.Uid, &tdapi.ChatFilter{
			ChatType: "direct",
			Paginator: tdapi.Paginator{
				Limit: 1,
			},
		})
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(chats) <= 2 {
			t.Error("chats number must be > 2")
		}

		for _, chat := range chats {
			if chat.ChatType != tdproto.DirectChatType {
				t.Error("invalid chat type:", chat.ChatType)
			}
		}
	})

	t.Run("groups", func(t *testing.T) {
		group, err := s.CreateGroup(team.Uid, tdapi.Group{
			DisplayName: "test group",
			Public:      false,
		})
		if err != nil {
			t.Fatalf("%+v", err)
		}

		t.Run("add member", func(t *testing.T) {
			member, err := s.AddGroupMember(team.Uid, group.Jid, newContact.Jid)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if member.Status != tdproto.GroupMember {
				t.Error("invalid status:", member.Status)
			}
		})

		t.Run("get members", func(t *testing.T) {
			members, err := s.GroupMembers(team.Uid, group.Jid)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if len(members) != 2 {
				t.Error("invalid groups number:", len(members))
			}
		})

		t.Run("remove member", func(t *testing.T) {
			if err := s.DropGroupMember(team.Uid, group.Jid, newContact.Jid); err != nil {
				t.Fatalf("%+v", err)
			}
		})

		t.Run("remove group", func(t *testing.T) {
			if err := s.DropGroup(team.Uid, group.Jid); err != nil {
				t.Fatalf("%+v", err)
			}
		})

		t.Run("group list smoke test", func(t *testing.T) {
			_, err := s.GetGroups(team.Uid)
			if err != nil {
				t.Fatalf("%+v", err)
			}
		})
	})
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(key + " variable not set")
	}
	return v
}
