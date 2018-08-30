package model

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aitour/scene/web/config"

	"github.com/aitour/scene/auth"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	sqlCreateUser = `
	INSERT INTO "ai_user" (email, password, create_at,  enabled, deleted) VALUES ($1, $2, now(), true, false) RETURNING id
	`

	sqlGetUserIdByEmail = `
	SELECT id from ai_user where email=$1
	`
	// sqlCheckActivationKey = `
	// SELECT COUNT(*) FROM ai_user WHERE email=$1 AND activate_key=$2 AND EXTRACT(EPOCH FROM (timestamp 'now' - create_at))/3600 < 24
	// `

	sqlSetUserPassword = `
	UPDATE ai_user set password=$2 where id=$1
	`

	sqlActivateUserAccount = `
	UPDATE ai_user set activated=true where email=$1
	`

	sqlBindOpenid = `
	INSERT INTO ai_user_oauth (user_id, platform, openid) select $1, $2, $3 where not exists (select 1 from ai_user_oauth where user_id=$1 and platform=$2 and openid=$3)
	`

	sqlGetUserBindwithOpenId = `
	SELECT user_id from ai_user_oauth where platform=$1 and openid=$2
	`

	sqlGetActivatedUserByEmail = `
	SELECT id, name, phone, create_at, password FROM "ai_user" WHERE email=$1 and enabled=true and activated=true and deleted=false
	`

	sqlGetActivatedUserById = `
	SELECT email, name, phone, create_at, password FROM "ai_user" WHERE id=$1 and enabled=true and activated=true and deleted=false
	`

	sqlSetUserAvatar = `
	WITH updated AS(
		UPDATE ai_user_profile 
		SET avatar=$2 
		WHERE user_id=$1 
		RETURNING *
	)
	INSERT INTO ai_user_profile (user_id, avatar)
	SELECT $1,$2
	WHERE NOT EXISTS (SELECT * FROM updated);
	`

	sqlSetUserLanguage = `
	WITH updated AS(
		UPDATE ai_user_profile 
		SET lang=$2 
		WHERE user_id=$1 
		RETURNING *
	)
	INSERT INTO ai_user_profile (user_id, lang)
	SELECT $1,$2
	WHERE NOT EXISTS (SELECT * FROM updated);
	`

	sqlSetUserNickName = `
	WITH updated AS(
		UPDATE ai_user_profile 
		SET nickname=$2 
		WHERE user_id=$1 
		RETURNING *
	)
	INSERT INTO ai_user_profile (user_id, nickname)
	SELECT $1,$2
	WHERE NOT EXISTS (SELECT * FROM updated);
	`

	sqlGetUserProfile = `
	SELECT lang,nickname,avatar FROM ai_user_profile WHERE user_id=$1
	`

	sqlGetOauthInfo = `
	SELECT platform, openid, access_token from ai_user_oauth where user_id=$1
	`
)

var (
	ErrorInvalidUserNameOrPassword = fmt.Errorf("invalid user name or password")
)

var (
	cfg *config.Config
	db  *sqlx.DB
)

func init() {
	cfg = config.GetConfig()
	if cfg == nil {
		log.Fatalln("unable to get app config")
	}
	var err error
	db, err = sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Db.Host, cfg.Db.Port, cfg.Db.User, cfg.Db.Password, cfg.Db.DbName))
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("db connected")
	}
}

type User struct {
	Id       int64
	Name     string
	Phone    string
	Password string
	Email    string
	CreateAt time.Time `db:"create_at"`
}

