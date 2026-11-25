package domain

import "slices"

const (
	PostTypeText = "text"
	PostTypeLink = "link"
)

type Post struct {
	Id       string    `json:"id"`
	Title    string    `json:"title"`
	Category string    `json:"category"`
	Author   User      `json:"author"`
	Score    int       `json:"score"`
	Votes    []Vote    `json:"votes"`
	Comments []Comment `json:"comments"`

	Created          string `json:"created"`
	Views            int    `json:"views"`
	UpvotePercentage int    `json:"upvotePercentage"`

	Type string  `json:"type"`
	Text *string `json:"text,omitempty"`
	Url  *string `json:"url,omitempty"`
}

func (p Post) Clone() Post {
	p.Votes = append(make([]Vote, 0, len(p.Votes)), p.Votes...)
	p.Comments = append(make([]Comment, 0, len(p.Comments)), p.Comments...)
	return p
}

func (p *Post) SetVote(userID string, value int) {
	for i := range p.Votes {
		if p.Votes[i].User == userID {
			p.Votes[i].Vote = value
			p.recount()
			return
		}
	}
	p.Votes = append(p.Votes, Vote{User: userID, Vote: value})
	p.recount()
}

func (p *Post) RemoveVote(userID string) {
	p.Votes = slices.DeleteFunc(p.Votes, func(v Vote) bool { return v.User == userID })
	p.recount()
}

func (p *Post) recount() {
	score, ups := 0, 0
	for _, v := range p.Votes {
		score += v.Vote
		if v.Vote > 0 {
			ups++
		}
	}
	p.Score = score
	if len(p.Votes) == 0 {
		p.UpvotePercentage = 0
		return
	}
	p.UpvotePercentage = 100 * ups / len(p.Votes)
}

func (p *Post) AddComment(c Comment) {
	p.Comments = append(p.Comments, c)
}

func (p *Post) CommentByID(id string) (Comment, bool) {
	for _, c := range p.Comments {
		if c.Id == id {
			return c, true
		}
	}
	return Comment{}, false
}

func (p *Post) RemoveComment(id string) bool {
	n := len(p.Comments)
	p.Comments = slices.DeleteFunc(p.Comments, func(c Comment) bool { return c.Id == id })
	return len(p.Comments) != n
}

func (p Post) CanBeDeletedBy(u User) bool {
	return u.Admin || p.Author.Id == u.Id
}

func (p Post) CanDeleteComment(c Comment, u User) bool {
	return u.Admin || c.Author.Id == u.Id || p.Author.Id == u.Id
}
