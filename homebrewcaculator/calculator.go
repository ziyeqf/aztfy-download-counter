package homebrewcaculator

type Span int

type calcFunc func(client DatabaseClient) ([]calcFunc, error)

type Calculator struct {
	spans []Span
	db    DatabaseClient
}

func NewCalculator(spans []Span, db DatabaseClient) Calculator {
	return Calculator{spans: spans, db: db}
}

func (c Calculator) Calc(idx int) {
	queue := Queue{}
	for _, span := range c.spans {
		queue.Enqueue(c.calcCountRight(idx, span))
		queue.Enqueue(c.calcTotalCountRight(idx, span))
		queue.Enqueue(c.calcTotalCountLeft(idx, span))
		queue.Enqueue(c.calcCountLeft(idx, span))
	}

	for !queue.IsEmpty() {
		calcFunc := queue.Dequeue()
		if calcFunc != nil {
			newFuncs, err := calcFunc(c.db)
			if err != nil {
				// TODO: collect the errors (errors.Join) and continue
				panic(err)
			}
			queue.Enqueue(newFuncs...)
		}
	}
}

// TODO Needs a ASCII diagram to explain what are count (left/right) and total count (left/right)

// Count -> ed(n)
// Total Count -> xd(n)

// ed(n) = xd(n) - xd(n-1) + ed(n-x)
func (c Calculator) calcCountRight(idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(idx)
		if err != nil {
			return nil, err
		}
		// Nothing to calc if already known
		if this.Count != -1 {
			return nil, nil
		}

		prev, err := client.Get(idx - 1)
		if err != nil {
			return nil, err
		}
		prevSpan, err := client.Get(idx - int(span))
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
		if err = client.Set(idx, this); err != nil {
			return nil, err
		}

		return c.postCalcCount(idx), nil
	}
}

func (c Calculator) postCalcCount(idx int) []calcFunc {
	var result []calcFunc

	for _, span := range c.spans {
		result = append(result,
			// Make the new idx as count left
			c.calcCountRight(idx+int(span), span),
			c.calcTotalCountLeft(idx+int(span)-1, span),
			c.calcTotalCountRight(idx+int(span), span),

			// Make the new idx as count right
			c.calcCountLeft(idx-int(span), span),
			c.calcTotalCountLeft(idx-1, span),
			c.calcTotalCountRight(idx, span),
		)
	}

	return result
}

func (c Calculator) postCalcTotalCount(idx int) []calcFunc {
	var result []calcFunc

	for _, span := range c.spans {
		result = append(result,
			// Make the new idx as total count left
			c.calcCountRight(idx+1, span),
			c.calcCountLeft(idx+1-int(span), span),
			c.calcTotalCountRight(idx+1, span),

			// Make the new idx as total count right
			c.calcCountRight(idx, span),
			c.calcCountLeft(idx-int(span), span),
			c.calcTotalCountLeft(idx-1, span),
		)
	}

	return result
}

// xd(n) = ed(n) + xd(n-1) - ed(n-x)
func (c Calculator) calcTotalCountRight(idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(idx)
		if err != nil {
			return nil, err
		}

		// Nothing to calc if already known
		if _, ok := this.TotalCounts[span]; ok {
			return nil, nil
		}

		prev, err := client.Get(idx - 1)
		if err != nil {
			return nil, err
		}
		prevSpan, err := client.Get(idx - int(span))
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
		if err = client.Set(idx, this); err != nil {
			return nil, err
		}

		return c.postCalcTotalCount(idx), nil
	}
}

// d(n-x) = xd(n) - xd(n-1) + ed(n)
func (c Calculator) calcCountLeft(idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(idx)
		if err != nil {
			return nil, err
		}
		// Nothing to calc if already known
		if this.Count != -1 {
			return nil, nil
		}

		nextSpan, err := client.Get(idx + int(span))
		if err != nil {
			return nil, err
		}
		nextSpanPrev, err := client.Get(idx + int(span) - 1)
		if err != nil {
			return nil, err
		}

		xdNextSpan, ok1 := nextSpan.TotalCounts[span]
		xdNextSpanPrev, ok2 := nextSpanPrev.TotalCounts[span]
		ok3 := nextSpan.Count != -1

		if !ok1 || !ok2 || !ok3 {
			return nil, nil
		}

		nextSpanPrev.Count = xdNextSpan - xdNextSpanPrev + nextSpan.Count
		if err = client.Set(idx, this); err != nil {
			return nil, err
		}

		return c.postCalcCount(idx), nil
	}
}

// xd(n-1) = xd(n) - ed(n) + d(n-x)
func (c Calculator) calcTotalCountLeft(idx int, span Span) calcFunc {
	return func(client DatabaseClient) ([]calcFunc, error) {
		this, err := client.Get(idx)
		if err != nil {
			return nil, err
		}

		// Nothing to calc if already known
		if _, ok := this.TotalCounts[span]; ok {
			return nil, nil
		}

		next, err := client.Get(idx + 1)
		if err != nil {
			return nil, err
		}
		nextPrevSpan, err := client.Get(idx - int(span) + 1)
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
		if err = client.Set(idx, this); err != nil {
			return nil, err
		}

		return c.postCalcTotalCount(idx), nil
	}
}
