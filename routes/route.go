package routes

import (
	"expspl/handle"
	"net/http"
)

func SetupRoutes(h *handle.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /Signup", h.Createuser) // Works
	mux.HandleFunc("POST /Signin", h.Signin)     //Works and returns JWT

	mux.HandleFunc("GET /user", h.UserProfile) //Works
	mux.HandleFunc("DELETE /user", h.Deleteuser)
	mux.HandleFunc("GET /dashbd", h.Dashboard)    //Works
	mux.HandleFunc("GET /friendlist", h.Findfrnd) //Works
	mux.HandleFunc("POST /friend", h.AddFriend)   // Works
	mux.HandleFunc("GET /friendreqs", h.FriendReqs)
	mux.HandleFunc("DELETE /friend", h.RemoveFriend)
	mux.HandleFunc("GET /xpense", h.FetchExpenses)
	mux.HandleFunc("POST /xpense", h.AddExpense) //Works with split
	mux.HandleFunc("PATCH /xpense", h.EditExpense)
	mux.HandleFunc("DELETE /xpense", h.DeleteExpense)
	mux.HandleFunc("PATCH /payments", h.Paid)
	mux.HandleFunc("POST /Creategroup", h.Creategroup) // Works

	return mux
}
