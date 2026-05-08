package handle

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	User_id  int    `json:"id"`
	Username string `json:"username"`
	Passwrd  string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"ph_no"`
}
type Handler struct {
	DB *sql.DB
}

var tokenkey = []byte(os.Getenv("SECRET_KEY"))

func checkcredentials(u *User) bool {
	if u.Passwrd == "" || u.Username == "" {
		fmt.Println("Missing credentials")
		return false
	}
	return true
}
func GenerateaToken(username string) (string, error) {
	claim := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	jwtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return jwtoken.SignedString(tokenkey)

}
func keyfunction(token *jwt.Token) (any, error) {
	return tokenkey, nil
}
func Verifyjwtoken(r *http.Request) (*jwt.Token, string, error) { // shd i return username of userid???

	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, "", fmt.Errorf("Missing token")
	}

	stringoken := strings.TrimPrefix(header, "Bearer ")

	token, err := jwt.Parse(stringoken, keyfunction)
	if err != nil {
		return nil, stringoken, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, "", fmt.Errorf("Invalid token")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return nil, "", fmt.Errorf("Missing Username")
	}
	return token, username, nil
}

// Handlers
func (h *Handler) Createuser(w http.ResponseWriter, r *http.Request) {
	var a_user User
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&a_user)
	if err != nil {
		log.Println("Error in request", err.Error())
		http.Error(w, "Missing Credentials", 401)
		return
	}
	if !(checkcredentials(&a_user)) {
		http.Error(w, "Missing Credentials", 500)
		return
	}
	if strings.TrimSpace(a_user.Email) == "" || strings.TrimSpace(a_user.Phone) == "" {
		http.Error(w, "Missing Information", 400)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(a_user.Passwrd), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Hash Error", 500)
		return
	}

	err = h.DB.QueryRowContext(r.Context(),
		"INSERT INTO users(username, password, name, email, ph_no, cr_date) VALUES($1,$2, $3, $4, $5, CURRENT_DATE) RETURNING id, name", a_user.Username, string(hash), a_user.Name, a_user.Email, a_user.Phone).Scan(&a_user.User_id, &a_user.Name)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) { // Handle UNIQUENSS OF USERNAME
			if pgErr.Code == "23505" {
				http.Error(w, "User alread exists", http.StatusConflict)
				return
			}
		}
		// if errors.As(err, &pgErr) { // Handle NULlITY OF USERNAME
		// 	if pgErr.Code == "23502" {
		// 		http.Error(w, "Username/Password/ Name is missing", http.StatusBadRequest)
		// 		return
		// 	}
		// }
		log.Println("Error in User creation/ Retrieval ", err)
		//fmt.Println(a_user.User_id, a_user.Username, a_user.Passwrd, a_user.Name)
		http.Error(w, "Internal server Error", 500)
		return
	}
	a_user.Passwrd = "#########"
	json.NewEncoder(w).Encode(map[string]string{
		"id": strconv.Itoa(a_user.User_id),
	})

}

func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	var u User
	var stored_pass string
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&u)
	if err != nil {
		log.Println("Error in request")
		http.Error(w, "Missing Credentials", 401)
		return
	}
	if !(checkcredentials(&u)) {
		http.Error(w, "Missing Credentials", 500)
		return
	}

	err = h.DB.QueryRowContext(r.Context(), "SELECT password FROM users where username=$1", u.Username).Scan(&stored_pass)
	if err != nil {
		log.Println("Invalid user", err.Error())
		http.Error(w, "Invalid User", http.StatusUnauthorized)
		return
	}
	//check password
	err = bcrypt.CompareHashAndPassword([]byte(stored_pass), []byte(u.Passwrd))
	if err != nil {
		log.Println("Invalid Password")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	//Generate JWT
	jwtoken, err := GenerateaToken(u.Username)
	if err != nil {
		http.Error(w, "Token Issues", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": jwtoken})
}

func (h *Handler) UserProfile(w http.ResponseWriter, r *http.Request) {
	//var u User
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}

	var rs struct {
		Name     string `json:"Name"`
		Username string `json:"Username"`
		Ph_no    string `json:"Phone number"`
		Email    string `json:"Email"`
		Date     string `json:"Date"`
		Grps     int    `json:"No_groups"`
	}
	//var rs res
	//json.NewDecoder(r.Body).Decode(&u)
	var dat time.Time
	rs.Username = username
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT name,email,ph_no,cr_date,grps from users where username=$1", rs.Username).Scan(&rs.Name, &rs.Email, &rs.Ph_no, &dat, &rs.Grps)
	if err != nil {
		log.Println("Details of", rs.Username, "missing", err.Error())
		http.Error(w, "Missing Details", 500)
		return
	}
	rs.Date = dat.Format("2006-01-02")

	json.NewEncoder(w).Encode(rs)

}

func (h *Handler) Findfrnd(w http.ResponseWriter, r *http.Request) { // Requires lot of changes. Sqitch to friends table for easier access
	//var u User
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}

	type res struct {
		Friends []string `json:"friendlst"`
	}
	var r1 res
	var user_id int
	err = h.DB.QueryRowContext(r.Context(), "SELECT id from users where username=$1", username).Scan(&user_id)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	rows, err := h.DB.QueryContext(r.Context(), "SELECT u.name from friends f join users u on u.id=f.user_id2 where f.user_id1=$1 AND accepted=true", user_id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "DB Issue", 500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var frnd string
		err = rows.Scan(&frnd)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}

		r1.Friends = append(r1.Friends, frnd)
	}
	rows2, err := h.DB.QueryContext(r.Context(), "SELECT u.name from friends f join users u on u.id=f.user_id1 where f.user_id2=$1 AND accepted=true", user_id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "DB Issue", 500)
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var frnd string
		err = rows2.Scan(&frnd)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}

		r1.Friends = append(r1.Friends, frnd)
	}
	fmt.Println("\n ", r1.Friends, " \n ")

	json.NewEncoder(w).Encode(r1)

}

func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) { //For both Initiating frnd req and accepting tht
	//var u User
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error(), username)
		return
	}

	type res struct {
		Friend_username string `json:"friendname"`
	}
	var friend_id, uid int
	var r1 res
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&r1)
	if err != nil {
		http.Error(w, "Missing JSON", http.StatusBadRequest)
		return
	}
	if username == r1.Friend_username {
		http.Error(w, "Cannot add yourself as friend", 400) //You are already your friend"
		return
	}
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", r1.Friend_username).Scan(&friend_id)
	if err != nil {
		log.Println("User not being fnd")
		http.Error(w, "Not found", 405)
		return
	}
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("User not being fnd")
		http.Error(w, "Not found", 405)
		return
	}
	// _, err = h.DB.ExecContext(r.Context(),
	// 	"UPDATE users set friends = friends || $1 where username=$2", []int32{int32(friend_id)}, username)
	// if err != nil {
	// 	var pgErr *pgconn.PgError

	// 	if errors.As(err, &pgErr) { // Handle UNIQUENSS OF USERNAME
	// 		if pgErr.Code == "23505" {
	// 			http.Error(w, "Friend already exists", http.StatusConflict)
	// 			return
	// 		}
	// 	}
	// 	log.Println("Unable to get friends", err.Error())

	// }// FOUND this not good and rying to migrate this one and groups out fo this but it seems hrd.
	var id1, id2 int
	if friend_id < uid {
		id1 = friend_id
		id2 = uid
	} else {
		id1 = uid
		id2 = friend_id
	}
	_, err = h.DB.ExecContext(r.Context(), // adding to friends colum
		"INSERT INTO friends(user_id1, user_id2, requestedby) values($1,$2,$3) ON CONFLICT(user_id1,user_id2) DO UPDATE SET accepted=true WHERE friends.requestedby!=EXCLUDED.requestedby", id1, id2, uid)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) { // Handle UNIQUENSS
			if pgErr.Code == "23505" {
				http.Error(w, "Friend alread exists", http.StatusConflict)
				return
			}
		}
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	//json.NewEncoder(w).Encode(r1)
}

func (h *Handler) FriendReqs(w http.ResponseWriter, r *http.Request) {
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error(), username)
		return
	}

	var uid int
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	type freqlist struct {
		Pendingreq []string `json:"pendingrequests"`
	}
	var f freqlist

	rows1, err := h.DB.QueryContext(r.Context(),
		"SELECT u.name FROM friends f JOIN users u ON u.id=f.user_id2 WHERE f.user_id1=$1 ", uid)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	found := false
	defer rows1.Close()
	for rows1.Next() {
		found = true
		var fndreq string
		err = rows1.Scan(&fndreq)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		f.Pendingreq = append(f.Pendingreq, fndreq)

	}
	rows2, err := h.DB.QueryContext(r.Context(),
		"SELECT u.name FROM friends f JOIN users u ON u.id=f.user_id1 WHERE f.user_id2=$1 ", uid)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		found = true
		var fndreq string
		err = rows2.Scan(&fndreq)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		f.Pendingreq = append(f.Pendingreq, fndreq)

	}
	if !found {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "No friends found",
		})
		return
	}
	json.NewEncoder(w).Encode(f)

}

func (h *Handler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error(), username)
		return
	}
	type remove struct {
		FndId int `json:"friendrequestid"`
	}
	type resp struct {
		Message string `json:"message"`
	}
	var reqf remove
	var resp1 resp
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&reqf)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	var uid int
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	var requestorID int
	var acc bool

	err = h.DB.QueryRowContext(r.Context(),
		"DELETE from friends WHERE id=$1 AND (user_id1=$2 OR user_id2=$2) RETURNING accepted, requestedby", reqf.FndId, uid).Scan(&acc, &requestorID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No such Friend request", http.StatusNotFound)
			return
		}

		http.Error(w, "Removal Not successful", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if requestorID != uid && !acc {
		resp1.Message = "Request Rejected Successfully"

	} else if !acc {
		resp1.Message = "Request Cancelled Successfully"
	} else {
		resp1.Message = "Friend Removed Successfully"
	}
	json.NewEncoder(w).Encode(resp1)

}

func (h *Handler) AddExpense(w http.ResponseWriter, r *http.Request) {
	//var u User]
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}

	type exp struct {
		Name        string    `json:"expense_name"`
		Description string    `json:"description"`
		Totalamt    float64   `json:"amount"`
		Date        string    `json:"date"`
		Paid        string    `json:"payer_username"`
		GroupName   string    `json:"groupname"`
		Splitop     int       `json:"splitoption"`
		Split       []float64 `json:"split"`
		Splitord    []string  `json:"splitorder"`
	}
	var e exp
	var id, paid_id, group_id int
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&e)
	if err != nil {
		log.Println("Error in JSON", err.Error())
		http.Error(w, "Missing JSON", http.StatusBadRequest)
		return
	}

	if e.Name == "" || e.Totalamt <= 0 {
		log.Println("MissingInfo")
		http.Error(w, "Missinginfo", 401)
		return
	}
	if len(e.Split) != len(e.Splitord) && e.Splitop == 1 {
		log.Println("Split order and SPlits dont match")
		http.Error(w, "Invalid Splits", 400)
		return
	} else if len(e.Split) > len(e.Splitord) {
		log.Println("Split order and SPlits dont match")
		http.Error(w, "Invalid Splits", 400)
		return
	}
	date, err := time.Parse("2006-01-02", e.Date)
	if err != nil {
		http.Error(w, "Wrong Date", 400)
		return
	}
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		log.Println("Creating Transaction")
		http.Error(w, "Internal Server Error", 500)
		return
	}

	defer tx.Rollback()
	err = tx.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", e.Paid).Scan(&paid_id)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	var uid int
	err = tx.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	var check []int
	err = tx.QueryRowContext(r.Context(),
		"SELECT members from groups where name=$1", e.GroupName).Scan(&check)
	if err != nil {
		log.Println("Check if is in group")
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	authorised, payerauth := false, false
	for _, id := range check {
		if id == uid {
			authorised = true

		}
		if id == paid_id {
			payerauth = true
		}
	}
	if !authorised {
		http.Error(w, "Unauthorized Group Access", http.StatusUnauthorized)
		return
	}
	if !payerauth {
		http.Error(w, "Payer not in Group", http.StatusBadRequest)
		return
	}
	err = tx.QueryRowContext(r.Context(),
		"SELECT id from groups where name = $1", e.GroupName).Scan(&group_id)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	mmap := make(map[int]bool)
	for _, i := range check {
		mmap[i] = true
	}

	for _, un := range e.Splitord { // Check if users exist in the group
		var spId int
		err := tx.QueryRowContext(r.Context(),
			"SELECT id FROM users WHERE username=$1", un).Scan(&spId)
		if err != nil {
			http.Error(w, "Split Users Not Found", 400)
			return
		}
		if !mmap[spId] {
			http.Error(w, "User not in Group", 400)
			return
		}
	}

	err = tx.QueryRowContext(r.Context(),
		"insert into expenses(name,description,amount,cr_date, paidby, group_id) values($1,$2,$3,$4,$5,$6) returning id", e.Name, e.Description, e.Totalamt, date, paid_id, group_id).Scan(&id)
	if err != nil {
		log.Println("Error in Expense insrt")
		http.Error(w, err.Error(), 500)
		return
	}

	// SPlit logic

	if e.Splitop == 1 { // which is ratio
		var sum float64
		for _, ratio := range e.Split {
			if ratio < 0 {
				http.Error(w, " Negative split not allowed", 400)
				return
			}
			sum += ratio
		}
		if sum == 0 {
			http.Error(w, "Invalid Ratio", 400)
			return
		}
		for i, ratio := range e.Split {
			e.Split[i] = (ratio / sum) * e.Totalamt
			fmt.Println(e.Split[i], e.Splitord[i])
		}
	} else {
		var sum float64
		for _, amt := range e.Split {
			if amt < 0 {
				http.Error(w, "Negative amount not allowed", 400)
				return
			}
			sum += amt
		}
		if sum > e.Totalamt {
			http.Error(w, "Split Amount Total Mismatch", 400)
			return
		} else if sum < e.Totalamt {
			diff := e.Totalamt - sum
			diffn := len(e.Splitord) - len(e.Split)
			if diffn > 0 {
				for i := len(e.Split); i < len(e.Splitord); i++ {
					e.Split = append(e.Split, diff/float64(diffn))
				}
			} else if diffn < 0 {
				http.Error(w, "Invalid Split", 400)
				return
			}

		} else {
			for len(e.Split) < len(e.Splitord) {
				e.Split = append(e.Split, 0.00)
			}
		}

	}
	for i, sp := range e.Splitord {
		if sp == e.Paid || e.Split[i] == 0 {
			continue
		}
		_, err = tx.ExecContext(r.Context(),
			"INSERT INTO transactions(id1,id2,amt1,remainingamt,expid) SELECT u.id,$1,$2,$2,$3 from users u where u.username=$4", paid_id, e.Split[i], id, sp) // id1 is the borrowe
		if err != nil {
			log.Println("Transaction updation issue for", sp, err.Error())
			http.Error(w, "Expense Creation failed", 500)
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println("Not ablt o Commit")
		http.Error(w, "InternalServer Error", 500)
		return
	}

	//

	json.NewEncoder(w).Encode(map[string]string{
		"expenseid": strconv.Itoa(id),
	})

}

func (h *Handler) EditExpense(w http.ResponseWriter, r *http.Request) {
	//var u User]
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}

	type exp struct {
		ExpenseID   int        `json:"expense_Id"`
		Name        *string    `json:"expense_name"`
		Description *string    `json:"description"`
		Totalamt    *float64   `json:"amount"`
		Date        *string    `json:"date"`
		Paid        *string    `json:"payer_username"`
		GroupName   *string    `json:"groupname"`
		Splitop     *int       `json:"splitoption"`
		Split       *[]float64 `json:"split"`
		Splitord    *[]string  `json:"splitorder"`
	}
	var e exp
	var group_id int
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&e)
	if err != nil {
		log.Println("Error in JSON", err.Error())
		http.Error(w, "Missing JSON", http.StatusBadRequest)
		return
	}
	if e.ExpenseID == 0 {
		http.Error(w, "Expense ID not Found", 400)
		return
	}
	//var flag bool
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT group_id FROM expenses WHERE id=$1", e.ExpenseID).Scan(&group_id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No such Expense", 404)
			return
		}
		http.Error(w, "internal Server Error", 500)
		return
	}
	// if !flag {
	// 	http.Error(w, "No such Expense", 404)
	// 	return
	// }
	var uid int
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}

	authorised := false

	err = h.DB.QueryRowContext(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM transactions WHERE expid=$1 AND (id1=$2 OR id2=$2))", e.ExpenseID, uid).Scan(&authorised)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	if !authorised {
		http.Error(w, "Unauthorized Expense Access", http.StatusUnauthorized)
		return
	}

	var update []string
	arg := []any{}
	i := 1
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer tx.Rollback()
	if e.Name != nil {
		if strings.TrimSpace(*e.Name) == "" {
			http.Error(w, "Wrong Name", 400)
			return
		}
		update = append(update, fmt.Sprintf("name=$%d", i))
		arg = append(arg, *e.Name)
		i++
	}
	if e.Description != nil {
		update = append(update, fmt.Sprintf("description=$%d", i))
		arg = append(arg, *e.Description)
		i++
	}
	if e.Totalamt != nil {
		if e.Split == nil || e.Splitop == nil || e.Splitord == nil {
			http.Error(w, "Split Conditionsmissing", http.StatusBadRequest)
			return

		}
		if *e.Totalamt < 0 {
			http.Error(w, "Invld Amount", 400)
			return
		}
		update = append(update, fmt.Sprintf("amount=$%d", i))
		arg = append(arg, *e.Totalamt)
		i++
	}
	if e.Date != nil {
		date, err := time.Parse("2006-01-02", *e.Date)
		if err != nil {
			http.Error(w, "Wrong Date", 400)
			return
		}
		update = append(update, fmt.Sprintf("cr_date=$%d", i))
		arg = append(arg, date)
		i++
	}
	if e.Paid != nil {
		http.Error(w, "Cannot Change Paid by", 404)
		return
		// if e.Split == nil || e.Splitop == nil || e.Splitord == nil {
		// 	http.Error(w, "Split Conditionsmissing", http.StatusBadRequest)
		// 	return
		// }
		// err = tx.QueryRowContext(r.Context(),
		// 	"SELECT id from users where username = $1", e.Paid).Scan(&paid_id)
		// if err != nil {
		// 	log.Println("Error ")
		// 	http.Error(w, err.Error(), 500)
		// 	return
		// }
		// update = append(update, fmt.Sprintf("paidby=$%d", i))
		// arg = append(arg, paid_id)
		// i++
	}

	if e.GroupName != nil {
		http.Error(w, "Cannot Change Group", 400)
		return
	}
	if len(update) == 0 {
		http.Error(w, "Nothing to Update", 400)
		return
	}
	query := fmt.Sprintf("UPDATE expenses SET %s WHERE id=$%d", strings.Join(update, ", "), i)
	arg = append(arg, e.ExpenseID)
	_, err = tx.ExecContext(r.Context(), query, arg...)
	if err != nil {
		log.Println("Error in Expense insrt")
		http.Error(w, err.Error(), 500)
		return
	}

	// SPlit logic
	if e.Totalamt != nil || e.Paid != nil { // If total amountis changed then split also must be gievn
		if *e.Splitop == 1 { // which is ratio
			var sum float64
			for _, ratio := range *e.Split {
				if ratio < 0 {
					http.Error(w, " Negative split not allowed", 400)
					return
				}
				sum += ratio
			}
			if sum == 0 {
				http.Error(w, "Invalid Ratio", 400)
				return
			}
			for i, ratio := range *e.Split {
				(*e.Split)[i] = (ratio / sum) * *e.Totalamt
				fmt.Println((*e.Split)[i], (*e.Splitord)[i])
			}
		} else {
			var sum float64
			for _, amt := range *e.Split {
				if amt < 0 {
					http.Error(w, "Negative amount not allowed", 400)
					return
				}
				sum += amt
			}
			if sum > *e.Totalamt {
				http.Error(w, "Split Amount Total Mismatch", 400)
				return
			} else if sum < *e.Totalamt {
				diff := *e.Totalamt - sum
				diffn := len(*e.Splitord) - len(*e.Split)
				if diffn > 0 {
					for i := len(*e.Split); i < len(*e.Splitord); i++ {
						*e.Split = append((*e.Split), diff/float64(diffn))
					}
				} else if diffn < 0 {
					http.Error(w, "Invalid Split", 400)
					return
				}

			} else {
				for len(*e.Split) < len(*e.Splitord) {
					*e.Split = append(*e.Split, 0.00)
				}
			}
		}
		for i, sp := range *e.Splitord {
			if sp == *e.Paid {
				continue
			}
			var borrwr_id int
			err = tx.QueryRowContext(r.Context(),
				"SELECT id FROM users WHERE username=$1", sp).Scan(&borrwr_id)
			_, err = tx.ExecContext(r.Context(),
				"UPDATE transactions SET amt1=$1, remainingamt=$1 WHERE expid=$2 AND id1=$3", (*e.Split)[i], e.ExpenseID, borrwr_id) // id1 is the borrowe
			if err != nil {
				log.Println("Transaction updation issue for", sp, err.Error())
				http.Error(w, "Expense Creation failed", 500)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}

		//
		w.WriteHeader(http.StatusAccepted)
		// json.NewEncoder(w).Encode(map[string]string{
		// 	"expenseid": strconv.Itoa(id),
		//})

	}
}

func (h *Handler) Paid(w http.ResponseWriter, r *http.Request) {
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}
	type payment struct {
		TransactionID int `json:"transactionid"`
		//ExpenseID     int     `json:"expenseid"`
		Amount float64 `json:"amount"`
	}

	var p payment
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&p)
	if err != nil {
		http.Error(w, "Misisng JSON", 400)
		return
	}
	var uid int
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	if p.Amount <= 0 {
		http.Error(w, "Wrong Amount", 400)
		return
	}
	if p.TransactionID == 0 { //p.ExpenseID == 0 &&
		http.Error(w, "Transaction/Expense ID Missing", http.StatusBadRequest)
		return
	}
	var amt float64
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	defer tx.Rollback()
	err = tx.QueryRowContext(r.Context(),
		"SELECT remainingamt FROM transactions WHERE id=$1 AND (id1=$2 OR id2=$2) FOR UPDATE", p.TransactionID, uid).Scan(&amt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No such Eexpense", http.StatusNotFound)
			return
		}
		http.Error(w, "internal Server Issue", 500)
		return
	}
	if math.Abs(amt-p.Amount) > 0.001 {
		if amt < p.Amount {
			http.Error(w, "Amount exceeds Balance", 400) //Maybe Instead can create a new expense with the same ID but revser borrower and Giver Ig
			return
		}
		_, err := tx.ExecContext(r.Context(),
			"UPDATE transactions SET remainingamt=remainingamt-$1 WHERE id=$2 AND  (id1=$3 OR id2=$3)", p.Amount, p.TransactionID, uid)

		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return

		}
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"Message": "Expense Partially Paid",
			"Amount":  fmt.Sprintf("%.2f", amt-p.Amount),
		})
	} else {
		_, err := tx.ExecContext(r.Context(),
			"UPDATE transactions SET remainingamt=0.00 WHERE id=$1 and (id1=$2 OR ID2=$2)", p.TransactionID, uid)

		if err != nil {
			http.Error(w, "Payment Updation Failed", 500)
			return
		}
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"Message": "Expense Fully Paid",
			"Amount":  "0.00",
		})
	}

}

func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}
	type expense struct {
		ExpID int `json:"expense_id"`
	}
	var e expense
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&e)
	if err != nil {
		http.Error(w, "Misisng JSON", 400)
		return
	}
	var uid int
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	var paidID int
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT paidby FROM expenses WHERE id=$1", e.ExpID).Scan(&paidID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No such Expense", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", 500)
		return
	}
	if paidID != uid {
		http.Error(w, "Only Expense Payer can Delete", http.StatusUnauthorized)
		return
	}
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer tx.Rollback()
	var count int //Partially Paid Rows
	err = tx.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM transactions WHERE expid=$1 AND remainingamt!=amt1", e.ExpID).Scan(&count)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	if count == 0 {
		_, err := tx.ExecContext(r.Context(),
			"DELETE FROM transactions where expid=$1", e.ExpID)
		//rows, err := res.RowsAffected() //INVALID EXPENSE IS ALREADY VALIDATED
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		_, err = tx.ExecContext(r.Context(),
			"DELETE from expenses WHERE id=$1", e.ExpID)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Deleted Expenses.Balances Restored.",
		})
		return

	}

	http.Error(w, "Cannot delete partially Settled Payments due to Payment Integrity", http.StatusMethodNotAllowed)
}

func (h *Handler) FetchExpenses(w http.ResponseWriter, r *http.Request) {
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}
	type resp struct {
		ExpenseID     int     `json:"expenseid"`
		TransactionID int     `json:"transactionid"`
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		Totalamt      float64 `json:"amount"`
		Remainingamt  float64 `json:"remaining"`
		Date          string  `json:"date"`
		Paid          string  `json:"payer_username"`
		GroupName     string  `json:"groupname"`
	}
	var r1 []resp
	var uid int
	err = h.DB.QueryRowContext(r.Context(),
		"SELECT id from users where username = $1", username).Scan(&uid)
	if err != nil {
		log.Println("Error ")
		http.Error(w, err.Error(), 500)
		return
	}
	var crdate time.Time
	status := r.URL.Query().Get("status")
	query := "SELECT e.id, e.name, e.description, e.cr_date, t.amt1,t.remainingamt, u.name, g.name, t.id from expenses e JOIN transactions t ON t.expid=e.id JOIN groups g ON g.id=e.group_id JOIN users u ON u.id =e.paidby WHERE "
	mode := r.URL.Query().Get("mode")
	switch mode {
	case "borrowed":
		query = query + "t.id1=$1"
	case "Paid":
		query = query + "t.id2=$1"
	default:
		query += "(t.id1=$1 OR t.id2=$1)"
	}
	switch status {
	case "pending":
		query = query + " AND t.remainingamt>0 ORDER BY e.cr_date DESC"
	case "paid":
		query += " AND t.remainingamt=0 ORDER BY e.cr_date DESC"
	default:
		query += " ORDER BY e.cr_date DESC"
	}
	rows, err := h.DB.QueryContext(r.Context(),
		query, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No Expenses Found", 404)
			return
		}
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tmp resp
		err = rows.Scan(&tmp.ExpenseID, &tmp.Name, &tmp.Description, &crdate, &tmp.Totalamt, &tmp.Remainingamt, &tmp.Paid, &tmp.GroupName, &tmp.TransactionID)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		tmp.Date = crdate.Format("2006-01-02")
		r1 = append(r1, tmp)

	}

	json.NewEncoder(w).Encode(r1)

}

func (h *Handler) Creategroup(w http.ResponseWriter, r *http.Request) {
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}

	type group struct {
		Name    string   `json:"group_name"`
		Members []string `json:"member_usernames"`
	}
	var g1 group
	var g_id int
	var m_ids []int
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&g1)
	if err != nil {
		http.Error(w, "Missing JSON", http.StatusBadRequest)
		return
	}

	if g1.Name == "" || len(g1.Members) < 1 {
		log.Println("Missing Group Info")
		http.Error(w, "Missing Group Information", http.StatusBadRequest)
		return
	} else if len(g1.Members) > 100 {
		http.Error(w, "User limit is 100", 400)
		return
		//g1.Members = g1.Members[:100]
	}
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer tx.Rollback()
	g1.Members = append(g1.Members, username)
	for _, m := range g1.Members {
		var tmp int
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		err = tx.QueryRowContext(r.Context(),
			"SELECT id from users where username = $1", m).Scan(&tmp)
		if err != nil {
			//http.Error(w, "No such User"+m, 400)
			continue
		}
		m_ids = append(m_ids, tmp)
	}

	err = tx.QueryRowContext(r.Context(),
		"INSERT into groups(name,members, created_by, cr_date) VALUES($1, $2, $3, CURRENT_DATE) RETURNING id", g1.Name, m_ids, m_ids[len(m_ids)-1]).Scan(&g_id)
	if err != nil {
		log.Println("Group creation Issue", err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	for _, id := range m_ids {
		res, err := tx.ExecContext(r.Context(),
			"UPDATE users SET grps=grps+1 WHERE id=$1", id)
		rows, err := res.RowsAffected()
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		if int(rows) != 1 {
			http.Error(w, "Group Mismatch", 400)
			return

		}
	}
	json.NewEncoder(w).Encode(map[string]string{
		"group_id": strconv.Itoa(g_id),
	})
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}
	var uid int
	err = h.DB.QueryRowContext(r.Context(), "SELECT id from users where username=$1", username).Scan(&uid)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	type transc struct {
		Oweto   map[string]float64 `json:"oweto"`   // You owe to somoeone
		Owefrom map[string]float64 `json:"owefrom"` // You owe fromsomeone or someone owes t you
	}
	type Response struct {
		Individual transc            `json:"individual"`
		GroupView  map[string]transc `json:"group"`
	}
	var t Response
	t.Individual.Owefrom = make(map[string]float64)
	t.Individual.Oweto = make(map[string]float64)
	t.GroupView = make(map[string]transc)
	rows1, err := h.DB.QueryContext(r.Context(),
		"SELECT u.name, t.remainingamt,  g.name from transactions t join users u on u.id=t.id2 join expenses e ON e.id=t.expid join groups g ON g.id=e.group_id where t.id1=$1", uid)
	if err != nil {
		log.Println("Transaction retrival thing", err.Error())
		http.Error(w, "Internal Server Erro", 500)
		return
	}
	defer rows1.Close()
	for rows1.Next() {
		var name, gname string
		var amt float64
		//var tid int
		err := rows1.Scan(&name, &amt, &gname)
		if err != nil {
			log.Println("Not adding to mp", err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}
		t.Individual.Oweto[name] += amt
		if _, exists := t.GroupView[gname]; !exists {
			t.GroupView[gname] = transc{
				Owefrom: make(map[string]float64),
				Oweto:   make(map[string]float64)}
		}
		t.GroupView[gname].Oweto[name] += amt
	}
	fmt.Println(t.Individual.Owefrom)
	rows2, err := h.DB.QueryContext(r.Context(),
		"SELECT u.name, t.remainingamt, g.name from transactions t join users u on u.id=t.id1 join expenses e ON e.id=t.expid join groups g ON g.id=e.group_id where t.id2=$1", uid)
	if err != nil {
		log.Println("Transaction retrival thing", err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var name, gname string
		var amt float64
		//var tid int
		err := rows2.Scan(&name, &amt, &gname)
		if err != nil {
			log.Println("Not adding to mp", err.Error())
			http.Error(w, "internal Server Error", 500)
			return
		}
		t.Individual.Owefrom[name] += amt
		if _, exists := t.GroupView[gname]; !exists {
			t.GroupView[gname] = transc{
				Owefrom: make(map[string]float64),
				Oweto:   make(map[string]float64)}
		}
		t.GroupView[gname].Owefrom[name] += amt
	}
	if r.URL.Query().Get("view") != "raw" {
		for person, amt := range t.Individual.Owefrom {
			if t.Individual.Oweto[person] > amt {
				t.Individual.Oweto[person] -= amt
				delete(t.Individual.Owefrom, person)
			} else if t.Individual.Oweto[person] < amt {
				t.Individual.Owefrom[person] -= t.Individual.Oweto[person]
				delete(t.Individual.Oweto, person)

			} else {
				delete(t.Individual.Owefrom, person)
				delete(t.Individual.Oweto, person)
			}

		}
		for gname, tc := range t.GroupView {
			for person, amt := range tc.Owefrom {
				if tc.Oweto[person] > amt {
					tc.Oweto[person] -= amt
					delete(tc.Owefrom, person)
				} else if tc.Oweto[person] < amt {
					tc.Owefrom[person] -= tc.Oweto[person]
					delete(tc.Oweto, person)

				} else {
					delete(tc.Owefrom, person)
					delete(tc.Oweto, person)
				}

			}
			t.GroupView[gname] = tc
		}
	}

	json.NewEncoder(w).Encode(t)

}

func (h *Handler) Deleteuser(w http.ResponseWriter, r *http.Request) {
	//var u User

	_, username, err := Verifyjwtoken(r) // request contains the token in headers
	if err != nil {
		http.Error(w, "Wrong Token", http.StatusBadRequest)
		return
	}
	var uid int
	err = h.DB.QueryRowContext(r.Context(), "SELECT id from users where username=$1", username).Scan(&uid)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// err := json.NewDecoder(r.Body).Decode(&u)
	// if u.User_id == 0 {
	// 	log.Println("User attempting to delete without id")
	// 	http.Error(w, "User ID not provided", http.StatusBadRequest)
	// 	return
	// }
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	var checkbor, checkgiv int
	err = tx.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM transactions WHERE remainingamt>0 AND id1=$1", uid).Scan(&checkbor)
	if checkbor > 0 {
		http.Error(w, "Cannot Delete with Unpaid Balance", http.StatusMethodNotAllowed)
		return
	}
	err = tx.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM transactions WHERE remainingamt>0 AND id2=$1", uid).Scan(&checkgiv)
	if checkgiv > 0 {
		http.Error(w, "Cannot Delete with Unsettled Balance", http.StatusMethodNotAllowed)
		return
	}
	_, err = tx.ExecContext(r.Context(), "DELETE from friends where (user_id1=$1 OR user_id2=$1)", uid)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, err = tx.ExecContext(r.Context(),
		"UPDATE groups SET members=array_remove(members,$1) WHERE $1=ANY(members)", uid)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, err = tx.ExecContext(r.Context(), "DELETE from users where id=$1", uid)
	if err != nil {
		log.Println("Deletion for", username, " Error")
		http.Error(w, "Deletion error", 500)
		log.Println("Error being : ", err)
		return
	}
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "User deleted",
		"username": username,
	})

}
