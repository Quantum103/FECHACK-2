package models

type TemplateData struct {
	StudentsWithTopics    []StudentWithTopic
	StudentsWithoutTopics []User
	FreeTopics            []Topic
}

type StudentWithTopic struct {
	User
	Topic Topic
}
