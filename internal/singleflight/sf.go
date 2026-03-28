package sf

import (
	"context"

	"golang.org/x/sync/singleflight"
)

// Group wraps singleflight.Group to deduplicate concurrent Redis reads
type Group struct {
	g singleflight.Group
}

/*

internally:
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

- mu → protects map
- m  → key → in-flight call
- call → holds result + waitgroup

*/

// Do ensures only one fn executes per key at a time.
// Concurrent callers with the same key wait and share the result.
// Returns early if ctx is cancelled (e.g. client disconnected).
func (g *Group) Do(ctx context.Context, key string, fn func() (any, error)) (any, error) {
	ch := g.g.DoChan(key, fn)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.Val, res.Err
	}
}
