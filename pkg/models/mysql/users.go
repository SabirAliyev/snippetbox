package mysql

import(
	"database/sql"
	"errors"
	"strings"

	"sabiraliyev.net/snippetbox/pkg/models"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct{
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `
	INSERT INTO users (name, email, hashed_password, created)
	VALUES($1, $2, $3, now())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var mySQLError *pq.Error
		if errors.As(err, &mySQLError) {
			// mySQLError.Number == 1062 &&
			if strings.Contains(mySQLError.Message, "users_uc_email") {
				return  models.ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error){
	var id int
	var hashedPassword []byte
	stmt := `
	SELECT id, hashed_password 
	FROM users 
	WHERE email = $1 AND active = TRUE`

	row := m.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Get(id int) (*models.User, error) {
	u := &models.User{}

	stmt := `
	SELECT id, name, email, created, active 
	FROM users
	WHERE id = $1`

	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows){
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	return u, nil
}

