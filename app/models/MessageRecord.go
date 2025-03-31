package models

type MessageRecord struct {
	ID          int    `db:"id"`
	Content     string `db:"content"`
	ConsumerID  string `db:"consumer_id"`
	ProcessedAt string `db:"processed_at"`
}
