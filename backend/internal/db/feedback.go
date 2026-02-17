package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/kube-rca/backend/internal/model"
)

func normalizeTargetType(targetType string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(targetType))
	if normalized != "incident" && normalized != "alert" {
		return "", fmt.Errorf("invalid target_type: %s", targetType)
	}
	return normalized, nil
}

func normalizeVoteType(voteType string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(voteType))
	if normalized != "up" && normalized != "down" && normalized != "none" {
		return "", fmt.Errorf("invalid vote_type: %s", voteType)
	}
	return normalized, nil
}

func (db *Postgres) EnsureFeedbackSchema() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS feedback_votes (
			id BIGSERIAL PRIMARY KEY,
			target_type TEXT NOT NULL CHECK (target_type IN ('incident', 'alert')),
			target_id TEXT NOT NULL,
			user_id BIGINT NOT NULL,
			vote_type TEXT NOT NULL CHECK (vote_type IN ('up', 'down')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(target_type, target_id, user_id)
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS feedback_comments (
			comment_id BIGSERIAL PRIMARY KEY,
			target_type TEXT NOT NULL CHECK (target_type IN ('incident', 'alert')),
			target_id TEXT NOT NULL,
			user_id BIGINT NOT NULL,
			author_login_id TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`CREATE INDEX IF NOT EXISTS feedback_votes_target_idx ON feedback_votes(target_type, target_id)`,
		`CREATE INDEX IF NOT EXISTS feedback_comments_target_idx ON feedback_comments(target_type, target_id, created_at)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(context.Background(), query); err != nil {
			return err
		}
	}
	return nil
}

func (db *Postgres) UpsertVote(targetType, targetID string, userID int64, voteType string) error {
	normalizedTargetType, err := normalizeTargetType(targetType)
	if err != nil {
		return err
	}
	normalizedVoteType, err := normalizeVoteType(voteType)
	if err != nil {
		return err
	}

	if normalizedVoteType == "none" {
		_, err := db.Pool.Exec(context.Background(),
			`DELETE FROM feedback_votes WHERE target_type = $1 AND target_id = $2 AND user_id = $3`,
			normalizedTargetType, targetID, userID,
		)
		return err
	}

	_, err = db.Pool.Exec(context.Background(), `
		INSERT INTO feedback_votes (target_type, target_id, user_id, vote_type)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (target_type, target_id, user_id)
		DO UPDATE SET vote_type = EXCLUDED.vote_type, updated_at = NOW()
	`, normalizedTargetType, targetID, userID, normalizedVoteType)
	return err
}

func (db *Postgres) CreateComment(targetType, targetID string, userID int64, authorLoginID, body string) (*model.FeedbackComment, error) {
	normalizedTargetType, err := normalizeTargetType(targetType)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO feedback_comments (target_type, target_id, user_id, author_login_id, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING comment_id, target_type, target_id, user_id, author_login_id, body, created_at
	`

	var comment model.FeedbackComment
	err = db.Pool.QueryRow(context.Background(), query, normalizedTargetType, targetID, userID, authorLoginID, body).Scan(
		&comment.CommentID,
		&comment.TargetType,
		&comment.TargetID,
		&comment.UserID,
		&comment.AuthorLoginID,
		&comment.Body,
		&comment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (db *Postgres) UpdateComment(targetType, targetID string, commentID, userID int64, body string) (*model.FeedbackComment, error) {
	normalizedTargetType, err := normalizeTargetType(targetType)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE feedback_comments
		SET body = $5, updated_at = NOW()
		WHERE comment_id = $1 AND target_type = $2 AND target_id = $3 AND user_id = $4
		RETURNING comment_id, target_type, target_id, user_id, author_login_id, body, created_at
	`

	var comment model.FeedbackComment
	err = db.Pool.QueryRow(context.Background(), query, commentID, normalizedTargetType, targetID, userID, body).Scan(
		&comment.CommentID,
		&comment.TargetType,
		&comment.TargetID,
		&comment.UserID,
		&comment.AuthorLoginID,
		&comment.Body,
		&comment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("comment not found or no permission")
		}
		return nil, err
	}
	return &comment, nil
}

func (db *Postgres) DeleteComment(targetType, targetID string, commentID, userID int64) error {
	normalizedTargetType, err := normalizeTargetType(targetType)
	if err != nil {
		return err
	}

	result, err := db.Pool.Exec(context.Background(), `
		DELETE FROM feedback_comments
		WHERE comment_id = $1 AND target_type = $2 AND target_id = $3 AND user_id = $4
	`, commentID, normalizedTargetType, targetID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found or no permission")
	}
	return nil
}

func (db *Postgres) GetComments(targetType, targetID string, limit int32) ([]model.FeedbackComment, error) {
	normalizedTargetType, err := normalizeTargetType(targetType)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}

	rows, err := db.Pool.Query(context.Background(), `
		SELECT comment_id, target_type, target_id, user_id, author_login_id, body, created_at
		FROM feedback_comments
		WHERE target_type = $1 AND target_id = $2
		ORDER BY created_at ASC
		LIMIT $3
	`, normalizedTargetType, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]model.FeedbackComment, 0)
	for rows.Next() {
		var c model.FeedbackComment
		if err := rows.Scan(
			&c.CommentID,
			&c.TargetType,
			&c.TargetID,
			&c.UserID,
			&c.AuthorLoginID,
			&c.Body,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (db *Postgres) GetVoteSummary(targetType, targetID string, userID int64) (upVotes int64, downVotes int64, myVote string, err error) {
	normalizedTargetType, err := normalizeTargetType(targetType)
	if err != nil {
		return 0, 0, "", err
	}

	err = db.Pool.QueryRow(context.Background(), `
		SELECT
			COALESCE(SUM(CASE WHEN vote_type = 'up' THEN 1 ELSE 0 END), 0) AS up_votes,
			COALESCE(SUM(CASE WHEN vote_type = 'down' THEN 1 ELSE 0 END), 0) AS down_votes
		FROM feedback_votes
		WHERE target_type = $1 AND target_id = $2
	`, normalizedTargetType, targetID).Scan(&upVotes, &downVotes)
	if err != nil {
		return 0, 0, "", err
	}

	var vote string
	err = db.Pool.QueryRow(context.Background(), `
		SELECT vote_type
		FROM feedback_votes
		WHERE target_type = $1 AND target_id = $2 AND user_id = $3
	`, normalizedTargetType, targetID, userID).Scan(&vote)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return upVotes, downVotes, "", nil
		}
		return 0, 0, "", err
	}

	if vote != "" {
		myVote = vote
	}
	return upVotes, downVotes, myVote, nil
}

func (db *Postgres) GetFeedbackSummary(targetType, targetID string, userID int64) (*model.FeedbackSummary, error) {
	upVotes, downVotes, myVote, err := db.GetVoteSummary(targetType, targetID, userID)
	if err != nil {
		return nil, err
	}

	comments, err := db.GetComments(targetType, targetID, 200)
	if err != nil {
		return nil, err
	}

	normalizedTargetType, err := normalizeTargetType(targetType)
	if err != nil {
		return nil, err
	}

	return &model.FeedbackSummary{
		TargetType: normalizedTargetType,
		TargetID:   targetID,
		UpVotes:    upVotes,
		DownVotes:  downVotes,
		MyVote:     myVote,
		Comments:   comments,
	}, nil
}
