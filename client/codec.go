package client

type Codec interface {
    MarshalJSON() ([]byte, error)
    UnmarshalJSON([]byte) error
}
