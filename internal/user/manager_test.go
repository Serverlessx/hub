package user

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/artifacthub/hub/internal/email"
	"github.com/artifacthub/hub/internal/hub"
	"github.com/artifacthub/hub/internal/tests"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestCheckAvailability(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg       string
			resourceKind string
			value        string
		}{
			{
				"invalid resource kind",
				"invalid",
				"value",
			},
			{
				"invalid value",
				"userAlias",
				"",
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				m := NewManager(nil, nil)
				_, err := m.CheckAvailability(context.Background(), tc.resourceKind, tc.value)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("database query succeeded", func(t *testing.T) {
		testCases := []struct {
			resourceKind string
			dbQuery      string
			available    bool
		}{
			{
				"userAlias",
				`select user_id from "user" where alias = $1`,
				true,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("resource kind: %s", tc.resourceKind), func(t *testing.T) {
				tc.dbQuery = fmt.Sprintf("select not exists (%s)", tc.dbQuery)
				db := &tests.DBMock{}
				db.On("QueryRow", tc.dbQuery, "value").Return(tc.available, nil)
				m := NewManager(db, nil)

				available, err := m.CheckAvailability(context.Background(), tc.resourceKind, "value")
				assert.NoError(t, err)
				assert.Equal(t, tc.available, available)
				db.AssertExpectations(t)
			})
		}
	})

	t.Run("database error", func(t *testing.T) {
		db := &tests.DBMock{}
		dbQuery := `select not exists (select user_id from "user" where alias = $1)`
		db.On("QueryRow", dbQuery, "value").Return(false, tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		available, err := m.CheckAvailability(context.Background(), "userAlias", "value")
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		assert.False(t, available)
		db.AssertExpectations(t)
	})
}

func TestCheckCredentials(t *testing.T) {
	dbQuery := `select user_id, password from "user" where email = $1 and password is not null`

	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg   string
			email    string
			password string
		}{
			{
				"email not provided",
				"",
				"password",
			},
			{
				"password not provided",
				"email",
				"",
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				m := NewManager(nil, nil)
				_, err := m.CheckCredentials(context.Background(), tc.email, tc.password)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("credentials provided not found in database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return(nil, pgx.ErrNoRows)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass")
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("error getting credentials from database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return(nil, tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass")
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		assert.Nil(t, output)
		db.AssertExpectations(t)
	})

	t.Run("invalid credentials provided", func(t *testing.T) {
		pw, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return([]interface{}{"userID", string(pw)}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass2")
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("valid credentials provided", func(t *testing.T) {
		pw, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return([]interface{}{"userID", string(pw)}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass")
		assert.NoError(t, err)
		assert.True(t, output.Valid)
		assert.Equal(t, "userID", output.UserID)
		db.AssertExpectations(t)
	})
}

func TestCheckSession(t *testing.T) {
	dbQuery := `
	select user_id, floor(extract(epoch from created_at))
	from session where session_id = $1
	`

	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg    string
			sessionID []byte
			duration  time.Duration
		}{
			{
				"session id not provided",
				nil,
				10,
			},
			{
				"session id not provided",
				[]byte(""),
				10,
			},
			{
				"duration not provided",
				[]byte("sessionID"),
				0,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				m := NewManager(nil, nil)
				_, err := m.CheckSession(context.Background(), tc.sessionID, tc.duration)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("session not found in database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return(nil, pgx.ErrNoRows)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("error getting session from database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return(nil, tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		assert.Nil(t, output)
		db.AssertExpectations(t)
	})

	t.Run("session has expired", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return([]interface{}{"userID", int64(1)}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("valid session", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return([]interface{}{
			"userID",
			time.Now().Unix(),
		}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.NoError(t, err)
		assert.True(t, output.Valid)
		assert.Equal(t, "userID", output.UserID)
		db.AssertExpectations(t)
	})
}

func TestDeleteSession(t *testing.T) {
	dbQuery := "delete from session where session_id = $1"

	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg    string
			sessionID []byte
			duration  time.Duration
		}{
			{
				"session id not provided",
				nil,
				10,
			},
			{
				"session id not provided",
				[]byte(""),
				10,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				m := NewManager(nil, nil)
				err := m.DeleteSession(context.Background(), tc.sessionID)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("valid input", func(t *testing.T) {
		testCases := []struct {
			description string
			dbResponse  interface{}
		}{
			{
				"session deleted successfully",
				nil,
			},
			{
				"error deleting session from database",
				tests.ErrFakeDatabaseFailure,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				db := &tests.DBMock{}
				db.On("Exec", dbQuery, []byte("sessionID")).Return(tc.dbResponse)
				m := NewManager(db, nil)

				err := m.DeleteSession(context.Background(), []byte("sessionID"))
				assert.Equal(t, tc.dbResponse, err)
				db.AssertExpectations(t)
			})
		}
	})
}

func TestGetProfileJSON(t *testing.T) {
	dbQuery := "select get_user_profile($1::uuid)"
	ctx := context.WithValue(context.Background(), hub.UserIDKey, "userID")

	t.Run("user id not found in ctx", func(t *testing.T) {
		m := NewManager(nil, nil)
		assert.Panics(t, func() {
			_, _ = m.GetProfileJSON(context.Background())
		})
	})

	t.Run("database query succeeded", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "userID").Return([]byte("dataJSON"), nil)
		m := NewManager(db, nil)

		data, err := m.GetProfileJSON(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []byte("dataJSON"), data)
		db.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "userID").Return(nil, tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		data, err := m.GetProfileJSON(ctx)
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		assert.Nil(t, data)
		db.AssertExpectations(t)
	})
}

func TestGetUserID(t *testing.T) {
	dbQuery := `select user_id from "user" where email = $1`

	t.Run("invalid input", func(t *testing.T) {
		m := NewManager(nil, nil)
		_, err := m.GetUserID(context.Background(), "")
		assert.True(t, errors.Is(err, ErrInvalidInput))
	})

	t.Run("database query succeeded", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return("userID", nil)
		m := NewManager(db, nil)

		userID, err := m.GetUserID(context.Background(), "email")
		assert.NoError(t, err)
		assert.Equal(t, "userID", userID)
		db.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return("", tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		userID, err := m.GetUserID(context.Background(), "email")
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		assert.Empty(t, userID)
		db.AssertExpectations(t)
	})
}

func TestRegisterSession(t *testing.T) {
	dbQuery := "select register_session($1::jsonb)"

	s := &hub.Session{
		UserID:    "00000000-0000-0000-0000-000000000001",
		IP:        "192.168.1.100",
		UserAgent: "Safari 13.0.5",
	}

	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg string
			userID string
		}{
			{
				"user id not provided",
				"",
			},
			{
				"invalid user id",
				"invalid",
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				m := NewManager(nil, nil)
				s := &hub.Session{UserID: tc.userID}
				_, err := m.RegisterSession(context.Background(), s)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("successful session registration", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, mock.Anything).Return([]byte("sessionID"), nil)
		m := NewManager(db, nil)

		sessionID, err := m.RegisterSession(context.Background(), s)
		assert.NoError(t, err)
		assert.Equal(t, []byte("sessionID"), sessionID)
		db.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, mock.Anything).Return(nil, tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		sessionID, err := m.RegisterSession(context.Background(), s)
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		assert.Nil(t, sessionID)
		db.AssertExpectations(t)
	})
}

func TestRegisterUser(t *testing.T) {
	dbQuery := "select register_user($1::jsonb)"

	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg  string
			user    *hub.User
			baseURL string
		}{
			{
				"alias not provided",
				&hub.User{},
				"http://baseurl.com",
			},
			{
				"email not provided",
				&hub.User{Alias: "user1"},
				"http://baseurl.com",
			},
			{
				"invalid base url",
				&hub.User{Alias: "user1", Email: "email"},
				"invalid",
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				es := &email.SenderMock{}
				es.On("SendEmail", mock.Anything).Return(nil)
				m := NewManager(nil, es)

				err := m.RegisterUser(context.Background(), tc.user, tc.baseURL)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("successful user registration in database", func(t *testing.T) {
		testCases := []struct {
			description         string
			emailSenderResponse error
		}{
			{
				"email verification code sent successfully",
				nil,
			},
			{
				"error sending email verification code",
				email.ErrFakeSenderFailure,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				db := &tests.DBMock{}
				db.On("QueryRow", dbQuery, mock.Anything).Return("emailVerificationCode", nil)
				es := &email.SenderMock{}
				es.On("SendEmail", mock.Anything).Return(tc.emailSenderResponse)
				m := NewManager(db, es)

				u := &hub.User{
					Alias:     "alias",
					FirstName: "first_name",
					LastName:  "last_name",
					Email:     "email@email.com",
					Password:  "password",
				}
				err := m.RegisterUser(context.Background(), u, "http://baseurl.com")
				assert.Equal(t, tc.emailSenderResponse, err)
				assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("password")))
				db.AssertExpectations(t)
				es.AssertExpectations(t)
			})
		}
	})

	t.Run("database error registering user", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, mock.Anything).Return("", tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		u := &hub.User{
			Alias: "alias",
			Email: "email@email.com",
		}
		err := m.RegisterUser(context.Background(), u, "http://baseurl.com")
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		db.AssertExpectations(t)
	})
}

func TestUpdatePassword(t *testing.T) {
	getPasswordDBQuery := `select password from "user" where user_id = $1 and password is not null`
	updatePasswordDBQuery := "select update_user_password($1::uuid, $2::text, $3::text)"
	ctx := context.WithValue(context.Background(), hub.UserIDKey, "userID")
	oldHashed, _ := bcrypt.GenerateFromPassword([]byte("old"), bcrypt.DefaultCost)

	t.Run("user id not found in ctx", func(t *testing.T) {
		m := NewManager(nil, nil)
		assert.Panics(t, func() {
			_ = m.UpdatePassword(context.Background(), "old", "new")
		})
	})

	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg string
			old    string
			new    string
		}{
			{
				"old password not provided",
				"",
				"new",
			},
			{
				"new password not provided",
				"old",
				"",
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				m := NewManager(nil, nil)
				err := m.UpdatePassword(ctx, tc.old, tc.new)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("database error getting user password", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", getPasswordDBQuery, "userID").Return("", tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		err := m.UpdatePassword(ctx, "old", "new")
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		db.AssertExpectations(t)
	})

	t.Run("invalid user password provided", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", getPasswordDBQuery, "userID").Return(string(oldHashed), nil)
		m := NewManager(db, nil)

		err := m.UpdatePassword(ctx, "old2", "new")
		assert.Error(t, err)
		db.AssertExpectations(t)
	})

	t.Run("database error updating password", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", getPasswordDBQuery, "userID").Return(string(oldHashed), nil)
		db.On("Exec", updatePasswordDBQuery, "userID", mock.Anything, mock.Anything).
			Return(tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		err := m.UpdatePassword(ctx, "old", "new")
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		db.AssertExpectations(t)
	})

	t.Run("successful password update", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", getPasswordDBQuery, "userID").Return(string(oldHashed), nil)
		db.On("Exec", updatePasswordDBQuery, "userID", mock.Anything, mock.Anything).Return(nil)
		m := NewManager(db, nil)

		err := m.UpdatePassword(ctx, "old", "new")
		assert.NoError(t, err)
		db.AssertExpectations(t)
	})
}

