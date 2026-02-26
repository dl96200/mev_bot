package marketdata

import "context"

type Source interface {
	Start(ctx context.Context) error
	Opportunities() <-chan Opportunity
}
