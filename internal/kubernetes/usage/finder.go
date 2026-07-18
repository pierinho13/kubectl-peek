package usage

import "context"

type Finder interface {
	Find(
		ctx context.Context,
		namespace string,
		secretName string,
	) Result
}
