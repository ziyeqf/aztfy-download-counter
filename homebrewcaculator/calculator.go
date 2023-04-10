package homebrewcaculator

import (
	"context"
	"errors"
	"log"
)

type Span int

type calcFunc func(client DatabaseClient) ([]calcFunc, error)

type Calculator struct {
	Logger *log.Logger
	spans  []Span
	db     *DatabaseClient
}

func NewCalculator(spans []Span, db *DatabaseClient, logger *log.Logger) Calculator {
	return Calculator{spans: spans, db: db, Logger: logger}
}

func (c Calculator) Calc(ctx context.Context, idx int) error {
	q := queue{}
	for _, span := range c.spans {
		q.enqueue(c.calcCountRight(ctx, idx, span))
		q.enqueue(c.calcTotalCountRight(ctx, idx, span))
		q.enqueue(c.calcTotalCountLeft(ctx, idx-1, span))
		q.enqueue(c.calcCountLeft(ctx, idx-int(span), span))
	}

	var errList error
	for !q.isEmpty() {
		calcFunc := q.dequeue()
		if calcFunc != nil {
			newFuncs, err := calcFunc(*c.db)
			if err != nil {
				errList = errors.Join(err)
				continue
			}
			q.enqueue(newFuncs...)
		}
	}

	return errList
}

// the core algorithm is: ed(n) = xd(n) - xd(n-1) + ed(n-x)
// ed(n) is the count at index n
// xd(n) is the total count of past `x` at index n
// So we can calculate one unknown variables from the other three known variables.
// ┌───────┬───┬───┬───┬───┐
// │days(n)│n-x│...│n-1│ n │
// ├───────┼───┼───┼───┼───┤
// │ed(n)  │ a │   │   │ b │
// ├───────┼───┼───┼───┼───┤
// │xd(n)  │   │   │ c │ d │
// └───────┴───┴───┴───┴───┘
// to make things more understandable, we call `a` as `Count Left`, b as `Count Right`, c as `Total Count Left`, d as `Total Count Right`.
// when we get a new `a`, it might become a new `b` or a new `a` with other `x`(spans).
// so we put all possible pattern into the queue, try to calculate the unknown variables.
// same for the `b` and `c` and `d`.
// a new `a` -> a new `b` or `a` with different `x`
// a new `b` -> a new `a` or `b` with different `x`
// a new `c` -> a new `d` or `c` with different `x`
// a nwe `d` -> a new `c` or `d` with different `x`

// ┌───────┬───┬───┬───┬───┐
// │days(n)│n-x│...│n-1│ n │
// ├───────┼───┼───┼───┼───┤
// │ed(n)  │ k │   │   │ ? │
// ├───────┼───┼───┼───┼───┤
// │xd(n)  │   │   │ k │ k │
// └───────┴───┴───┴───┴───┘
// `k` means known.
// ed(n) = xd(n) - xd(n-1) + ed(n-x)
func (c Calculator) calcCountRight(ctx context.Context, idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(ctx, idx)
		if err != nil {
			return nil, err
		}
		// Nothing to calc if already known
		if this.Count != -1 {
			return nil, nil
		}

		prev, err := client.Get(ctx, idx-1)
		if err != nil {
			return nil, err
		}
		prevSpan, err := client.Get(ctx, idx-int(span))
		if err != nil {
			return nil, err
		}

		xdN, ok1 := this.TotalCounts[span]
		xdNPrev, ok2 := prev.TotalCounts[span]
		ok3 := prevSpan.Count != -1

		if !ok1 || !ok2 || !ok3 {
			return nil, nil
		}

		this.Count = xdN - xdNPrev + prevSpan.Count
		if err = client.Set(ctx, idx, this); err != nil {
			return nil, err
		}

		return c.postCalcCount(ctx, idx), nil
	}
}

// ┌───────┬───┬───┬───┬───┐
// │days(n)│n-x│...│n-1│ n │
// ├───────┼───┼───┼───┼───┤
// │ed(n)  │ k │   │   │ k │
// ├───────┼───┼───┼───┼───┤
// │xd(n)  │   │   │ k │ ? │
// └───────┴───┴───┴───┴───┘
// `k` means known.
// xd(n) = ed(n) + xd(n-1) - ed(n-x)
func (c Calculator) calcTotalCountRight(ctx context.Context, idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(ctx, idx)
		if err != nil {
			return nil, err
		}

		// Nothing to calc if already known
		if _, ok := this.TotalCounts[span]; ok {
			return nil, nil
		}

		prev, err := client.Get(ctx, idx-1)
		if err != nil {
			return nil, err
		}
		prevSpan, err := client.Get(ctx, idx-int(span))
		if err != nil {
			return nil, err
		}

		ok1 := this.Count != -1
		_, ok2 := prev.TotalCounts[span]
		ok3 := prevSpan.Count != -1

		if !ok1 || !ok2 || !ok3 {
			return nil, nil
		}

		this.TotalCounts[span] = this.Count + prev.TotalCounts[span] - prevSpan.Count
		if err = client.Set(ctx, idx, this); err != nil {
			return nil, err
		}

		return c.postCalcTotalCount(ctx, idx), nil
	}
}