func TestUpdateProfile(t *testing.T) {
	dbQuery := "select update_user_profile($1::uuid, $2::jsonb)"
	ctx := context.WithValue(context.Background(), hub.UserIDKey, "userID")

	t.Run("user id not found in ctx", func(t *testing.T) {
		m := NewManager(nil, nil)
		assert.Panics(t, func() {
			_ = m.UpdateProfile(context.Background(), &hub.User{})
		})
	})

	t.Run("invalid input", func(t *testing.T) {
		testCases := []struct {
			errMsg string
			user   *hub.User
		}{
			{
				"alias not provided",
				&hub.User{},
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.errMsg, func(t *testing.T) {
				m := NewManager(nil, nil)
				err := m.UpdateProfile(ctx, tc.user)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("database query succeeded", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("Exec", dbQuery, "userID", mock.Anything).Return(nil)
		m := NewManager(db, nil)

		err := m.UpdateProfile(ctx, &hub.User{Alias: "user1"})
		assert.NoError(t, err)
		db.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("Exec", dbQuery, "userID", mock.Anything).Return(tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		err := m.UpdateProfile(ctx, &hub.User{Alias: "user1"})
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		db.AssertExpectations(t)
	})
}

func TestVerifyEmail(t *testing.T) {
	dbQuery := "select verify_email($1::uuid)"

	t.Run("invalid input", func(t *testing.T) {
		m := NewManager(nil, nil)
		_, err := m.VerifyEmail(context.Background(), "")
		assert.True(t, errors.Is(err, ErrInvalidInput))
	})

	t.Run("successful email verification", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "emailVerificationCode").Return(true, nil)
		m := NewManager(db, nil)

		verified, err := m.VerifyEmail(context.Background(), "emailVerificationCode")
		assert.NoError(t, err)
		assert.True(t, verified)
		db.AssertExpectations(t)
	})

	t.Run("database error verifying email", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "emailVerificationCode").Return(false, tests.ErrFakeDatabaseFailure)
		m := NewManager(db, nil)

		verified, err := m.VerifyEmail(context.Background(), "emailVerificationCode")
		assert.Equal(t, tests.ErrFakeDatabaseFailure, err)
		assert.False(t, verified)
		db.AssertExpectations(t)
	})
}
