package model

type Controller interface {
	Run(<-chan struct{})
	HasSynced() bool
}