// ┌───────┬───┬───┬───┬───┐
// │days(n)│n-x│...│n-1│ n │
// ├───────┼───┼───┼───┼───┤
// │ed(n)  │ ? │   │   │ k │
// ├───────┼───┼───┼───┼───┤
// │xd(n)  │   │   │ k │ k │
// └───────┴───┴───┴───┴───┘
// `k` means known.
// d(n-x) = xd(n-1) - xd(n) + ed(n)
func (c Calculator) calcCountLeft(ctx context.Context, idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(ctx, idx)
		if err != nil {
			return nil, err
		}
		// Nothing to calc if already known
		if this.Count != -1 {
			return nil, nil
		}

		nextSpan, err := client.Get(ctx, idx+int(span))
		if err != nil {
			return nil, err
		}
		nextSpanPrev, err := client.Get(ctx, idx+int(span)-1)
		if err != nil {
			return nil, err
		}

		xdNextSpan, ok1 := nextSpan.TotalCounts[span]
		xdNextSpanPrev, ok2 := nextSpanPrev.TotalCounts[span]
		ok3 := nextSpan.Count != -1

		if !ok1 || !ok2 || !ok3 {
			return nil, nil
		}

		this.Count = xdNextSpanPrev - xdNextSpan + nextSpan.Count
		if err = client.Set(ctx, idx, this); err != nil {
			return nil, err
		}

		return c.postCalcCount(ctx, idx), nil
	}
}

// ┌───────┬───┬───┬───┬───┐
// │days(n)│n-x│...│ n │n+1│
// ├───────┼───┼───┼───┼───┤
// │ed(n)  │ k │   │   │ k │
// ├───────┼───┼───┼───┼───┤
// │xd(n)  │   │   │ ? │ k │
// └───────┴───┴───┴───┴───┘
// `k` means known.
// xd(n-1) = xd(n) - ed(n) + d(n-x)
func (c Calculator) calcTotalCountLeft(ctx context.Context, idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(ctx, idx)
		if err != nil {
			return nil, err
		}

		// Nothing to calc if already known
		if _, ok := this.TotalCounts[span]; ok {
			return nil, nil
		}

		next, err := client.Get(ctx, idx+1)
		if err != nil {
			return nil, err
		}
		nextPrevSpan, err := client.Get(ctx, idx-int(span)+1)
		if err != nil {
			return nil, err
		}

		xdNext, ok1 := next.TotalCounts[span]
		ok2 := nextPrevSpan.Count != -1
		ok3 := next.Count != -1

		if !ok1 || !ok2 || !ok3 {
			return nil, nil
		}
		this.TotalCounts[span] = xdNext - next.Count + nextPrevSpan.Count
		if err = client.Set(ctx, idx, this); err != nil {
			return nil, err
		}

		return c.postCalcTotalCount(ctx, idx), nil
	}
}

func (c Calculator) postCalcCount(ctx context.Context, idx int) []calcFunc {
	var result []calcFunc

	for _, span := range c.spans {
		result = append(result,
			// Make the new idx as count left
			c.calcCountRight(ctx, idx+int(span), span),
			c.calcTotalCountLeft(ctx, idx+int(span)-1, span),
			c.calcTotalCountRight(ctx, idx+int(span), span),

			// Make the new idx as count right
			c.calcCountLeft(ctx, idx-int(span), span),
			c.calcTotalCountLeft(ctx, idx-1, span),
			c.calcTotalCountRight(ctx, idx, span),
		)
	}

	return result
}

func (c Calculator) postCalcTotalCount(ctx context.Context, idx int) []calcFunc {
	var result []calcFunc

	for _, span := range c.spans {
		result = append(result,
			// Make the new idx as total count left
			c.calcCountRight(ctx, idx+1, span),
			c.calcCountLeft(ctx, idx+1-int(span), span),
			c.calcTotalCountRight(ctx, idx+1, span),

			// Make the new idx as total count right
			c.calcCountRight(ctx, idx, span),
			c.calcCountLeft(ctx, idx-int(span), span),
			c.calcTotalCountLeft(ctx, idx-1, span),
		)
	}

	return result
}
