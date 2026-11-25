package v1

import "net/http"

func (h *PostHandler) Upvote(w http.ResponseWriter, r *http.Request)   { h.vote(w, r, 1) }
func (h *PostHandler) Downvote(w http.ResponseWriter, r *http.Request) { h.vote(w, r, -1) }
func (h *PostHandler) Unvote(w http.ResponseWriter, r *http.Request)   { h.vote(w, r, 0) }

func (h *PostHandler) vote(w http.ResponseWriter, r *http.Request, value int) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	post, err := h.uc.Vote(r.Context(), user, r.PathValue("postId"), value)
	if err != nil {
		fail(w, "vote", err)
		return
	}
	writeJSON(w, http.StatusOK, post)
}
