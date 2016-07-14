package source

type VideoSource interface {
	Open() error
	Close()
	Output() chan []byte
}
