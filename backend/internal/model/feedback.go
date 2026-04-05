package model

import "time"

type FeedbackComment struct {
	CommentID     int64     `json:"comment_id"`
	TargetType    string    `json:"target_type"`
	TargetID      string    `json:"target_id"`
	UserID        int64     `json:"user_id"`
	AuthorLoginID string    `json:"author_login_id"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
}

type FeedbackSummary struct {
	TargetType string            `json:"target_type"`
	TargetID   string            `json:"target_id"`
	UpVotes    int64             `json:"up_votes"`
	DownVotes  int64             `json:"down_votes"`
	MyVote     string            `json:"my_vote,omitempty"` // up | down | ""
	Comments   []FeedbackComment `json:"comments"`
}

type CreateFeedbackCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type UpdateFeedbackCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type VoteFeedbackRequest struct {
	// up | down | none (none이면 기존 투표 취소)
	VoteType string `json:"vote_type" binding:"required"`
}

type VoteFeedbackResponse struct {
	Status string `json:"status"`
}

type FeedbackCommentMutationResponse struct {
	Status    string `json:"status"`
	CommentID int64  `json:"comment_id"`
}
