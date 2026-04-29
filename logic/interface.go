package logic

type Attribute interface {
	Active(data map[string]string) error
}
