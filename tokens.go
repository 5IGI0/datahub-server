package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type TokenRow struct {
	Id              uint32         `db:"id"`
	Token           string         `db:"token"`
	Ratelimit       sql.NullInt64  `db:"ratelimit"`
	RatelimitWindow sql.NullInt64  `db:"ratelimit_window"`
	Flags           uint64         `db:"flags"`
	Comment         sql.NullString `db:"comment"`
	ExpiredAt       sql.NullInt64  `db:"expired_at"`
}

var ApiAccessDeniedErr = errors.New("access denied")
var ApiInvalidTokenErr = errors.Join(ApiAccessDeniedErr, errors.New("invalid token"))
var ApiFeedPermError = errors.Join(ApiAccessDeniedErr, errors.New("not allowed to interact with feed endpoints"))
var ApiExpiredTokenErr = errors.Join(ApiAccessDeniedErr, errors.New("token expired"))
var ApiRateLimitErr = errors.New("ratelimit exceeded")

type RatelimitType struct {
	window_id int64
	req_count int64
}

var RatelimitMap = make(map[string]RatelimitType)
var RatelimitLock = sync.Mutex{}

func check_token_perms(token_flags uint64, endpoint_flags int) error {

	if (token_flags & TOKEN_ADMIN) != 0 {
		return nil
	}

	/* if the token is admin we don't go here */
	if (endpoint_flags & API_ADMIN) != 0 {
		return ApiAccessDeniedErr
	}

	/* only specific tokens can interect with feed-related endpoints. */
	if (endpoint_flags&API_FEED) != 0 && (token_flags&TOKEN_FEED) == 0 {
		return ApiFeedPermError
	}

	return nil
}

func check_token_ratelimit(w http.ResponseWriter, token string, token_rt int64, token_rt_win int64) error {
	if token_rt <= 0 {
		return ApiRateLimitErr
	} else if token_rt_win <= 0 {
		return nil
	}

	RatelimitLock.Lock()
	defer RatelimitLock.Unlock()

	r, _ := RatelimitMap[token]
	timestamp := time.Now().Unix()
	window_id := (timestamp / token_rt_win)

	if r.window_id != window_id {
		RatelimitMap[token] = RatelimitType{
			window_id: window_id,
			req_count: 1}
		return nil
	} else if r.req_count >= token_rt {
		w.Header().Add("Retry-After", fmt.Sprint(token_rt_win-(timestamp%token_rt_win)))
		return ApiRateLimitErr
	}

	r.req_count++
	RatelimitMap[token] = r

	return nil
}

func ValidateToken(w http.ResponseWriter, r *http.Request, endpoint_flags int) error {
	is_anonymous, token := GetTokenFromRequest(r)

	if !is_anonymous {
		if token == "" {
			return ApiInvalidTokenErr
		}

		row := GlobalContext.Database.QueryRowx("SELECT ratelimit, ratelimit_window, flags, expired_at FROM tokens WHERE token = ?", token)

		var token_row TokenRow
		if err := row.StructScan(&token_row); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ApiInvalidTokenErr
			}
			return err
		}

		if token_row.ExpiredAt.Valid && token_row.ExpiredAt.Int64 < time.Now().UTC().Unix() {
			return ApiExpiredTokenErr
		}

		if err := check_token_perms(token_row.Flags, endpoint_flags); err != nil {
			return err
		}

		if !token_row.Ratelimit.Valid || !token_row.RatelimitWindow.Valid {
			return nil
		}

		return check_token_ratelimit(w, token_row.Token, token_row.Ratelimit.Int64, token_row.RatelimitWindow.Int64)
	} else {
		if err := check_token_perms(0, endpoint_flags); err != nil {
			return err
		}

		return check_token_ratelimit(w, token, GlobalContext.DefaultRateLimit, GlobalContext.DefaultRateLimitWindow)
	}
}

func GetTokenFromRequest(r *http.Request) (bool, string) {
	auth_header := r.Header.Get("Authorization")

	if auth_header != "" {
		if !strings.HasPrefix(auth_header, "Bearer ") || len(auth_header) != (36+7 /* UUID + "Bearer " */) {
			return false, ""
		}

		return false, auth_header[7:]
	} else {
		var token string
		if GlobalContext.ForwardedFromHdr == "" {
			split := strings.Split(r.RemoteAddr, ":")
			token = "anonymous:" + strings.Join(split[:len(split)-1], ":")
		} else {
			token = "anonymous:" + r.Header.Get(GlobalContext.ForwardedFromHdr)
		}

		return true, token
	}
}

func ApiTokenInfo(_ http.ResponseWriter, r *http.Request) (any, int, string, error) {
	is_anonymous, token := GetTokenFromRequest(r)

	return map[string]any{
		"anonymous": is_anonymous,
		"token":     token,
	}, 200, "", nil
}

func ApiTokenCreate(_ http.ResponseWriter, r *http.Request) (any, int, string, error) {
	var input struct {
		Comment         string   `json:"comment"`
		ExpiresIn       *int64   `json:"expires_in"`
		Flags           []string `json:"flags"`
		Ratelimit       *int64   `json:"ratelimit"`
		RatelimitWindow *int64   `json:"ratelimit_window"`
	}

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	var token_row TokenRow

	if len(input.Comment) != 0 {
		token_row.Comment.Valid = true
		token_row.Comment.String = input.Comment
	}

	if input.ExpiresIn != nil && *input.ExpiresIn >= 0 {
		token_row.ExpiredAt.Valid = true
		token_row.ExpiredAt.Int64 = time.Now().UTC().Unix() + *input.ExpiresIn
	}

	for _, flag := range input.Flags {
		int_flag := map[string]uint64{
			"admin": TOKEN_ADMIN,
			"feed":  TOKEN_FEED}[flag]

		if int_flag == 0 {
			return nil, http.StatusBadRequest, "INVALID_FLAG", errors.New("invalid flag")
		}

		token_row.Flags |= int_flag
	}

	if (input.Ratelimit != nil || input.RatelimitWindow != nil) && (input.Ratelimit == nil || input.Ratelimit == nil) {
		return nil, http.StatusBadRequest, "INVALID_RATELIMIT", errors.New("you need to provide ratelimit and ratelimit window")
	}

	if input.Ratelimit != nil && *input.Ratelimit >= 0 {
		token_row.Ratelimit.Valid = true
		token_row.Ratelimit.Int64 = *input.Ratelimit
	}

	if input.RatelimitWindow != nil && *input.RatelimitWindow >= 0 {
		token_row.Ratelimit.Valid = true
		token_row.Ratelimit.Int64 = *input.RatelimitWindow
	}

	token_uuid, _ := uuid.NewRandom()
	token_row.Token = token_uuid.String()

	q, v := squirrel.Insert("tokens").SetMap(map[string]any{
		"comment":          token_row.Comment,
		"token":            token_row.Token,
		"ratelimit":        token_row.Ratelimit,
		"ratelimit_window": token_row.RatelimitWindow,
		"expired_at":       token_row.ExpiredAt,
		"flags":            token_row.Flags,
	}).MustSql()

	_, e := GlobalContext.Database.Exec(q, v...)

	if e != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", e
	}

	return token_row.Token, 200, "", nil
}
