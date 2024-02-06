package digiconfig

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Setter provides an interface to set config items
//
//counterfeiter:generate . Setter
type Setter interface {
	// Set should set an item into persistent config store.
	Set(key, value string)
}

type SetterFunc func(key, value string)

func (f SetterFunc) Set(key, value string) {
	f(key, value)
}

// Getter provides an interface to get config items
//
//counterfeiter:generate . Getter
type Getter interface {
	// Get should get an item with the key passed in and return
	// the value. If the item is found then it should return true,
	// otherwise false.
	Get(key string) (value string, ok bool)
}

type GetterFunc func(key string) (value string, ok bool)

func (f GetterFunc) Get(key string) (string, bool) {
	return f(key)
}
