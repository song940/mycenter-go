package models

import (
	"database/sql"
	"time"
)

type Post struct {
	Id        int
	UserId    int
	Content   string
	CreatedAt time.Time
}

func CreatePost(db *sql.DB, userId int, content string) (post *Post, err error) {
	post = &Post{
		UserId:  userId,
		Content: content,
	}
	sql := `INSERT INTO posts (user_id, content) VALUES (?, ?) RETURNING id, created_at`
	err = db.QueryRow(sql, post.UserId, post.Content).Scan(&post.Id, &post.CreatedAt)
	return
}

func GetPosts(db *sql.DB, userId int) (posts []Post, err error) {
	sql := `SELECT id, user_id, content, created_at FROM posts WHERE user_id = ?`
	rows, err := db.Query(sql, userId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := Post{}
		err = rows.Scan(&post.Id, &post.UserId, &post.Content, &post.CreatedAt)
		if err != nil {
			return
		}
		posts = append(posts, post)
	}
	return
}
