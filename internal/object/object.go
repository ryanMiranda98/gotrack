package object

// Object - blob | tree | commit | tag

type Object interface {
	Serialize() string
	Deserialize(data string) error
}
