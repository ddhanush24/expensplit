package routes

import (
	"expspl/handle"
	"net/http"
)

func SetupRoutes(h *handle.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /Signup", h.Createuser) // Works
	mux.HandleFunc("POST /Signin", h.Signin)     //Works and returns JWT

	mux.HandleFunc("GET /user", h.Middleware(h.UserProfile)) //Works
	mux.HandleFunc("DELETE /user", h.Middleware(h.Deleteuser))
	mux.HandleFunc("GET /dashbd", h.Middleware(h.Dashboard))    //Works
	mux.HandleFunc("GET /friendlist", h.Middleware(h.Findfrnd)) //Works
	mux.HandleFunc("POST /friend", h.Middleware(h.AddFriend))   // Works
	mux.HandleFunc("GET /friendreqs", h.Middleware(h.FriendReqs))
	mux.HandleFunc("DELETE /friend", h.Middleware(h.RemoveFriend))
	mux.HandleFunc("GET /xpense", h.Middleware(h.FetchExpenses))
	mux.HandleFunc("POST /xpense", h.Middleware(h.AddExpense)) //Works with split
	mux.HandleFunc("PATCH /xpense", h.Middleware(h.EditExpense))
	mux.HandleFunc("DELETE /xpense", h.Middleware(h.DeleteExpense))
	mux.HandleFunc("PATCH /payments", h.Middleware(h.Paid))
	mux.HandleFunc("POST /Creategroup", h.Middleware(h.Creategroup)) // Works

	return mux
}
