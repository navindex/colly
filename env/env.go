package env

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// ------------------------------------------------------------------------

type environment struct {
	prefix string
	values map[string]string
	dict   map[string]string
}

// ------------------------------------------------------------------------

// NewFromMap returns a pointer to a newly created environment structure.
// It is based on a map where the keys will be filtered by a prefix.
// An optional dictionary can be given to convert the keys.
func NewFromMap(prefix string, values map[string]string, dict map[string]string) *environment {
	env := &environment{
		prefix: prefix,
		values: map[string]string{},
	}

	env.SetDictionary(dict)

	skip := len(env.prefix)

	for k, v := range values {
		if !strings.HasPrefix(k, env.prefix) {
			continue
		}

		key := k[skip:]
		if _, present := dict[key]; present {
			key = dict[key]
		}

		env.values[key] = v
	}

	return env
}

// ------------------------------------------------------------------------

// NewFromOSEnv returns a pointer to a newly created environment structure.
// It is based on the OS environment settings where the keys will be filtered by a prefix.
// An optional dictionary can be given to convert the keys.
func NewFromOSEnv(prefix string, dict map[string]string) *environment {
	values := map[string]string{}

	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, prefix) {
			continue
		}

		pair := strings.SplitN(v, "=", 2)

		values[pair[0]] = pair[1]
	}

	return NewFromMap(prefix, values, dict)
}

// ------------------------------------------------------------------------

// NewFromOSEnv returns a pointer to a newly created environment structure.
// It is based on a content of an (tipycally .env) file where the keys will be filtered by a prefix.
// An optional dictionary can be given to convert the keys.
func NewFromFile(prefix string, path string, dict map[string]string) (*environment, error) {
	values, err := godotenv.Read(path)
	if err != nil {
		return nil, err
	}

	return NewFromMap(prefix, values, dict), nil
}

// ------------------------------------------------------------------------

// Set sets a value named by the key. It overrides any existing value stored with the same key.
// Set does not check for the prefix.
func (e *environment) Set(key string, value string) {
	if _, present := e.dict[key]; present {
		key = e.dict[key]
	}

	e.values[key] = value
}

// ------------------------------------------------------------------------

// SetPrefixed sets a value named by the key if the key starts with the prefix.
// It overrides any existing value stored with the same key.
func (e *environment) SetPrefixed(key, value string) {
	if !strings.HasPrefix(value, e.prefix) {
		return
	}

	e.Set(key[len(e.prefix):], value)
}

// ------------------------------------------------------------------------

// Unset unsets a value named by the key.
func (e *environment) Unset(key string) {
	delete(e.values, key)
}

// ------------------------------------------------------------------------

// SetDictionary sets the dictionary that will be used to convert the keys.
func (e *environment) SetDictionary(dict map[string]string) {
	if dict == nil {
		dict = map[string]string{}
	}

	e.dict = dict
}

// ------------------------------------------------------------------------

// SetPrefix sets the prefix that will be used to check the keys in SetPrefixed method.
func (e *environment) SetPrefix(prefix string) {
	e.prefix = prefix
}

// ------------------------------------------------------------------------

// Values returns the key/value pairs stored in the environment structure.
func (e *environment) Values() map[string]string {
	return e.values
}