type UserProfile struct {
	Id       int64  `json:"-"`
	Lang     string `json:"lang"`
	NickName string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func CreateUser(email string, password string, language string) (*User, error) {
	createAt := time.Now()

	//hash password
	hashPwd := auth.HashAndSalt([]byte(password))
	user := &User{
		Email:    email,
		Password: hashPwd,
		CreateAt: createAt,
	}

	row := db.QueryRow(sqlCreateUser, user.Email, user.Password)
	if row == nil {
		return nil, fmt.Errorf("query error")
	}
	err := row.Scan(&user.Id)
	if err != nil {
		return nil, err
	}

	if len(language) > 0 {
		err = SetUserProfile(user.Id, map[string]string{
			"lang": language,
		})
		if err != nil {
			return user, err
		}
	}
	return user, err
}

func BindOpenId(platform string, email string, openid string) (*User, error) {
	//检查是否存在用该email注册的用户
	user := &User{}
	db.Get(&user.Id, sqlGetUserIdByEmail, email)
	if user.Id == 0 {
		var err error
		if user, err = CreateUser(email, "", ""); err != nil {
			return nil, err
		}
		ActivateUser(email)
	}

	//关联openid
	_, err := db.Exec(sqlBindOpenid, user.Id, platform, openid)
	return user, err
}

func VerifyUser(email string, password string) (*User, error) {
	var user User
	row := db.QueryRowx(sqlGetActivatedUserByEmail, email)
	var uid int64
	var name, phone sql.NullString
	if err := row.Scan(&uid, &name, &phone, &user.CreateAt, &user.Password); err != nil {
		return nil, err
	}

	checkPass := auth.ComparePasswords(user.Password, []byte(password))
	if !checkPass {
		return nil, ErrorInvalidUserNameOrPassword
	}
	user.Id = uid
	user.Password = ""
	user.Name = name.String
	user.Phone = name.String
	return &user, nil
}

func VerifyUserById(uid int64, password string) (*User, error) {
	var user User
	row := db.QueryRowx(sqlGetActivatedUserById, uid)
	var email, name, phone sql.NullString
	if err := row.Scan(&email, &name, &phone, &user.CreateAt, &user.Password); err != nil {
		return nil, err
	}

	checkPass := auth.ComparePasswords(user.Password, []byte(password))
	if !checkPass {
		return nil, ErrorInvalidUserNameOrPassword
	}
	user.Id = uid
	user.Email = email.String
	user.Password = ""
	user.Name = name.String
	user.Phone = name.String
	return &user, nil
}

func VerifyUserByOpenId(platform, openid string) (*User, error) {
	var user User
	err := db.Get(&user.Id, sqlGetUserBindwithOpenId, platform, openid)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func SetUserPassword(uid int64, password string) error {
	//hash password
	hashPwd := auth.HashAndSalt([]byte(password))
	_, err := db.Exec(sqlSetUserPassword, uid, hashPwd)
	return err
}

func ActivateUser(email string) bool {
	result, err := db.Exec(sqlActivateUserAccount, email)
	if err != nil {
		log.Printf("activate user error:%v", err)
		return false
	}
	count, err := result.RowsAffected()
	if err != nil {
		log.Printf("activate user error:%v", err)
	}
	return err == nil && count == 1
}

// -- user profile --
func GetUserProfile(uid int64) (*UserProfile, error) {
	var profile UserProfile
	var lang, nickname, avatar sql.NullString
	row := db.QueryRow(sqlGetUserProfile, uid)
	if err := row.Scan(&lang, &nickname, &avatar); err != nil {
		return nil, err
	}
	if lang.Valid {
		profile.Lang = lang.String
	}
	if nickname.Valid {
		profile.NickName = nickname.String
	}
	if avatar.Valid {
		profile.Avatar = avatar.String
	}
	return &profile, nil
}

func SetUserAvatar(uid int64, avatar string) error {
	return nil
}

func SetUserProfile(uid int64, props map[string]string) error {
	acceptProps := map[string]string{
		"avatar":   sqlSetUserAvatar,
		"nickname": sqlSetUserNickName,
		"lang":     sqlSetUserLanguage,
	}
	tx := db.MustBegin()
	for key, value := range props {
		sqlStat, ok := acceptProps[key]
		if !ok {
			tx.Rollback()
			return fmt.Errorf("prop %s not recognize", key)
		}
		if len(value) > 0 {
			if _, err := tx.Exec(sqlStat, uid, value); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

type OAuthInfo struct {
	Platform    string
	OpenId      string
	AccessToken string
}

func GetOAuthInfo(uid int64) ([]*OAuthInfo, error) {
	var oauthInfo []*OAuthInfo
	err := db.Select(&oauthInfo, sqlGetOauthInfo, uid)
	return oauthInfo, err
}
