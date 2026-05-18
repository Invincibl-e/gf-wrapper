package render

// Resolver resolves placeholders for one source, such as env, file, or docker-secret.
type Resolver interface {
	Source() string
	Resolve(key string) (value string, ok bool, err error)
}
