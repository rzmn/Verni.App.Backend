package ydbStorage

import (
	"context"
	"fmt"
	"log"
	"path"

	"golang.org/x/crypto/bcrypt"

	"verni/internal/storage"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	yc "github.com/ydb-platform/ydb-go-yc"
)

type Storage struct {
	db  *ydb.Driver
	ctx context.Context
}

func NewUnauthorized(storagePath string) (storage.Storage, error) {
	const op = "storage.ydb.NewUnauthorized"
	ctx := context.Background()
	db, err := ydb.Open(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Printf("%s: connection created", op)
	return new(ctx, db)
}

func New(storagePath string, keyPath string) (storage.Storage, error) {
	const op = "storage.ydb.New"
	ctx := context.Background()
	db, err := ydb.Open(ctx, storagePath, yc.WithServiceAccountKeyFileCredentials(keyPath))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Printf("%s: connection created", op)
	return new(ctx, db)
}

func (s *Storage) CreateTables() error {
	const op = "storage.ydb.CreateTables"
	dbName := s.db.Name()
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			if err := s.CreateTable(ctx, path.Join(dbName, "credentials"),
				options.WithColumn("id", types.TypeText),
				options.WithColumn("email", types.TypeText),
				options.WithColumn("password", types.TypeText),
				options.WithPrimaryKeyColumn("id"),
			); err != nil {
				log.Printf("%s: create credentials: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "userInfos"),
				options.WithColumn("id", types.TypeText),
				options.WithColumn("displayName", types.TypeText),
				options.WithColumn("avatarId", types.Optional(types.TypeText)),
				options.WithPrimaryKeyColumn("id"),
			); err != nil {
				log.Printf("%s: create users: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "emails"),
				options.WithColumn("email", types.TypeText),
				options.WithColumn("confirmed", types.TypeBool),
				options.WithColumn("validationToken", types.Optional(types.TypeText)),
				options.WithPrimaryKeyColumn("email"),
			); err != nil {
				log.Printf("%s: create emails: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "avatars"),
				options.WithColumn("id", types.TypeText),
				options.WithColumn("data", types.TypeText),
				options.WithPrimaryKeyColumn("id"),
			); err != nil {
				log.Printf("%s: create avatars: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "pushTokens"),
				options.WithColumn("id", types.TypeText),
				options.WithColumn("token", types.TypeText),
				options.WithPrimaryKeyColumn("id"),
			); err != nil {
				log.Printf("%s: create pushTokens: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "tokens"),
				options.WithColumn("id", types.TypeText),
				options.WithColumn("token", types.TypeText),
				options.WithPrimaryKeyColumn("id"),
			); err != nil {
				log.Printf("%s: create tokens: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "friendRequests"),
				options.WithColumn("sender", types.TypeText),
				options.WithColumn("target", types.TypeText),
				options.WithPrimaryKeyColumn("sender", "target"),
			); err != nil {
				log.Printf("%s: create friendRequests: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "friends"),
				options.WithColumn("friendA", types.TypeText),
				options.WithColumn("friendB", types.TypeText),
				options.WithPrimaryKeyColumn("friendA", "friendB"),
			); err != nil {
				log.Printf("%s: create friends: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "spendings"),
				options.WithColumn("id", types.TypeText),
				options.WithColumn("dealId", types.TypeText),
				options.WithColumn("cost", types.TypeInt64),
				options.WithColumn("counterparty", types.TypeText),
				options.WithPrimaryKeyColumn("id"),
			); err != nil {
				log.Printf("%s: create spendings: unexpected err %v", op, err)
				return err
			}
			if err := s.CreateTable(ctx, path.Join(dbName, "deals"),
				options.WithColumn("id", types.TypeText),
				options.WithColumn("timestamp", types.TypeInt64),
				options.WithColumn("details", types.TypeText),
				options.WithColumn("cost", types.TypeInt64),
				options.WithColumn("currency", types.TypeText),
				options.WithPrimaryKeyColumn("id"),
			); err != nil {
				log.Printf("%s: create deals: unexpected err %v", op, err)
				return err
			}
			return nil
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func new(ctx context.Context, db *ydb.Driver) (storage.Storage, error) {
	const op = "storage.ydb.new"
	return &Storage{db: db, ctx: ctx}, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *Storage) GetAccountInfo(uid storage.UserId) (*storage.ProfileInfo, error) {
	const op = "storage.ydb.GetAccountInfo"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res    result.Result
		result *storage.ProfileInfo
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $id AS Text;
SELECT 
	c.email,
	e.confirmed,
	u.displayName,
	u.avatarId
FROM 
	credentials c
	JOIN emails e ON c.email = e.email
	JOIN userInfos u ON c.id = u.id
WHERE 
	c.id = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var email string
				var confirmed bool
				var displayName string
				var avatarId *string
				err = res.Scan(&email, &confirmed, &displayName, &avatarId)
				if err != nil {
					return err
				}
				info := storage.ProfileInfo{
					User: storage.User{
						Id:          uid,
						DisplayName: displayName,
						Avatar: storage.Avatar{
							Id: (*storage.AvatarId)(avatarId),
						},
						FriendStatus: storage.FriendStatusMe,
					},
					Email:         email,
					EmailVerified: confirmed,
				}
				result = &info
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return nil, err
	}
	return result, nil
}

func (s *Storage) GetUserId(email string) (*storage.UserId, error) {
	const op = "storage.ydb.GetUserId"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res    result.Result
		uid    *storage.UserId
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $email AS Text;
SELECT 
	id 
FROM 
	credentials 
WHERE 
	email = $email;`,
			table.NewQueryParameters(
				table.ValueParam("$email", types.TextValue(email)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var value string
				err = res.Scan(&value)
				if err != nil {
					return err
				}
				var uidValue = storage.UserId(value)
				uid = &uidValue
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return nil, err
	}
	return uid, nil
}

func (s *Storage) StoreEmailValidationToken(email string, token string) error {
	const op = "storage.ydb.StoreEmailValidationToken"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $email AS Text;
DECLARE $validationToken AS Text;
UPSERT INTO 
	emails(email, confirmed, validationToken) 
VALUES($email, False, $validationToken);`,
			table.NewQueryParameters(
				table.ValueParam("$email", types.TextValue(email)),
				table.ValueParam("$validationToken", types.TextValue(token)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) ExtractEmailValidationToken(email string) (*string, error) {
	const op = "storage.ydb.ExtractEmailValidationToken"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
		token *string
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $email AS Text;
SELECT 
	validationToken 
FROM 
	emails 
WHERE 
	email = $email AND validationToken IS NOT NULL;`,
			table.NewQueryParameters(
				table.ValueParam("$email", types.TextValue(email)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var value *string
				err = res.Scan(&value)
				if err != nil {
					return err
				}
				token = value
			}
		}
		if err := res.Err(); err != nil {
			return err
		}
		if token == nil {
			return nil
		}
		_, res, err = s.Execute(ctx, writeTx, `
DECLARE $email AS Text;
UPSERT INTO 
	emails(email, confirmed, validationToken) 
VALUES($email, False, NULL);`,
			table.NewQueryParameters(
				table.ValueParam("$email", types.TextValue(email)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return token, err
	}
	return token, nil
}

func (s *Storage) ValidateEmail(email string) error {
	const op = "storage.ydb.ValidateEmail"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $email AS Text;
UPSERT INTO 
	emails(email, confirmed, validationToken) 
VALUES($email, True, NULL);`,
			table.NewQueryParameters(
				table.ValueParam("$email", types.TextValue(email)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) UpdateEmail(uid storage.UserId, email string) error {
	const op = "storage.ydb.UpdateEmail"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $email AS Text;
UPDATE 
	credentials
SET 
	email = $email
WHERE
	id = $id;

DELETE FROM
	emails
WHERE
	email = $email;

UPSERT INTO 
	emails(email, confirmed, validationToken) 
VALUES($email, False, NULL);
`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
				table.ValueParam("$email", types.TextValue(email)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) IsEmailExists(email string) (bool, error) {
	const op = "storage.ydb.IsEmailExists"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		exists bool
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, readTx, `
DECLARE $email AS Text;
SELECT 
	True
FROM 
	emails 
WHERE 
	email = $email;`,
			table.NewQueryParameters(table.ValueParam("$email", types.TextValue(email))),
		)
		if err != nil {
			return err
		}
		exists = false
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				if err := res.Scan(&exists); err != nil {
					return err
				}
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return false, err
	}
	return exists, nil
}

func (s *Storage) IsUserExists(uid storage.UserId) (bool, error) {
	const op = "storage.ydb.IsUserExists"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		exists bool
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, readTx, `
DECLARE $id AS Text;
SELECT 
	True
FROM 
	credentials 
WHERE 
	id = $id;`,
			table.NewQueryParameters(table.ValueParam("$id", types.TextValue(string(uid)))),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				if err := res.Scan(&exists); err != nil {
					return err
				}
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return false, err
	}
	return exists, nil
}

func (s *Storage) CheckCredentials(credentials storage.UserCredentials) (bool, error) {
	const op = "storage.ydb.CheckCredentials"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res    result.Result
		passed bool
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $email AS Text;
SELECT 
	password 
FROM 
	credentials 
WHERE 
	email = $email;`,
			table.NewQueryParameters(table.ValueParam("$email", types.TextValue(credentials.Email))),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var passwordHash string
				err = res.Scan(&passwordHash)
				if err != nil {
					return err
				}
				if checkPasswordHash(credentials.Password, passwordHash) {
					passed = true
				}
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return false, err
	}
	return passed, nil
}

func (s *Storage) CheckPasswordForId(uid storage.UserId, password string) (bool, error) {
	const op = "storage.ydb.CheckPasswordForId"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res    result.Result
		passed bool
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $id AS Text;
SELECT 
	password 
FROM 
	credentials 
WHERE 
	id = $id;`,
			table.NewQueryParameters(table.ValueParam("$id", types.TextValue(string(uid)))),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var passwordHash string
				err = res.Scan(&passwordHash)
				if err != nil {
					return err
				}
				if checkPasswordHash(password, passwordHash) {
					passed = true
				}
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return false, err
	}
	return passed, nil
}

func (s *Storage) UpdatePasswordForId(uid storage.UserId, password string) error {
	const op = "storage.ydb.UpdatePasswordForId"
	log.Printf("%s: start", op)
	passwordHash, err := hashPassword(password)
	if err != nil {
		log.Printf("%s: cannot hash password %v", op, err)
		return err
	}
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err = s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $password AS Text;
UPDATE 
	credentials
SET 
	password = $password
WHERE
	id = $id;
`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
				table.ValueParam("$password", types.TextValue(passwordHash)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) StoreCredentials(uid storage.UserId, credentials storage.UserCredentials) error {
	const op = "storage.ydb.StoreCredentials"
	log.Printf("%s: start", op)
	passwordHash, err := hashPassword(credentials.Password)
	if err != nil {
		log.Printf("%s: cannot hash password %v", op, err)
		return err
	}
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err = s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $email AS Text;
DECLARE $password AS Text;
INSERT INTO 
	credentials(id, email, password) 
VALUES($id, $email, $password);

INSERT INTO 
	userInfos(id, displayName, avatarId) 
VALUES($id, $email, NULL);

INSERT INTO 
	emails(email, confirmed, validationToken) 
VALUES($email, False, NULL);
`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
				table.ValueParam("$email", types.TextValue(credentials.Email)),
				table.ValueParam("$password", types.TextValue(passwordHash)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) StorePushToken(uid storage.UserId, token string) error {
	const op = "storage.ydb.StorePushToken"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $token AS Text;
UPSERT INTO 
	pushTokens(id, token) 
VALUES($id, $token);`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
				table.ValueParam("$token", types.TextValue(token)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) GetPushToken(uid storage.UserId) (*string, error) {
	const op = "storage.ydb.GetPushToken"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res    result.Result
		token  *string
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $id AS Text;
SELECT 
	token 
FROM 
	pushTokens 
WHERE 
	id = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var value string
				err = res.Scan(&value)
				if err != nil {
					return err
				}
				token = &value
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return nil, err
	}
	return token, nil
}

func (s *Storage) StoreDisplayName(uid storage.UserId, displayName string) error {
	const op = "storage.ydb.StoreDisplayName"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $displayName AS Text;
UPDATE 
	userInfos
SET 
	displayName = $displayName
WHERE
	id = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
				table.ValueParam("$displayName", types.TextValue(displayName)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) GetAvatarsBase64(aids []storage.AvatarId) (map[storage.AvatarId]storage.AvatarData, error) {
	const op = "storage.ydb.GetAvatarsBase64"
	log.Printf("%s: start", op)

	var (
		readTx     = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		unfiltered []types.Value
		result     map[storage.AvatarId]storage.AvatarData
	)
	for i := 0; i < len(aids); i++ {
		unfiltered = append(unfiltered, types.TextValue(string(aids[i])))
	}

	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, res, err := s.Execute(ctx, readTx, `
DECLARE $aids AS List<Text>;
SELECT 
	id, data
FROM 
	avatars
WHERE 
	id in $aids;`,
				table.NewQueryParameters(
					table.ValueParam("$aids", types.ListValue(unfiltered...)),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()

			result = map[storage.AvatarId]storage.AvatarData{}

			for res.NextResultSet(ctx) {
				for res.NextRow() {
					var id string
					var data string
					res.ScanNamed()
					if err := res.Scan(&id, &data); err != nil {
						return err
					}
					result[storage.AvatarId(id)] = storage.AvatarData{
						Base64Data: &data,
					}
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return result, err
	}
	return result, nil
}

func (s *Storage) StoreAvatarBase64(uid storage.UserId, data string) (storage.AvatarId, error) {
	const op = "storage.ydb.StoreAvatar"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
		avatarId = uuid.New().String()
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $avatarId AS Text;
DECLARE $data AS Text;
INSERT INTO 
	avatars(id, data) 
VALUES($avatarId, $data);`,
			table.NewQueryParameters(
				table.ValueParam("$avatarId", types.TextValue(avatarId)),
				table.ValueParam("$data", types.TextValue(data)),
			),
		)
		if err != nil {
			return err
		}
		res.Close()

		_, res, err = s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $avatarId AS Text;
UPDATE 
	userInfos
SET 
	avatarId = $avatarId
WHERE
	id = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
				table.ValueParam("$avatarId", types.TextValue(avatarId)),
			),
		)
		if err != nil {
			return err
		}
		res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return "", err
	}
	return storage.AvatarId(avatarId), nil
}

func (s *Storage) GetRefreshToken(uid storage.UserId) (*string, error) {
	const op = "storage.ydb.GetRefreshToken"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res    result.Result
		token  *string
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $id AS Text;
SELECT 
	token 
FROM 
	tokens 
WHERE 
	id = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var value string
				err = res.Scan(&value)
				if err != nil {
					return err
				}
				token = &value
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return nil, err
	}
	return token, nil
}

func (s *Storage) StoreRefreshToken(token string, uid storage.UserId) error {
	const op = "storage.ydb.StoreRefreshToken"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $token AS Text;
UPSERT INTO 
	tokens(id, token) 
VALUES($id, $token);`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
				table.ValueParam("$token", types.TextValue(token)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) RemoveRefreshToken(uid storage.UserId) error {
	const op = "storage.ydb.RemoveRefreshToken"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DELETE FROM 
	tokens 
WHERE 
	id = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(uid))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) StoreFriendRequest(sender storage.UserId, target storage.UserId) error {
	const op = "storage.ydb.StoreFriendRequest"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $sender AS Text;
DECLARE $target AS Text;
UPSERT INTO 
	friendRequests(sender, target) 
VALUES($sender, $target);`,
			table.NewQueryParameters(
				table.ValueParam("$sender", types.TextValue(string(sender))),
				table.ValueParam("$target", types.TextValue(string(target))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) HasFriendRequest(sender storage.UserId, target storage.UserId) (bool, error) {
	const op = "storage.ydb.HasFriendRequest"
	log.Printf("%s: start", op)
	var (
		readTx     = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res        result.Result
		hasRequest bool
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $sender AS Text;
DECLARE $target AS Text;
SELECT 
	True 
FROM 
	friendRequests 
WHERE 
	sender = $sender AND target = $target;`,
			table.NewQueryParameters(
				table.ValueParam("$sender", types.TextValue(string(sender))),
				table.ValueParam("$target", types.TextValue(string(target))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				if err := res.Scan(&hasRequest); err != nil {
					return err
				}
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return false, err
	}
	return hasRequest, nil
}

func (s *Storage) RemoveFriendRequest(sender storage.UserId, target storage.UserId) error {
	const op = "storage.ydb.RemoveFriendRequest"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $sender AS Text;
DECLARE $target AS Text;
DELETE FROM 
	friendRequests 
WHERE 
	sender = $sender and target = $target;`,
			table.NewQueryParameters(
				table.ValueParam("$sender", types.TextValue(string(sender))),
				table.ValueParam("$target", types.TextValue(string(target))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) StoreFriendship(friendA storage.UserId, friendB storage.UserId) error {
	const op = "storage.ydb.StoreFriendship"
	log.Printf("%s: start", op)
	if friendA > friendB {
		friendA, friendB = friendB, friendA
	}
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $friendA AS Text;
DECLARE $friendB AS Text;
UPSERT INTO 
	friends(friendA, friendB) 
VALUES($friendA, $friendB);`,
			table.NewQueryParameters(
				table.ValueParam("$friendA", types.TextValue(string(friendA))),
				table.ValueParam("$friendB", types.TextValue(string(friendB))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) HasFriendship(friendA storage.UserId, friendB storage.UserId) (bool, error) {
	const op = "storage.ydb.HasFriendship"
	log.Printf("%s: start", op)
	if friendA > friendB {
		friendA, friendB = friendB, friendA
	}
	var (
		readTx        = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		res           result.Result
		hasFriendship bool
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, readTx, `
DECLARE $friendA AS Text;
DECLARE $friendB AS Text;
SELECT 
	True 
FROM 
	friends 
WHERE 
	friendA = $friendA AND friendB = $friendB;`,
			table.NewQueryParameters(
				table.ValueParam("$friendA", types.TextValue(string(friendA))),
				table.ValueParam("$friendB", types.TextValue(string(friendB))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				if err := res.Scan(&hasFriendship); err != nil {
					return err
				}
			}
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return false, err
	}
	return hasFriendship, nil
}

func (s *Storage) RemoveFriendship(friendA storage.UserId, friendB storage.UserId) error {
	const op = "storage.ydb.RemoveFriendship"
	log.Printf("%s: start", op)
	if friendA > friendB {
		friendA, friendB = friendB, friendA
	}
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $friendA AS Text;
DECLARE $friendB AS Text;
DELETE FROM 
	friends 
WHERE 
	friendA = $friendA AND friendB = $friendB;`,
			table.NewQueryParameters(
				table.ValueParam("$friendA", types.TextValue(string(friendA))),
				table.ValueParam("$friendB", types.TextValue(string(friendB))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) GetFriends(uid storage.UserId) ([]storage.UserId, error) {
	const op = "storage.ydb.GetFriends"
	log.Printf("%s: start", op)
	var (
		readTx  = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		friends []storage.UserId
	)
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, res, err := s.Execute(ctx, readTx, `
DECLARE $user AS Text;
SELECT 
	friendA, friendB 
FROM 
	friends 
WHERE 
	friendA = $user OR friendB = $user;`,
				table.NewQueryParameters(
					table.ValueParam("$user", types.TextValue(string(uid))),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					var friendA, friendB string
					if err := res.Scan(&friendA, &friendB); err != nil {
						return err
					}
					if friendA == string(uid) {
						friends = append(friends, storage.UserId(friendB))
					} else {
						friends = append(friends, storage.UserId(friendA))
					}
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return friends, err
	}
	return friends, nil
}

func (s *Storage) GetIncomingRequests(uid storage.UserId) ([]storage.UserId, error) {
	const op = "storage.ydb.GetIncomingRequests"
	log.Printf("%s: start", op)
	var (
		readTx  = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		senders []storage.UserId
	)
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, res, err := s.Execute(ctx, readTx, `
DECLARE $target AS Text;
SELECT 
	sender 
FROM 
	friendRequests 
WHERE 
	target = $target;`,
				table.NewQueryParameters(
					table.ValueParam("$target", types.TextValue(string(uid))),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					var sender string
					if err := res.Scan(&sender); err != nil {
						return err
					}
					senders = append(senders, storage.UserId(sender))
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return senders, err
	}
	return senders, nil
}

func (s *Storage) GetPendingRequests(uid storage.UserId) ([]storage.UserId, error) {
	const op = "storage.ydb.GetIncomingRequests"
	log.Printf("%s: start", op)
	var (
		readTx  = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		targets []storage.UserId
	)
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, res, err := s.Execute(ctx, readTx, `
DECLARE $sender AS Text;
SELECT 
	target 
FROM 
	friendRequests 
WHERE 
	sender = $sender;`,
				table.NewQueryParameters(
					table.ValueParam("$sender", types.TextValue(string(uid))),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					var target string
					if err := res.Scan(&target); err != nil {
						return err
					}
					targets = append(targets, storage.UserId(target))
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return targets, err
	}
	return targets, nil
}

func (s *Storage) getUsersWithoutFriendRelationInfo(ids []storage.UserId) ([]storage.User, error) {
	const op = "storage.ydb.getUsersWithoutFriendRelationInfo"
	log.Printf("%s: start args: %v", op, ids)
	var (
		readTx     = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		unfiltered []types.Value
		uids       []storage.User
	)
	for i := 0; i < len(ids); i++ {
		unfiltered = append(unfiltered, types.TextValue(string(ids[i])))
	}

	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, res, err := s.Execute(ctx, readTx, `
DECLARE $uids AS List<Text>;
SELECT 
	id, displayName, avatarId
FROM 
	userInfos
WHERE 
	id in $uids;`,
				table.NewQueryParameters(
					table.ValueParam("$uids", types.ListValue(unfiltered...)),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					var displayName string
					var id string
					var avatarId *string
					res.ScanNamed()
					if err := res.Scan(&id, &displayName, &avatarId); err != nil {
						return err
					}
					uids = append(uids, storage.User{
						Id:          storage.UserId(id),
						DisplayName: displayName,
						Avatar: storage.Avatar{
							Id: (*storage.AvatarId)(avatarId),
						},
						FriendStatus: storage.FriendStatusNo,
					})
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return uids, err
	}
	return uids, nil
}

func (s *Storage) GetUsers(sender storage.UserId, ids []storage.UserId) ([]storage.User, error) {
	const op = "storage.ydb.GetUsers"
	log.Printf("%s: start args: %v", op, ids)
	if len(ids) == 0 {
		return []storage.User{}, nil
	}
	users, err := s.getUsersWithoutFriendRelationInfo(ids)
	if err != nil {
		log.Printf("%s: filter invalid ids failed err %v", op, err)
		return []storage.User{}, err
	}
	log.Printf("%s: found %d users", op, len(users))

	pendingFriends, err := s.GetPendingRequests(sender)
	if err != nil {
		log.Printf("%s: get pending friends failed %v", op, err)
	}
	pendingFriendsMap := map[storage.UserId]bool{}
	for i := 0; i < len(pendingFriends); i++ {
		pendingFriendsMap[pendingFriends[i]] = true
	}
	incomingFriends, err := s.GetIncomingRequests(sender)
	if err != nil {
		log.Printf("%s: get incoming friends failed %v", op, err)
	}
	incomingFriendsMap := map[storage.UserId]bool{}
	for i := 0; i < len(incomingFriends); i++ {
		incomingFriendsMap[incomingFriends[i]] = true
	}
	friends, err := s.GetFriends(sender)
	if err != nil {
		log.Printf("%s: get friends failed %v", op, err)
	}
	friendsMap := map[storage.UserId]bool{}
	for i := 0; i < len(friends); i++ {
		friendsMap[friends[i]] = true
	}

	for i := 0; i < len(users); i++ {
		var status storage.FriendStatus
		if users[i].Id == sender {
			status = storage.FriendStatusMe
		} else if pendingFriendsMap[users[i].Id] {
			status = storage.FriendStatusOutgoingRequest
		} else if incomingFriendsMap[users[i].Id] {
			status = storage.FriendStatusIncomingRequest
		} else if friendsMap[users[i].Id] {
			status = storage.FriendStatusFriends
		} else {
			status = storage.FriendStatusNo
		}
		users[i].FriendStatus = status
	}
	return users, nil
}

func (s *Storage) SearchUsers(sender storage.UserId, query string) ([]storage.User, error) {
	const op = "storage.ydb.SearchUsers"
	log.Printf("%s: start", op)
	if len(query) == 0 {
		return []storage.User{}, nil
	}
	var (
		readTx  = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		targets []storage.UserId
	)
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, res, err := s.Execute(ctx, readTx, fmt.Sprintf(`
DECLARE $query AS Text;
SELECT 
	id 
FROM 
	userInfos
WHERE 
	displayName LIKE '%s%%' or displayName = $query ORDER BY id;`, query),
				table.NewQueryParameters(
					table.ValueParam("$query", types.TextValue(query)),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					var sender string
					if err := res.Scan(&sender); err != nil {
						return err
					}
					targets = append(targets, storage.UserId(sender))
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return []storage.User{}, err
	}
	return s.GetUsers(sender, targets)
}

func (s *Storage) InsertDeal(deal storage.Deal) (storage.DealId, error) {
	const op = "storage.ydb.InsertDeal"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
		dealId = storage.DealId(uuid.New().String())
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $timestamp AS Int64;
DECLARE $details AS Text;
DECLARE $cost AS Int64;
DECLARE $currency AS Text;
INSERT INTO 
	deals(id, timestamp, details, cost, currency) 
VALUES($id, $timestamp, $details, $cost, $currency);`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(dealId))),
				table.ValueParam("$timestamp", types.Int64Value(deal.Timestamp)),
				table.ValueParam("$details", types.TextValue(deal.Details)),
				table.ValueParam("$cost", types.Int64Value(int64(deal.Cost))),
				table.ValueParam("$currency", types.TextValue(deal.Currency)),
			),
		)
		if err != nil {
			return err
		}

		defer res.Close()
		for i := 0; i < len(deal.Spendings); i++ {
			_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DECLARE $dealId AS Text;
DECLARE $cost AS Int64;
DECLARE $counterparty AS Text;
INSERT INTO 
	spendings(id, dealId, cost, counterparty) 
VALUES($id, $dealId, $cost, $counterparty);`,
				table.NewQueryParameters(
					table.ValueParam("$id", types.TextValue(uuid.New().String())),
					table.ValueParam("$dealId", types.TextValue(string(dealId))),
					table.ValueParam("$cost", types.Int64Value(int64(deal.Spendings[i].Cost))),
					table.ValueParam("$counterparty", types.TextValue(string(deal.Spendings[i].UserId))),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return dealId, err
	}
	return dealId, nil
}

func (s *Storage) GetDeal(did storage.DealId) (*storage.IdentifiableDeal, error) {
	const op = "storage.ydb.GetDeal"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		deal   *storage.IdentifiableDeal
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, readTx, `
DECLARE $id AS Text;
SELECT 
	d.id, 
	d.timestamp,
	d.details,
	d.cost,
	d.currency,
	s.cost,
	s.counterparty
FROM 
	deals d
	JOIN spendings s ON s.dealId = d.id
WHERE 
	d.id = $id;`,
			table.NewQueryParameters(table.ValueParam("$id", types.TextValue(string(did)))),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.NextResultSet(ctx) {
			for res.NextRow() {
				var _deal storage.IdentifiableDeal
				if deal == nil {
					_deal = storage.IdentifiableDeal{}
				} else {
					_deal = *deal
				}
				var cost int64
				var counterparty string
				err = res.Scan(
					&_deal.Id,
					&_deal.Timestamp,
					&_deal.Details,
					&_deal.Cost,
					&_deal.Currency,
					&cost,
					&counterparty)
				_deal.Spendings = append(_deal.Spendings, storage.Spending{UserId: storage.UserId(counterparty), Cost: cost})
				deal = &_deal
			}
		}
		if err := res.Err(); err != nil {
			return err
		}
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return nil, err
	}
	return deal, nil
}

func (s *Storage) RemoveDeal(did storage.DealId) error {
	const op = "storage.ydb.RemoveDeal"
	log.Printf("%s: start", op)
	var (
		writeTx = table.TxControl(
			table.BeginTx(
				table.WithSerializableReadWrite(),
			),
			table.CommitTx(),
		)
	)
	err := s.db.Table().Do(s.ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DELETE FROM 
	deals 
WHERE 
	id = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(did))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		_, res, err = s.Execute(ctx, writeTx, `
DECLARE $id AS Text;
DELETE FROM 
	spendings 
WHERE 
	dealId = $id;`,
			table.NewQueryParameters(
				table.ValueParam("$id", types.TextValue(string(did))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		return res.Err()
	})
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return err
	}
	return nil
}

func (s *Storage) GetDeals(counterparty1 storage.UserId, counterparty2 storage.UserId) ([]storage.IdentifiableDeal, error) {
	const op = "storage.ydb.GetDeals"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(
			table.BeginTx(
				table.WithOnlineReadOnly(),
			),
			table.CommitTx(),
		)
		deals = []storage.IdentifiableDeal{}
	)
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			var (
				res result.Result
			)
			_, res, err = s.Execute(ctx, readTx, `
DECLARE $counterparty1 AS Text;
DECLARE $counterparty2 AS Text;
SELECT
  s1.dealId,
  s1.counterparty,
  s1.cost,
  s2.counterparty,
  s2.cost,
  d.timestamp,
  d.details,
  d.cost,
  d.currency
FROM
  spendings s1
  JOIN spendings s2 ON s1.dealId = s2.dealId
  JOIN deals d ON s1.dealId = d.id
WHERE
  s1.counterparty = $counterparty1 AND s2.counterparty = $counterparty2;`,
				table.NewQueryParameters(
					table.ValueParam("$counterparty1", types.TextValue(string(counterparty1))),
					table.ValueParam("$counterparty2", types.TextValue(string(counterparty2))),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			log.Printf("> select_simple_transaction:\n")
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					deal := storage.IdentifiableDeal{}
					deal.Spendings = make([]storage.Spending, 2)
					var uid1 string
					var uid2 string
					var did string
					err = res.Scan(
						&did,
						&uid1,
						&deal.Spendings[0].Cost,
						&uid2,
						&deal.Spendings[1].Cost,
						&deal.Timestamp,
						&deal.Details,
						&deal.Cost,
						&deal.Currency)
					if err != nil {
						return err
					}
					deal.Id = storage.DealId(did)
					deal.Spendings[0].UserId = storage.UserId(uid1)
					deal.Spendings[1].UserId = storage.UserId(uid2)
					deals = append(deals, deal)
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return deals, err
	}
	return deals, nil
}

func (s *Storage) GetCounterpartiesForDeal(did storage.DealId) ([]storage.UserId, error) {
	const op = "storage.ydb.GetCounterpartiesForDeal"
	log.Printf("%s: start", op)
	var (
		readTx         = table.TxControl(table.BeginTx(table.WithOnlineReadOnly()), table.CommitTx())
		counterparties []storage.UserId
	)
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, res, err := s.Execute(ctx, readTx, `
DECLARE $dealId AS Text;
SELECT 
	counterparty 
FROM 
	spendings 
WHERE 
	dealId = $dealId;`,
				table.NewQueryParameters(
					table.ValueParam("$dealId", types.TextValue(string(did))),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					var counterparty string
					if err := res.Scan(&counterparty); err != nil {
						return err
					}
					counterparties = append(counterparties, storage.UserId(counterparty))
				}
			}
			return res.Err()
		},
	)
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return counterparties, err
	}
	return counterparties, nil
}

func (s *Storage) GetCounterparties(target storage.UserId) ([]storage.SpendingsPreview, error) {
	const op = "storage.ydb.GetCounterparties"
	log.Printf("%s: start", op)
	var (
		readTx = table.TxControl(
			table.BeginTx(
				table.WithOnlineReadOnly(),
			),
			table.CommitTx(),
		)
		spendingsMap = map[string]storage.SpendingsPreview{}
	)
	err := s.db.Table().Do(s.ctx,
		func(ctx context.Context, s table.Session) (err error) {
			var (
				res      result.Result
				currency string
				user     string
				cost     int64
			)
			_, res, err = s.Execute(ctx, readTx, `
DECLARE $target AS Text;
SELECT
  d.currency,
  s2.counterparty,
  s1.cost
FROM
  deals d
  JOIN spendings s1 ON s1.dealId = d.id
  JOIN spendings s2 ON s2.dealId = d.id
WHERE 
  s1.counterparty = $target AND s2.counterparty != $target;`,
				table.NewQueryParameters(
					table.ValueParam("$target", types.TextValue(string(target))),
				),
			)
			if err != nil {
				return err
			}
			defer res.Close()
			for res.NextResultSet(ctx) {
				for res.NextRow() {
					err = res.Scan(
						&currency,
						&user,
						&cost,
					)
					if err != nil {
						return err
					}
					_, ok := spendingsMap[user]
					if !ok {
						spendingsMap[user] = storage.SpendingsPreview{
							Counterparty: user,
							Balance:      map[string]int64{},
						}
					}
					spendingsMap[user].Balance[currency] += cost
				}
			}
			return res.Err()
		},
	)
	spendings := make([]storage.SpendingsPreview, 0, len(spendingsMap))
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return spendings, err
	}

	for _, value := range spendingsMap {
		spendings = append(spendings, value)
	}
	return spendings, nil
}

func (s *Storage) Close() {
	log.Printf("storage closed")
	s.db.Close(s.ctx)
}
