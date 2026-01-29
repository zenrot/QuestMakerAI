package api

type QuestionsRequest struct {
	Topic          string `json:"topic"`
	QuestionNumber int    `json:"question_number"`
}
